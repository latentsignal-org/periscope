# Context Session Visualizer Product Specification

## Document Status

- Status: Draft for implementation handoff
- Author: Codex
- Date: 2026-04-18
- Product area: Session browser, analytics, context engineering

## Executive Summary

The Context Session Visualizer is a session-level observability and decision
support feature for coding-agent workflows.

Its purpose is to help users understand how a live or historical coding-agent
session consumed its context window over time, how that context composition
affects session quality, and what session-management action the user should take
next.

The feature is built on the thesis that success in coding-agent workflows
depends not only on model quality and prompt quality, but on context quality.
Long sessions accumulate useful knowledge, but they also accumulate irrelevant
tool output, failed attempts, stale assumptions, debugging tangents, and
compressed summaries. As a session grows, the agent's effective intelligence
often degrades because the context becomes diluted, noisy, or misaligned with
the current objective.

The product should therefore do two jobs:

1. Show an x-ray of how session context was spent and how it evolved.
2. Recommend concrete context-shaping actions such as continue, rewind, fork,
   compact, or delegate to a subagent.

This feature should make context engineering visible, measurable, and
actionable.

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

## Product Vision

Build a context observability and context intervention layer for coding-agent
sessions.

The user should be able to inspect a session and immediately understand:

- how full the session context is,
- what is occupying that context,
- which turns added durable value versus ephemeral noise,
- where context quality likely degraded,
- what action will best preserve progress while restoring clarity.

The product should not merely report token usage. It should help the user make
better branching decisions at every meaningful turn of a session.

## Core Product Thesis

The Context Session Visualizer is founded on these principles:

1. Session success depends on context engineering.
2. Context size alone is not enough; context quality matters more.
3. As context grows, relevance concentration often decreases.
4. Every turn is a branching point where the user could continue, rewind, fork,
   compact, or delegate.
5. Tangents, retries, and long tool traces create hidden context drag.
6. Session-management operations should be guided by evidence, not intuition.
7. A separate analysis agent with a clean context window can often diagnose the
   primary session better than the primary session can diagnose itself.

## Product Goals

### Primary Goals

- Give users a detailed view of how session context grew over time.
- Attribute context growth to concrete sources such as prompts, file reads, tool
  results, summaries, and agent replies.
- Identify likely context-health problems such as tangents, retries, repeated
  reads, compaction loss, and stale branches.
- Recommend the best next session-management action.
- Help users preserve model intelligence by acting before context quality
  collapses.

### Secondary Goals

- Make session-management concepts legible to advanced users.
- Improve confidence in operations like rewind and compact.
- Create a foundation for automated coaching and future proactive assistance.
- Allow retrospective analysis of successful and failed sessions to improve user
  habits.

### Non-Goals

- Do not attempt to reconstruct exact provider-side prompt packing with perfect
  fidelity if the provider does not expose it.
- Do not replace the primary transcript viewer.
- Do not automatically mutate user sessions in v1.
- Do not act as a universal tokenizer for every provider at perfect parity.
- Do not attempt to solve general agent evaluation across all dimensions; focus
  specifically on context health and session branching quality.

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

The feature consists of two tightly coupled subsystems:

### 1. Context Session Visualizer

A visual, inspectable breakdown of session context history and current state.

### 2. Context Advisor

A secondary analysis layer, optionally powered by a separate agent, that
interprets the session and recommends actions.

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

## Product Requirements

## Functional Requirements

### FR1: Show current session context state

The product must show a current-state summary of the session, including:

- estimated context tokens in use,
- estimated percent of available context consumed,
- available remaining context,
- source breakdown by category,
- current session health status,
- count of significant context-shaping events.

### FR2: Show context growth over time

The product must display context growth turn by turn across the session.

This should include:

- per-turn context delta,
- cumulative context estimate,
- growth spikes,
- category attribution for each spike,
- markers for compaction, rewind, fork points, and subagent events when known.

### FR3: Attribute context to sources

The product must classify context consumption into meaningful categories. Initial
categories should include:

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

