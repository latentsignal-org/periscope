# Context Session Visualizer Roadmap

> **For agentic workers:** Implement in sequence unless explicitly parallelized.
> Keep the first shipping slice narrow. Prefer read-only observability before
> recommendation sophistication, and recommendation sophistication before
> advisor-agent integration.

**Goal:** Break the Context Session Visualizer into MVP, v1, and v2 delivery
phases with concrete implementation tickets for this codebase.

**Reference spec:** [`docs/context-session-visualizer-spec.md`](./context-session-visualizer-spec.md)

**Tech stack:** Go, SQLite, Svelte 5, TypeScript

---

## Delivery Strategy

The feature should ship in three layers:

1. **MVP:** useful read-only context x-ray for one session.
2. **v1:** actionable recommendations and tighter integration with the existing
   transcript and session browser.
3. **v2:** advisor-agent analysis, more advanced heuristics, and broader
   agent-specific fidelity.

The main risk is overbuilding the intelligence layer before the data model and
UI are solid. The implementation should therefore move in this order:

1. Build normalized context-analysis data structures.
2. Expose them through backend APIs.
3. Render a clear context tab in the frontend.
4. Add evidence-backed recommendations.
5. Add advisor-agent integration only after the base UX is already useful
   without it.

---

## MVP

## MVP Goal

Ship a useful session-level context x-ray that answers:

- how full is this session,
- what categories are consuming context,
- how did context grow over time,
- where are the obvious spikes and degraded zones.

MVP is explicitly **read-only**. It does not need advisor-agent analysis, and it
does not need perfect branching recommendations. It should already be valuable
for users trying to understand session growth.

## MVP Scope

### Included

- Session-level context summary.
- Estimated context usage and percent full.
- Category breakdown of context sources.
- Timeline of context growth per turn or per message window.
- Highlighting of large growth spikes.
- Simple heuristic health state: healthy, watch, degraded, critical.
- Dedicated UI entry point in the session detail experience.

### Excluded

- Advisor-agent execution.
- Copyable action text.
- Exact rewind/fork/compact recommendations.
- Multi-session comparisons.
- Cross-agent parity for every source agent.

## MVP File Map

| File / Area | Action | Responsibility |
| --- | --- | --- |
| `internal/db/` | Modify | Add read-side context analysis queries / helpers |
| `internal/server/` | Modify | Add context API endpoints |
| `frontend/src/lib/api/` | Modify | Add client types and request helpers |
| `frontend/src/lib/stores/` | Add | Context analysis store |
| `frontend/src/lib/components/` | Add | Context summary, breakdown, timeline UI |
| `frontend/src/App.svelte` | Modify | Route or tab integration |

## MVP Tickets

### Ticket MVP-1: Define normalized context taxonomy

**Goal:** Create a minimal shared taxonomy for context sources that can be
computed from existing session/message/tool data.

**Deliverables:**

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

**Implementation notes:**

- Start with a Go enum or string constants under a new package such as
  `internal/contextview` or `internal/analysis/context`.
- Reuse existing `source_type`, `source_subtype`, `tool_calls`, and
  `tool_result_events` data where possible.
- Map imperfectly when needed, but always preserve a fallback category.

**Acceptance criteria:**

- Every message/tool-derived unit can be assigned to one category.
- Unknown cases fall into `other` rather than failing.

### Ticket MVP-2: Create backend context summary model

**Goal:** Define the API payload shapes for session context analysis.

**Deliverables:**

- Summary response struct
- Composition row struct
- Timeline row struct
- Health summary struct

**Suggested response shape:**

- `session_id`
- `estimated_context_tokens`
- `estimated_context_percent`
- `estimated_remaining_tokens`
- `estimate_confidence`
- `health_state`
- `health_reasons`
- `composition[]`
- `timeline[]`

**Acceptance criteria:**

- Payload is stable enough to support frontend work.
- API clearly labels estimated fields.

### Ticket MVP-3: Compute session context composition

**Goal:** Produce a category-level context estimate for a single session.

**Implementation notes:**

