package db

import (
	"context"
	"fmt"
)

// Stats holds database-wide statistics.
type Stats struct {
	SessionCount    int     `json:"session_count"`
	MessageCount    int     `json:"message_count"`
	ProjectCount    int     `json:"project_count"`
	MachineCount    int     `json:"machine_count"`
	EarliestSession *string `json:"earliest_session"`
}

// rootSessionFilter is the WHERE clause shared by session list
// and stats to exclude sub-agent, fork, and trashed sessions.
const rootSessionFilter = `message_count > 0
	AND relationship_type NOT IN ('subagent', 'fork')
	AND deleted_at IS NULL`

// FileBackedSessionCount returns the number of root sessions
// synced from files (excludes OpenCode, which is DB-backed).
// Used by ResyncAll to decide whether empty file discovery
// should abort the swap.
func (db *DB) FileBackedSessionCount(
	ctx context.Context,
) (int, error) {
	var count int
	err := db.getReader().QueryRowContext(ctx,
		`SELECT COUNT(*) FROM sessions
		 WHERE agent != 'opencode'
		 AND `+rootSessionFilter,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf(
			"counting file-backed sessions: %w", err,
		)
	}
	return count, nil
}

// GetStats returns database statistics, counting only root
// sessions with messages (matching the session list filter).
func (db *DB) GetStats(
	ctx context.Context,
	excludeOneShot, excludeAutomated bool,
) (Stats, error) {
	filter := rootSessionFilter
	if excludeOneShot {
		if !excludeAutomated {
			filter += " AND (user_message_count > 1 OR is_automated = 1)"
		} else {
			filter += " AND user_message_count > 1"
		}
	}
	if excludeAutomated {
		filter += " AND is_automated = 0"
	}
	query := fmt.Sprintf(`
		SELECT
			(SELECT COUNT(*) FROM sessions
			 WHERE %s),
			(SELECT COALESCE(SUM(message_count), 0)
			 FROM sessions WHERE %s),
			(SELECT COUNT(DISTINCT project) FROM sessions
			 WHERE %s),
			(SELECT COUNT(DISTINCT machine) FROM sessions
			 WHERE %s),
			(SELECT MIN(COALESCE(
				NULLIF(started_at, ''), created_at
			 )) FROM sessions
			 WHERE %s)`,
		filter, filter, filter, filter, filter)

	var s Stats
	err := db.getReader().QueryRowContext(ctx, query).Scan(
		&s.SessionCount,
		&s.MessageCount,
		&s.ProjectCount,
		&s.MachineCount,
		&s.EarliestSession,
	)
	if err != nil {
		return Stats{}, fmt.Errorf("fetching stats: %w", err)
	}
	return s, nil
}
