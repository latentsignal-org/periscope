package db

import (
	"context"
	"testing"
)

func TestBackfillIsAutomatedBidirectional(t *testing.T) {
	d := testDB(t)

	// Seed a false negative: single-turn roborev session with
	// is_automated = 0 (simulates pre-migration data).
	insertSession(t, d, "missed", "proj", func(s *Session) {
		fm := "You are a code reviewer. Review the code."
		s.FirstMessage = &fm
		s.MessageCount = 3
		s.UserMessageCount = 1
	})
	// Force is_automated to 0 to simulate pre-migration state.
	_, err := d.getWriter().Exec(
		"UPDATE sessions SET is_automated = 0 WHERE id = 'missed'",
	)
	requireNoError(t, err, "force missed to 0")

	// Seed a stale false positive: multi-turn session that was
	// previously marked automated under old broad rules.
	insertSession(t, d, "stale", "proj", func(s *Session) {
		fm := "# Fix Request for login flow"
		s.FirstMessage = &fm
		s.MessageCount = 10
		s.UserMessageCount = 5
	})
	_, err = d.getWriter().Exec(
		"UPDATE sessions SET is_automated = 1 WHERE id = 'stale'",
	)
	requireNoError(t, err, "force stale to 1")

	// Clear the marker so the backfill will run.
	_, err = d.getWriter().Exec(
		"DELETE FROM stats WHERE key = 'is_automated_backfill_v1'",
	)
	requireNoError(t, err, "clear marker")

	// Run backfill.
	d.mu.Lock()
	err = d.backfillIsAutomatedLocked(d.getWriter())
	d.mu.Unlock()
	requireNoError(t, err, "first backfill run")

	ctx := context.Background()

	// False negative should now be set.
	missed, err := d.GetSession(ctx, "missed")
	requireNoError(t, err, "get missed")
	if !missed.IsAutomated {
		t.Error("missed session should be automated after backfill")
	}

	// Stale false positive should now be cleared.
	stale, err := d.GetSession(ctx, "stale")
	requireNoError(t, err, "get stale")
	if stale.IsAutomated {
		t.Error("stale session should not be automated after backfill")
	}
}

func TestBackfillIsAutomatedMarkerIdempotent(t *testing.T) {
	d := testDB(t)

	// Seed a roborev session.
	insertSession(t, d, "review", "proj", func(s *Session) {
		fm := "You are a code reviewer. Review the code."
		s.FirstMessage = &fm
		s.MessageCount = 3
		s.UserMessageCount = 1
	})

	// Clear the marker and run backfill.
	_, err := d.getWriter().Exec(
		"DELETE FROM stats WHERE key = 'is_automated_backfill_v1'",
	)
	requireNoError(t, err, "clear marker")

	d.mu.Lock()
	err = d.backfillIsAutomatedLocked(d.getWriter())
	d.mu.Unlock()
	requireNoError(t, err, "first run")

	// Manually corrupt the session to verify second run is a no-op.
	_, err = d.getWriter().Exec(
		"UPDATE sessions SET is_automated = 0 WHERE id = 'review'",
	)
	requireNoError(t, err, "corrupt")

	// Second run should be a no-op (marker present).
	d.mu.Lock()
	err = d.backfillIsAutomatedLocked(d.getWriter())
	d.mu.Unlock()
	requireNoError(t, err, "second run")

	ctx := context.Background()
	review, err := d.GetSession(ctx, "review")
	requireNoError(t, err, "get review")
	if review.IsAutomated {
		t.Error("second run should be no-op; is_automated should still be 0")
	}
}
