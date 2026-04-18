# Context Session Visualizer MVP Implementation Plan

> **For agentic workers:** Execute this plan in order. Keep the scope strictly
> within MVP. Do not add advisor-agent execution, branch-point recommendations,
> or action-copy UI in this phase. The target is a read-only context x-ray for a
> single session.

**Goal:** Implement the MVP of the Context Session Visualizer in this codebase.

**Reference docs:**

- [`docs/context-session-visualizer-spec.md`](./context-session-visualizer-spec.md)
- [`docs/context-session-visualizer-roadmap.md`](./context-session-visualizer-roadmap.md)

**Tech stack:** Go, SQLite, Svelte 5, TypeScript

---

## MVP Definition

The MVP is complete when a user can open a session and see:

- an estimated context usage summary,
- a context source/category breakdown,
- a context growth timeline,
- compaction markers on that timeline,
- a simple session health label with short reasons.

The MVP must be:

- read-only,
- session-scoped,
- estimate-driven but explicitly labeled,
- useful even without advisor-agent support.

The MVP must not include:

- branch-point detection,
- rewind/fork/compact recommendations,
- advisor-agent analysis,
- automated session actions.

---

## Architecture

## Backend

Add a new context-analysis read path that derives context metrics from existing
session data:

- `sessions`
- `messages`
- `tool_calls`
- `tool_result_events`

The implementation should prefer exact token data when available and fall back
to approximations when it is not.

The implementation should also reuse already-persisted session-level signals
instead of recomputing them from scratch. In particular, MVP health should start
from existing `sessions` columns such as:

- `health_score`
- `health_grade`
- `compaction_count`
- `peak_context_tokens`
- `context_pressure_max`
- `tool_failure_signal_count`
- `tool_retry_count`
- `edit_churn_count`

## Frontend

Add a session-level Context view in the existing session detail flow. The UI
should load summary and timeline data for the active session and render them via
new Svelte components.

## Key Design Constraint

Do not attempt perfect provider-side prompt reconstruction. The data model
should always distinguish exact versus estimated values.

---

## File Map

| File | Action | Responsibility |
| --- | --- | --- |
| `internal/db/context_analysis.go` | Add | Core context analysis queries and aggregation |
| `internal/db/context_analysis_test.go` | Add | Unit tests for composition, timeline, health |
| `internal/server/context.go` | Add | HTTP handlers for context endpoints |
| `internal/server/server.go` | Modify | Route registration |
| `frontend/src/lib/api/types/context.ts` | Add | Context response types |
| `frontend/src/lib/api/types/index.ts` | Modify | Export context types |
| `frontend/src/lib/api/client.ts` | Modify | Context fetch methods |
| `frontend/src/lib/stores/context.svelte.ts` | Add | Client-side state and loading |
| `frontend/src/lib/components/context/ContextSummaryCard.svelte` | Add | Summary panel |
| `frontend/src/lib/components/context/ContextCompositionChart.svelte` | Add | Category breakdown |
| `frontend/src/lib/components/context/ContextTimeline.svelte` | Add | Timeline view |
| `frontend/src/lib/components/context/ContextPage.svelte` | Add | Page composition |
| `frontend/src/App.svelte` | Modify | Session-level mode integration |
| `frontend/src/lib/components/content/MessageList.svelte` or adjacent session UI | Modify | Add Context view toggle if needed |
| `frontend/src/lib/stores/context.test.ts` | Add | Store tests |

---

## Data Model

## API Types To Introduce

### `ContextCompositionEntry`

- `category: string`
- `label: string`
- `tokens: number`
- `percent: number`
- `is_estimated: boolean`

### `ContextTimelineEntry`

- `index: number`
- `ordinal_start: number`
- `ordinal_end: number`
- `timestamp: string | null`
- `delta_tokens: number`
- `cumulative_tokens: number`
- `dominant_category: string`
- `spike_level: "none" | "medium" | "high"`
- `is_compact_boundary: boolean`
- `is_estimated: boolean`

### `ContextHealthSummary`

- `state: "healthy" | "watch" | "degraded" | "critical"`
- `reasons: string[]`

### `SessionContextResponse`

