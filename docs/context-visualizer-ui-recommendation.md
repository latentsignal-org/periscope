# Context Visualizer UI Recommendation

## Document Status

- Status: Recommended MVP UI direction
- Date: 2026-04-18
- Depends on:
  [`context-session-visualizer-spec.md`](./context-session-visualizer-spec.md),
  [`context-session-visualizer-roadmap.md`](./context-session-visualizer-roadmap.md),
  [`context-session-visualizer-mvp-plan.md`](./context-session-visualizer-mvp-plan.md),
  [`v1-ui-spec.md`](./v1-ui-spec.md)

## Summary

The best first-shipping UI for the Context Session Visualizer is not the full
stacked-bar Option C. It is a simpler **Option B+**:

- summary card at the top,
- composition chart beneath it,
- explicit scrollable message/turn rows beneath that,
- inline event markers for compaction, spikes, and subagents,
- click-to-transcript as a first-class behavior.

This recommendation is intentionally conservative. The first version of this
feature should optimize for explanation, trust, and durability under imperfect
data. It should not optimize for maximum visual density.

## Recommendation

## Recommended MVP Layout: Option B+

Option B+ is a refinement of the "Turn-by-Turn Table with Event Markers" idea.
It keeps the core strengths of Option B:

- explicit boundaries,
- strong transcript linkage,
- simple rendering model,
- good long-session behavior,
- low implementation risk.

It adds only two things from later options:

- a compact summary card above the list,
- a lightweight composition chart above the list.

It does **not** adopt per-row stacked category bars as the default MVP
representation.

## Why this is the right MVP

### 1. The feature’s first job is explanation

Users need to answer:

- where did context grow,
- which message or turn caused it,
- where did compaction happen,
- where did a spike happen,
- what kind of thing dominated that row.

Option B+ answers those directly.

### 2. It is robust to imperfect categorization

Early context attribution will be approximate, especially across multiple
agents. A stacked-per-row composition UI makes that approximation look more
precise than it really is.

Option B+ is safer because it only requires:

- one dominant category,
- one delta size,
- one cumulative total,
- event markers.

That is easier to trust and easier to explain.

### 3. It maps cleanly to the current data model

The current implementation is still closer to **per-message** granularity than
true cross-agent **per-turn** granularity. Option B+ tolerates that well.

The UI can explicitly say:

- “Context growth per message”

without creating a mismatch between presentation and actual reconstructed data.

### 4. It scales better for long sessions

Long sessions are exactly where this feature matters most. Dense stacked bars
and annotation-heavy rows can become visually noisy very quickly.

Option B+ keeps the scan pattern simple:

- one row,
- one bar,
- one dominant label,
- one or more event markers.

### 5. It keeps the transcript connection strong

The killer interaction for this feature is:

- click the suspicious row,
- jump to the transcript,
- inspect what happened there.

Option B+ makes that interaction central instead of secondary.

## Why not Option C for MVP

Option C is the strongest medium-term direction, but it is slightly too
ambitious for the first release.

### Problems with Option C too early

- Per-row stacked bars imply a level of categorization fidelity that the system
  may not yet deserve.
- Row density increases fast once you add:
  - stacked segments,
  - spike markers,
  - compaction markers,
  - subagent markers,
  - inline annotations.
- The resulting UI can feel “busy” before the underlying heuristics are mature.
- It is harder to identify the primary signal because every row tries to say too
  much.

Option C should come after the team has confidence in:

- category attribution quality,
- compaction semantics,
- row interaction patterns,
- user demand for richer per-row breakdowns.

## Why not Option A

Option A is too shallow for the actual user need.

It shows:

- how big the session is,
- the overall shape of growth,
- aggregate composition.

But it hides:

- exactly where bad growth happened,
- exactly where compaction happened,
- exactly which row deserves inspection.

That makes it useful as a dashboard, but weak as a context-engineering tool.

## Why not Option D for MVP

Option D is a good v1 or v2 direction, not a good MVP.

It adds:

- synchronized chart/detail interaction,
- more layout complexity,
- more narrow-width stress,
- more state management complexity,
- higher implementation and maintenance cost.

The product should earn that complexity only after proving that users return to
this feature often enough to justify it.

## Recommended MVP Layout Structure

## Section 1: Summary Header

Purpose:

- communicate session health immediately,
- show size and occupancy,
- make uncertainty explicit.

