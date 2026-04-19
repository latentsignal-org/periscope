# Periscope Product Specification

## Document Status

- Status: Draft for implementation handoff
- Author: Codex
- Date: 2026-04-18
- Product area: Context engineering, optimization

## Summary

Periscope is a context session visualizer and improver for coding-agent
workflows. It provides session-level observability and decision support so users
can understand how a live or historical coding-agent session consumed its
context window over time, how that context composition affects session quality,
and what session-management action to take next.

Periscope is built on the thesis that success in coding-agent workflows depends
not only on model quality and prompt quality but on context quality. Long
sessions accumulate useful knowledge, but they also accumulate irrelevant tool
output, failed attempts, stale assumptions, debugging tangents, and compressed
summaries. As a session grows, the agent's effective intelligence often degrades
because the context becomes diluted, noisy, or misaligned with the current
objective.

The product ships in two phases:

- **V1 — Context Visualizer.** An x-ray of how session context was spent and how
  it evolved. Descriptive only: no analysis or recommendations.
- **V2 — Context Guidance.** An interpretation layer on top of V1: compute
  signals, classify session health, detect branch points, and recommend a
  concrete next action — continue, rewind, fork, compact, or delegate to a
  subagent.

Available context capacity is session-specific. Periscope should infer the
maximum usable context window for each session from the best available signal:
agent, model, provider-reported limits when exposed, and the user's current
plan or entitlement tier when that affects model limits. Occupancy percentages,
remaining budget, and threshold-based guidance must be calculated against this
session-specific maximum rather than a single global default.

Together, the two phases turn context engineering from intuition into something
visible, measurable, and actionable.

### Non-Goals

- Do not reconstruct exact provider-side prompt packing with perfect fidelity if
  the provider does not expose it.
- Do not replace the primary transcript viewer.
- Do not automatically mutate user sessions.
- Do not act as a universal tokenizer for every provider at perfect parity.
- Do not attempt general agent evaluation across all dimensions; focus
  specifically on context health and session branching quality.

## Problem Statement

Current coding-agent tools expose either a very shallow context meter or no
context introspection at all. A user may see a single number like "23.7k /
200.0k" or a coarse category breakdown, but they typically cannot answer the
questions that matter operationally:

- What interactions caused the biggest growth in context?
- Which portion of context is still relevant to the current task?
- Which tool outputs consumed context but are now dead weight?
- Where did the session go on a tangent?
- When did the session become unhealthy enough that continuing was the wrong
  move?
- Would rewind have been better than "try something else"?
- Would a fresh session, targeted compact, or subagent have preserved better
  model performance?

Without this visibility, users manage sessions by intuition. That leads to:

- wasted context budget,
- degraded model performance late in sessions,
- compactions performed too late or with poor focus,
- continued work on top of polluted context,
- unnecessary rereading of files after avoidable fresh starts,
- poor use of subagents,
- weak handoffs between session branches.

The user needs a system that turns hidden context dynamics into explicit signals
and recommended actions.

## Target Users

### Primary Users

- Developers using long-running coding-agent sessions.
- Power users of Claude Code, Codex, Cursor, Gemini CLI, and similar tools.
- Users managing large refactors, debugging sessions, or multi-step build work.

### Secondary Users

- Team leads reviewing agent workflows.
- Researchers investigating session success/failure dynamics.
- Advanced users optimizing cost, speed, and session hygiene.

## User Needs

The user wants to:

- see where session context is going,
- understand why a session feels less effective,
- know which parts of the session are still worth preserving,
- know whether to continue or branch,
- get a suggested action with rationale,
- receive ready-to-use commands or prompts for that action,
- avoid repeating mistakes that poison context.

## Key Concepts

### Context Window

Everything the model can attend to while generating the next response:
instructions, conversation history, tool calls, tool outputs, file contents,
summaries, and embedded metadata.

### Context Quality

The degree to which occupied context remains relevant, coherent, and useful for
the current objective.

### Context Drag

