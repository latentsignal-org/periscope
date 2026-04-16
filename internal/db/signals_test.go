package db

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestUpdateSessionSignals(t *testing.T) {
	d := testDB(t)
	ctx := context.Background()

	insertSession(t, d, "sig-1", "proj", func(s *Session) {
		s.MessageCount = 5
	})

	update := SessionSignalUpdate{
		ToolFailureSignalCount: 3,
		ToolRetryCount:         2,
		EditChurnCount:         1,
		ConsecutiveFailureMax:  4,
		Outcome:                "completed",
		OutcomeConfidence:      "high",
		EndedWithRole:          "assistant",
		FinalFailureStreak:     0,
		SignalsPendingSince:    nil,
		CompactionCount:        2,
		ContextPressureMax:     Ptr(0.85),
		HealthScore:            Ptr(72),
		HealthGrade:            Ptr("B"),
	}
	if err := d.UpdateSessionSignals("sig-1", update); err != nil {
		t.Fatalf("UpdateSessionSignals: %v", err)
	}

	got, err := d.GetSessionFull(ctx, "sig-1")
	if err != nil {
		t.Fatalf("GetSessionFull: %v", err)
	}
	if got == nil {
		t.Fatal("session not found after update")
	}

	checks := []struct {
		name string
		got  any
		want any
	}{
		{"ToolFailureSignalCount", got.ToolFailureSignalCount, 3},
		{"ToolRetryCount", got.ToolRetryCount, 2},
		{"EditChurnCount", got.EditChurnCount, 1},
		{"ConsecutiveFailureMax", got.ConsecutiveFailureMax, 4},
		{"Outcome", got.Outcome, "completed"},
		{"OutcomeConfidence", got.OutcomeConfidence, "high"},
		{"EndedWithRole", got.EndedWithRole, "assistant"},
		{"FinalFailureStreak", got.FinalFailureStreak, 0},
		{"CompactionCount", got.CompactionCount, 2},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}

	if got.SignalsPendingSince != nil {
		t.Errorf(
			"SignalsPendingSince = %v, want nil",
			*got.SignalsPendingSince,
		)
	}
	if got.ContextPressureMax == nil || *got.ContextPressureMax != 0.85 {
		t.Errorf(
			"ContextPressureMax = %v, want 0.85",
			got.ContextPressureMax,
		)
	}
	if got.HealthScore == nil || *got.HealthScore != 72 {
		t.Errorf("HealthScore = %v, want 72", got.HealthScore)
	}
	if got.HealthGrade == nil || *got.HealthGrade != "B" {
		t.Errorf("HealthGrade = %v, want B", got.HealthGrade)
	}

	// Update again with pending since set and nullable fields
	// cleared.
	pending := "2024-06-01T00:00:00Z"
	update2 := SessionSignalUpdate{
		Outcome:             "unknown",
		OutcomeConfidence:   "low",
		SignalsPendingSince: &pending,
	}
	if err := d.UpdateSessionSignals("sig-1", update2); err != nil {
		t.Fatalf("UpdateSessionSignals (2nd): %v", err)
	}

	got2, err := d.GetSessionFull(ctx, "sig-1")
	if err != nil {
		t.Fatalf("GetSessionFull (2nd): %v", err)
	}

	// Verify signals_pending_since is loaded by GetSessionFull
	// (was previously absent from the column lists).
	if got2.SignalsPendingSince == nil ||
		*got2.SignalsPendingSince != pending {
		t.Errorf(
			"SignalsPendingSince = %v, want %q",
			got2.SignalsPendingSince, pending,
		)
	}

	pendingIDs, err := d.PendingSignalSessions(
		ctx, "2024-07-01T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("PendingSignalSessions: %v", err)
	}
	if len(pendingIDs) != 1 || pendingIDs[0] != "sig-1" {
		t.Errorf(
			"PendingSignalSessions = %v, want [sig-1]",
			pendingIDs,
		)
	}

	if got2.ContextPressureMax != nil {
		t.Errorf(
			"ContextPressureMax = %v, want nil",
			*got2.ContextPressureMax,
		)
	}
	if got2.HealthScore != nil {
		t.Errorf("HealthScore = %v, want nil", *got2.HealthScore)
	}
	if got2.HealthGrade != nil {
		t.Errorf("HealthGrade = %v, want nil", *got2.HealthGrade)
	}
}