- `session_id: string`
- `estimated_context_tokens: number`
- `estimated_context_percent: number | null`
- `estimated_remaining_tokens: number | null`
- `max_context_tokens: number | null`
- `estimate_confidence: "high" | "medium" | "low"`
- `health: ContextHealthSummary`
- `composition: ContextCompositionEntry[]`
- `timeline: ContextTimelineEntry[]`

---

## Context Estimation Rules

## Rule 1: Prefer explicit token data

When a message has token usage or explicit context/output token fields, use them
first.

Possible sources already present in the DB:

- `messages.token_usage`
- `messages.context_tokens`
- `messages.output_tokens`
- `sessions.peak_context_tokens`

## Rule 2: Fall back to content length approximation

When token data is absent:

- estimate from `content_length`,
- use a simple approximation constant that is documented in code comments,
- keep the estimate method consistent across summary and timeline views.

Suggested initial approximation:

- `estimated_tokens = ceil(content_length / 4)`

This can be tuned later; consistency matters more than perfect accuracy in MVP.

## Rule 3: Treat context as accumulated session mass

MVP should model context as cumulative session content added over time, not as an
attempt to subtract pruned or packed-away content after compaction. Compaction
awareness can come later.

## Rule 4: Split categories using existing metadata

Suggested initial classification logic:

- `thinking` if `has_thinking` is true or content parser/source metadata implies
  thinking content.
- `tool_call` for structured tool-call metadata in `tool_calls`.
- `tool_result` from `tool_result_events` when available, else from
  `tool_calls.result_content_length`.
- `user` and `assistant` from message roles.
- `file_read` and `search` inferred from tool categories/names where possible.
- `subagent` for child-session or subagent-related tool metadata.
- `other` as fallback.

## Rule 5: Use peak context as the primary source for percent-full when present

Not all agents expose a reliable maximum. The response must support
`max_context_tokens = null` and therefore `estimated_context_percent = null`.

For MVP:

- use `sessions.peak_context_tokens` as the first-class source when it reflects
  the session's observed context high-water mark,
- use any explicit max-context metadata from parsed session data when available,
- otherwise allow agent-specific defaults only if confidence is reasonable,
- otherwise leave percent/remaining as null.

Claude-first support is acceptable.

## Rule 6: Reuse persisted session health signals

MVP health should not duplicate the existing sync-time signal pipeline. Instead:

- treat persisted `health_score` and `health_grade` as the baseline health
  assessment,
- use `compaction_count`, `context_pressure_max`, `tool_retry_count`, and
  `edit_churn_count` as additional explanatory inputs,
- add occupancy-based logic only where the persisted data does not already cover
  the concern.

This keeps the new Context view aligned with the rest of the product.

---

## Implementation Tasks

### Task 1: Backend — Add context analysis package and types

**Files:**

- Add: `internal/db/context_analysis.go`
- Add: `internal/db/context_analysis_test.go`

- [ ] **Step 1: Define core Go structs**

Add response-oriented structs in `internal/db/context_analysis.go`:

- `ContextCompositionEntry`
- `ContextTimelineEntry`
- `ContextHealthSummary`
- `SessionContextSummary`
- `SessionContextTimeline`

Keep JSON tags aligned with the API payload names listed above.

- [ ] **Step 2: Define category constants**

Create string constants for:

- `system`
- `user`
- `assistant`
- `thinking`
- `tool_call`
- `tool_result`
- `file_read`
- `search`
- `summary`
- `subagent`
- `other`

- [ ] **Step 3: Add helper functions**

Add private helpers for:

- token estimation from a message row,
- token estimation from tool-result content length,
- category labeling,
- confidence labeling,
- spike classification,
- health-state classification that layers on top of persisted session signals.

**Acceptance criteria:**

- File compiles.
- Types are usable by both DB and server layers.

### Task 2: Backend — Implement session context composition query

**Files:**

- Modify: `internal/db/context_analysis.go`

- [ ] **Step 1: Add a message-loading query for one session**

Implement a read function that fetches the session’s messages with enough fields
to classify and estimate tokens:

- `id`
- `ordinal`
- `role`
- `timestamp`
- `has_thinking`
- `has_tool_use`
- `content_length`
- `token_usage`
- `context_tokens`
- `output_tokens`
- `source_type`
- `source_subtype`
- `is_compact_boundary`

