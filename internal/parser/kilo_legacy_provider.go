package parser

import (
	"os"
	"path/filepath"
	"strings"
)

// newKiloLegacyProviderFactory creates a provider factory for
// Kilo (legacy). Kilo stores sessions as task directories under
// <root>/tasks/<taskId>/ with task_metadata.json (always
// present), ui_messages.json (transcript), and
// api_conversation_history.json (Claude-format). The primary
// source is task_metadata.json; the other two are folded into
// the composite fingerprint via kiloLegacyFingerprintSource.
func newKiloLegacyProviderFactory(def AgentDef) ProviderFactory {
	return NewSingleFileProviderFactory(
		def,
		kiloLegacyProviderCapabilities(),
		func(cfg ProviderConfig) singleFileSourceSet {
			return NewSingleFileSourceSet(
				def.Type,
				cfg.Roots,
				WithFileDiscovery(func(root string) []singleFileMatch {
					return kiloLegacyDiscoverFiles(root)
				}),
				WithFileWatchRoots(func(roots []string) []WatchRoot {
					return kiloLegacyWatchRoots(roots)
				}),
				WithFileChangedPathClassifier(
					func(root, path string, allowMissing bool) (
						singleFileMatch, bool,
					) {
						return kiloLegacyClassifyPath(
							root, path, allowMissing,
						)
					},
				),
				WithFileLookup(func(root, rawID string) (singleFileMatch, bool) {
					return kiloLegacyFindFile(root, rawID)
				}),
				WithFileFingerprint(func(src singleFileSource) (SourceFingerprint, error) {
					return kiloLegacyFingerprintSource(src.Path)
				}),
				WithFileParse(func(src singleFileSource, req ParseRequest) ([]ParseResult, []string, error) {
					return kiloLegacyParseFile(src, req)
				}),
			)
		},
	)
}

// kiloLegacyDiscoverFiles finds all Kilo (legacy) session directories
// under a root. root is the globalStorage directory (e.g.
// ~/Library/.../kilocode.kilo-code). Sessions live under
// <root>/tasks/<taskId>/, identified by the presence of
// task_metadata.json.
func kiloLegacyDiscoverFiles(root string) []singleFileMatch {
	tasksDir := filepath.Join(root, "tasks")
	entries, err := os.ReadDir(tasksDir)
	if err != nil {
		return nil
	}
	var matches []singleFileMatch
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Reject symlinked task directories. Symlinks can
		// resolve outside the configured root, so remote sync
		// must not archive files from an unexpected location.
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		// Skip _index.json and other underscore-prefixed
		// metadata files. Sessions are timestamped UUIDs.
		if strings.HasPrefix(entry.Name(), "_") ||
			strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		taskDir := filepath.Join(tasksDir, entry.Name())
		metadataPath := filepath.Join(taskDir, "task_metadata.json")
		if !IsRegularFile(metadataPath) {
			continue
		}
		matches = append(matches, singleFileMatch{
			Path: metadataPath,
		})
	}
	return matches
}

// kiloLegacyWatchRoots returns the watch plan for the configured
// roots. Each root watches its `tasks/` subtree for any change
// to the three Kilo session files. A single watch per root keeps
// debouncing simple.
func kiloLegacyWatchRoots(roots []string) []WatchRoot {
	out := make([]WatchRoot, 0, len(roots))
	for _, root := range roots {
		tasksDir := filepath.Join(root, "tasks")
		out = append(out, WatchRoot{
			Path:      tasksDir,
			Recursive: true,
			IncludeGlobs: []string{
				"task_metadata.json",
				"ui_messages.json",
				"api_conversation_history.json",
			},
			DebounceKey: "kilo-legacy:sessions:" + root,
		})
	}
	return out
}

