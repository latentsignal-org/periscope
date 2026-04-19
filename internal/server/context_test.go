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