This taxonomy should be extensible and agent-specific mapping rules must be
allowed.

### FR4: Surface context-health signals

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
- divergence between recent task focus and older context mass,
- large branches later abandoned,
- high ratio of intermediate output to durable conclusions.

### FR5: Highlight branching points

The product must identify meaningful points in the session where an alternative
action would likely have been beneficial.

For each highlighted branch point, the UI should show:

- the message or turn index,
- the likely reason it matters,
- the recommended action,
- the estimated benefit,
- the confidence level.

### FR6: Recommend next best action

The product must generate a recommendation for what the user should do now.

At minimum, the system must support recommending:

- continue in current session,
- rewind to a specific point,
- compact with a specific focus,
- fork a new session with a handoff,
- start a fresh session with a distilled brief,
- use a subagent for the next work chunk.

### FR7: Provide operational guidance

Recommendations must be actionable. The product should provide:

- a concrete explanation,
- the exact command or control to use when the source agent supports it,
- suggested prompt text or handoff text,
- suggested compact focus instructions where relevant,
- suggested rewind reprompt,
- suggested fork brief,
- suggested subagent task wording.

### FR8: Support historical and current sessions

The feature should work for:

- historical sessions already indexed in the database,
- the currently active session when live updates are available.

### FR9: Support agent-specific interpretation

The system must support agent-specific logic for:

- context category mapping,
- compaction event recognition,
- rewind or fork semantics,
- subagent event detection,
- token estimation quality,
- recommended command text.

### FR10: Support advisor-agent analysis

The system should optionally invoke a separate analysis agent with a fresh
context window to analyze the primary session.

The side agent must:

- receive a structured summary of the primary session,
- diagnose context health,
- explain why the session is healthy or unhealthy,
- recommend a context-shaping operation,
- produce ready-to-use action text,
- avoid importing the entire raw session when a compressed representation is
  sufficient.

## User Experience Requirements

### UX1: Session-level entry point

The visualizer should be accessible from the session detail view as a dedicated
tab, panel, or route-level mode.

Candidate labels:

- Context
- Context Health
- Session X-Ray

### UX2: Immediate legibility

Within a few seconds, the user should be able to answer:

- How full is this session?
- What is consuming the context?
- Is this session still healthy?
- What should I do next?

### UX3: Drill-down capability

The user must be able to inspect:

- turn-level context changes,
- category-level consumption,
- a branch-point explanation,
- recommended action details.

### UX4: Action-oriented summary

The top of the view should include a compact decision card:

- Session health: healthy, watch, degraded, critical
- Primary issue: tangent, tool noise, retry loop, compaction risk, stale branch
- Recommended action
- Confidence
- CTA text or copyable command

### UX5: Avoid false authority

The system must distinguish:

- measured facts,
- heuristics,
- inferred judgments,
- advisor-agent opinions.

Confidence and uncertainty should be explicit.

## Proposed User Interface

## Primary Layout

### Section A: Context Summary Header

Displays:

- current estimated context usage,
- percent full,
- remaining budget,
- health status,
- most likely recommended action,
- timestamp of last update.

### Section B: Context Composition Breakdown

A stacked bar or treemap showing major context categories.

The user should be able to:

- hover or click a category,
- see token estimate and percentage,
- navigate to representative turns that contributed to it.

### Section C: Context Timeline

A chronological view of turns showing:

- per-turn additions,
- cumulative context,
- session-management events,
- tool-heavy sections,
- tangent windows,
- suspected degraded zones.

This is the core x-ray.

### Section D: Branch Points

A list of recommended or historically meaningful branch points:

- "Rewind recommended after turn 42"
- "Fork recommended after task pivot at turn 68"
- "Compact should have happened before turn 103"
- "Use subagent for verification branch at turn 119"

### Section E: Advisor Panel

Displays either rule-based analysis, agent-based analysis, or both:

- diagnosis,
- rationale,
- recommendation,
- suggested command text,
- suggested handoff text,
- confidence.

### Section F: Evidence Inspector

Shows why the system made the recommendation:

- turns involved,
- categories implicated,
- repeated tools,
- repeated reads,
- compaction boundaries,
- high-growth spikes.

## Example Decision Outputs

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
- Branch point
- Session-management event
- Context-health signal
- Recommendation

### Proposed Data Model

#### ContextSegment

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
- `is_ephemeral_candidate`
- `is_currently_relevant_estimate`
- `branch_group`
- `created_at`

#### ContextTurnDelta

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

#### ContextBranchPoint

Fields:

- `session_id`
- `turn_index`
- `branch_type`
- `reason_code`
- `reason_text`
- `recommended_action`
- `confidence`
- `action_payload`

#### ContextRecommendation

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

## Signal Framework

The system should compute a set of context-health signals. These do not need to
be perfect in v1, but they must be explainable.

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

#### 3. Drift Signals

- topic shift from original task,
- task pivot without session reset,
- low overlap between recent turns and early retained context.

#### 4. Retry Signals

- repeated attempts on same failing approach,
- repeated edits to same files without convergence,
- repeated tool failures,
- repeated reversions or churn.

#### 5. Compression Signals

- number of compactions,
- time since last compaction,
- compaction near high-noise state,
- likely summary loss after a pivot.

#### 6. Branching Signals

- identifiable tangent start,
- abandoned side branch,
- points where knowledge stabilized before experimentation resumed,
- points where subagent use would have contained disposable output.

### Health States

The system should classify overall session health into:

- Healthy
- Watch
- Degraded
- Critical

This label should derive from a score plus major trigger conditions.

### Example Heuristic Score Inputs

- occupancy percentile,
- recent growth acceleration,
- noise-to-conclusion ratio,
- retry density,
- drift intensity,
- compaction risk,
- stale branch weight,
- tool failure density.

## Recommendation Engine

The recommendation engine should combine:

- deterministic heuristics,
- interpretable rule-based logic,
- optional advisor-agent analysis.

### Recommendation Pipeline

1. Reconstruct or estimate context state.
2. Compute turn-level deltas and signal values.
3. Detect branching points.
4. Score candidate actions.
5. Produce a primary recommendation and alternatives.
6. Attach evidence and suggested commands.

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

## Advisor-Agent Specification

The Context Advisor is an optional higher-order agent that analyzes the primary
session from outside the session.

### Purpose

Use a fresh context window to reason about the health of the primary session
without being degraded by the primary session's own overloaded context.

### Inputs

The advisor should receive a structured representation rather than the full raw
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

### Outputs

The advisor must return:

- diagnosis,
- primary recommendation,
- alternative options,
- rationale,
- exact commands where agent-specific commands exist,
- suggested prompt or handoff text,
- confidence,
- cited evidence from the structured input.

### Advisor Safety Constraints

- Do not allow the advisor to issue destructive commands automatically.
- Clearly separate advisor opinion from measured telemetry.
- Do not imply certainty where evidence is weak.
- Provide fallback recommendations when confidence is low.

### Example Advisor Output

"Recommendation: rewind to turn 84. The session retains useful repo-reading
context up to turn 84, but turns 85-104 represent a failed tangent into
deployment warnings that do not support the active auth bug. Rewind and re-prompt
with: 'Ignore the deployment warning branch. Focus on auth middleware in
files X and Y. We ruled out approach A because constraint B.'"

## Agent-Specific Considerations

Different coding agents expose different observability surfaces. The system must
handle partial fidelity by agent.

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

The product should define:

- a normalized context taxonomy,
- a per-agent mapping layer,
- a confidence score for each context estimate.

## Data Requirements

The implementation should reuse existing `agentsview` session and message data
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

### New Derived Data Needed

- context timeline aggregates,
- context category attribution,
- turn-level context deltas,
- branch-point detections,
- recommendation objects,
- advisor outputs,
- optional cached token estimates.

## Storage Strategy

### Option A: Mostly computed on read

Advantages:

- fewer schema changes,
- easy iteration,
- no migration burden for early versions.

Disadvantages:

- more expensive reads,
- harder to compare repeated analysis runs.

