package db

import (
	"context"
	"fmt"
	"strings"
)

// SecretScanCandidateFilter selects sessions to scan. When OnlyStale is true,
// only sessions whose secrets_rules_version != CurrentVersion are returned
// (resumable backfill); otherwise all sessions matching the project/agent/date
// filters are returned (forced rescan). Sessions with no messages are excluded.
type SecretScanCandidateFilter struct {
	CurrentVersion string
	OnlyStale      bool
	Project        string
	Agent          string
	DateFrom       string
	DateTo         string
}

// SecretScanCandidates returns session IDs needing a secret scan, oldest first
// (a stable order for resumable batched processing).
func (db *DB) SecretScanCandidates(
	ctx context.Context, f SecretScanCandidateFilter,
) ([]string, error) {
	preds := []string{"message_count > 0", "deleted_at IS NULL"}
	var args []any
	if f.OnlyStale {
		preds = append(preds, "secrets_rules_version != ?")
		args = append(args, f.CurrentVersion)
	}
	if f.Project != "" {
		preds = append(preds, "project = ?")
		args = append(args, f.Project)
	}
	if f.Agent != "" {
		preds = append(preds, "agent = ?")
		args = append(args, f.Agent)
	}
	if f.DateFrom != "" {
		preds = append(preds, "date(COALESCE(NULLIF(started_at, ''), created_at)) >= ?")
		args = append(args, f.DateFrom)
	}
	if f.DateTo != "" {
		preds = append(preds, "date(COALESCE(NULLIF(started_at, ''), created_at)) <= ?")
		args = append(args, f.DateTo)
	}
	rows, err := db.getReader().QueryContext(ctx,
		"SELECT id FROM sessions WHERE "+strings.Join(preds, " AND ")+
			" ORDER BY COALESCE(NULLIF(started_at, ''), created_at) ASC, id ASC",
		args...)
	if err != nil {
		return nil, fmt.Errorf("secret scan candidates: %w", err)
	}
	defer rows.Close()
	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan candidate id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
