# Periscope V1 Implementation Plan

## Document Status

- Status: Draft for implementation
- Date: 2026-04-18
- Depends on:
  [`periscope-spec.md`](./periscope-spec.md),
  [`v1-ui-spec.md`](./v1-ui-spec.md)

## Objective

Deliver Periscope V1 as a descriptive context visualizer for historical and
active coding-agent sessions. V1 satisfies FR1, FR2, FR3, FR8, and FR9 from
the product spec.

V1 must:

- show current session context state,
- show context growth over time,
- attribute context to normalized source categories,
- support historical and active sessions with SSE,
- support every agent registered in `internal/parser/types.go`.

V1 must not include guidance, health scoring, recommendation cards, or
branch-point analysis.

## Scope Summary

Periscope V1 ships a standalone route at `/context/:sessionId` with optional
embedding inside the existing session detail experience later.

The product should let a user answer two questions quickly:

1. How full is this session?
1. What is consuming the context?

It should then let the user inspect turn-by-turn growth, spikes, and
compaction boundaries.

## UI Plan

### Chosen UI

Use the spec-mandated page structure:

1. `ContextSummaryCard`
1. `ContextCompositionChart`
1. `ContextTimeline`

For the timeline, adopt Option C from
[`v1-ui-spec.md`](/Users/ann/dev/periscope/docs/v1-ui-spec.md):
explicit turn rows, per-turn stacked category bars, inline annotations, and a
strong compaction divider.

### Why This UI

- It matches the current spec directly.
- It works well in a standalone route and in future embedded contexts.
- It preserves turn-level detail without requiring a chart library.
- It stays usable in JetBrains webviews and narrow layouts.
- It leaves a clean future path to add a linked growth chart later.

### UI Requirements by Section

#### Section A: Summary Header

Show:

- estimated context tokens in use,
- max context window for the session,
- whether the max window is measured, inferred, or unknown,
- percent consumed,
- remaining budget,
- token estimate provenance,
- agent, model, and plan when known,
- live status and last update time.

The summary should visually communicate occupancy without implying V2 diagnosis.

#### Section B: Composition Breakdown

Use a stacked horizontal bar and ranked legend. Each category row should show:

- category name,
- token estimate,
- percentage of current context,
- estimate provenance,
- sample turns or jump links.

#### Section C: Timeline

Each visible turn row should show:

- turn index,
- delta tokens,
- cumulative tokens,
- dominant category,
- stacked category composition for that turn,
- event markers for spike, compaction, rewind, fork, and subagent events when
  known,
- optional annotation text for notable spikes or tool/file-heavy turns.

Compaction is rendered as a hard reset boundary. Pre-compaction history is not
shown in V1.

## Data and Computation Plan

### Context Capacity Resolution

Percent-full calculations depend on a session-specific max context window.
Implement a resolver with this precedence:

1. recorded provider or agent session limit,
1. known agent + model mapping,
1. plan-aware model limit when plan signal exists,
1. conservative agent/model-family default,
1. unknown.

The resolver must return both the numeric value and the provenance label.

### Token Estimation

For each relevant message or artifact:

1. use recorded token usage when present,
1. otherwise estimate from byte length using parser-specific ratios,
1. attach measured vs estimated provenance,
1. classify into a normalized V1 source category.

### V1 Source Categories

Start with the spec taxonomy:

- system prompt and tool definitions,
- user messages,
- assistant messages,
- thinking or reasoning blocks when available,
- file reads or fetched code content,
- tool calls,
- tool outputs,
- search results and grep-like output,
- summaries and compacted handoffs,
- subagent outputs,
- deferred or hidden payloads when exposed,
- free space.

### Post-Compaction Handling

For V1:

- detect the latest compaction event if present,
- start the visible timeline at that event,
- represent the compacted summary as a single visible segment or seed row,
- compute cumulative totals from that point forward only.

## Backend Work Plan

### 1. Parser Metadata Audit

Audit every agent in `internal/parser/types.go` for:

- token usage availability,
- model detection,
- context-limit metadata availability,
- plan or entitlement hints,
- compaction detection,
- rewind or fork detection,
- subagent detection,
- bytes-per-token fallback ratio.

Fill in missing parser metadata where practical.

### 2. Normalization Layer

Build a normalization layer that can turn agent-specific session records into a
common V1 context model:

- `ContextCapacity`
- `ContextSummary`
- `ContextCategoryBreakdownItem`
- `ContextTimelineRow`

### 3. Context Summary Endpoint

Implement `GET /api/v1/sessions/{id}/context`.

Recommended response shape:

- `summary`
- `capacity`
- `composition`
- `supports`
- `warnings`

### 4. Context Timeline Endpoint

Implement `GET /api/v1/sessions/{id}/context/timeline`.

Recommended response shape:

- `timeline`
- `markers`
- `supports`
- `warnings`

### 5. Live Updates

Add SSE updates for active sessions.

Initial event types:

- `context_summary_updated`
- `context_timeline_updated`
- `session_status_changed`

The first version can send coarse refresh events instead of fine-grained diffs.

## Frontend Work Plan

### Route

Add `/context/:sessionId` as a dedicated route with a link back to `/`.

### Components

Build:

- `ContextPage`
- `ContextSummaryCard`
- `ContextCompositionChart`
- `ContextTimeline`

### Interaction Requirements

- Clicking a turn jumps to the transcript.
- Clicking a category highlights or filters representative turns.
- Live sessions visibly update without stealing scroll position.
- Capacity uncertainty is visible inline near the max-window value.

### Layout Priorities

- Optimize for vertical scanning.
- Keep typography and spacing dense enough for power users.
- Avoid synchronized split panes in the initial release.
- Make compaction boundaries visually unmistakable.

## Delivery Sequence

### Milestone 1: Data Plumbing

- parser metadata audit,
- token fallback support,
- capacity resolver,
- source-category normalization.

Exit criteria:

- one V1 summary can be computed for every supported agent family.

### Milestone 2: API Surface

- summary endpoint,
- timeline endpoint,
- provenance fields,
- post-compaction handling,
- warning and support metadata.

Exit criteria:

- the UI can be built entirely from stable V1 API responses.

### Milestone 3: Historical Session UI

- route shell,
- summary card,
- composition chart,
- Option C timeline,
- transcript jump behavior.

Exit criteria:

- historical sessions are inspectable end to end.

### Milestone 4: Live Session UX

- SSE wiring,
- live refresh handling,
- last-updated and live-state indicators,
- scroll-preserving updates.

Exit criteria:

- active sessions stay readable while updating.

### Milestone 5: Cross-Agent Hardening

- verify support across every registered agent,
- improve low-fidelity handling,
- tune spike attribution and category mapping,
- tighten empty and partial states.

Exit criteria:

- V1 is ready to ship without obvious unsupported-agent gaps.

## Testing Plan

### Backend Tests

- token-estimation fallback tests,
- capacity-resolution precedence tests,
- category-mapping tests,
- compaction truncation tests,
- spike-detection tests,
- parser-fixture coverage per agent family.

### Frontend Tests

- component rendering tests,
- click-to-transcript tests,
- composition interaction tests,
- active-session update tests,
- narrow-layout verification.

### Manual Verification

Verify with:

- a short measured-token session,
- a long estimated-token session,
- a session with compaction,
- an active session receiving SSE updates,
- at least one session per supported agent family,
- a session with plan-aware inferred capacity,
- a session with unknown capacity.

## Risks

### Plan-Aware Capacity Can Be Incomplete

Mitigation:

- show provenance explicitly,
- prefer conservative defaults,
- avoid presenting inferred capacity as certain.

### Agent Fidelity Varies

Mitigation:

- centralize taxonomy,
- keep agent-specific mapping hooks,
- surface support gaps as warnings.

### Long Timelines Can Get Heavy

Mitigation:

- keep the initial row rendering simple,
- use efficient scrolling if needed,
- defer linked charts until after the baseline ship.

### Live Updates Can Be Distracting

Mitigation:

- preserve scroll position,
- avoid aggressive reflow,
- keep update indicators subtle but visible.

## Out of Scope for V1

- health states,
- branch-point detection,
- recommendation engine,
- guidance panel,
- evidence inspector,
- guidance-agent execution,
- any analysis that feeds output back into the live parent session context.

## Recommended First Build Order

1. parser audit across supported agents,
1. context capacity resolver,
1. token attribution and category mapping,
1. summary and timeline APIs,
1. standalone `/context/:sessionId` route,
1. summary and composition UI,
1. Option C timeline,
1. transcript jump behavior,
1. SSE live updates,
1. cross-agent polish.