### Option B: Persist derived context-analysis artifacts

Advantages:

- faster UI,
- easier diffing across runs,
- advisor output caching.

Disadvantages:

- more schema work,
- invalidation complexity.

### Recommendation

For v1:

- compute lightweight context metrics on read,
- persist only advisor outputs and expensive derived analyses if needed,
- add caching once UX and signal definitions stabilize.

## API Requirements

The backend should expose dedicated context-analysis endpoints.

### Proposed Endpoints

#### `GET /api/v1/sessions/{id}/context`

Returns current summary and context composition.

#### `GET /api/v1/sessions/{id}/context/timeline`

Returns per-turn or per-window context growth data.

#### `GET /api/v1/sessions/{id}/context/branch-points`

Returns detected branch points and suggested actions.

#### `GET /api/v1/sessions/{id}/context/recommendation`

Returns the primary recommendation and rationale.

#### `POST /api/v1/sessions/{id}/context/analyze`

Triggers fresh analysis, optionally including advisor-agent execution.

#### `GET /api/v1/sessions/{id}/context/advisor`

Returns latest advisor result if one exists.

### Response Requirements

Every response should clearly separate:

- measured values,
- estimated values,
- inferred values,
- model-generated advisory text.

## Frontend Requirements

### UI Components

Candidate component set:

- `ContextSummaryCard`
- `ContextCompositionChart`
- `ContextTimeline`
- `ContextBranchPointList`
- `ContextRecommendationCard`
- `ContextAdvisorPanel`
- `ContextEvidenceDrawer`

### Interaction Requirements

- Clicking a turn in the timeline should jump to the transcript.
- Clicking a branch point should highlight the implicated turns.
- Recommendation actions should provide copyable command and prompt text.
- The UI should make live updates obvious when the session is active.

### Filtering and Modes

The user should be able to switch between:

- current-state view,
- full timeline view,
- recent degradation view,
- branch-point view,
- advisor view.

## Algorithm and Heuristic Guidance

The system does not need perfect semantic understanding in v1, but it must be
useful and explainable.

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

### Flow 1: Live degraded session

1. User opens an active session.
2. Context view shows 78% occupancy and rapid recent growth.
3. Timeline highlights the last 24 turns as a tool-heavy debugging tangent.
4. Branch-point card says: "Rewind to turn 91."
5. Recommendation card provides rewind guidance and a reprompt.
6. User copies the suggested rewind text and acts in the primary tool.

### Flow 2: Compact before failure

1. User sees session at 63% occupancy.
2. Context remains coherent but a large implementation branch has accumulated.
3. Recommendation says compact proactively.
4. Product offers compact focus text.
5. User compacts before context rot worsens.

### Flow 3: Fork after task pivot

1. Session began as feature implementation.
2. User later switched to documentation and deployment.
3. Visualizer detects low overlap between current task and early session mass.
4. Recommendation says start a new session with a prepared handoff brief.

### Flow 4: Advisor review

1. User requests advisor analysis.
2. Secondary agent receives structured summary.
3. Advisor returns diagnosis, recommendation, and copyable action text.
4. UI displays advisor output alongside system heuristics.

## Success Metrics

### Product Success Metrics

- users open the context view during long sessions,
- users adopt suggested branch actions,
- users report higher confidence in session-management decisions,
- users recover from degraded sessions faster,
- fewer very-long sessions end in obvious retry loops or compaction failures.

### Behavioral Metrics

- rate of rewind/fork/compact actions after viewing the tool,
- reduction in average context occupancy at action time,
- reduced continuation of low-value tangents,
- increased use of subagents where appropriate.

### Quality Metrics

- recommendation acceptance rate,
- user-rated usefulness,
- branch-point precision,
- advisor agreement with user action,
- false-positive rate for unhealthy-session warnings.

## Constraints and Risks

### Constraint: Exact context fidelity may be unavailable

Many agents do not expose the exact provider-side packed context. The product
must work with best-effort reconstruction and clearly label estimates.

### Risk: Overconfident recommendations