Content:

- estimated context usage,
- percent full if available,
- remaining budget if available,
- health badge,
- short reasons,
- small explanatory footnote for percent-full semantics.

This should be a compact card, not a dashboard grid.

## Section 2: Composition Overview

Purpose:

- answer “what is filling this session?”

Content:

- one stacked horizontal composition bar,
- legend/list of category totals,
- token count and percentage per category.

This belongs above the timeline, not embedded into every row.

## Section 3: Context Timeline List

Purpose:

- answer “where did growth happen?”

Representation:

- one row per message or turn boundary,
- delta bar,
- cumulative total,
- dominant category,
- event markers.

Required fields per row:

- ordinal or index,
- delta tokens,
- cumulative total,
- dominant category,
- `▲` spike marker,
- `↻` compaction marker when relevant,
- `●` subagent marker when relevant.

Optional secondary text:

- timestamp,
- short event explanation,
- short note like “tool-heavy” or “read-heavy”.

## Section 4: Transcript Jump Behavior

Purpose:

- convert the visualizer into a navigational tool.

Behavior:

- clicking a row jumps to the corresponding transcript area,
- clicking a compaction row jumps to the compaction boundary,
- clicking a spike row jumps to the row that created the spike.

This is more important than adding more visual density.

## Recommended Visual Hierarchy

The row should visually prioritize:

1. delta magnitude,
2. cumulative total,
3. event marker,
4. dominant category,
5. secondary note.

The UI should avoid making all categories equally loud. The user primarily wants
to find the suspicious row, not parse a miniature chart on every line.

## Recommended Row Design

Each row should include:

- left: ordinal/index,
- middle: horizontal delta bar,
- right: numeric delta and cumulative total,
- tag area: dominant category and event markers.

Example:

```text
#42   ████████████           +14.7k   127.4k total   tool_result   ▲
```

Compaction rows should appear as explicit dividers, not tiny badges tucked into
normal rows.

Example:

```text
────────────── ↻ Context Compacted ──────────────
```

This is clearer than embedding a small “compact” pill in the row itself.

## Copy Recommendation

The UI should be explicit about what it is showing.

Use:

- “Context growth per message” if the backend is still message-based
- “Context growth per turn” only when the system truly aggregates at turn level
- “Estimated session context usage”
- “Percent full based on agent-reported peak context when available”

Avoid:

- “Exact context”
- “Turn” when the row is actually a message
- overly authoritative labels that hide estimate quality

## MVP Interaction Rules

## Required

- switch between Transcript and Context
- view summary
- view composition
- scroll timeline
- click a row to jump to transcript

## Nice to have

- hover tooltip for delta details
- highlight selected row after transcript jump
- small filter for “show spikes only” or “show compactions only”

## Not for MVP

- synchronized chart + list
- inline per-row stacked category bars
- nested annotations under every row
- rich evidence drawers

## Suggested Component Structure

For the current codebase, the recommended component structure is:

- `ContextSummaryCard`
- `ContextCompositionChart`
- `ContextTimeline`
- `ContextTimelineRow`
- `ContextCompactDivider`

If desired, `ContextTimelineRow` and `ContextCompactDivider` can remain internal
sub-components until the row design stabilizes.

## Evolution Path

## MVP

Use Option B+:

- summary card,
- composition chart,
- row list,
- event markers,
- transcript jump.

## v1

Upgrade toward Option C selectively:

- allow richer inline row annotations,
- optionally show per-row secondary breakdown on demand,
- add branch-point highlighting,
- add recommendation card.

The key is that per-row richness should be progressive, not mandatory.

## v2

Evolve toward Option D:

- persistent growth chart,
- chart/list synchronization,
- richer comparative inspection,
- advisor overlays.

## Decision

The recommended UI direction is:

- **MVP:** Option B+
- **v1:** B+ with selective Option C enhancements
- **v2:** Option D if usage validates the added complexity

This gives the feature the best chance to be:

- understandable,
- trustworthy,
- implementable,
- and useful under real session conditions.

## Final Recommendation

Do not ship the first version as the full stacked-row Option C.

Ship the first version as a cleaner, more conservative Option B+ that emphasizes:

- explicit boundaries,
- event visibility,
- transcript linkage,
- and trustworthy interpretation.

Then earn the right to add richer per-row composition once the data quality and
user demand justify it.
