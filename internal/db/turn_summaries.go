package db

import (
	"context"
	"database/sql"
	"fmt"
)

// TurnSummary is one LLM-generated summary for a single turn.
type TurnSummary struct {
	SessionID     string
	TurnIndex     int
	StartOrdinal  int
	EndOrdinal    int
	ContentHash   string
	Summary       string
	Intent        string
	Outcome       string
	Topic         string
	FilesTouched  string // JSON array
	Tags          string // JSON array
	Model         string
	PromptVersion int
	CreatedAt     string
}

// UpsertTurnSummary inserts a summary, or no-ops if a row with the
// same (session_id, turn_index, content_hash) already exists.
func (db *DB) UpsertTurnSummary(s TurnSummary) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	_, err := db.getWriter().Exec(`
		INSERT OR IGNORE INTO context_turn_summaries (
			session_id, turn_index, start_ordinal, end_ordinal,
			content_hash, summary, intent, outcome, topic,
			files_touched, tags, model, prompt_version
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.SessionID, s.TurnIndex, s.StartOrdinal, s.EndOrdinal,
		s.ContentHash, s.Summary, s.Intent, s.Outcome, s.Topic,
		s.FilesTouched, s.Tags, s.Model, s.PromptVersion,
	)
	if err != nil {
		return fmt.Errorf(
			"upsert turn summary %s:%d: %w",
			s.SessionID, s.TurnIndex, err,
		)
	}
	return nil
}

// ListTurnSummaries returns all summaries for a session ordered by
// turn_index. When multiple rows share a turn_index (content
// re-hashed), only the most recent is returned.
func (db *DB) ListTurnSummaries(
	ctx context.Context, sessionID string,
) ([]TurnSummary, error) {
	rows, err := db.getReader().QueryContext(ctx, `
		SELECT session_id, turn_index, start_ordinal, end_ordinal,
		       content_hash, summary, intent, outcome, topic,
		       files_touched, tags, model, prompt_version, created_at
		FROM context_turn_summaries
		WHERE session_id = ?
		  AND id IN (
		    SELECT MAX(id) FROM context_turn_summaries
		    WHERE session_id = ?
		    GROUP BY turn_index
		  )
		ORDER BY turn_index`,
		sessionID, sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"list turn summaries %s: %w", sessionID, err,
		)
	}
	defer rows.Close()
	var out []TurnSummary
	for rows.Next() {
		var s TurnSummary
		if err := rows.Scan(
			&s.SessionID, &s.TurnIndex, &s.StartOrdinal, &s.EndOrdinal,
			&s.ContentHash, &s.Summary, &s.Intent, &s.Outcome, &s.Topic,
			&s.FilesTouched, &s.Tags, &s.Model, &s.PromptVersion,
			&s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan turn summary: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// HasTurnSummary reports whether a summary with the given content
// hash already exists. Used by the worker to skip fresh generations.
func (db *DB) HasTurnSummary(
	ctx context.Context,
	sessionID string, turnIndex int, contentHash string,
) (bool, error) {
	var id int64
	err := db.getReader().QueryRowContext(ctx, `
		SELECT id FROM context_turn_summaries
		WHERE session_id = ? AND turn_index = ? AND content_hash = ?
		LIMIT 1`,
		sessionID, turnIndex, contentHash,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf(
			"has turn summary %s:%d: %w",
			sessionID, turnIndex, err,
		)
	}
	return true, nil
}
