package db

import (
	"context"
	"fmt"
	"strings"
)

// SecretFindingFilter narrows a findings listing. Empty fields = no filter.
type SecretFindingFilter struct {
	Project    string
	Agent      string
	DateFrom   string
	DateTo     string
	Rule       string
	Confidence string // definite | candidate | "" (all)
	// RulesVersions, when non-empty, limits rows to findings produced by one
	// of the currently accepted scanner versions. This lets service callers
	// hide stale findings after rule/fixture-deny changes before a backfill has
	// rewritten old rows.
	RulesVersions []string
	Limit         int
	Cursor        int
}

// SecretFindingRow is a finding enriched with its session's project/agent.
type SecretFindingRow struct {
	SecretFinding
	Project string `json:"project"`
	Agent   string `json:"agent"`
}

// SecretFindingPage is a page of findings.
type SecretFindingPage struct {
	Findings   []SecretFindingRow `json:"findings"`
	NextCursor int                `json:"next_cursor,omitempty"`
}

// ListSecretFindings returns findings joined to their sessions, filtered.
func (db *DB) ListSecretFindings(
	ctx context.Context, f SecretFindingFilter,
) (SecretFindingPage, error) {
	if f.Limit <= 0 || f.Limit > MaxContentSearchLimit {
		f.Limit = DefaultContentSearchLimit
	}
	var preds []string
	var args []any
	add := func(p string, v any) { preds = append(preds, p); args = append(args, v) }
	if f.Project != "" {
		add("s.project = ?", f.Project)
	}
	if f.Agent != "" {
		add("s.agent = ?", f.Agent)
	}
	if f.Rule != "" {
		add("sf.rule_name = ?", f.Rule)
	}
	if f.Confidence != "" && f.Confidence != "all" {
		add("sf.confidence = ?", f.Confidence)
	}
	if len(f.RulesVersions) > 0 {
		placeholders := make([]string, 0, len(f.RulesVersions))
		for _, v := range f.RulesVersions {
			if v == "" {
				continue
			}
			placeholders = append(placeholders, "?")
			args = append(args, v)
		}
		if len(placeholders) > 0 {
			preds = append(preds,
				"sf.rules_version IN ("+strings.Join(placeholders, ",")+")")
		}
	}
	if f.DateFrom != "" {
		add("date(COALESCE(NULLIF(s.started_at, ''), s.created_at)) >= ?", f.DateFrom)
	}
	if f.DateTo != "" {
		add("date(COALESCE(NULLIF(s.started_at, ''), s.created_at)) <= ?", f.DateTo)
	}
	where := "s.deleted_at IS NULL"
	if len(preds) > 0 {
		where += " AND " + strings.Join(preds, " AND ")
	}
	query := `
		SELECT sf.session_id, sf.rule_name, sf.confidence, sf.location_kind,
			sf.message_ordinal, sf.call_index, sf.event_index,
			sf.match_start, sf.match_end, sf.match_index,
			sf.redacted_match, sf.rules_version, s.project, s.agent
		FROM secret_findings sf JOIN sessions s ON s.id = sf.session_id
		WHERE ` + where + `
		ORDER BY julianday(COALESCE(NULLIF(s.ended_at, ''),
				NULLIF(s.started_at, ''), s.created_at)) DESC,
			sf.session_id, sf.message_ordinal, sf.match_start,
			sf.match_index, sf.id
		LIMIT ? OFFSET ?`
	args = append(args, f.Limit+1, f.Cursor)

	rows, err := db.getReader().QueryContext(ctx, query, args...)
	if err != nil {
		return SecretFindingPage{}, fmt.Errorf("listing findings: %w", err)
	}
	defer rows.Close()
	out := make([]SecretFindingRow, 0, f.Limit+1)
	for rows.Next() {
		var r SecretFindingRow
		if err := rows.Scan(&r.SessionID, &r.RuleName, &r.Confidence,
			&r.LocationKind, &r.MessageOrdinal, &r.CallIndex, &r.EventIndex,
			&r.MatchStart, &r.MatchEnd, &r.MatchIndex, &r.RedactedMatch,
			&r.RulesVersion, &r.Project, &r.Agent); err != nil {
			return SecretFindingPage{}, fmt.Errorf("scan finding row: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return SecretFindingPage{}, err
	}
	page := SecretFindingPage{Findings: out}
	if len(out) > f.Limit {
		page.Findings = out[:f.Limit]
		page.NextCursor = f.Cursor + f.Limit
	}
	return page, nil
}
