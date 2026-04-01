# Roborev Detection & Filter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Detect automated roborev review/fix sessions via
first_message heuristics, hide them by default, and flip single-turn
sessions to shown-by-default with a "Hide single-turn" toggle.

**Architecture:** Add `is_automated` boolean column to the sessions
table (SQLite + PG). Compute it from `first_message` during upsert
using prefix-matching heuristics. Add `ExcludeAutomated` filter
alongside existing `ExcludeOneShot`. Frontend flips single-turn
default and adds automated toggle.

**Tech Stack:** Go (SQLite, PostgreSQL), Svelte 5 (TypeScript)

---

## File Structure

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `internal/db/automated.go` | Detection heuristic function |
| Create | `internal/db/automated_test.go` | Tests for detection |
| Modify | `internal/db/schema.sql` | Add `is_automated` column |
| Modify | `internal/db/db.go` | Add migration + backfill |
| Modify | `internal/db/sessions.go` | Struct, cols, scan, upsert, filter |
| Modify | `internal/db/analytics.go` | `ExcludeAutomated` in `AnalyticsFilter` |
| Modify | `internal/db/stats.go` | `excludeAutomated` param on `GetStats` |
| Modify | `internal/db/filter_test.go` | Filter tests for automated exclusion |
| Modify | `internal/server/sessions.go` | `include_automated` API param |
| Modify | `internal/server/analytics.go` | `include_automated` API param |
| Modify | `internal/server/events.go` | `include_automated` on metadata endpoints |
| Modify | `internal/server/server_test.go` | API tests |
| Modify | `internal/postgres/schema.go` | PG DDL + migration |
| Modify | `internal/postgres/sessions.go` | PG cols, scan, filter |
| Modify | `internal/postgres/analytics.go` | PG analytics filter |
| Modify | `internal/postgres/push.go` | Sync `is_automated` to PG |
| Modify | `frontend/src/lib/stores/sessions.svelte.ts` | Filters + default flip |
| Modify | `frontend/src/lib/stores/analytics.svelte.ts` | Analytics filter |
| Modify | `frontend/src/lib/components/sidebar/SessionList.svelte` | Toggle UI |
| Modify | `frontend/src/lib/components/analytics/ActiveFilters.svelte` | Filter chips |
| Modify | `frontend/src/lib/api/client.ts` | API params |
| Modify | `frontend/src/App.svelte` | URL sync |

---

### Task 1: Detection heuristic + tests

**Files:**
- Create: `internal/db/automated.go`
- Create: `internal/db/automated_test.go`

- [ ] **Step 1: Write tests for the detection heuristic**

Create `internal/db/automated_test.go`:

```go
package db

import "testing"

func TestIsAutomatedSession(t *testing.T) {
	tests := []struct {
		name         string
		firstMessage string
		want         bool
	}{
		{
			"EmptyMessage",
			"",
			false,
		},
		{
			"NormalUserPrompt",
			"fix the login bug",
			false,
		},
		{
			"RoborevReviewPrompt",
			"You are a code reviewer. Review the code changes shown below.\n\n## Changes\n...",
			true,
		},
		{
			"RoborevReviewPromptExact",
			"You are a code reviewer. Review the code changes shown below.",
			true,
		},
		{
			"RoborevFixPrompt",
			"# Fix Request\n\nAn analysis was performed and produced the following findings:\n...",
			true,
		},
		{
			"RoborevFixPromptExact",
			"# Fix Request",
			true,
		},
		{
			"SimilarButNotReview",
			"You are a code reviewer but I need help",
			false,
		},
		{
			"FixInNormalContext",
			"Fix the request handler",
			false,
		},
		{
			"NilSafeViaPointer",
			"",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAutomatedSession(tt.firstMessage)
			if got != tt.want {
				t.Errorf(
					"IsAutomatedSession(%q) = %v, want %v",
					tt.firstMessage, got, tt.want,
				)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/wesm/.superset/worktrees/agentsview/feat/roborev-detection && CGO_ENABLED=1 go test -tags fts5 ./internal/db/ -run TestIsAutomatedSession -v`