The burden imposed by accumulated but low-value context. Examples include stale
investigation branches, repeated tool failures, excessive logs, and irrelevant
file reads.

### Context Rot

Performance degradation caused by large or noisy context, especially when the
model must attend across too much irrelevant material.

### Branching Point

A point after a completed turn where the user could:

- continue,
- rewind,
- start fresh,
- compact,
- fork,
- spawn or use a subagent.

### Context-Shaping Operation

Any user or system action that changes future context composition:

- rewind,
- clear or fresh session,
- compact,
- fork session,
- delegate to subagent,
- summarize and hand off.

## Product Scope

The product ships in two phases that map to V1 and V2.

### V1: Context Visualizer

A visual, inspectable breakdown of session context history and current state. V1
is descriptive only — it shows what happened and what is currently in the
window, with no analysis, scoring, or recommendations.

V1 must deliver:

- A standalone `/context/:sessionId` route with a link back to the agentsview
  home. The same route may also be embedded as a tab inside the agentsview
  session detail view, and is the surface exposed inside JetBrains IDEs.
- A **summary header** showing estimated context tokens in use, estimated
  percent of the available window consumed, remaining budget, and the inferred
  maximum context window used for the calculation.
- A **composition breakdown** of context by source category (system prompt and
  tool definitions, user messages, assistant messages, file reads, tool calls,
  tool outputs, summaries and compacted handoffs, subagent outputs, and the
  other categories enumerated in FR3).
- A **turn-by-turn timeline** showing per-turn context delta, cumulative context
  estimate, growth spikes, category attribution per spike, and markers for
  compaction, rewind, fork, and subagent events when known.
- **Token estimation** as specified below: prefer recorded token usage from the
  agent's session files, fall back to byte-count with a per-agent ratio.
- **Context-capacity estimation** as specified below: determine the session's
  maximum context window from agent, model, provider metadata, and user-plan
  signals when available; label the result as measured, inferred, or unknown.
- **Live updates over SSE** for active sessions so the visualizer keeps current
  as turns arrive.
- **Support for all agents** registered in `internal/parser/types.go`.
- **Post-compaction context only.** When a session has been compacted, the
  timeline begins at the compaction event and treats the compacted summary as a
  single source segment.

### V2: Context Guidance

An interpretation layer built on top of V1 that turns the visualized data into
diagnoses and recommended actions.

V2 must deliver:

- **Context-health signals** across occupancy, noise, retry, compression, and
  branching families.
- **Health classification** into Healthy / Watch / Degraded / Critical.
- **Branch-point detection** with reason, recommended action, estimated benefit,
  and confidence.
- **Recommendation engine** that produces one primary action plus alternatives:
  continue, rewind, compact, fork, fresh session, or subagent.
- **Operational guidance text**: copyable command, suggested prompt, suggested
  compact focus, suggested rewind reprompt, suggested fork brief, suggested
  subagent task wording.
- **Evidence inspector** that traces every recommendation back to the turns,
  categories, and signals that produced it.
- **Relevance estimation** as a heuristic: recency × file-overlap ×
  topic-overlap.
- **Optional guidance agent**: a separate agent invoked with a fresh context
  window to analyze the primary session from outside.
- **Live-session isolation**: when guidance or analysis is run against an
  active session, it must execute in a separate subagent or equivalent fresh
  analysis session so the primary live session does not accumulate the analysis
  prompt, evidence bundle, or generated guidance output in its own context.

## Jobs To Be Done

### JTBD 1: Diagnose a session that feels degraded

When a long coding-agent session starts producing weaker results, the user wants
to see why the session degraded and what accumulated context is likely causing
the drop in quality, so they can recover without guessing.

### JTBD 2: Decide what to do next

At the end of a turn, the user wants to know whether to continue, rewind,
compact, fork, or delegate, so they can preserve useful context and avoid making
the session worse.

### JTBD 3: Recover from a tangent

When the agent has explored an unhelpful branch, the user wants to identify the
last clean point before the tangent and receive a ready-to-use rewind or handoff
instruction.

### JTBD 4: Prepare a fresh branch intelligently