- Use existing message content length and token usage when present.
- Prefer token usage fields if available.
- Fall back to approximate token estimation when exact values are missing.
- Count tool outputs separately when reconstructable from `tool_result_events`
  and/or `tool_calls.result_content_length`.

**Suggested location:**

- `internal/db/context_analysis.go`
- tests in `internal/db/context_analysis_test.go`

**Acceptance criteria:**

- API returns non-empty composition for sessions with messages.
- Categories sum to a plausible session-level total.
- Test coverage exists for sessions with mixed messages and tool calls.

### Ticket MVP-4: Compute context growth timeline

**Goal:** Show cumulative growth over time.

**Implementation notes:**

- Base the first version on message ordinals.
- Aggregate per ordinal or merge into turn windows if message-level granularity
  is too noisy.
- For each timeline row, compute:
  - `index`
  - `timestamp`
  - `delta_tokens_estimate`
  - `cumulative_tokens_estimate`
  - `dominant_category`
  - `spike_level`

**Acceptance criteria:**

- Timeline shows monotonically increasing cumulative context.
- Growth spikes are marked when a row exceeds configurable thresholds.

### Ticket MVP-5: Compute minimal health state

**Goal:** Expose a simple health status from observable session behavior.

**Initial heuristics:**

- Healthy: low occupancy, modest growth.
- Watch: moderate occupancy or recent spike.
- Degraded: high occupancy, repeated spikes, or heavy tool-result dominance.
- Critical: near-threshold occupancy with sustained noisy growth.

**Acceptance criteria:**

- Summary card returns one of four health states.
- Reasons are included as short explanatory strings.

### Ticket MVP-6: Add context analysis HTTP endpoints

**Goal:** Expose context data to the frontend.

**Endpoints:**

- `GET /api/v1/sessions/{id}/context`
- `GET /api/v1/sessions/{id}/context/timeline`

**Suggested files:**

- `internal/server/context.go`
- register routes in `internal/server/server.go`

**Acceptance criteria:**

- Endpoints return JSON for valid sessions.
- Missing sessions return `404`.
- Error shape matches existing server conventions.

### Ticket MVP-7: Add frontend API types and client methods

**Goal:** Make context analysis available to the SPA.

**Suggested files:**

- `frontend/src/lib/api/types/context.ts`
- export from `frontend/src/lib/api/types/index.ts`
- add methods in `frontend/src/lib/api/client.ts`

**Acceptance criteria:**

- Frontend can fetch both summary and timeline.
- TypeScript types compile cleanly.

### Ticket MVP-8: Add context analysis store

**Goal:** Provide reactive client-side loading and caching.

**Suggested file:**

- `frontend/src/lib/stores/context.svelte.ts`

**Store responsibilities:**

- current session context summary
- timeline data
- loading state
- error state
- reload on active session change

**Acceptance criteria:**

- Context data loads when the user enters the context view for a session.
- Store resets correctly when the active session changes.

### Ticket MVP-9: Build Context Summary Card

**Goal:** Provide immediate legibility.

**UI content:**

- estimated context usage
- percent full
- remaining budget
- health state
- short explanation

**Suggested file:**

- `frontend/src/lib/components/context/ContextSummaryCard.svelte`

**Acceptance criteria:**

- A user can tell in seconds how full and how healthy the session is.

### Ticket MVP-10: Build Context Composition visualization

**Goal:** Show what is consuming context.

**Suggested representation:**

- stacked horizontal bar in v1,
- optional legend/list below,
- hover or click for details.

**Suggested file:**

- `frontend/src/lib/components/context/ContextCompositionChart.svelte`

**Acceptance criteria:**

- Users can inspect category percentages and raw estimates.

### Ticket MVP-11: Build Context Timeline visualization

**Goal:** Show how context grew over time.

**Suggested file:**

- `frontend/src/lib/components/context/ContextTimeline.svelte`

**Acceptance criteria:**

- Timeline is readable for both short and long sessions.
- Spikes are visually obvious.

### Ticket MVP-12: Integrate context view into session detail UX