Expected: FAIL — `IsAutomatedSession` not defined.

- [ ] **Step 3: Implement detection function**

Create `internal/db/automated.go`:

```go
package db

import "strings"

// automatedPrefixes are first_message prefixes that identify
// automated (roborev) review and fix sessions. Matched
// case-sensitively against the start of first_message.
var automatedPrefixes = []string{
	"You are a code reviewer. Review the code changes shown below.",
	"# Fix Request\n",
}

// IsAutomatedSession returns true if the first message
// matches a known automated review/fix prompt pattern.
func IsAutomatedSession(firstMessage string) bool {
	for _, prefix := range automatedPrefixes {
		if strings.HasPrefix(firstMessage, prefix) {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `CGO_ENABLED=1 go test -tags fts5 ./internal/db/ -run TestIsAutomatedSession -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/db/automated.go internal/db/automated_test.go
git commit -m "feat: add IsAutomatedSession detection heuristic"
```

---

### Task 2: Add `is_automated` column to SQLite

**Files:**
- Modify: `internal/db/schema.sql:2-26` (add column to CREATE TABLE)
- Modify: `internal/db/db.go:264-372` (add migration + backfill)
- Modify: `internal/db/sessions.go:26-52` (column lists)
- Modify: `internal/db/sessions.go:68-80` (scanSessionRow)
- Modify: `internal/db/sessions.go:83-107` (Session struct)
- Modify: `internal/db/sessions.go:523-577` (UpsertSession)

- [ ] **Step 1: Add column to schema.sql**

Add `is_automated INTEGER NOT NULL DEFAULT 0` after `has_peak_context_tokens` (line 23) in `internal/db/schema.sql`.

- [ ] **Step 2: Add migration + backfill in db.go**

In `migrateColumns()`, add a new migration entry:

```go
{
	"sessions", "is_automated",
	"ALTER TABLE sessions ADD COLUMN is_automated INTEGER NOT NULL DEFAULT 0",
},
```

After column migrations, add a backfill step that updates existing
sessions:

```go
// Backfill is_automated for existing sessions.
if err := db.backfillIsAutomatedLocked(w); err != nil {
	return err
}
```

Add the backfill function:

```go
func (db *DB) backfillIsAutomatedLocked(w *sql.DB) error {
	var count int
	if err := w.QueryRow(
		`SELECT count(*) FROM sessions
		 WHERE is_automated = 0
		   AND first_message IS NOT NULL
		   AND (first_message LIKE 'You are a code reviewer. Review the code changes shown below.%'
		     OR first_message LIKE '# Fix Request' || X'0A' || '%')`,
	).Scan(&count); err != nil {
		return fmt.Errorf("probing automated backfill: %w", err)
	}
	if count == 0 {
		return nil
	}
	_, err := w.Exec(
		`UPDATE sessions SET is_automated = 1
		 WHERE first_message IS NOT NULL
		   AND (first_message LIKE 'You are a code reviewer. Review the code changes shown below.%'
		     OR first_message LIKE '# Fix Request' || X'0A' || '%')`,
	)
	if err != nil {
		return fmt.Errorf("backfilling is_automated: %w", err)
	}
	log.Printf("migration: backfilled is_automated for %d sessions", count)
	return nil
}
```

- [ ] **Step 3: Add `IsAutomated` field to Session struct**

In `internal/db/sessions.go`, add to Session struct (after
`HasPeakContextTokens`):

```go
IsAutomated bool `json:"is_automated"`
```

- [ ] **Step 4: Update column lists**

Update `sessionBaseCols` to include `is_automated` after
`has_peak_context_tokens`:

```go
const sessionBaseCols = `id, project, machine, agent,
	first_message, display_name, started_at, ended_at,
	message_count, user_message_count,
	parent_session_id, relationship_type,
	total_output_tokens, peak_context_tokens,
	has_total_output_tokens, has_peak_context_tokens,
	is_automated,
	deleted_at, created_at`