When starting a new session or fork, the user wants a concise, accurate handoff
containing only the relevant learnings, constraints, files, and next steps.

### JTBD 5: Learn better session habits

When reviewing past work, the user wants to see what context-management choices
correlated with success or failure, so they can improve how they work with
coding agents.

## Functional Requirements

V1 requirements are descriptive; V2 requirements add interpretation. Numbering
preserves the original draft for cross-reference.

### V1 Requirements

#### FR1: Show current session context state

The product must show a current-state summary of the session, including:

- estimated context tokens in use,
- inferred maximum context window for this session,
- estimated percent of available context consumed,
- available remaining context,
- source breakdown by category.

#### FR2: Show context growth over time

The product must display context growth turn by turn across the session.

This must include:

- per-turn context delta,
- cumulative context estimate,
- growth spikes,
- category attribution for each spike,
- markers for compaction, rewind, fork points, and subagent events when known.

#### FR3: Attribute context to sources

The product must classify context consumption into meaningful categories.
Initial categories include:

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
- deferred or hidden tool payloads when exposed by the source agent,
- free space.

This taxonomy must be extensible, and agent-specific mapping rules must be
allowed.

#### FR8: Support historical and current sessions

The feature must work for:

- historical sessions already indexed in the database,
- the currently active session, with live updates pushed over SSE.

For every session, V1 must determine the best available maximum context window
before computing percentages. The determination should use, in priority order:

1. provider- or agent-recorded context limits attached to the session,
1. model-specific limits known for the detected agent/model pair,
1. user-plan or entitlement-aware limits when the same model exposes different
   windows by plan,
1. conservative agent-level defaults when no better signal exists.

If the exact session limit cannot be proven, the UI and API must label the
window size as inferred rather than measured.

#### FR9: Support agent-specific interpretation

The system must support agent-specific logic for:

- context category mapping,
- compaction event recognition,
- rewind or fork semantics,
- subagent event detection,
- token estimation quality.

V1 must support every agent registered in `internal/parser/types.go`.

### V2 Requirements

#### FR4: Surface context-health signals

The product must compute interpretable signals that help estimate session
quality. Candidate signals include:

- sustained high context occupancy,
- rapid context growth in a short turn span,
- long debugging or exploration tangents,
- repeated file rereads,
- repeated failed tool attempts,
- excessive low-value tool output,
- retry loops,
- compaction frequency,
- suspected compaction loss,
- large branches later abandoned,
- high ratio of intermediate output to durable conclusions.

#### FR5: Highlight branching points

The product must identify meaningful points in the session where an alternative
action would likely have been beneficial.

For each highlighted branch point, the UI must show:

- the message or turn index,
- the likely reason it matters,
- the recommended action,
- the estimated benefit,
- the confidence level.

#### FR6: Recommend next best action

The product must generate a recommendation for what the user should do now.

At minimum, the system must support recommending:

- continue in current session,
- rewind to a specific point,
- compact with a specific focus,
- fork a new session with a handoff,
- start a fresh session with a distilled brief,
- use a subagent for the next work chunk.

#### FR7: Provide operational guidance

Recommendations must be actionable. The product must provide:

- a concrete explanation,
- the exact command or control to use when the source agent supports it,
- suggested prompt text or handoff text,
- suggested compact focus instructions where relevant,
- suggested rewind reprompt,
- suggested fork brief,
- suggested subagent task wording.

#### FR10: Support guidance-agent analysis

The system should optionally invoke a separate analysis agent with a fresh
context window to analyze the primary session.

The guidance agent must:

- receive a structured summary of the primary session,
- diagnose context health,
- explain why the session is healthy or unhealthy,
- recommend a context-shaping operation,
- produce ready-to-use action text,
- avoid importing the entire raw session when a compressed representation is
  sufficient.
- when analyzing a live session, run outside the primary session context so the
  analysis itself does not pollute the live session's context window.

## User Experience Requirements

### UX1: Standalone /context route