**Goal:** Add an entry point without disrupting the existing transcript flow.

**Options:**

- New tab in the center column.
- Drawer or alternate route state.

**Recommendation:**

- Add a session-level tab mode rather than a full new top-level route.

**Acceptance criteria:**

- Users can switch between transcript and context view for a session.
- Existing session navigation still works.

### Ticket MVP-13: Add backend tests

**Goal:** Lock the first version to deterministic behavior.

**Test cases:**

- mixed categories,
- tool-heavy session,
- low-data session,
- missing token-usage fallback,
- monotonic cumulative timeline,
- health state classification.

### Ticket MVP-14: Add frontend tests

**Goal:** Prevent regressions in the new view.

**Test cases:**

- summary card rendering,
- composition rendering,
- timeline rendering,
- active session switch behavior,
- loading and error states.

## MVP Exit Criteria

MVP is complete when:

- a user can open a session,
- view a context summary,
- see a breakdown of what is consuming context,
- inspect how context grew over time,
- and understand whether the session is broadly healthy or degraded.

---

## v1

## v1 Goal

Move from observability to decision support.

v1 should tell the user not only what happened, but what to do next. The core
deliverable is a recommendation engine with evidence-backed, actionable
guidance.

## v1 Scope

### Included

- Branch-point detection.
- Recommendation generation.
- Actionable guidance text.
- Transcript linkage from timeline and branch points.
- Better health signals.
- Copyable command/prompt text for supported agents.

### Excluded

- Full advisor-agent execution.
- Cross-session pattern learning.
- Advanced semantic relevance scoring.

## v1 Tickets

### Ticket V1-1: Define branch-point model

**Goal:** Represent candidate decision points in the session.

**Fields:**

- `turn_index`
- `timestamp`
- `reason_code`
- `reason_text`
- `recommended_action`
- `confidence`
- `evidence_refs`

**Acceptance criteria:**

- Backend can emit zero or more branch points for a session.

### Ticket V1-2: Implement tangent detection heuristic

**Goal:** Find low-value branches that likely polluted context.

**Signals to combine:**

- topic/file-focus shift,
- repeated tool attempts,
- heavy output growth,
- abrupt later abandonment,
- lack of reuse in later turns.

**Acceptance criteria:**

- At least obvious tangent cases are detected in tests.

### Ticket V1-3: Implement rewind-opportunity detection

**Goal:** Detect stable pre-branch points where rewind would preserve useful
context while dropping noise.

**Acceptance criteria:**

- Sessions with a clear failed tangent produce a rewind recommendation at the
  last stable point before the tangent.

### Ticket V1-4: Implement compact-opportunity detection

**Goal:** Detect when the session is large but still coherent enough that a
focused compact is preferable to a fresh session.

**Acceptance criteria:**

- High-occupancy but coherent sessions can yield compact recommendations rather
  than generic warnings.

### Ticket V1-5: Implement fork/fresh-session detection

**Goal:** Detect when the task has pivoted and old context is no longer aligned
enough.

**Acceptance criteria:**

- Sessions with a substantial task pivot can emit a fork/start-fresh
  recommendation.

### Ticket V1-6: Implement subagent-opportunity detection

**Goal:** Detect next-step work that likely belongs in a subagent.

**Examples:**

- verification pass,
- documentation generation,
- research into another codebase,
- high-output search/exploration task.

**Acceptance criteria:**

- Recommendation engine can output `subagent` as a candidate action with
  rationale.

### Ticket V1-7: Build recommendation engine

**Goal:** Convert signals into ranked candidate actions.

**Output:**

- primary recommendation,
- alternatives,
- rationale,
- confidence,
- evidence refs.

**Suggested file:**

- `internal/contextview/recommendation.go`

**Acceptance criteria:**

- Engine consistently returns one primary action from:
  - continue
  - rewind
  - compact
  - fork
  - fresh_session
  - subagent

### Ticket V1-8: Generate actionable text payloads

**Goal:** Recommendations must be operational, not abstract.

**Outputs by action type:**

