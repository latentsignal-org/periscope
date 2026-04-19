package db

import (
	"context"
	"testing"
)

// TestTurnSummaryRoundTrip covers upsert + list + has-check and the
// INSERT OR IGNORE idempotence guarantee on (session, turn, hash).
func TestTurnSummaryRoundTrip(t *testing.T) {
	d := testDB(t)
	insertSession(t, d, "sess-1", "proj")

	ctx := context.Background()

	ts := TurnSummary{
		SessionID:     "sess-1",
		TurnIndex:     1,
		StartOrdinal:  1,
		EndOrdinal:    3,
		ContentHash:   "hashA",
		Summary:       "did a thing",
		Intent:        "implement",
		Outcome:       "progress",
		Topic:         "auth refactor",
		FilesTouched:  `["auth.go"]`,
		Tags:          `["retry-loop"]`,
		Model:         "claude-haiku-4-5",
		PromptVersion: 1,
	}
	if err := d.UpsertTurnSummary(ts); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	// Duplicate upsert with same (session,turn,hash) must no-op.
	if err := d.UpsertTurnSummary(ts); err != nil {
		t.Fatalf("duplicate upsert: %v", err)
	}

	got, err := d.ListTurnSummaries(ctx, "sess-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListTurnSummaries len = %d, want 1 (dedup by unique key)",
			len(got))
	}
	if got[0].Summary != ts.Summary ||
		got[0].Topic != ts.Topic ||
		got[0].PromptVersion != 1 {
		t.Errorf("round-trip mismatch: %+v", got[0])
	}
	if got[0].CreatedAt == "" {
		t.Errorf("CreatedAt default not populated")
	}

	ok, err := d.HasTurnSummary(ctx, "sess-1", 1, "hashA")
	if err != nil || !ok {
		t.Errorf("HasTurnSummary exact = (%v,%v), want (true,nil)", ok, err)
	}
	ok, err = d.HasTurnSummary(ctx, "sess-1", 1, "hashB")
	if err != nil || ok {
		t.Errorf("HasTurnSummary mismatch hash = (%v,%v), want (false,nil)", ok, err)
	}

	// New content hash on the same turn = new row kept, list returns
	// the most recent (highest id).
	ts2 := ts
	ts2.ContentHash = "hashB"
	ts2.Summary = "did another thing"
	if err := d.UpsertTurnSummary(ts2); err != nil {
		t.Fatalf("upsert new hash: %v", err)
	}
	got, err = d.ListTurnSummaries(ctx, "sess-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("ListTurnSummaries len = %d, want 1 (latest per turn)",
			len(got))
	}
	if got[0].Summary != "did another thing" {
		t.Errorf("expected latest summary, got %q", got[0].Summary)
	}
}

// TestIsSessionStarred reflects star/unstar state.
func TestIsSessionStarred(t *testing.T) {
	d := testDB(t)
	insertSession(t, d, "sess-1", "proj")
	ctx := context.Background()

	ok, err := d.IsSessionStarred(ctx, "sess-1")
	if err != nil || ok {
		t.Fatalf("initial IsSessionStarred = (%v,%v), want (false,nil)", ok, err)
	}
	if _, err := d.StarSession("sess-1"); err != nil {
		t.Fatalf("StarSession: %v", err)
	}
	ok, err = d.IsSessionStarred(ctx, "sess-1")
	if err != nil || !ok {
		t.Fatalf("after star, IsSessionStarred = (%v,%v), want (true,nil)", ok, err)
	}
	if err := d.UnstarSession("sess-1"); err != nil {
		t.Fatalf("UnstarSession: %v", err)
	}
	ok, err = d.IsSessionStarred(ctx, "sess-1")
	if err != nil || ok {
		t.Fatalf("after unstar, IsSessionStarred = (%v,%v), want (false,nil)", ok, err)
	}
}