Periscope is served at `/context/:sessionId`. The page links back to the
agentsview home at `/`. The same view may also be embedded as a tab inside the
agentsview session detail view.

The route must be safe to expose inside a JetBrains IDE webview as the primary
surface for using Periscope from inside the IDE.

### UX2: Immediate legibility

Within a few seconds, V1 must let the user answer:

- How full is this session?
- What is consuming the context?

V2 must additionally let the user answer:

- Is this session still healthy?
- What should I do next?

### UX3: Drill-down capability

V1 users must be able to inspect:

- turn-level context changes,
- category-level consumption.

V2 adds:

- branch-point explanations,
- recommended action details.

### UX4: Action-oriented summary (V2)

The top of the V2 view must include a compact decision card:

- Session health: healthy, watch, degraded, critical
- Primary issue: tangent, tool noise, retry loop, compaction risk, stale branch
- Recommended action
- Confidence
- CTA text or copyable command

### UX5: Avoid false authority (V2)

The system must distinguish:

- measured facts,
- heuristics,
- inferred judgments,
- guidance-agent opinions.

Confidence and uncertainty must be explicit.

## Proposed User Interface

V1 uses a fixed three-section page layout:

1. `ContextSummaryCard`
1. `ContextCompositionChart`
1. `ContextTimeline`

The page structure is a vertical stack. The timeline section uses explicit
turn-row rendering with per-turn stacked category bars, inline annotations, and
strong compaction dividers. This is the chosen V1 UI, not an open design
question.

### Section A: Context Summary Header (V1)

Displays:

- current estimated context usage,
- percent full,
- remaining budget,
- timestamp of last update.

V2 adds health status and the most likely recommended action to this header.

### Section B: Context Composition Breakdown (V1)

A stacked bar or treemap showing major context categories.

The user must be able to:

- hover or click a category,
- see token estimate and percentage,
- navigate to representative turns that contributed to it.

### Section C: Context Timeline (V1)

A chronological view of turns showing:

- per-turn additions,
- cumulative context,
- session-management events,
- tool-heavy sections.

The V1 timeline must render as explicit rows rather than only as an aggregate
chart. Each row should show the turn index, delta, cumulative estimate,
dominant category, and a stacked per-turn category composition bar. Inline
annotation rows should call out spikes, compaction boundaries, and subagent
events when known.

This is the core x-ray. V2 overlays tangent windows and suspected degraded
zones.

### Section D: Branch Points (V2)

A list of recommended or historically meaningful branch points:

- "Rewind recommended after turn 42"
- "Fork recommended after task pivot at turn 68"
- "Compact should have happened before turn 103"
- "Use subagent for verification branch at turn 119"

### Section E: Guidance Panel (V2)

Displays either rule-based analysis, agent-based analysis, or both:

- diagnosis,
- rationale,
- recommendation,
- suggested command text,
- suggested handoff text,
- confidence.

### Section F: Evidence Inspector (V2)

Shows why the system made the recommendation:

- turns involved,
- categories implicated,
- repeated tools,
- repeated reads,
- compaction boundaries,
- high-growth spikes.

## Example Decision Outputs (V2)

### Example: Continue

"Continue in the current session. Recent context growth is modest, the task has
remained coherent, and the last 20 turns are still aligned with the active
objective."

### Example: Rewind

"Rewind to turn 57. The subsequent 19 turns are dominated by a failed debugging
branch and tool retry loop. Preserve the file-reading context up to turn 57 and
re-prompt with the refined constraint."

### Example: Compact

"Compact now, but focus on the auth refactor and exclude the exploratory test
debugging branch. The session is large but still contains useful state that
would be expensive to rebuild."

### Example: Fork

"Fork a new session. The session pivoted from implementation to deployment and
only part of the earlier context remains relevant. Start fresh with the handoff
below."

### Example: Subagent

"Use a subagent for verification. The next task will generate substantial
intermediate output that is unlikely to be useful in the parent session."

## Context Model

The product needs an internal model of session context, even if it is only an
estimate.

### Required Concepts