- [ ] **Step 2: Add tool-call / tool-result loading**

Load:

- `tool_calls` rows for the session,
- `tool_result_events` rows for the session.

If needed, aggregate tool-result rows by message ordinal or message linkage for
timeline purposes.

- [ ] **Step 3: Aggregate composition**

Compute category totals across the session:

- estimate tokens for each message,
- estimate tokens for tool calls/results,
- classify each unit,
- sum per category,
- compute percentages against session total.

- [ ] **Step 4: Return summary**

Implement a DB method, e.g.:

```go
func (db *DB) GetSessionContextSummary(
	ctx context.Context, sessionID string,
) (*SessionContextSummary, error)
```

**Acceptance criteria:**

- Summary returns composition entries sorted by descending tokens.
- Total estimated context tokens is > 0 for non-empty sessions.
- Empty sessions produce valid empty payloads without crashing.

### Task 3: Backend — Implement context timeline

**Files:**

- Modify: `internal/db/context_analysis.go`

- [ ] **Step 1: Choose timeline granularity**

Use message ordinals as the base unit for MVP.

If the session is very large, optionally coalesce adjacent ordinals into windows
of 5 or 10 on the server side, but do not add adaptive bucketing unless needed
for performance.

- [ ] **Step 2: Compute per-row delta**

For each ordinal or window:

- sum newly added estimated tokens,
- determine dominant category,
- accumulate into `cumulative_tokens`,
- assign spike level based on threshold,
- carry through `is_compact_boundary` when any included message is a compact
  boundary.

- [ ] **Step 3: Add DB method**

Implement:

```go
func (db *DB) GetSessionContextTimeline(
	ctx context.Context, sessionID string,
) (*SessionContextTimeline, error)
```

**Spike suggestion:**

- `high` if delta >= 2.5x median non-zero delta
- `medium` if delta >= 1.5x median non-zero delta
- otherwise `none`

**Acceptance criteria:**

- Cumulative values never decrease.
- Timeline rows are in chronological/ordinal order.
- Dominant category is always set.
- Compaction boundaries are visible as timeline markers.

### Task 4: Backend — Implement simple health classification

**Files:**

- Modify: `internal/db/context_analysis.go`

- [ ] **Step 1: Define initial health heuristics**

Use:

- persisted session health and signal fields first,
- occupancy percent where available,
- recent spike count,
- tool-result dominance ratio,
- total context mass.

**Suggested initial rules:**

- if `health_grade` is already present, use it as the primary bucket seed,
- raise severity when occupancy or spike behavior is worse than the persisted
  score suggests,
- do not downgrade severe persisted signal states merely because current
  occupancy looks modest,
- `critical`
  - percent >= 85, or
  - percent unknown but very large total plus repeated high spikes
- `degraded`
  - percent >= 65, or
  - high tool-result dominance plus repeated spikes
- `watch`
  - percent >= 40, or
  - one recent high spike
- `healthy`
  - otherwise

- [ ] **Step 2: Emit short reasons**

Examples:

- `"High estimated context occupancy"`
- `"Recent growth spikes increased context rapidly"`
- `"Tool results dominate session growth"`
- `"Existing session health signals indicate degraded execution quality"`
- `"Compaction boundaries occurred during this session"`
- `"Context growth is currently modest"`

**Acceptance criteria:**

- Every summary includes a health state and at least one reason.
- Rules are deterministic and testable.

### Task 5: Backend — Add tests

**Files:**

- Add: `internal/db/context_analysis_test.go`

- [ ] **Step 1: Create fixture sessions**

Add tests covering:

- a simple message-only session,
- a session with thinking content,
- a tool-heavy session,
- a session with missing token usage that needs approximation,
- a session with large spikes.

- [ ] **Step 2: Validate composition**

Assert:

- expected categories appear,
- total tokens are plausible,
- percentages add up approximately.

- [ ] **Step 3: Validate timeline**

Assert:

- correct row ordering,
- non-decreasing cumulative values,
- spike levels appear on high-growth rows.

- [ ] **Step 4: Validate health**

Assert:

- different inputs map to expected health states.

**Suggested command:**

```bash
CGO_ENABLED=1 go test -tags fts5 ./internal/db/ -run 'TestSessionContext' -v
```

### Task 6: Server — Add context endpoint