- rewind: suggested reprompt
- compact: suggested compact focus text
- fork/fresh: suggested handoff brief
- subagent: suggested delegation instruction

**Acceptance criteria:**

- API payload contains copyable action text for non-continue recommendations.

### Ticket V1-9: Add recommendation endpoints

**Endpoints:**

- `GET /api/v1/sessions/{id}/context/branch-points`
- `GET /api/v1/sessions/{id}/context/recommendation`

**Acceptance criteria:**

- Frontend can fetch branch points and recommendation independently.

### Ticket V1-10: Build Branch Points UI

**Goal:** Let users inspect decision points and jump back into the transcript.

**Suggested file:**

- `frontend/src/lib/components/context/ContextBranchPointList.svelte`

**Acceptance criteria:**

- Clicking a branch point highlights or jumps to relevant session content.

### Ticket V1-11: Build Recommendation Card UI

**Goal:** Put the decision at the top of the experience.

**Content:**

- recommended action,
- confidence,
- short rationale,
- copyable text,
- alternative actions.

**Suggested file:**

- `frontend/src/lib/components/context/ContextRecommendationCard.svelte`

**Acceptance criteria:**

- A user can copy a suggested next-step prompt or command directly from the UI.

### Ticket V1-12: Add transcript linkage

**Goal:** Tie x-ray evidence back to the transcript.

**Behaviors:**

- click timeline point -> scroll transcript,
- click branch point -> highlight implicated ordinals,
- click evidence item -> reveal transcript neighborhood.

**Acceptance criteria:**

- Context view and transcript view feel connected rather than separate tools.

### Ticket V1-13: Add agent-specific action rendering

**Goal:** Recommendations should use the source agent’s actual semantics where
known.

**Examples:**

- Claude Code: `/rewind`, `/compact`, `/clear`
- other agents: generic handoff/start-fresh phrasing when no exact command is
  supported

**Acceptance criteria:**

- At least Claude Code receives tailored action text.
- Other agents degrade gracefully with generic guidance.

### Ticket V1-14: Add E2E coverage

**Goal:** Ensure the main user flow works.

**Scenarios:**

- open session,
- switch to context view,
- inspect timeline,
- inspect recommendation,
- jump into transcript,
- copy suggested action text.

## v1 Exit Criteria

v1 is complete when:

- the product can identify meaningful branch points,
- provide a primary recommendation,
- explain the rationale,
- and give the user concrete text for acting on that recommendation.

---

## v2

## v2 Goal

Add higher-order analysis via a fresh-context advisor agent and deepen the
quality of the recommendation system.

v2 is where the feature becomes both a visualizer and a context coach.

## v2 Scope

### Included

- Manual advisor-agent execution.
- Structured advisor input bundle.
- Advisor result persistence or caching.
- Richer heuristics and confidence modeling.
- More agent-specific mappings.
- Optional live warnings for active sessions.

### Excluded

- Automatic execution of recommended actions.
- Fully autonomous session management.

## v2 Tickets

### Ticket V2-1: Define advisor input schema

**Goal:** Create a structured representation of a session for external analysis.

**Include:**

- session metadata,
- context summary,
- composition,
- timeline summary,
- branch points,
- signals,
- recent turn snippets,
- key files,
- current recommendation.

**Acceptance criteria:**

- Advisor input can be serialized deterministically.

### Ticket V2-2: Define advisor output schema

**Goal:** Standardize analysis results from the side agent.

**Fields:**

- diagnosis,
- recommendation,
- confidence,
- rationale,
- alternatives,
- suggested action text,
- evidence references.

**Acceptance criteria:**

- Output can be rendered by frontend without free-form parsing hacks.

### Ticket V2-3: Implement advisor orchestration backend

**Goal:** Trigger a side-agent run on demand.

**Implementation paths:**

- reuse existing insight-generation infrastructure if appropriate,
- or create a dedicated context-advisor execution path.

**Acceptance criteria:**

- A user can request advisor analysis and retrieve the result asynchronously or
  synchronously, depending on implementation.

### Ticket V2-4: Build Advisor Panel UI