- Context source
- Context segment
- Turn
- Branch point (V2)
- Session-management event
- Context-health signal (V2)
- Recommendation (V2)

### Proposed Data Model

#### ContextSegment (V1)

Represents a piece of session context attributable to a source.

Fields:

- `session_id`
- `turn_index`
- `message_id` or equivalent
- `source_type`
- `source_subtype`
- `origin_agent`
- `token_estimate`
- `content_length`
- `branch_group`
- `created_at`

V2 adds:

- `is_ephemeral_candidate`
- `is_currently_relevant_estimate`

#### ContextTurnDelta (V1)

Represents context added or transformed during a turn.

Fields:

- `session_id`
- `turn_index`
- `user_input_tokens_estimate`
- `assistant_output_tokens_estimate`
- `tool_tokens_estimate`
- `file_read_tokens_estimate`
- `summary_tokens_estimate`
- `cumulative_context_estimate`
- `primary_label`
- `notes`

#### ContextBranchPoint (V2)

Fields:

- `session_id`
- `turn_index`
- `branch_type`
- `reason_code`
- `reason_text`
- `recommended_action`
- `confidence`
- `action_payload`

#### ContextRecommendation (V2)

Fields:

- `session_id`
- `generated_at`
- `engine`
- `action`
- `summary`
- `rationale`
- `confidence`
- `suggested_command`
- `suggested_prompt`
- `evidence_refs`

## Token Estimation

Token counts must be estimated for every context segment. The pipeline:

1. **Prefer recorded token usage.** When the source agent's session file exposes
   per-message token counts (typically in `token_usage` fields on assistant or
   system messages), use those values directly.
1. **Fall back to byte count plus per-agent ratio.** When token counts are
   absent, estimate tokens from content byte length using an agent-specific
   bytes-per-token ratio. Each parser in `internal/parser/` declares its own
   ratio; the parser registry in `internal/parser/types.go` is extended to carry
   it.

Estimation accuracy is a best-effort signal. Every API and UI response that
includes a token count must mark whether the count is **measured** (from
recorded usage) or **estimated** (from the byte-count fallback).

## Context Capacity Estimation

Periscope must estimate the maximum available context window for each session so
occupancy percentages and remaining-budget calculations are grounded in the
actual session limit.

Capacity detection pipeline:

1. **Prefer recorded limits.** If the source agent or provider records the
   session's maximum context window, use it directly.
1. **Otherwise infer from agent + model.** Map the detected agent and model to
   a known maximum window.
1. **Adjust for user plan when relevant.** If the same model can expose
   different context limits depending on the user's subscription, plan, or
   entitlement tier, Periscope should use the best available plan signal to pick
   the correct limit.
1. **Fall back conservatively.** If exact limits remain unavailable, use a
   conservative default for that agent/model family and mark it as inferred.

Every API and UI response that includes occupancy percentage, remaining budget,
or threshold projections must also expose whether the maximum context window was
**measured**, **inferred**, or **unknown**.

## Compaction Handling

Periscope shows post-compaction context only. When a session has been compacted,
the timeline begins at the compaction event and treats the compacted summary as
a single source segment. Pre-compaction segments are not reconstructed.

This is an explicit V1 simplification, and the same constraint applies to V2.
Reconstructing pre-compaction context across all agents is out of scope.

## Signal Framework (V2)

The system must compute a set of context-health signals. These do not need to be
perfect, but they must be explainable.

### Signal Families

#### 1. Occupancy Signals

- current percent of window used,
- rate of recent growth,
- projected turns until threshold crossings.

#### 2. Noise Signals

- tool-output-heavy turns,
- repeated search or grep output,
- long raw logs,
- duplicated or near-duplicated reads.

#### 3. Retry Signals

- repeated attempts on same failing approach,
- repeated edits to same files without convergence,
- repeated tool failures,
- repeated reversions or churn.

#### 4. Compression Signals

- number of compactions,
- time since last compaction,
- compaction near high-noise state,
- likely summary loss after a pivot.

#### 5. Branching Signals