**Files:**

- Add: `internal/server/context.go`
- Modify: `internal/server/server.go`

- [ ] **Step 1: Add a single context handler**

Implement:

- `GET /api/v1/sessions/{id}/context`

Behavior:

- fetch summary and timeline together,
- return `404` if session missing,
- return JSON on success.

- [ ] **Step 2: Register route**

Add route registration in `internal/server/server.go`.

**Acceptance criteria:**

- Endpoints behave like existing API handlers.
- They use the same timeout/error conventions as the rest of the server.
- Frontend only needs one request to render the MVP view.

### Task 7: Server — Add handler tests

**Files:**

- Add: `internal/server/context_test.go`

- [ ] **Step 1: Test happy path**

Verify the endpoint returns `200` for a seeded session.

- [ ] **Step 2: Test not found**

Verify the endpoint returns `404` for unknown sessions.

- [ ] **Step 3: Test payload basics**

Verify summary contains:

- `session_id`
- `composition`
- `health`
- `timeline`

**Suggested command:**

```bash
CGO_ENABLED=1 go test -tags fts5 ./internal/server/ -run 'TestContext' -v
```

### Task 8: Frontend — Add context API types

**Files:**

- Add: `frontend/src/lib/api/types/context.ts`
- Modify: `frontend/src/lib/api/types/index.ts`

- [ ] **Step 1: Define TS interfaces**

Mirror the backend payloads:

- `ContextCompositionEntry`
- `ContextTimelineEntry`
- `ContextHealthSummary`
- `SessionContextResponse`

- [ ] **Step 2: Export from type index**

Add exports from `frontend/src/lib/api/types/index.ts`.

**Acceptance criteria:**

- TypeScript builds without circular import issues.

### Task 9: Frontend — Add API client methods

**Files:**

- Modify: `frontend/src/lib/api/client.ts`

- [ ] **Step 1: Add summary method**

Implement:

```ts
export function getSessionContext(sessionId: string): Promise<SessionContextResponse>
```

**Acceptance criteria:**

- Method uses existing `fetchJSON` patterns.
- Errors propagate using existing API error handling.

### Task 10: Frontend — Add context store

**Files:**

- Add: `frontend/src/lib/stores/context.svelte.ts`
- Add: `frontend/src/lib/stores/context.test.ts`

- [ ] **Step 1: Create store state**

Store should hold:

- `summary`
- `timeline`
- `loading`
- `error`
- `sessionId`

- [ ] **Step 2: Add load methods**

Implement:

- `loadSession(id: string)`
- `reload()`
- `clear()`

- [ ] **Step 3: Add tests**

Cover:

- initial empty state,
- successful load,
- session switch reset,
- error handling.

### Task 11: Frontend — Build summary card component

**Files:**

- Add: `frontend/src/lib/components/context/ContextSummaryCard.svelte`

- [ ] **Step 1: Render the key metrics**

Show:

- estimated context tokens,
- percent full or "unknown",
- remaining tokens or "unknown",
- health state badge,
- reasons.

- [ ] **Step 2: Keep estimates explicit**

Display copy such as:

- `"Estimated context usage"`
- `"Percent full unavailable for this agent"` when needed.

**Acceptance criteria:**

- Summary is readable and compact.
- Unknown values are handled gracefully.

### Task 12: Frontend — Build composition chart

**Files:**

- Add: `frontend/src/lib/components/context/ContextCompositionChart.svelte`

- [ ] **Step 1: Render a stacked bar**

For each category:

- width by percent,
- hover label or visible legend,
- token and percent display.

- [ ] **Step 2: Add a legend/table**

Make the raw numbers visible below the chart.

**Acceptance criteria:**

- User can identify the largest categories immediately.

### Task 13: Frontend — Build timeline component

**Files:**

- Add: `frontend/src/lib/components/context/ContextTimeline.svelte`

- [ ] **Step 1: Render chronological rows or mini-bars**

Show:

- index range,
- delta tokens,
- cumulative tokens,
- dominant category,
- spike highlighting,
- compaction markers.

- [ ] **Step 2: Keep the first version simple**

Avoid a complex chart library if a CSS/SVG implementation is enough.

**Acceptance criteria:**

- Spikes are visually obvious.
- Long sessions remain readable.

