//go:build pgtest

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/wesm/agentsview/internal/db"
)

// TestStoreGetAnalyticsSignals exercises the PG implementation
// of GetAnalyticsSignals end to end: seed signals on local
// rows, push to PG, then read them back through the Store and
// confirm the aggregated response matches what the SQLite
// implementation would have produced over the same data.
func TestStoreGetAnalyticsSignals(t *testing.T) {
	pgURL := testPGURL(t)
	cleanPGSchema(t, pgURL)
	t.Cleanup(func() { cleanPGSchema(t, pgURL) })

	local := testDB(t)
	ps, err := New(
		pgURL, "agentsview", local,
		"signals-test-machine", true,
		SyncOptions{},
	)
	if err != nil {
		t.Fatalf("creating sync: %v", err)
	}
	defer ps.Close()

	ctx := context.Background()
	if err := ps.EnsureSchema(ctx); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	score := 90
	grade := "A"
	pressure := 0.42
	started := time.Now().UTC().Add(-1 * time.Hour).
		Format(time.RFC3339)
	first := "hi"

	for _, id := range []string{"sig-1", "sig-2"} {
		sess := db.Session{
			ID:           id,
			Project:      "proj",
			Machine:      "local",
			Agent:        "claude",
			FirstMessage: &first,
			StartedAt:    &started,
			MessageCount: 4,
		}
		if err := local.UpsertSession(sess); err != nil {
			t.Fatalf("upsert %s: %v", id, err)
		}
		if err := local.UpdateSessionSignals(
			id,
			db.SessionSignalUpdate{
				Outcome:                "completed",
				OutcomeConfidence:      "high",
				EndedWithRole:          "assistant",
				HasToolCalls:           true,
				HasContextData:         true,
				ToolFailureSignalCount: 1,
				CompactionCount:        1,
				MidTaskCompactionCount: 1,
				ContextPressureMax:     &pressure,
				HealthScore:            &score,
				HealthGrade:            &grade,
			},
		); err != nil {
			t.Fatalf("UpdateSessionSignals %s: %v", id, err)
		}
	}

	if _, err := ps.Push(ctx, false, nil); err != nil {
		t.Fatalf("push: %v", err)
	}

	store, err := NewStore(pgURL, "agentsview", true)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer store.Close()

	// Empty AnalyticsFilter must be accepted -- exercises the
	// sentinel-bound path in analyticsUTCRange that earlier
	// produced "T00:00:00Z" and tripped PG's TIMESTAMPTZ cast.
	resp, err := store.GetAnalyticsSignals(
		ctx, db.AnalyticsFilter{},
	)
	if err != nil {
		t.Fatalf("GetAnalyticsSignals: %v", err)
	}

	if resp.ScoredSessions != 2 {
		t.Errorf(
			"ScoredSessions = %d, want 2",
			resp.ScoredSessions,
		)
	}
	if resp.GradeDistribution["A"] != 2 {
		t.Errorf(
			"GradeDistribution[A] = %d, want 2",
			resp.GradeDistribution["A"],
		)
	}
	if resp.OutcomeDistribution["completed"] != 2 {
		t.Errorf(
			"OutcomeDistribution[completed] = %d, want 2",
			resp.OutcomeDistribution["completed"],
		)
	}
	if resp.AvgHealthScore == nil ||
		*resp.AvgHealthScore != 90 {
		t.Errorf(
			"AvgHealthScore = %v, want 90", resp.AvgHealthScore,
		)
	}
	if resp.ContextHealth.MidTaskCompactionCount != 2 {
		t.Errorf(
			"MidTaskCompactionCount = %d, want 2",
			resp.ContextHealth.MidTaskCompactionCount,
		)
	}
	if resp.ToolHealth.TotalFailureSignals != 2 {
		t.Errorf(
			"TotalFailureSignals = %d, want 2",
			resp.ToolHealth.TotalFailureSignals,
		)
	}
	if len(resp.ByAgent) != 1 ||
		resp.ByAgent[0].Agent != "claude" ||
		resp.ByAgent[0].SessionCount != 2 {
		t.Errorf("ByAgent = %+v, want [claude / 2]", resp.ByAgent)
	}
	if len(resp.ByProject) != 1 ||
		resp.ByProject[0].Project != "proj" ||
		resp.ByProject[0].SessionCount != 2 {
		t.Errorf(
			"ByProject = %+v, want [proj / 2]", resp.ByProject,
		)
	}
}