// kiloLegacyClassifyPath maps a changed file path back to the
// owning task's task_metadata.json. Paths are shaped like
//
//	<root>/tasks/<taskId>/task_metadata.json
//	<root>/tasks/<taskId>/ui_messages.json
//	<root>/tasks/<taskId>/api_conversation_history.json
//
// All three names accept; we resolve them to the same
// task_metadata.json anchor so a single reparse picks up the
// freshest snapshot of the session.
func kiloLegacyClassifyPath(
	root, path string, allowMissing bool,
) (singleFileMatch, bool) {
	root = filepath.Clean(root)
	path = filepath.Clean(path)

	tasksRoot := filepath.Join(root, "tasks")
	rel, err := filepath.Rel(tasksRoot, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return singleFileMatch{}, false
	}
	parts := strings.Split(rel, string(filepath.Separator))
	// <taskId>/<file>
	if len(parts) != 2 {
		return singleFileMatch{}, false
	}
	taskID := parts[0]
	filename := parts[1]
	switch filename {
	case "task_metadata.json",
		"ui_messages.json",
		"api_conversation_history.json":
	default:
		return singleFileMatch{}, false
	}
	if strings.HasPrefix(taskID, "_") ||
		strings.HasPrefix(taskID, ".") {
		return singleFileMatch{}, false
	}
	metadataPath := filepath.Join(
		tasksRoot, taskID, "task_metadata.json",
	)
	if allowMissing {
		return singleFileMatch{Path: metadataPath}, true
	}
	if IsRegularFile(metadataPath) {
		return singleFileMatch{Path: metadataPath}, true
	}
	return singleFileMatch{}, false
}

// kiloLegacyFindFile resolves a raw task ID (the UUID basename of
// the task directory) to its task_metadata.json source.
func kiloLegacyFindFile(root, rawID string) (singleFileMatch, bool) {
	metadataPath := filepath.Join(
		root, "tasks", rawID, "task_metadata.json",
	)
	if IsRegularFile(metadataPath) {
		return singleFileMatch{Path: metadataPath}, true
	}
	return singleFileMatch{}, false
}

// kiloLegacyParseFile parses a single Kilo (legacy) session from a
// task directory. The source anchor is task_metadata.json;
// ui_messages.json and api_conversation_history.json are
// folded into the composite fingerprint but parsed by
// parseKiloLegacySession.
func kiloLegacyParseFile(
	src singleFileSource, req ParseRequest,
) ([]ParseResult, []string, error) {
	taskDir := filepath.Dir(src.Path)
	sess, msgs, err := parseKiloLegacySession(
		taskDir, req.Source.ProjectHint, req.Machine,
	)
	if err != nil {
		return nil, nil, err
	}
	if sess == nil {
		return nil, nil, nil
	}
	if req.Fingerprint.Size > 0 {
		sess.File.Size = req.Fingerprint.Size
	}
	if req.Fingerprint.MTimeNS > 0 {
		sess.File.Mtime = req.Fingerprint.MTimeNS
	}
	if req.Fingerprint.Hash != "" {
		sess.File.Hash = req.Fingerprint.Hash
	}
	return []ParseResult{{
		Session:     *sess,
		Messages:    msgs,
		UsageEvents: sess.UsageEvents,
	}}, nil, nil
}

func kiloLegacyProviderCapabilities() Capabilities {
	return Capabilities{
		Source: jsonlFileProviderSourceCapabilities(),
		Content: ContentCapabilities{
			FirstMessage:         CapabilitySupported,
			SessionName:          CapabilitySupported,
			Cwd:                  CapabilitySupported,
			Thinking:             CapabilitySupported,
			Model:                CapabilityNotApplicable,
			ToolCalls:            CapabilitySupported,
			ToolResults:          CapabilitySupported,
			ToolResultEvents:     CapabilitySupported,
			Subagents:            CapabilityNotApplicable,
			AggregateUsageEvents: CapabilitySupported,
			Relationships:        CapabilityNotApplicable,
			TerminationStatus:    CapabilitySupported,
			MalformedLineCount:   CapabilityNotApplicable,
		},
	}
}
