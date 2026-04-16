// ABOUTME: Tests for sync engine helper functions.
// ABOUTME: Covers pairToolResults and related conversion logic.
package sync

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/wesm/agentsview/internal/db"
)

func openTestDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.Open(
		filepath.Join(t.TempDir(), "test.db"),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

// fakeFileInfo implements os.FileInfo for test use.
type fakeFileInfo struct {
	size  int64
	mtime int64 // UnixNano
}

func (f fakeFileInfo) Name() string      { return "test" }
func (f fakeFileInfo) Size() int64       { return f.size }
func (f fakeFileInfo) Mode() os.FileMode { return 0 }
func (f fakeFileInfo) ModTime() time.Time {
	return time.Unix(0, f.mtime)
}
func (f fakeFileInfo) IsDir() bool { return false }
func (f fakeFileInfo) Sys() any    { return nil }

func TestFilterEmptyMessages(t *testing.T) {
	tests := []struct {
		name string
		msgs []db.Message
		want []db.Message
	}{
		{
			name: "removes empty-content user message after pairing",
			msgs: []db.Message{
				{
					Role:    "assistant",
					Content: "Let me read the file.",
					ToolCalls: []db.ToolCall{
						{ToolUseID: "t1", ToolName: "Read"},
					},
				},
				{
					Role:    "user",
					Content: "",
					ToolResults: []db.ToolResult{
						{ToolUseID: "t1", ContentLength: 500},
					},
				},
			},
			want: []db.Message{
				{
					Role:    "assistant",
					Content: "Let me read the file.",
					ToolCalls: []db.ToolCall{
						{ToolUseID: "t1", ToolName: "Read", ResultContentLength: 500},
					},
				},
			},
		},
		{
			name: "keeps user message with real content",
			msgs: []db.Message{
				{
					Role:    "assistant",
					Content: "Here is the result.",
					ToolCalls: []db.ToolCall{
						{ToolUseID: "t1", ToolName: "Bash"},
					},
				},
				{
					Role:    "user",
					Content: "",
					ToolResults: []db.ToolResult{
						{ToolUseID: "t1", ContentLength: 100},
					},
				},
				{
					Role:    "user",
					Content: "Thanks, now do something else.",
				},
			},
			want: []db.Message{
				{
					Role:    "assistant",
					Content: "Here is the result.",
					ToolCalls: []db.ToolCall{
						{ToolUseID: "t1", ToolName: "Bash", ResultContentLength: 100},
					},
				},
				{
					Role:    "user",
					Content: "Thanks, now do something else.",
				},
			},
		},
		{
			name: "whitespace-only content treated as empty",
			msgs: []db.Message{
				{
					Role:    "assistant",
					Content: "Reading...",
					ToolCalls: []db.ToolCall{
						{ToolUseID: "t1", ToolName: "Read"},
					},
				},
				{
					Role:    "user",
					Content: "   \n\t  ",
					ToolResults: []db.ToolResult{
						{ToolUseID: "t1", ContentLength: 300},
					},
				},
			},
			want: []db.Message{
				{
					Role:    "assistant",
					Content: "Reading...",
					ToolCalls: []db.ToolCall{
						{ToolUseID: "t1", ToolName: "Read", ResultContentLength: 300},
					},
				},
			},
		},
		{
			name: "preserves empty assistant message",
			msgs: []db.Message{
				{
					Role:    "assistant",
					Content: "",
				},
			},
			want: []db.Message{
				{
					Role:    "assistant",
					Content: "",
				},
			},
		},
		{
			name: "only removes user messages with tool results",
			msgs: []db.Message{
				{
					Role:    "assistant",
					Content: "",
				},
				{
					Role:    "user",
					Content: "",
				},
			},
			want: []db.Message{
				{
					Role:    "assistant",
					Content: "",
				},
				{
					Role:    "user",
					Content: "",
				},
			},
		},
		{
			name: "no messages returns empty",
			msgs: nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pairAndFilter(tt.msgs, nil)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("pairAndFilter() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPostFilterCounts(t *testing.T) {
	type counts struct {
		Total int
		User  int
	}
	tests := []struct {
		name string
		msgs []db.Message
		want counts
	}{
		{
			name: "mixed roles",
			msgs: []db.Message{
				{Role: "user", Content: "hello"},
				{Role: "assistant", Content: "hi"},
				{Role: "user", Content: "thanks"},
			},
			want: counts{Total: 3, User: 2},
		},
		{
			name: "no user messages",
			msgs: []db.Message{
				{Role: "assistant", Content: "hi"},
			},
			want: counts{Total: 1, User: 0},
		},
		{
			name: "empty slice",
			msgs: nil,
			want: counts{Total: 0, User: 0},
		},
		{
			name: "all user messages",
			msgs: []db.Message{
				{Role: "user", Content: "a"},
				{Role: "user", Content: "b"},
			},
			want: counts{Total: 2, User: 2},
		},
		{
			name: "system messages excluded from user count",
			msgs: []db.Message{
				{Role: "user", Content: "hello", IsSystem: false},
				{Role: "user", Content: "system notice", IsSystem: true},
				{Role: "assistant", Content: "hi"},
				{Role: "user", Content: "[Turn finished: endTurn]", IsSystem: true},
			},
			want: counts{Total: 4, User: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, user := postFilterCounts(tt.msgs)
			got := counts{Total: total, User: user}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("postFilterCounts() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPairToolResults(t *testing.T) {
	tests := []struct {
		name string
		msgs []db.Message
		want []db.Message
	}{
		{
			name: "basic pairing across messages",
			msgs: []db.Message{
				{ToolCalls: []db.ToolCall{
					{ToolUseID: "t1", ToolName: "Read"},
					{ToolUseID: "t2", ToolName: "Grep"},
				}},
				{ToolResults: []db.ToolResult{
					{ToolUseID: "t1", ContentLength: 100},
					{ToolUseID: "t2", ContentLength: 200},
				}},
			},
			want: []db.Message{
				{ToolCalls: []db.ToolCall{
					{ToolUseID: "t1", ToolName: "Read", ResultContentLength: 100},
					{ToolUseID: "t2", ToolName: "Grep", ResultContentLength: 200},
				}},
				{ToolResults: []db.ToolResult{
					{ToolUseID: "t1", ContentLength: 100},
					{ToolUseID: "t2", ContentLength: 200},
				}},
			},
		},
		{
			name: "unmatched tool_result ignored",
			msgs: []db.Message{
				{ToolCalls: []db.ToolCall{
					{ToolUseID: "t1", ToolName: "Read"},
				}},
				{ToolResults: []db.ToolResult{
					{ToolUseID: "t1", ContentLength: 50},
					{ToolUseID: "t_unknown", ContentLength: 999},
				}},
			},
			want: []db.Message{
				{ToolCalls: []db.ToolCall{
					{ToolUseID: "t1", ToolName: "Read", ResultContentLength: 50},
				}},
				{ToolResults: []db.ToolResult{
					{ToolUseID: "t1", ContentLength: 50},
					{ToolUseID: "t_unknown", ContentLength: 999},
				}},
			},
		},
		{
			name: "unmatched tool_call keeps zero",
			msgs: []db.Message{
				{ToolCalls: []db.ToolCall{
					{ToolUseID: "t1", ToolName: "Read"},
					{ToolUseID: "t2", ToolName: "Bash"},
				}},
				{ToolResults: []db.ToolResult{
					{ToolUseID: "t1", ContentLength: 42},
				}},
			},
			want: []db.Message{
				{ToolCalls: []db.ToolCall{
					{ToolUseID: "t1", ToolName: "Read", ResultContentLength: 42},
					{ToolUseID: "t2", ToolName: "Bash", ResultContentLength: 0},
				}},
				{ToolResults: []db.ToolResult{
					{ToolUseID: "t1", ContentLength: 42},
				}},
			},
		},
		{
			name: "empty messages",
			msgs: nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pairToolResults(tt.msgs, nil)
			if diff := cmp.Diff(tt.want, tt.msgs); diff != "" {
				t.Errorf("pairToolResults() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPairToolResultsContent(t *testing.T) {
	ampToolResultText := "line 1\nline \"2\" output"
	ampToolResultRaw := "\"line 1\\nline \\\"2\\\" output\""

	tests := []struct {
		name    string
		msgs    []db.Message
		blocked map[string]bool
		want    []db.Message
	}{
		{
			name: "content stored for non-blocked category",
			msgs: []db.Message{
				{ToolCalls: []db.ToolCall{
					{ToolUseID: "t1", ToolName: "Bash", Category: "Bash"},
				}},
				{ToolResults: []db.ToolResult{
					{ToolUseID: "t1", ContentLength: 42, ContentRaw: `"output text"`},
				}},
			},
			blocked: map[string]bool{"Read": true, "Glob": true},
			want: []db.Message{
				{ToolCalls: []db.ToolCall{
					{ToolUseID: "t1", ToolName: "Bash", Category: "Bash",
						ResultContentLength: 42, ResultContent: "output text"},
				}},
				{ToolResults: []db.ToolResult{
					{ToolUseID: "t1", ContentLength: 42, ContentRaw: `"output text"`},
				}},
			},
		},
		{
			name: "content blocked for Read category",
			msgs: []db.Message{
				{ToolCalls: []db.ToolCall{
					{ToolUseID: "t1", ToolName: "Read", Category: "Read"},
				}},
				{ToolResults: []db.ToolResult{
					{ToolUseID: "t1", ContentLength: 5000, ContentRaw: `"file data"`},
				}},
			},
			blocked: map[string]bool{"Read": true, "Glob": true},
			want: []db.Message{
				{ToolCalls: []db.ToolCall{
					{ToolUseID: "t1", ToolName: "Read", Category: "Read",
						ResultContentLength: 5000, ResultContent: ""},
				}},
				{ToolResults: []db.ToolResult{
					{ToolUseID: "t1", ContentLength: 5000, ContentRaw: `"file data"`},
				}},
			},
		},
		{
			name: "nil blocked map stores all content",
			msgs: []db.Message{
				{ToolCalls: []db.ToolCall{
					{ToolUseID: "t1", ToolName: "Read", Category: "Read"},
				}},
				{ToolResults: []db.ToolResult{
					{ToolUseID: "t1", ContentLength: 100, ContentRaw: `"file content"`},
				}},
			},
			blocked: nil,
			want: []db.Message{
				{ToolCalls: []db.ToolCall{
					{ToolUseID: "t1", ToolName: "Read", Category: "Read",
						ResultContentLength: 100, ResultContent: "file content"},
				}},
				{ToolResults: []db.ToolResult{
					{ToolUseID: "t1", ContentLength: 100, ContentRaw: `"file content"`},
				}},
			},
		},
		{
			// Mirrors ContentRaw produced by parser.extractAmpToolResults
			// (JSON-marshaled plain-text output).
			name: "amp: marshaled tool result text decodes into ResultContent",
			msgs: []db.Message{
				{ToolCalls: []db.ToolCall{
					{ToolUseID: "t1", ToolName: "Bash", Category: "Bash"},
				}},
				{ToolResults: []db.ToolResult{
					{
						ToolUseID:     "t1",
						ContentLength: len(ampToolResultText),
						ContentRaw:    ampToolResultRaw,
					},
				}},
			},
			blocked: nil,
			want: []db.Message{
				{ToolCalls: []db.ToolCall{
					{
						ToolUseID: "t1", ToolName: "Bash", Category: "Bash",
						ResultContentLength: len(ampToolResultText),
						ResultContent:       ampToolResultText,
					},
				}},
				{ToolResults: []db.ToolResult{
					{
						ToolUseID:     "t1",
						ContentLength: len(ampToolResultText),
						ContentRaw:    ampToolResultRaw,
					},
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pairToolResults(tt.msgs, tt.blocked)
			if diff := cmp.Diff(tt.want, tt.msgs); diff != "" {
				t.Errorf("pairToolResults() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPairToolResultEventSummaries(t *testing.T) {
	tests := []struct {
		name    string
		msgs    []db.Message
		blocked map[string]bool
		want    []db.Message
	}{
		{
			name: "single event becomes summary",
			msgs: []db.Message{{
				ToolCalls: []db.ToolCall{{
					ToolUseID: "call_wait",
					ToolName:  "wait",
					Category:  "Other",
					ResultEvents: []db.ToolResultEvent{{
						ToolUseID:     "call_wait",
						AgentID:       "agent-1",
						Source:        "wait_output",
						Status:        "completed",
						Content:       "Finished successfully",
						ContentLength: len("Finished successfully"),
					}},
				}},
			}},
			want: []db.Message{{
				ToolCalls: []db.ToolCall{{
					ToolUseID:           "call_wait",
					ToolName:            "wait",
					Category:            "Other",
					ResultContentLength: len("Finished successfully"),
					ResultContent:       "Finished successfully",
					ResultEvents: []db.ToolResultEvent{{
						ToolUseID:     "call_wait",
						AgentID:       "agent-1",
						Source:        "wait_output",
						Status:        "completed",
						Content:       "Finished successfully",
						ContentLength: len("Finished successfully"),
					}},
				}},
			}},
		},
		{
			name: "multi-agent latest summary keeps one line per agent",
			msgs: []db.Message{{
				ToolCalls: []db.ToolCall{{
					ToolUseID: "call_wait",
					ToolName:  "wait",
					Category:  "Other",
					ResultEvents: []db.ToolResultEvent{
						{
							ToolUseID:     "call_wait",
							AgentID:       "agent-a",
							Source:        "wait_output",
							Status:        "completed",
							Content:       "First finished",
							ContentLength: len("First finished"),
						},
						{
							ToolUseID:     "call_wait",
							AgentID:       "agent-b",
							Source:        "subagent_notification",
							Status:        "completed",
							Content:       "Second finished",
							ContentLength: len("Second finished"),
						},
					},
				}},
			}},
			want: []db.Message{{
				ToolCalls: []db.ToolCall{{
					ToolUseID:           "call_wait",
					ToolName:            "wait",
					Category:            "Other",
					ResultContentLength: len("agent-a:\nFirst finished\n\nagent-b:\nSecond finished"),
					ResultContent:       "agent-a:\nFirst finished\n\nagent-b:\nSecond finished",
					ResultEvents: []db.ToolResultEvent{
						{
							ToolUseID:     "call_wait",
							AgentID:       "agent-a",
							Source:        "wait_output",
							Status:        "completed",
							Content:       "First finished",
							ContentLength: len("First finished"),
						},
						{
							ToolUseID:     "call_wait",
							AgentID:       "agent-b",
							Source:        "subagent_notification",
							Status:        "completed",
							Content:       "Second finished",
							ContentLength: len("Second finished"),
						},
					},
				}},
			}},
		},
		{
			name: "blocked category keeps length but drops summary content",
			msgs: []db.Message{{
				ToolCalls: []db.ToolCall{{
					ToolUseID: "call_read",
					ToolName:  "Read",
					Category:  "Read",
					ResultEvents: []db.ToolResultEvent{{
						ToolUseID:     "call_read",
						Source:        "wait_output",
						Status:        "completed",
						Content:       "secret file body",
						ContentLength: len("secret file body"),
					}},
				}},
			}},
			blocked: map[string]bool{"Read": true},
			want: []db.Message{{
				ToolCalls: []db.ToolCall{{
					ToolUseID:           "call_read",
					ToolName:            "Read",
					Category:            "Read",
					ResultContentLength: len("secret file body"),
					ResultContent:       "",
					ResultEvents:        nil,
				}},
			}},
		},
		{
			name: "mixed anonymous and multi-agent content keeps both",
			msgs: []db.Message{{
				ToolCalls: []db.ToolCall{{
					ToolUseID: "call_wait",
					ToolName:  "wait",
					Category:  "Other",
					ResultEvents: []db.ToolResultEvent{
						{
							ToolUseID:     "call_wait",
							AgentID:       "agent-a",
							Source:        "wait_output",
							Status:        "completed",
							Content:       "First finished",
							ContentLength: len("First finished"),
						},
						{
							ToolUseID:     "call_wait",
							AgentID:       "agent-b",
							Source:        "wait_output",
							Status:        "completed",
							Content:       "Second finished",
							ContentLength: len("Second finished"),
						},
						{
							ToolUseID:     "call_wait",
							Source:        "subagent_notification",
							Status:        "completed",
							Content:       "Detached note",
							ContentLength: len("Detached note"),
						},
					},
				}},
			}},
			want: []db.Message{{
				ToolCalls: []db.ToolCall{{
					ToolUseID:           "call_wait",
					ToolName:            "wait",
					Category:            "Other",
					ResultContentLength: len("agent-a:\nFirst finished\n\nagent-b:\nSecond finished\n\nDetached note"),
					ResultContent:       "agent-a:\nFirst finished\n\nagent-b:\nSecond finished\n\nDetached note",
					ResultEvents: []db.ToolResultEvent{
						{
							ToolUseID:     "call_wait",
							AgentID:       "agent-a",
							Source:        "wait_output",
							Status:        "completed",
							Content:       "First finished",
							ContentLength: len("First finished"),
						},
						{
							ToolUseID:     "call_wait",
							AgentID:       "agent-b",
							Source:        "wait_output",
							Status:        "completed",
							Content:       "Second finished",
							ContentLength: len("Second finished"),
						},
						{
							ToolUseID:     "call_wait",
							Source:        "subagent_notification",
							Status:        "completed",
							Content:       "Detached note",
							ContentLength: len("Detached note"),
						},
					},
				}},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pairToolResultEventSummaries(tt.msgs, tt.blocked)
			if diff := cmp.Diff(tt.want, tt.msgs); diff != "" {
				t.Fatalf("pairToolResultEventSummaries() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestApplyRemoteRewrites(t *testing.T) {
	tests := []struct {
		name         string
		prefix       string
		rewriter     func(string) string
		sess         db.Session
		msgs         []db.Message
		wantSessID   string
		wantParent   *string
		wantFilePath *string
		wantMsgSess  string // expected SessionID on messages
		wantSubs     []string
		wantEvSubs   []string
	}{
		{
			name:   "no prefix is no-op",
			prefix: "",
			sess: db.Session{
				ID: "abc",
			},
			msgs: []db.Message{
				{SessionID: "abc"},
			},
			wantSessID:  "abc",
			wantMsgSess: "abc",
		},
		{
			name:   "all fields prefixed",
			prefix: "host~",
			sess: db.Session{
				ID:              "abc",
				ParentSessionID: strPtr("parent-1"),
				FilePath:        strPtr("/tmp/file"),
			},
			msgs: []db.Message{
				{
					SessionID: "abc",
					ToolCalls: []db.ToolCall{
						{
							SessionID:         "abc",
							SubagentSessionID: "sub-1",
							ResultEvents: []db.ToolResultEvent{
								{SubagentSessionID: "ev-1"},
								{SubagentSessionID: ""},
							},
						},
						{SessionID: "abc"},
					},
				},
			},
			wantSessID:   "host~abc",
			wantParent:   strPtr("host~parent-1"),
			wantFilePath: strPtr("/tmp/file"),
			wantMsgSess:  "host~abc",
			wantSubs:     []string{"host~sub-1", ""},
			wantEvSubs:   []string{"host~ev-1", ""},
		},
		{
			name:   "path rewriter applied",
			prefix: "box~",
			rewriter: func(p string) string {
				return "box:" + p
			},
			sess: db.Session{
				ID:       "x",
				FilePath: strPtr("/remote/path"),
			},
			msgs:         nil,
			wantSessID:   "box~x",
			wantFilePath: strPtr("box:/remote/path"),
		},
		{
			name:   "nil parent stays nil",
			prefix: "h~",
			sess: db.Session{
				ID: "z",
			},
			wantSessID: "h~z",
			wantParent: nil,
		},
		{
			name:   "empty parent stays empty",
			prefix: "h~",
			sess: db.Session{
				ID:              "z",
				ParentSessionID: strPtr(""),
			},
			wantSessID: "h~z",
			wantParent: strPtr(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Engine{
				idPrefix:     tt.prefix,
				pathRewriter: tt.rewriter,
			}
			e.applyRemoteRewrites(&tt.sess, tt.msgs)

			if tt.sess.ID != tt.wantSessID {
				t.Errorf(
					"ID = %q, want %q",
					tt.sess.ID, tt.wantSessID,
				)
			}
			if diff := cmp.Diff(
				tt.wantParent, tt.sess.ParentSessionID,
			); diff != "" {
				t.Errorf("ParentSessionID %s", diff)
			}
			if tt.wantFilePath != nil {
				if diff := cmp.Diff(
					tt.wantFilePath, tt.sess.FilePath,
				); diff != "" {
					t.Errorf("FilePath %s", diff)
				}
			}
			for _, m := range tt.msgs {
				if m.SessionID != tt.wantMsgSess {
					t.Errorf(
						"msg SessionID = %q, want %q",
						m.SessionID, tt.wantMsgSess,
					)
				}
			}
			var gotSubs, gotEvSubs []string
			for _, m := range tt.msgs {
				for _, tc := range m.ToolCalls {
					gotSubs = append(
						gotSubs, tc.SubagentSessionID,
					)
					for _, ev := range tc.ResultEvents {
						gotEvSubs = append(
							gotEvSubs,
							ev.SubagentSessionID,
						)
					}
				}
			}
			if diff := cmp.Diff(
				tt.wantSubs, gotSubs,
			); diff != "" {
				t.Errorf("SubagentSessionIDs %s", diff)
			}
			if diff := cmp.Diff(
				tt.wantEvSubs, gotEvSubs,
			); diff != "" {
				t.Errorf("ResultEvent SubagentSessionIDs %s", diff)
			}
		})
	}
}

func TestShouldSkipFileWithIDPrefix(t *testing.T) {
	database := openTestDB(t)

	// Store a session with prefixed ID and file metadata.
	sess := db.Session{
		ID:       "host~abc-123",
		Project:  "test",
		Machine:  "host",
		Agent:    "claude",
		FilePath: strPtr("host:/remote/session.jsonl"),
		FileSize: int64Ptr(1024),
		FileMtime: int64Ptr(
			int64(1700000000000000000),
		),
	}
	if err := database.UpsertSession(sess); err != nil {
		t.Fatal(err)
	}
	// data_version is no longer persisted by UpsertSession;
	// stamp it explicitly so the skip check sees a current
	// row.
	if err := database.SetSessionDataVersion(
		sess.ID, db.CurrentDataVersion(),
	); err != nil {
		t.Fatal(err)
	}

	// Engine with IDPrefix should find the session.
	e := &Engine{
		db:       database,
		idPrefix: "host~",
	}
	got := e.shouldSkipFile(
		"abc-123",
		fakeFileInfo{size: 1024, mtime: 1700000000000000000},
	)
	if !got {
		t.Error("shouldSkipFile should return true")
	}

	// Engine WITHOUT IDPrefix should NOT find it.
	e2 := &Engine{db: database}
	got2 := e2.shouldSkipFile(
		"abc-123",
		fakeFileInfo{size: 1024, mtime: 1700000000000000000},
	)
	if got2 {
		t.Error(
			"shouldSkipFile without prefix should return false",
		)
	}
}

func TestShouldSkipByPathWithRewriter(t *testing.T) {
	database := openTestDB(t)

	// Store a session with rewritten file path.
	sess := db.Session{
		ID:       "host~codex:abc",
		Project:  "test",
		Machine:  "host",
		Agent:    "codex",
		FilePath: strPtr("host:/remote/codex/abc.jsonl"),
		FileSize: int64Ptr(2048),
		FileMtime: int64Ptr(
			int64(1700000000000000000),
		),
	}
	if err := database.UpsertSession(sess); err != nil {
		t.Fatal(err)
	}
	if err := database.SetSessionDataVersion(
		sess.ID, db.CurrentDataVersion(),
	); err != nil {
		t.Fatal(err)
	}

	rewriter := func(p string) string {
		return "host:" + p
	}

	// Engine with PathRewriter should find the session.
	e := &Engine{
		db:           database,
		pathRewriter: rewriter,
	}
	got := e.shouldSkipByPath(
		"/remote/codex/abc.jsonl",
		fakeFileInfo{size: 2048, mtime: 1700000000000000000},
	)
	if !got {
		t.Error("shouldSkipByPath should return true")
	}

	// Without rewriter, lookup misses.
	e2 := &Engine{db: database}
	got2 := e2.shouldSkipByPath(
		"/remote/codex/abc.jsonl",
		fakeFileInfo{size: 2048, mtime: 1700000000000000000},
	)
	if got2 {
		t.Error(
			"shouldSkipByPath without rewriter should " +
				"return false",
		)
	}
}

func TestBlockedCategorySet(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		check string
		want  bool
	}{
		{"exact match", []string{"Read"}, "Read", true},
		{"lowercase normalized", []string{"read"}, "Read", true},
		{"uppercase normalized", []string{"GLOB"}, "Glob", true},
		{"trimmed", []string{" Read "}, "Read", true},
		{"empty entry skipped", []string{""}, "Read", false},
		{"nil input", nil, "Read", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := blockedCategorySet(tt.input)
			got := m[tt.check]
			if got != tt.want {
				t.Errorf(
					"blockedCategorySet(%v)[%q] = %v, want %v",
					tt.input, tt.check, got, tt.want,
				)
			}
		})
	}
}
