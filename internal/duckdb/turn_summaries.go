package duckdb

import (
	"context"

	"go.kenn.io/agentsview/internal/db"
)

// UpsertTurnSummary is not supported on the read-only DuckDB mirror.
func (s *Store) UpsertTurnSummary(_ db.TurnSummary) error {
	return db.ErrReadOnly
}

// ListTurnSummaries returns an empty slice; turn summaries are
// local-only and are not mirrored into DuckDB.
func (s *Store) ListTurnSummaries(
	_ context.Context, _ string,
) ([]db.TurnSummary, error) {
	return []db.TurnSummary{}, nil
}

// HasTurnSummary always returns false on the DuckDB mirror.
func (s *Store) HasTurnSummary(
	_ context.Context, _ string, _ int, _ string,
) (bool, error) {
	return false, nil
}