### Task 14: Frontend — Build Context page wrapper

**Files:**

- Add: `frontend/src/lib/components/context/ContextPage.svelte`

- [ ] **Step 1: Compose child components**

Render:

- summary card,
- composition chart,
- timeline,
- loading/error states.

- [ ] **Step 2: Wire to store**

Read data from `context` store.

**Acceptance criteria:**

- Page works with a single prop or session context derived from global stores.

### Task 15: Frontend — Integrate Context view into session UX

**Files:**

- Modify: `frontend/src/App.svelte`
- Modify: one or more existing session layout/content components as needed

- [ ] **Step 1: Choose a session-level toggle**

Add a simple mode switch between:

- Transcript
- Context

Use the smallest integration that fits the existing UX patterns.

- [ ] **Step 2: Load context when the active session is selected**

When entering context mode:

- call `context.loadSession(activeSessionId)`

When leaving or clearing session:

- call `context.clear()`

- [ ] **Step 3: Preserve current transcript behavior**

Do not regress:

- message loading,
- activity minimap,
- session selection,
- routing.

**Acceptance criteria:**

- User can toggle into Context view for the active session.
- Session switching updates the context view correctly.

### Task 16: Frontend — Add component tests

**Files:**

- Add tests adjacent to new components or store tests only, depending on the
  project’s current testing conventions.

- [ ] **Step 1: Summary card rendering test**

- [ ] **Step 2: Composition chart rendering test**

- [ ] **Step 3: Timeline rendering test**

Include a compaction-marker rendering case.

- [ ] **Step 4: Context page loading/error state test**

**Suggested command:**

```bash
cd frontend && npm test -- --run context
```

Adjust to the repo’s actual test command format if needed.

### Task 17: Validation and finish

- [ ] **Step 1: Run Go tests**

```bash
CGO_ENABLED=1 go test -tags fts5 ./internal/db/ ./internal/server/
```

- [ ] **Step 2: Run frontend tests**

```bash
cd frontend && npm test
```

- [ ] **Step 3: Run frontend build**

```bash
cd frontend && npm run build
```

- [ ] **Step 4: Manual verification**

Run the app and verify:

1. open a session,
2. switch to Context view,
3. summary is visible,
4. composition chart renders,
5. timeline renders,
6. a different session updates the view.

---

## Suggested Commit Boundaries

To keep review clean, split work into focused commits:

1. `feat: add backend context analysis models and queries`
2. `test: add coverage for session context analysis`
3. `feat: add session context API endpoints`
4. `feat: add frontend context api types and store`
5. `feat: add session context view components`
6. `feat: integrate context view into session detail ui`

Do not squash unrelated work into these commits.

---

## Risks and Scope Guards

## Risk 1: Overcomplicating token estimation

Guardrail:

- use a simple consistent estimator for missing token data,
- document limitations,
- do not block MVP on perfect accuracy.

## Risk 2: UI sprawl

Guardrail:

- keep MVP to summary, composition, and timeline only.

## Risk 3: Hidden coupling with transcript layout

Guardrail:

- integrate using a simple session-level mode toggle,
- do not refactor the whole layout for MVP.

## Risk 4: Pulling recommendation logic into MVP

Guardrail:

- health state is allowed,
- action recommendation is not.

---

## Explicit Out-of-Scope Work For MVP

Do not implement any of the following in this plan:

- branch-point detection,
- rewind suggestions,
- compact focus generation,
- fork handoff generation,
- subagent recommendations,
- advisor-agent orchestration,
- proactive live warnings,
- cross-session analytics for context health.

These belong to later phases.

---

## MVP Acceptance Checklist

- [ ] Session context endpoint returns summary and timeline together
- [ ] Summary returns category breakdown and health state
- [ ] Timeline returns cumulative growth data and compaction markers
- [ ] Frontend can fetch and render the single payload
- [ ] User can access Context view from a session
- [ ] Estimates are explicitly labeled
- [ ] Tests cover core backend and frontend behavior

## Final Guidance

If the implementation starts slipping, preserve value by keeping:

1. summary card,
2. composition chart,
3. timeline,
4. health label.

Cut polish before cutting clarity. The MVP only needs to answer:

- how much context is in this session,
- what is filling it,
- and how quickly it has been growing.
