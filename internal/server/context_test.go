package server

import (
	"testing"

	"github.com/wesm/agentsview/internal/db"
)

func TestComputeSessionContextView_TrimsToLatestCompaction(t *testing.T) {
	session := db.Session{ID: "sess-1", Agent: "claude"}
	msgs := []db.Message{
		{SessionID: "sess-1", Ordinal: 1, Role: "user", ContentLength: 20},
		{
			SessionID:         "sess-1",
			Ordinal:           2,
			Role:              "system",
			ContentLength:     80,
			IsCompactBoundary: true,
			ContextTokens:     1200,
			HasContextTokens:  true,
		},
		{
			SessionID:        "sess-1",
			Ordinal:          3,
			Role:             "assistant",
			ContentLength:    40,
			ContextTokens:    1500,
			HasContextTokens: true,
			Model:            "claude-sonnet-4-5",
		},
	}

	view := computeSessionContextView(session, msgs)

	if got := len(view.Timeline); got != 2 {
		t.Fatalf("timeline rows = %d, want 2", got)
	}
	if got := view.Summary.VisibleSinceOrdinal; got != 2 {
		t.Fatalf("visible_since_ordinal = %d, want 2", got)
	}
	if !view.Supports.CompactionTrimmed {
		t.Fatal("supports.compaction_trimmed = false, want true")
	}
	if got := view.Summary.TokensInUse; got != 1500 {
		t.Fatalf("tokens_in_use = %d, want 1500", got)
	}
}

func TestComputeSessionContextView_InfersCapacityFromModel(t *testing.T) {
	session := db.Session{ID: "sess-1", Agent: "codex"}
	msgs := []db.Message{
		{
			SessionID:        "sess-1",
			Ordinal:          1,
			Role:             "assistant",
			ContentLength:    40,
			ContextTokens:    1000,
			HasContextTokens: true,
			Model:            "gpt-4o",
		},
	}

	view := computeSessionContextView(session, msgs)

	if got := view.Capacity.MaxTokens; got != 128000 {
		t.Fatalf("capacity.max_tokens = %d, want 128000", got)
	}
	if got := view.Capacity.Provenance; got != contextProvenanceInferred {
		t.Fatalf("capacity.provenance = %q, want %q", got, contextProvenanceInferred)
	}
	if !view.Summary.RemainingKnown {
		t.Fatal("summary.remaining_known = false, want true")
	}
}

func TestComputeSessionContextView_GroupsTimelineByTurn(t *testing.T) {
	session := db.Session{ID: "sess-1", Agent: "codex"}
	msgs := []db.Message{
		{
			SessionID:        "sess-1",
			Ordinal:          1,
			Role:             "user",
			Content:          "Investigate why the tests are failing after the refactor.",
			ContentLength:    58,
			ContextTokens:    100,
			HasContextTokens: true,
		},
		{
			SessionID:        "sess-1",
			Ordinal:          2,
			Role:             "assistant",
			Content:          "I will inspect the failing files first.",
			ContentLength:    38,
			ContextTokens:    240,
			HasContextTokens: true,
			HasToolUse:       true,
			ToolCalls: []db.ToolCall{
				{
					ToolName:  "Read",
					Category:  "Read",
					InputJSON: `{"file_path":"internal/server/context.go"}`,
				},
			},
		},
		{
			SessionID:        "sess-1",
			Ordinal:          3,
			Role:             "assistant",
			Content:          "The issue is a stale assertion in the test fixture.",
			ContentLength:    52,
			ContextTokens:    320,
			HasContextTokens: true,
		},
		{
			SessionID:        "sess-1",
			Ordinal:          4,
			Role:             "user",
			Content:          "Patch it and rerun the targeted tests.",
			ContentLength:    37,
			ContextTokens:    420,
			HasContextTokens: true,
		},
		{
			SessionID:        "sess-1",
			Ordinal:          5,
			Role:             "assistant",
			Content:          "Patched. The targeted test now passes.",
			ContentLength:    38,
			ContextTokens:    520,
			HasContextTokens: true,
			HasToolUse:       true,
			ToolCalls: []db.ToolCall{
				{
					ToolName:  "Bash",
					Category:  "Bash",
					InputJSON: `{"command":"go test ./internal/server -run TestComputeSessionContextView"}`,
				},
			},
		},
	}

	view := computeSessionContextView(session, msgs)

	if got := len(view.Timeline); got != 2 {
		t.Fatalf("len(timeline) = %d, want 2 turns", got)
	}

	first := view.Timeline[0]
	if first.StartOrdinal != 1 || first.EndOrdinal != 3 {
		t.Fatalf("first turn ordinals = %d-%d, want 1-3", first.StartOrdinal, first.EndOrdinal)
	}
	if first.UserMessage == nil || first.UserMessage.Ordinal != 1 {
		t.Fatalf("first turn user message = %+v, want ordinal 1", first.UserMessage)
	}
	if first.AssistantMessage == nil || first.AssistantMessage.Ordinal != 3 {
		t.Fatalf("first turn assistant message = %+v, want ordinal 3", first.AssistantMessage)
	}
	if len(first.ToolCalls) != 1 || first.ToolCalls[0].Ordinal != 2 {
		t.Fatalf("first turn tool calls = %+v, want one tool at ordinal 2", first.ToolCalls)
	}
	if len(first.Entries) != 4 {
		t.Fatalf("len(first.entries) = %d, want 4", len(first.Entries))
	}

	second := view.Timeline[1]
	if second.StartOrdinal != 4 || second.EndOrdinal != 5 {
		t.Fatalf("second turn ordinals = %d-%d, want 4-5", second.StartOrdinal, second.EndOrdinal)
	}
	if second.UserMessage == nil || second.UserMessage.Ordinal != 4 {
		t.Fatalf("second turn user message = %+v, want ordinal 4", second.UserMessage)
	}
	if second.AssistantMessage == nil || second.AssistantMessage.Ordinal != 5 {
		t.Fatalf("second turn assistant message = %+v, want ordinal 5", second.AssistantMessage)
	}
	if len(second.ToolCalls) != 1 || second.ToolCalls[0].ToolName != "Bash" {
		t.Fatalf("second turn tool calls = %+v, want Bash", second.ToolCalls)
	}
	if len(second.Entries) != 3 {
		t.Fatalf("len(second.entries) = %d, want 3", len(second.Entries))
	}
}