// TestUpdateSessionSignalsBumpsLocalModifiedAt ensures that
// signal updates bump local_modified_at so the session is
// re-selected by the next pg push. Without this bump, sessions
// backfilled by BackfillSignals (e.g. after a PG schema
// migration adds new signal columns) would never propagate to
// PG-backed deployments.
func TestUpdateSessionSignalsBumpsLocalModifiedAt(t *testing.T) {
	d := testDB(t)
	ctx := context.Background()

	insertSession(t, d, "lm-1", "proj")

	// Snapshot local_modified_at after the initial upsert.
	beforeRow, err := d.GetSessionFull(ctx, "lm-1")
	if err != nil {
		t.Fatalf("GetSessionFull: %v", err)
	}
	if beforeRow == nil {
		t.Fatal("session not found before update")
	}
	before := ""
	if beforeRow.LocalModifiedAt != nil {
		before = *beforeRow.LocalModifiedAt
	}

	// SQLite's strftime('now') ticks at millisecond precision.
	// Sleep a few ms so a re-set produces a strictly later value.
	time.Sleep(5 * time.Millisecond)

	if err := d.UpdateSessionSignals("lm-1", SessionSignalUpdate{
		ToolFailureSignalCount: 1,
		Outcome:                "completed",
		OutcomeConfidence:      "high",
		EndedWithRole:          "assistant",
	}); err != nil {
		t.Fatalf("UpdateSessionSignals: %v", err)
	}

	afterRow, err := d.GetSessionFull(ctx, "lm-1")
	if err != nil {
		t.Fatalf("GetSessionFull (after): %v", err)
	}
	if afterRow.LocalModifiedAt == nil ||
		*afterRow.LocalModifiedAt == "" {
		t.Fatal("local_modified_at not set after signal update")
	}
	if *afterRow.LocalModifiedAt <= before {
		t.Errorf(
			"local_modified_at not bumped: before=%q after=%q",
			before, *afterRow.LocalModifiedAt,
		)
	}
}

func TestPendingSignalSessions(t *testing.T) {
	d := testDB(t)
	ctx := context.Background()

	cutoff := "2024-06-01T12:00:00Z"

	// Session with pending_since before cutoff -- should match.
	insertSession(t, d, "ps-old", "proj")
	old := SessionSignalUpdate{
		Outcome:             "unknown",
		OutcomeConfidence:   "low",
		SignalsPendingSince: Ptr("2024-06-01T10:00:00Z"),
	}
	if err := d.UpdateSessionSignals("ps-old", old); err != nil {
		t.Fatalf("UpdateSessionSignals ps-old: %v", err)
	}

	// Session with pending_since after cutoff -- should NOT match.
	insertSession(t, d, "ps-new", "proj")
	newer := SessionSignalUpdate{
		Outcome:             "unknown",
		OutcomeConfidence:   "low",
		SignalsPendingSince: Ptr("2024-06-01T14:00:00Z"),
	}
	if err := d.UpdateSessionSignals("ps-new", newer); err != nil {
		t.Fatalf("UpdateSessionSignals ps-new: %v", err)
	}

	// Session with no pending_since -- should NOT match.
	insertSession(t, d, "ps-none", "proj")

	ids, err := d.PendingSignalSessions(ctx, cutoff)
	if err != nil {
		t.Fatalf("PendingSignalSessions: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("got %d IDs, want 1", len(ids))
	}
	if ids[0] != "ps-old" {
		t.Errorf("got ID %q, want ps-old", ids[0])
	}
}

// TestBackfillSignalsMarkerOnlyOnSuccess guards the
// completion-marker contract: the one-shot marker must only be
// set when every session was processed successfully. Partial
// runs (e.g. a concurrent resync that disconnects the DB
// mid-backfill) must leave the marker unset so the next
// startup retries.
func TestBackfillSignalsMarkerOnlyOnSuccess(t *testing.T) {
	d := testDB(t)
	ctx := context.Background()

	insertSession(t, d, "ok-1", "p")
	insertSession(t, d, "ok-2", "p")
	insertSession(t, d, "fail-1", "p")

	// One session fails -- marker must NOT be set.
	err := d.BackfillSignals(
		ctx,
		func(_ context.Context, id string) error {
			if id == "fail-1" {
				return fmt.Errorf("simulated failure")
			}
			return nil
		},
	)
	if err == nil {
		t.Fatal("expected error from partial backfill, got nil")
	}

	// Marker check: a second BackfillSignals call must NOT
	// short-circuit since the marker is unset.
	calls := 0
	err = d.BackfillSignals(
		ctx,
		func(_ context.Context, _ string) error {
			calls++
			return nil
		},
	)
	if err != nil {
		t.Fatalf("retry: %v", err)
	}
	if calls != 3 {
		t.Errorf(
			"second backfill saw %d sessions, want 3 "+
				"(marker should not be set after partial run)",
			calls,
		)
	}

	// Now the marker should be set; a third call short-circuits.
	calls = 0
	err = d.BackfillSignals(
		ctx,
		func(_ context.Context, _ string) error {
			calls++
			return nil
		},
	)
	if err != nil {
		t.Fatalf("third call: %v", err)
	}
	if calls != 0 {
		t.Errorf(
			"third backfill saw %d sessions, want 0 "+
				"(marker should be set after clean run)",
			calls,
		)
	}
}