- identifiable tangent start,
- abandoned side branch,
- points where knowledge stabilized before experimentation resumed,
- points where subagent use would have contained disposable output.

### Health States

The system must classify overall session health into:

- Healthy
- Watch
- Degraded
- Critical

This label derives from a score plus major trigger conditions.

### Example Heuristic Score Inputs

- occupancy percentile,
- recent growth acceleration,
- noise-to-conclusion ratio,
- retry density,
- compaction risk,
- stale branch weight,
- tool failure density.

## Recommendation Engine (V2)

The recommendation engine combines:

- deterministic heuristics,
- interpretable rule-based logic,
- optional guidance-agent analysis.

### Recommendation Pipeline

1. Reconstruct or estimate context state.
1. Compute turn-level deltas and signal values.
1. Detect branching points.
1. Score candidate actions.
1. Produce a primary recommendation and alternatives.
1. Attach evidence and suggested commands.

### Candidate Actions and Selection Logic

#### Continue

Recommend when:

- recent context remains coherent,
- growth is moderate,
- no significant tangent or retry loop is active,
- rebuilding context elsewhere would be wasteful.

#### Rewind

Recommend when:

- a recent branch is low-value or failed,
- the pre-branch context remains useful,
- the user can preserve useful reads while dropping bad experimentation.

#### Compact

Recommend when:

- the session is large but still directionally coherent,
- preserving context is valuable,
- a focused summary can retain the right state.

#### Fork or Fresh Session

Recommend when:

- task focus changed materially,
- old context is partially useful but no longer aligned enough,
- compaction would likely preserve the wrong things,
- the user benefits from a user-authored handoff.

#### Subagent

Recommend when:

- the next work chunk will produce lots of disposable output,
- only the conclusion needs to return,
- the parent context should remain clean.

## Guidance-Agent Specification (V2)

The Context Guidance agent is an optional higher-order agent that analyzes the
primary session from outside the session.

### Purpose

Use a fresh context window to reason about the health of the primary session
without being degraded by the primary session's own overloaded context.

### Inputs

The guidance agent receives a structured representation rather than the full raw
transcript by default. Candidate input bundle:

- session metadata,
- current occupancy estimate,
- context category breakdown,
- context timeline summary,
- top growth spikes,
- detected branch points,
- detected compactions,
- recent turns,
- key files involved,
- task summary,
- computed signals,
- snippets of evidence around suspected problems.

For active sessions, this bundle must be passed to a separate subagent or
equivalent fresh analysis session rather than appended to the live parent
session.

### Outputs

The guidance agent must return:

- diagnosis,
- primary recommendation,
- alternative options,
- rationale,
- exact commands where agent-specific commands exist,
- suggested prompt or handoff text,
- confidence,
- cited evidence from the structured input.

### Safety Constraints

- Do not allow the guidance agent to issue destructive commands automatically.
- Clearly separate guidance-agent opinion from measured telemetry.
- Do not imply certainty where evidence is weak.
- Provide fallback recommendations when confidence is low.
- Do not run live-session guidance inline inside the primary session if doing so
  would enlarge or contaminate the primary session's context.

### Example Guidance Output

"Recommendation: rewind to turn 84. The session retains useful repo-reading
context up to turn 84, but turns 85-104 represent a failed tangent into
deployment warnings that do not support the active auth bug. Rewind and
re-prompt with: 'Ignore the deployment warning branch. Focus on auth middleware
in files X and Y. We ruled out approach A because constraint B.'"

## Agent-Specific Considerations

Different coding agents expose different observability surfaces. The product
must handle partial fidelity by agent. V1 supports every agent registered in
`internal/parser/types.go`.

### Claude Code

Likely strongest support for:

- compaction events,
- rewind semantics,
- subagent detection,
- tool usage richness,
- source categories visible in the session UI.

### Codex and Similar Tools

Likely strong support for:

- session transcript parsing,
- tool-call attribution,
- subagent/delegation traces where present,
- token estimates from messages or tool metadata.

### General Requirement

The product must define:

- a normalized context taxonomy,
- a per-agent mapping layer,
- a confidence score for each context estimate.

## Data Requirements

The implementation must reuse existing `agentsview` session and message data
where possible and extend the schema only where needed.

### Existing Data Likely Reusable

- sessions,
- messages,
- tool calls,
- tool result events,
- source metadata,
- token usage on messages,
- signal-related session metadata,
- subagent session relationships when present.

V1 implementation should build on top of this existing data model rather than
introducing a parallel session-ingestion pipeline.

### New Derived Data Needed

- context timeline aggregates (V1),
- context category attribution (V1),
- turn-level context deltas (V1),
- branch-point detections (V2),
- recommendation objects (V2),
- guidance-agent outputs (V2),
- optional cached token estimates.

## Storage Strategy

Lightweight context metrics are computed on read. Persistence is added only for
expensive derived analyses (such as guidance-agent outputs in V2) and only when
caching becomes necessary for UX.

Trade-offs we accept:

- **Pro:** fewer schema changes, easy iteration, no migration burden for early
  versions.
- **Con:** more expensive reads, harder to compare repeated analysis runs.

Caching is added once UX and signal definitions stabilize.

## API Requirements

The backend exposes dedicated context-analysis endpoints. V1 endpoints return
descriptive data only; V2 endpoints return interpretation.

### V1 Endpoints

#### `GET /api/v1/sessions/{id}/context`

Returns current summary and context composition.

#### `GET /api/v1/sessions/{id}/context/timeline`

Returns per-turn or per-window context growth data.

### V2 Endpoints

#### `GET /api/v1/sessions/{id}/context/branch-points`

Returns detected branch points and suggested actions.

#### `GET /api/v1/sessions/{id}/context/recommendation`

Returns the primary recommendation and rationale.

#### `POST /api/v1/sessions/{id}/context/analyze`

Triggers fresh analysis, optionally including guidance-agent execution.

#### `GET /api/v1/sessions/{id}/context/guidance`

Returns the latest guidance-agent result if one exists.

### Response Requirements

Every response must clearly separate:

- measured values,
- estimated values,
- inferred session-capacity values,
- inferred values,
- model-generated guidance text.

## Frontend Requirements

### Route

Periscope is served at `/context/:sessionId`. The page includes:

- a link to the agentsview home (`/`),
- embedding as a tab inside the agentsview session detail view in the first V1
  pass,
- safe exposure as a JetBrains IDE webview.

### UI Components

V1 components:

- `ContextSummaryCard`
- `ContextCompositionChart`
- `ContextTimeline`

`ContextTimeline` is the canonical V1 interaction surface. A linked dual-panel
growth chart may be added later, but it is not part of the initial V1 ship.

V2 components (additive):

- `ContextBranchPointList`
- `ContextRecommendationCard`
- `ContextGuidancePanel`
- `ContextEvidenceDrawer`

### Interaction Requirements

- Transcript jump linkage is explicitly deferred from the initial V1 ship. The
  timeline should preserve stable turn/message identifiers so transcript
  navigation can be added later without reworking the data model.
- (V2) Clicking a branch point must highlight the implicated turns.
- (V2) Recommendation actions must provide copyable command and prompt text.
- The UI must make live updates obvious when the session is active.

### Filtering and Modes

The user must be able to switch between:

- current-state view (V1),
- full timeline view (V1),
- recent degradation view (V2),
- branch-point view (V2),
- guidance view (V2).

## Algorithm and Heuristic Guidance (V2)

The system does not need perfect semantic understanding, but it must be useful
and explainable.

### Suggested Initial Heuristics

#### H1: Large low-value tool output

Flag windows where tool-result token volume is high but subsequent turns do not
reference those results materially.

#### H2: Retry loop

Flag repeated similar tool calls or edits over a short span with no clear
convergence.

#### H3: Tangent detection

Detect topic or file-focus shifts that are later abandoned.

#### H4: Rewind opportunity

When a low-value branch follows a stable knowledge-acquisition phase, propose a
rewind point at the last stable point.