Bad advice here can be costly. The system must remain transparent and cautious.

### Risk: High implementation complexity

Token attribution, branch detection, and multi-agent support can expand quickly.
Scope discipline is required.

### Risk: Advisor contamination

If the advisor receives too much raw transcript, it may become expensive and less
focused. Structured summaries should be preferred.

## Privacy and Security

- All analysis should remain local by default.
- Advisor-agent invocation should use the user's existing local agent tools when
  possible.
- No session data should be sent to remote services without explicit user
  action.
- Cached advisor outputs should be stored as local session artifacts.

## Rollout Plan

### Phase 1: Read-only visualizer

Deliver:

- context summary,
- composition breakdown,
- turn timeline,
- basic health signals,
- branch-point heuristics,
- static recommendation card.

### Phase 2: Operational guidance

Deliver:

- copyable command text,
- suggested rewind prompts,
- suggested compact focus text,
- fork handoff generation.

### Phase 3: Advisor-agent integration

Deliver:

- structured analysis bundle,
- external advisor run,
- advisor results panel,
- caching and re-run controls.

### Phase 4: Deeper automation

Potential future work:

- proactive warnings,
- auto-generated handoff drafts,
- compare two candidate branch strategies,
- recommend subagent boundaries before a task starts.

## Open Questions

- Which agents expose enough metadata for high-fidelity context reconstruction?
- Should advisor recommendations be persisted or always recomputed?
- How should "relevance" be estimated without adding heavy semantic models?
- How should the UI distinguish context size from context usefulness?
- Should the initial release be Claude-first, then generalized?
- How much of the recommendation engine should be deterministic versus
  agent-generated?

## Recommended v1 Scope

To keep the first version tractable, v1 should focus on:

- a single session detail experience,
- historical plus live analysis for agents with strong parse support,
- estimated context growth timeline,
- source-category breakdown,
- heuristic health classification,
- rewind/compact/fork/subagent recommendations,
- copyable action text,
- optional advisor execution behind a manual button.

Avoid for v1:

- fully automatic action execution,
- cross-session optimization,
- heavy semantic topic modeling,
- exact packed-context reconstruction for every agent,
- team-wide dashboards for context health.

## Implementation Handoff

## Suggested Workstreams

### Workstream 1: Context Data Layer

- define normalized context categories,
- implement token-estimation utilities,
- build turn-level context delta computation,
- expose derived analysis structs.

### Workstream 2: Recommendation Engine

- implement signal calculations,
- build rule-based action scorer,
- generate recommendation payloads,
- create evidence attachments.

### Workstream 3: Backend API

- add context-analysis endpoints,
- add caching or persistence where needed,
- support live refresh for active sessions.

### Workstream 4: Frontend Experience

- create context tab or route,
- build summary, timeline, branch-point, and recommendation components,
- integrate transcript jump links and copy actions.

### Workstream 5: Advisor Integration

- define advisor input schema,
- define advisor output schema,
- implement advisor execution path,
- render advisor result safely with confidence and evidence.

## Suggested Deliverables for the Implementing Agent

1. Data model and API design proposal
2. Backend context-analysis implementation
3. Frontend context visualizer UI
4. Recommendation engine
5. Advisor-agent integration
6. Tests for signal logic and recommendation output
7. Documentation for supported agents and accuracy caveats

## Acceptance Criteria

The feature is complete for v1 when:

- a user can open a session and see estimated context usage and composition,
- a user can inspect context growth turn by turn,
- a user can see at least one evidence-backed recommendation,
- the recommendation includes actionable text,
- the system distinguishes measured versus inferred values,
- the UI supports active-session refresh,
- at least one agent is supported well enough to make the feature genuinely
  useful.

## Final Product Statement

The Context Session Visualizer is a context x-ray and context coach for
coding-agent sessions.

It helps users see how context was spent, where it stopped helping, and what to
do next to preserve model performance. By combining context observability,
branch-point detection, and actionable session-management guidance, it turns
session health from intuition into an explicit, operable part of the coding
workflow.