```

Similarly update `sessionPruneCols` and `sessionFullCols`.

- [ ] **Step 5: Update scanSessionRow**

Add `&s.IsAutomated` to the Scan call after `&s.HasPeakContextTokens`:

```go
func scanSessionRow(rs rowScanner) (Session, error) {
	var s Session
	err := rs.Scan(
		&s.ID, &s.Project, &s.Machine, &s.Agent,
		&s.FirstMessage, &s.DisplayName, &s.StartedAt, &s.EndedAt,
		&s.MessageCount, &s.UserMessageCount,
		&s.ParentSessionID, &s.RelationshipType,
		&s.TotalOutputTokens, &s.PeakContextTokens,
		&s.HasTotalOutputTokens, &s.HasPeakContextTokens,
		&s.IsAutomated,
		&s.DeletedAt, &s.CreatedAt,
	)
	return s, err
}
```

Also update the two manual Scan calls in `GetSessionFull` and
`ListSessionsModifiedBetween` — add `&s.IsAutomated` in the
same position (after `&s.HasPeakContextTokens`).

- [ ] **Step 6: Update UpsertSession to set is_automated**

Add `is_automated` to the INSERT column list and the ON CONFLICT
UPDATE set. Compute it from `FirstMessage`:

```go
isAutomated := false
if s.FirstMessage != nil {
	isAutomated = IsAutomatedSession(*s.FirstMessage)
}
```

Add `is_automated` as a column in the INSERT and the ON CONFLICT
DO UPDATE SET clause. Pass `isAutomated` as a bind parameter.

- [ ] **Step 7: Also update scanPruneCandidate in FindPruneCandidates**

The `FindPruneCandidates` function scans `sessionPruneCols` manually
(not via `scanSessionRow`). Add `&s.IsAutomated` to that Scan in
the correct position.

- [ ] **Step 8: Run all tests**

Run: `CGO_ENABLED=1 go test -tags fts5 ./internal/db/ -v`
Expected: PASS (existing tests exercise UpsertSession, scan, etc.)

- [ ] **Step 9: Commit**

```bash
git add internal/db/schema.sql internal/db/db.go internal/db/sessions.go
git commit -m "feat: add is_automated column to sessions table"
```

---

### Task 3: Add `ExcludeAutomated` filter to SQLite queries

**Files:**
- Modify: `internal/db/sessions.go:186-362` (SessionFilter, buildSessionFilter)
- Modify: `internal/db/analytics.go:44-156` (AnalyticsFilter, buildWhere)
- Modify: `internal/db/stats.go:46-66` (GetStats)
- Modify: `internal/db/sessions.go:898-994` (GetProjects, GetAgents, GetMachines)

- [ ] **Step 1: Write filter tests**

Add to `internal/db/filter_test.go`:

```go
func TestSessionFilterExcludeAutomated(t *testing.T) {
	d := testDB(t)

	insertSession(t, d, "normal", "proj", func(s *Session) {
		s.MessageCount = 3
		s.UserMessageCount = 1
	})
	insertSession(t, d, "review", "proj", func(s *Session) {
		fm := "You are a code reviewer. Review the code changes shown below.\n\n## Changes"
		s.FirstMessage = &fm
		s.MessageCount = 3
		s.UserMessageCount = 1
	})
	insertSession(t, d, "fix", "proj", func(s *Session) {
		fm := "# Fix Request\nAn analysis was performed"
		s.FirstMessage = &fm
		s.MessageCount = 3
		s.UserMessageCount = 1
	})
	insertSession(t, d, "multi", "proj", func(s *Session) {
		s.MessageCount = 10
		s.UserMessageCount = 5
	})

	tests := []struct {
		name            string
		excludeAutomated bool
		want            []string
	}{
		{
			"IncludeAll",
			false,
			[]string{"normal", "review", "fix", "multi"},
		},
		{
			"ExcludeAutomated",
			true,
			[]string{"normal", "multi"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := SessionFilter{
				ExcludeAutomated: tt.excludeAutomated,
			}
			requireSessions(t, d, f, tt.want)
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=1 go test -tags fts5 ./internal/db/ -run TestSessionFilterExcludeAutomated -v`
Expected: FAIL — `ExcludeAutomated` field not found.

- [ ] **Step 3: Add `ExcludeAutomated` to SessionFilter**

In `internal/db/sessions.go`, add to `SessionFilter`:

```go
ExcludeAutomated bool // exclude sessions where is_automated = 1
```

- [ ] **Step 4: Add filter predicate to buildSessionFilter**

In `buildSessionFilter`, after the `ExcludeOneShot` handling block
(around line 324), add:

```go
if f.ExcludeAutomated {
	filterPreds = append(filterPreds, "is_automated = 0")
}
```

No special `IncludeChildren` handling needed — automated sessions
are never parents.

- [ ] **Step 5: Run test to verify it passes**

Run: `CGO_ENABLED=1 go test -tags fts5 ./internal/db/ -run TestSessionFilterExcludeAutomated -v`
Expected: PASS

- [ ] **Step 6: Add `ExcludeAutomated` to AnalyticsFilter**

In `internal/db/analytics.go`, add to `AnalyticsFilter`:

```go
ExcludeAutomated bool // exclude automated sessions
```

In `buildWhere`, after the `ExcludeOneShot` block (line 146-147),
add:

```go
if f.ExcludeAutomated {
	preds = append(preds, "is_automated = 0")
}
```

- [ ] **Step 7: Update GetStats, GetProjects, GetAgents, GetMachines**

Change their signature from `excludeOneShot bool` to accept a
struct:

```go
type MetadataFilter struct {
	ExcludeOneShot   bool
	ExcludeAutomated bool
}
```

Or simpler: add a second `bool` parameter `excludeAutomated` to
each function and add the `AND is_automated = 0` predicate when
true. Follow the existing `excludeOneShot` pattern.

- [ ] **Step 8: Run all DB tests**

Run: `CGO_ENABLED=1 go test -tags fts5 ./internal/db/ -v`
Expected: PASS

- [ ] **Step 9: Commit**

```bash
git add internal/db/sessions.go internal/db/analytics.go \
       internal/db/stats.go internal/db/filter_test.go
git commit -m "feat: add ExcludeAutomated filter to session queries"
```

---

### Task 4: Update API handlers

**Files:**
- Modify: `internal/server/sessions.go:10-93` (handleListSessions)
- Modify: `internal/server/analytics.go:49-125` (parseAnalyticsFilter)
- Modify: `internal/server/events.go:319-378` (stats/projects/agents/machines handlers)

- [ ] **Step 1: Add `include_automated` param to handleListSessions**

In `internal/server/sessions.go`, after `includeOneShot` (line 58):

```go
includeAutomated := q.Get("include_automated") == "true"
```

Add to filter struct:

```go
ExcludeAutomated: !includeAutomated,
```

- [ ] **Step 2: Add `include_automated` param to parseAnalyticsFilter**

In `internal/server/analytics.go`, after `includeOneShot` (line 110):

```go
includeAutomated := q.Get("include_automated") == "true"
```

Add to return struct:

```go
ExcludeAutomated: !includeAutomated,
```

- [ ] **Step 3: Add `include_automated` to metadata handlers**

In `internal/server/events.go`, update `handleGetStats`,
`handleListProjects`, `handleListMachines`, `handleListAgents`
to parse `include_automated` and pass `excludeAutomated` alongside
`excludeOneShot`.

- [ ] **Step 4: Run server tests**

Run: `CGO_ENABLED=1 go test -tags fts5 ./internal/server/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/server/sessions.go internal/server/analytics.go \
       internal/server/events.go
git commit -m "feat: add include_automated API parameter"
```

---

### Task 5: Update PostgreSQL support

**Files:**
- Modify: `internal/postgres/schema.go:26-46` (coreDDL)
- Modify: `internal/postgres/sessions.go:29-35` (pgSessionCols)
- Modify: `internal/postgres/sessions.go:51-86` (scanPGSession)
- Modify: `internal/postgres/sessions.go:113-239` (buildPGSessionFilter)
- Modify: `internal/postgres/analytics.go:65-119` (buildAnalyticsWhere)
- Modify: `internal/postgres/push.go:589-660` (pushSession)

- [ ] **Step 1: Add column to PG DDL**

In `internal/postgres/schema.go`, add to sessions table DDL
(after `has_peak_context_tokens`):

```sql
is_automated BOOLEAN NOT NULL DEFAULT FALSE,
```

- [ ] **Step 2: Add schema migration**

In the migration section of `schema.go` (wherever columns are
added to existing tables), add an `ALTER TABLE sessions ADD COLUMN
is_automated BOOLEAN NOT NULL DEFAULT FALSE` migration.

- [ ] **Step 3: Update pgSessionCols and scanPGSession**

Add `is_automated` to `pgSessionCols` after
`has_peak_context_tokens`. Add `&s.IsAutomated` to the Scan call
in `scanPGSession`.

- [ ] **Step 4: Update buildPGSessionFilter**

After the `ExcludeOneShot` block, add:

```go
if f.ExcludeAutomated {
	filterPreds = append(filterPreds, "is_automated = FALSE")
}
```

- [ ] **Step 5: Update buildAnalyticsWhere**

After the `ExcludeOneShot` block, add:

```go
if f.ExcludeAutomated {
	preds = append(preds,
		"is_automated = FALSE")
}
```

- [ ] **Step 6: Update pushSession**

Add `is_automated` to the INSERT columns, values, ON CONFLICT
UPDATE SET, and WHERE IS DISTINCT FROM clauses. Compute from
`sess.FirstMessage`:

```go
isAutomated := false
if sess.FirstMessage != nil {
	isAutomated = db.IsAutomatedSession(*sess.FirstMessage)
}
```

- [ ] **Step 7: Run PG compilation check**

Run: `CGO_ENABLED=1 go build -tags fts5 ./internal/postgres/...`
Expected: compiles cleanly (full PG tests require a running
PostgreSQL instance).

- [ ] **Step 8: Commit**

```bash
git add internal/postgres/schema.go internal/postgres/sessions.go \
       internal/postgres/analytics.go internal/postgres/push.go
git commit -m "feat: add is_automated column to PostgreSQL schema"
```

---

### Task 6: Frontend - flip defaults and add automated filter

**Files:**
- Modify: `frontend/src/lib/api/types/core.ts` (add `is_automated`)
- Modify: `frontend/src/lib/api/client.ts` (add `include_automated` params)
- Modify: `frontend/src/lib/stores/sessions.svelte.ts` (filter changes)
- Modify: `frontend/src/lib/stores/analytics.svelte.ts` (filter changes)
- Modify: `frontend/src/lib/components/sidebar/SessionList.svelte` (toggle UI)
- Modify: `frontend/src/lib/components/analytics/ActiveFilters.svelte` (filter chips)
- Modify: `frontend/src/App.svelte` (URL sync)

- [ ] **Step 1: Add `is_automated` to TypeScript Session type**

In `frontend/src/lib/api/types/core.ts`, add:

```typescript
is_automated: boolean;
```

- [ ] **Step 2: Add `include_automated` to API client**

Update `getProjects`, `getMachines`, `getAgents`, `getStats` to
accept `include_automated?: boolean` alongside `include_one_shot`.

- [ ] **Step 3: Update sessions store**

In `frontend/src/lib/stores/sessions.svelte.ts`:

1. Add `includeAutomated: boolean` to `Filters` interface
   (default `false`)
2. Change `includeOneShot` default from `false` to `true`
3. In `apiParams`, add `include_automated: f.includeAutomated ||
   undefined`
4. Update `hasActiveFilters`: replace `f.includeOneShot` with
   `!f.includeOneShot` (since default is now true, NOT including
   is the active filter). Add `f.includeAutomated`.
5. Add `setIncludeAutomatedFilter(include: boolean)` method
6. Update `clearSessionFilters` to reset both filters
7. Update `invalidateFilterCaches` to pass both params
8. Update `initFromParams` to parse `include_automated`
9. Update `loadProjects`, `loadAgents`, `loadMachines` to pass
   both filter params

- [ ] **Step 4: Update analytics store**

In `frontend/src/lib/stores/analytics.svelte.ts`:

1. Add `includeAutomated: boolean = $state(false)`
2. Update `params` getter to include `include_automated`
3. Update `hasActiveFilters` with the same logic changes

- [ ] **Step 5: Update SessionList filter toggles**

In `frontend/src/lib/components/sidebar/SessionList.svelte`:

1. Change "Include single-turn" toggle to "Hide single-turn"
   with inverted logic (ON = `includeOneShot: false`)
2. Add new "Include automated" toggle below it (ON =
   `includeAutomated: true`)

- [ ] **Step 6: Update ActiveFilters chips**

In `frontend/src/lib/components/analytics/ActiveFilters.svelte`:

1. Change "Single-turn included" chip to "Single-turn hidden"
   (shown when `!includeOneShot`, since default is now true)
2. Add "Automated included" chip (shown when
   `includeAutomated`)

- [ ] **Step 7: Update App.svelte URL sync**

In `frontend/src/App.svelte`, add `include_automated` to the URL
param sync (same pattern as `include_one_shot`). Since
`includeOneShot` is now true by default, always include it in URL
params when false: `if (!f.includeOneShot)
p.include_one_shot = "false"`.

Wait — for backwards compatibility, keep the existing URL param
behavior: `include_one_shot=true` means include. The default now
sends `include_one_shot=true` because the frontend default flipped.
When user toggles "Hide single-turn", it stops sending the param
(server default excludes).

- [ ] **Step 8: Build frontend**

Run: `cd frontend && npm run build`
Expected: no TypeScript errors

- [ ] **Step 9: Commit**

```bash
git add frontend/src/lib/api/types/core.ts \
       frontend/src/lib/api/client.ts \
       frontend/src/lib/stores/sessions.svelte.ts \
       frontend/src/lib/stores/analytics.svelte.ts \
       frontend/src/lib/components/sidebar/SessionList.svelte \
       frontend/src/lib/components/analytics/ActiveFilters.svelte \
       frontend/src/App.svelte
git commit -m "feat: flip single-turn default, add automated filter UI"
```

---

### Task 7: Update server tests

**Files:**
- Modify: `internal/server/server_test.go`

- [ ] **Step 1: Update existing tests that use `include_one_shot`**

Several tests use `include_one_shot=true` to include single-turn
sessions. These remain correct — the server-side default hasn't
changed (server still excludes one-shot unless told otherwise).
Only the frontend default changed.

Add a test for the new `include_automated` parameter:

```go
// Verify include_automated=true shows automated sessions.
// Create a session with a roborev first_message, verify it's
// excluded by default and included with include_automated=true.
```

- [ ] **Step 2: Run all server tests**

Run: `CGO_ENABLED=1 go test -tags fts5 ./internal/server/ -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/server/server_test.go
git commit -m "test: add server tests for include_automated param"
```

---

### Task 8: Final verification

- [ ] **Step 1: Run all Go tests**

Run: `CGO_ENABLED=1 go test -tags fts5 ./... -v`
Expected: PASS

- [ ] **Step 2: Run lints**

Run: `go vet ./... && go fmt ./...`
Expected: clean

- [ ] **Step 3: Build full binary**

Run: `make build`
Expected: builds successfully with embedded frontend

- [ ] **Step 4: Commit any final fixes**

---

## Key Design Decisions

1. **`is_automated` as stored column, not computed**: Allows
   indexed filtering without per-row computation. Computed from
   `first_message` during upsert; backfilled via SQL LIKE for
   existing rows.

2. **Server-side defaults unchanged**: The `include_one_shot`
   API parameter behavior is preserved. Only the frontend
   default flips from "exclude" to "include". This avoids
   breaking API consumers.

3. **No `dataVersion` bump needed**: The column is added via
   migration + backfill. No need for a full resync since the
   detection is based on existing `first_message` data.

4. **Detection heuristics are prefix-based**: Simple
   `strings.HasPrefix` matching against known prompt patterns.
   Easy to extend by adding entries to `automatedPrefixes`.