**Goal:** Show side-agent analysis distinctly from system heuristics.

**Suggested file:**

- `frontend/src/lib/components/context/ContextAdvisorPanel.svelte`

**Acceptance criteria:**

- UI explicitly labels advisor output as agent-generated analysis.
- Confidence and evidence are visible.

### Ticket V2-5: Add advisor re-run and caching controls

**Goal:** Prevent unnecessary reruns and give the user control.

**Features:**

- run analysis,
- show last run timestamp,
- refresh analysis,
- display stale state if session changed.

### Ticket V2-6: Improve relevance and drift heuristics

**Goal:** Move beyond simple occupancy and spike logic.

**Potential improvements:**

- lexical similarity windows,
- file-focus persistence,
- branch abandonment scoring,
- durable-conclusion vs intermediate-output ratio.

### Ticket V2-7: Add proactive live warnings

**Goal:** Surface “watch” or “degraded” warnings during active sessions.

**Examples:**

- “Context grew 18% in the last 6 turns.”
- “Tool-result output is dominating recent growth.”
- “Likely tangent detected.”

**Acceptance criteria:**

- Warnings are informational and dismissible.
- No automatic action is taken.

### Ticket V2-8: Broaden agent-specific support

**Goal:** Improve fidelity beyond the best-supported agent.

**Tasks:**

- map compaction semantics where available,
- map subagent semantics where available,
- improve token estimation confidence by agent,
- improve action text by agent.

### Ticket V2-9: Add evaluation harness

**Goal:** Measure recommendation usefulness on saved sessions.

**Ideas:**

- fixture sessions with expected branch points,
- compare engine recommendation vs expected action,
- score precision and recall on tangent detection.

## v2 Exit Criteria

v2 is complete when:

- users can request fresh-context advisor analysis,
- advisor output is integrated safely into the UX,
- and the recommendation quality is measurably better than the rule-only
  baseline.

---

## Sequencing Recommendations

## Recommended Order

1. MVP-1 through MVP-6
2. MVP-7 through MVP-12
3. MVP-13 and MVP-14
4. V1-1 through V1-9
5. V1-10 through V1-14
6. V2-1 through V2-5
7. V2-6 through V2-9

## Safe Parallelization

Can be parallelized after payload shapes stabilize:

- backend API implementation vs frontend component scaffolding,
- context composition vs timeline computation,
- branch-point engine vs frontend recommendation components,
- advisor schema definition vs advisor UI scaffolding.

Should not be parallelized too early:

- taxonomy definition and payload naming,
- recommendation payload schema,
- transcript-linking contract between frontend and backend.

---

## Suggested Milestone Acceptance

## Milestone A: MVP Demo

The demo should show:

- one session open in transcript view,
- switch to Context tab,
- summary card,
- context composition chart,
- context timeline,
- health state explanation.

## Milestone B: v1 Demo

The demo should show:

- context view with branch points,
- a primary recommendation,
- evidence supporting the recommendation,
- click-through into transcript,
- copyable action text.

## Milestone C: v2 Demo

The demo should show:

- system recommendation,
- advisor-agent recommendation,
- visible distinction between the two,
- user ability to re-run advisor analysis.

---

## Implementation Notes For The Agent

- Prefer deriving from existing `sessions`, `messages`, `tool_calls`, and
  `tool_result_events` before introducing new tables.
- Keep all estimates explicitly labeled as estimates.
- Start Claude-first if agent-specific fidelity forces prioritization.
- Do not block the UI on advisor-agent support.
- Treat recommendation quality as a product concern, not just an engineering
  concern; explanations matter as much as the action label.

## Final Guidance

If scope pressure appears, cut in this order:

1. Cut advisor-agent integration before cutting recommendation quality.
2. Cut recommendation sophistication before cutting the core context timeline.
3. Cut cross-agent breadth before cutting depth for the best-supported agent.

The feature only becomes truly useful when the user can understand both
**what happened to the session context** and **what to do next**. MVP should
solve the first problem. v1 should solve the second. v2 should make the advice
substantially smarter.