#### H5: Compact opportunity

When occupancy is high but task focus remains coherent, propose compact with a
focused summary target.

#### H6: Fresh fork opportunity

When the session has pivoted from one task to another and earlier context is now
only partially relevant, recommend a fork or fresh session.

#### H7: Subagent opportunity

When the next logical task is high-output but conclusion-oriented, recommend
delegation to a subagent.

## Example End-to-End Flows

### Flow 1: Live degraded session (V2)

1. User opens an active session.
1. Context view shows 78% occupancy and rapid recent growth.
1. Timeline highlights the last 24 turns as a tool-heavy debugging tangent.
1. Branch-point card says: "Rewind to turn 91."
1. Recommendation card provides rewind guidance and a reprompt.
1. User copies the suggested rewind text and acts in the primary tool.

### Flow 2: Compact before failure (V2)

1. User sees session at 63% occupancy.
1. Context remains coherent but a large implementation branch has accumulated.
1. Recommendation says compact proactively.
1. Product offers compact focus text.
1. User compacts before context rot worsens.

### Flow 3: Fork after task pivot (V2)

1. Session began as feature implementation.
1. User later switched to documentation and deployment.
1. Visualizer detects low overlap between current task and early session mass.
1. Recommendation says start a new session with a prepared handoff brief.

### Flow 4: Guidance review (V2)

1. User requests guidance-agent analysis.
1. Secondary agent receives structured summary.
1. Guidance returns diagnosis, recommendation, and copyable action text.
1. UI displays guidance output alongside system heuristics.

## Success Metrics

### Product Success Metrics

- users open the context view during long sessions,
- users adopt suggested branch actions (V2),
- users report higher confidence in session-management decisions,
- users recover from degraded sessions faster,
- fewer very-long sessions end in obvious retry loops or compaction failures.

### Behavioral Metrics

- rate of rewind/fork/compact actions after viewing the tool,
- reduction in average context occupancy at action time,
- reduced continuation of low-value tangents,
- increased use of subagents where appropriate.

### Quality Metrics

- recommendation acceptance rate (V2),
- user-rated usefulness,
- branch-point precision (V2),
- guidance-agent agreement with user action (V2),
- false-positive rate for unhealthy-session warnings (V2).

## Constraints and Risks

### Constraint: Exact context fidelity may be unavailable

Many agents do not expose the exact provider-side packed context. The product
must work with best-effort reconstruction and clearly label estimates.

### Risk: Overconfident recommendations

Bad advice here can be costly. The system must remain transparent and cautious.

### Risk: High implementation complexity

Token attribution, branch detection, and multi-agent support can expand quickly.
Scope discipline is required.

### Risk: Guidance contamination

If the guidance agent receives too much raw transcript, it may become expensive
and less focused. Structured summaries must be preferred.

## Privacy and Security

- All analysis must remain local by default.
- Guidance-agent invocation must use the user's existing local agent tools when
  possible.
- No session data may be sent to remote services without explicit user action.
- Cached guidance outputs must be stored as local session artifacts.

## Rollout Plan

### V1: Context Visualizer

Deliver the visualizer described in Product Scope > V1, satisfying FR1, FR2,
FR3, FR8, and FR9. No analysis, scoring, or recommendation surfaces. Live
updates are wired through SSE; all agents in `internal/parser/types.go` are
supported.

### V2: Context Guidance

Deliver the guidance layer described in Product Scope > V2, satisfying FR4
through FR7 and FR10. Includes signal computation, health classification,
branch-point detection, recommendation engine, copyable action text, evidence
inspector, relevance heuristic, and the optional guidance agent with structured
input bundle, fresh-context analysis, and result panel with caching.

Detailed implementation plans for each phase are produced separately.

## Final Product Statement

Periscope is a context x-ray and context coach for coding-agent sessions.

It helps users see how context was spent, where it stopped helping, and what to
do next to preserve model performance. By combining context observability,
branch-point detection, and actionable session-management guidance, it turns
session health from intuition into an explicit, operable part of the coding
workflow.
