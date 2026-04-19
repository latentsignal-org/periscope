// Package summarize converts stored session messages into per-turn
// bundles and asks the LLM to produce compact summaries.
//
// The package is deliberately narrow: turn grouping, content hashing,
// prompt/parse logic, and a background worker. Rewind/compact banner
// text generation is layered on top in Phase B.
package summarize

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"github.com/wesm/agentsview/internal/db"
)

// PromptVersion is bumped when the summariser prompt or output schema
// changes. Bumping invalidates previously stored summaries (by not
// matching on the new content_hash).
const PromptVersion = 1

// TurnBundle represents one turn's worth of content for summarisation.
type TurnBundle struct {
	TurnIndex    int
	StartOrdinal int
	EndOrdinal   int

	UserMessage      string
	AssistantMessage string
	Thinking         string
	ToolCalls        []ToolCallBundle
}

// ToolCallBundle is a single tool invocation's input + brief result.
type ToolCallBundle struct {
	ToolName  string
	Category  string
	InputText string
	Result    string
	Status    string
}

// BuildTurns groups a session's messages into turn bundles. Turn
// boundaries match the logic used by server/context.go: a new turn
// begins at every non-system, non-compact-boundary user message, and
// immediately after a compaction boundary.
//
// The messages must be sorted ascending by ordinal.
func BuildTurns(msgs []db.Message) []TurnBundle {
	var turns []TurnBundle
	var cur *TurnBundle
	flush := func() {
		if cur == nil {
			return
		}
		turns = append(turns, *cur)
		cur = nil
	}
	for _, m := range msgs {
		if m.IsCompactBoundary || m.SourceSubtype == "compact_boundary" {
			flush()
			continue
		}
		// System prompts don't participate in a turn. Skip them so
		// a leading system message doesn't produce an empty turn.
		if m.IsSystem {
			continue
		}
		startsNew := cur == nil || m.Role == "user"
		if startsNew {
			flush()
			cur = &TurnBundle{
				StartOrdinal: m.Ordinal,
				EndOrdinal:   m.Ordinal,
			}
		}
		cur.EndOrdinal = m.Ordinal
		switch {
		case m.Role == "user":
			cur.UserMessage = appendSnippet(
				cur.UserMessage, stripContent(m.Content),
			)
		case m.Role == "assistant":
			thinking, assistant := splitThinking(m.Content)
			if thinking != "" {
				cur.Thinking = appendSnippet(cur.Thinking, thinking)
			}
			if assistant != "" {
				cur.AssistantMessage = appendSnippet(
					cur.AssistantMessage, stripToolBlocks(assistant),
				)
			}
		}
		for _, tc := range m.ToolCalls {
			cur.ToolCalls = append(cur.ToolCalls, ToolCallBundle{
				ToolName:  tc.ToolName,
				Category:  tc.Category,
				InputText: compactInput(tc.InputJSON),
				Result:    firstToolResult(tc),
				Status:    toolStatus(tc),
			})
		}
	}
	flush()
	for i := range turns {
		turns[i].TurnIndex = i + 1
	}
	return turns
}

// ContentHash returns a stable hash of a turn's user/assistant text,
// tool-call signatures, and the current prompt version. Used as the
// cache key for a summary.
func ContentHash(b TurnBundle) string {
	h := sha256.New()
	writeKV := func(k, v string) {
		h.Write([]byte(k))
		h.Write([]byte{0})
		h.Write([]byte(v))
		h.Write([]byte{0})
	}
	writeKV("v", itoa(PromptVersion))
	writeKV("user", b.UserMessage)
	writeKV("assistant", b.AssistantMessage)
	writeKV("thinking", b.Thinking)
	sigs := make([]string, 0, len(b.ToolCalls))
	for _, tc := range b.ToolCalls {
		sigs = append(sigs, tc.ToolName+"|"+tc.Category+"|"+tc.InputText)
	}
	sort.Strings(sigs)
	for _, s := range sigs {
		writeKV("tool", s)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// --- text helpers ---

const maxFieldLen = 4000
const maxToolInput = 600
const maxToolResult = 600

func appendSnippet(base, more string) string {
	more = strings.TrimSpace(more)
	if more == "" {
		return base
	}
	if base == "" {
		return truncate(more, maxFieldLen)
	}
	return truncate(base+"\n"+more, maxFieldLen)
}

func stripContent(content string) string {
	return strings.TrimSpace(content)
}

func splitThinking(content string) (thinking, assistant string) {
	const startTag = "[Thinking]\n"
	const endTag = "\n[/Thinking]"
	remaining := content
	for {
		start := strings.Index(remaining, startTag)
		if start < 0 {
			assistant += remaining
			break
		}
		assistant += remaining[:start]
		remaining = remaining[start+len(startTag):]
		end := strings.Index(remaining, endTag)
		if end < 0 {
			thinking += remaining
			break
		}
		thinking += remaining[:end]
		remaining = remaining[end+len(endTag):]
	}
	return strings.TrimSpace(thinking), strings.TrimSpace(assistant)
}

// stripToolBlocks removes inline "[ToolName] ..." prose markers used
// by the transcript renderer. Kept simple — mirrors what the server's
// message-preview logic does.
func stripToolBlocks(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	skip := false
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "[") &&
			strings.Contains(trim, "]") &&
			looksLikeToolMarker(trim) {
			skip = true
			continue
		}
		if trim == "" {
			skip = false
			out = append(out, line)
			continue
		}
		if skip {
			continue
		}
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func looksLikeToolMarker(line string) bool {
	tags := []string{
		"[Tool", "[Read", "[Write", "[Edit", "[Bash", "[Glob",
		"[Grep", "[Task", "[Agent", "[Skill", "[apply_patch",
		"[exec_command", "[shell_command", "[view_image",
		"[SendMessage", "[Question", "[Todo List",
		"[Entering Plan Mode", "[Exiting Plan Mode", "[update_plan",
	}
	for _, t := range tags {
		if strings.HasPrefix(line, t) {
			return true
		}
	}
	return false
}

func compactInput(input string) string {
	s := strings.TrimSpace(input)
	s = strings.ReplaceAll(s, "\n", " ")
	return truncate(s, maxToolInput)
}

func firstToolResult(tc db.ToolCall) string {
	if s := strings.TrimSpace(tc.ResultContent); s != "" {
		return truncate(s, maxToolResult)
	}
	for _, ev := range tc.ResultEvents {
		if s := strings.TrimSpace(ev.Content); s != "" {
			return truncate(s, maxToolResult)
		}
	}
	return ""
}

func toolStatus(tc db.ToolCall) string {
	if len(tc.ResultEvents) > 0 {
		return tc.ResultEvents[len(tc.ResultEvents)-1].Status
	}
	return ""
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
