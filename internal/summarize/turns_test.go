package summarize

import (
	"testing"

	"github.com/wesm/agentsview/internal/db"
)

// TestBuildTurnsGroupsByUserMessage asserts the turn boundaries
// match the Context timeline: a new turn begins at every non-system
// user message, and compaction boundaries also split.
func TestBuildTurnsGroupsByUserMessage(t *testing.T) {
	msgs := []db.Message{
		{Ordinal: 1, Role: "system", IsSystem: true, Content: "sys"},
		{Ordinal: 2, Role: "user", Content: "hello"},
		{Ordinal: 3, Role: "assistant", Content: "hi"},
		{Ordinal: 4, Role: "user", Content: "next"},
		{Ordinal: 5, Role: "assistant", Content: "ok"},
	}
	turns := BuildTurns(msgs)
	if len(turns) != 2 {
		t.Fatalf("len(turns) = %d, want 2", len(turns))
	}
	if turns[0].UserMessage != "hello" || turns[0].AssistantMessage != "hi" {
		t.Errorf("turn 1 = %+v", turns[0])
	}
	if turns[1].UserMessage != "next" || turns[1].AssistantMessage != "ok" {
		t.Errorf("turn 2 = %+v", turns[1])
	}
	if turns[0].TurnIndex != 1 || turns[1].TurnIndex != 2 {
		t.Errorf("TurnIndex = %d,%d", turns[0].TurnIndex, turns[1].TurnIndex)
	}
}

// TestBuildTurnsCompactBoundary ensures compact boundaries flush
// the previous turn so downstream summaries don't span the gap.
func TestBuildTurnsCompactBoundary(t *testing.T) {
	msgs := []db.Message{
		{Ordinal: 1, Role: "user", Content: "pre"},
		{Ordinal: 2, Role: "assistant", Content: "resp"},
		{Ordinal: 3, Role: "user", SourceSubtype: "compact_boundary", IsCompactBoundary: true, Content: "boundary"},
		{Ordinal: 4, Role: "user", Content: "post"},
		{Ordinal: 5, Role: "assistant", Content: "after"},
	}
	turns := BuildTurns(msgs)
	if len(turns) != 2 {
		t.Fatalf("len(turns) = %d, want 2", len(turns))
	}
	if turns[1].UserMessage != "post" {
		t.Errorf("post-compact turn = %q", turns[1].UserMessage)
	}
}

// TestContentHashStability ensures the hash is stable across calls
// and changes when the content changes.
func TestContentHashStability(t *testing.T) {
	b := TurnBundle{
		TurnIndex:        1,
		UserMessage:      "hello",
		AssistantMessage: "world",
		ToolCalls: []ToolCallBundle{
			{ToolName: "Read", Category: "Read", InputText: "foo"},
		},
	}
	h1 := ContentHash(b)
	h2 := ContentHash(b)
	if h1 != h2 {
		t.Fatalf("hash not stable: %s vs %s", h1, h2)
	}
	b.UserMessage = "hello!"
	if h3 := ContentHash(b); h3 == h1 {
		t.Fatalf("hash should change with user message: %s", h3)
	}
}

// TestContentHashToolCallOrderInsensitive ensures tool-call ordering
// in the bundle does not affect the hash — necessary because tool
// calls may arrive in varying order on re-reads.
func TestContentHashToolCallOrderInsensitive(t *testing.T) {
	a := TurnBundle{
		UserMessage: "x",
		ToolCalls: []ToolCallBundle{
			{ToolName: "A", InputText: "1"},
			{ToolName: "B", InputText: "2"},
		},
	}
	b := TurnBundle{
		UserMessage: "x",
		ToolCalls: []ToolCallBundle{
			{ToolName: "B", InputText: "2"},
			{ToolName: "A", InputText: "1"},
		},
	}
	if ContentHash(a) != ContentHash(b) {
		t.Fatalf("tool-call order should not affect hash")
	}
}
