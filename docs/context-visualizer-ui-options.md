# Context Visualizer UI Layout Options

## Document Status

- Status: Design exploration updated for current spec
- Date: 2026-04-18
- Depends on: [`periscope-spec.md`](./periscope-spec.md)

## Purpose

This document evaluates candidate UI layouts for the Periscope V1 Context
Visualizer. It updates the earlier design exploration to match the current
product spec, especially:

- V1 is descriptive only,
- the visualizer is available at `/context/:sessionId`,
- V1 requires a distinct summary header, composition breakdown, and timeline,
- occupancy percentages depend on a session-specific max context window,
- post-compaction history only is shown,
- live sessions update via SSE.

## Shared Assumptions

All options assume:

- the view can stand alone and can also be embedded later in session detail,
- turn boundaries are first-class visual elements,
- compaction events, subagent spawns, and growth spikes are marked inline,
- data is estimate-driven and labeled accordingly,
- the UI must expose whether token counts and max-window values are measured,
  inferred, estimated, or unknown,
- V1 does not show health labels, recommendation copy, or guidance controls.

## Shared Visual Language

The following markers are consistent across all options:

| Marker | Meaning                                   |
|--------|-------------------------------------------|
| `▲`    | Growth spike (delta >= 2.5x median delta) |
| `↻`    | Compaction boundary                       |
| `●`    | Subagent spawn                            |
| `▓`    | User content                              |
| `░`    | Assistant / thinking content              |
| `█`    | Tool output / file read content           |

## Option A: Full-Page Vertical Stack

### Concept

Three stacked sections fill the page: summary card, composition bar, and a
continuous area-style timeline chart. The timeline uses turn indices on the
x-axis and cumulative tokens on the y-axis.

### Strengths

- Simple top-to-bottom reading order.
- The area chart gives an immediate gestalt of session growth shape.
- Composition bar provides a quick "what is eating context" answer.
- Aligns naturally with the spec's three-section structure.

### Weaknesses

- Turn boundaries are implicit rather than explicit.
- Hard to show per-turn event detail without hover or tooltips.
- Long sessions compress the x-axis and lose inspectability.
- Weak fit for click-to-transcript behavior.

### Fit with Current Spec

Reasonable as a page shell, weak as the primary timeline treatment.

## Option B: Turn-by-Turn Table with Event Markers

### Concept

Every turn is a discrete row in a scrollable table. Each row shows the turn
index, token delta, cumulative total, dominant category, and any event markers.

### Strengths

- Turn boundaries are explicit and impossible to miss.
- Events are inline with the turn that caused them.
- Scales well for long sessions via scrolling.
- Easy to implement and easy to connect to transcript navigation.

### Weaknesses

- No strong gestalt view of overall growth shape.
- Per-turn category composition is missing unless extra UI is added.
- Feels repetitive for sessions with many low-delta turns.

### Fit with Current Spec

Strong timeline baseline, but incomplete unless paired with a separate
composition chart and a stronger per-turn composition treatment.

## Option C: Compact Vertical Timeline with Stacked Category Bars

### Concept

Each turn is a row, and the main row bar is a stacked horizontal bar showing
that turn's category mix. Annotation sub-rows call out spikes, compactions, or
subagent events.

### Strengths

- Turn boundaries are first-class.
- Per-turn category breakdown is visible without hover.
- Events appear inline with context annotations.
- Compaction resets are visually dramatic and easy to spot.
- No charting library is required.
- Strong match for click-to-transcript behavior.

### Weaknesses

- Slightly more complex than a plain table.
- Long sessions still require substantial scrolling.
- Small stacked segments can get visually tight in narrow widths.

### Fit with Current Spec

Best timeline design for V1. It complements, but does not replace, the
spec-required current-state composition chart.

## Option D: Dual-Panel Chart + Turn Detail

### Concept

Split the page into a growth chart panel and a synchronized turn-detail panel.
Chart clicks and row clicks stay linked.

### Strengths

- Best combined view of overall shape and per-turn detail.
- Strong inspection workflow once implemented.
- Natural evolution path after the first release.

### Weaknesses

- Highest implementation complexity.
- More fragile in narrow layouts and IDE webviews.
- Requires charting work plus synchronized interaction state.

### Fit with Current Spec

Attractive, but too expensive for the first V1 ship given the backend work
needed for cross-agent normalization, capacity inference, and SSE updates.

## Comparison Matrix

| Criterion                             | A: Full Stack | B: Turn Table | C: Stacked Bars | D: Dual Panel |
|---------------------------------------|---------------|---------------|-----------------|---------------|
| Matches spec page structure           | High          | Medium        | Medium          | Medium        |
| Turn boundaries visible               | Implicit      | Explicit      | Explicit        | Explicit      |
| Per-turn category breakdown           | No            | No            | Yes             | Partial       |
| Overall growth shape                  | Yes           | No            | Partial         | Yes           |
| Event markers inline                  | Weak          | Yes           | Yes             | Yes           |
| Works well in narrow webviews         | Medium        | High          | High            | Low-Medium    |
| Click-to-transcript ready             | Hard          | Easy          | Easy            | Easy          |
| Implementation complexity             | Medium        | Low           | Low-Medium      | High          |
| Requires chart library                | Yes           | No            | No              | Yes           |
| Best fit for first V1 ship            | No            | Maybe         | Yes             | No            |

## Evaluation Against Latest Spec

The current spec changes the recommendation in a few important ways:

1. V1 needs a standalone page, not just a tab in a three-column shell.
1. V1 requires separate summary and composition sections, so the timeline should
   not try to do all the work itself.
1. The summary must surface max-context provenance and uncertainty, which fits
   naturally in a dedicated summary card.
1. V1 is descriptive only, so visual language that implies diagnosis should be
   avoided.
1. Post-compaction-only rendering simplifies the timeline and makes strong
   compaction dividers more useful.

## Recommendation

### Recommended V1 Layout: Spec Stack + Option C Timeline

The best V1 UI is a combination rather than one option verbatim:

1. `ContextSummaryCard`
1. `ContextCompositionChart`
1. `ContextTimeline` using Option C's stacked-bar turn rows

This is the right choice because:

1. It satisfies the updated spec directly.
1. It keeps the page legible in standard browser widths and JetBrains webviews.
1. It provides turn-level inspectability without chart-library complexity.
1. It preserves a clear path to a richer linked chart later.

### Evolution Path

After the first V1 ship, the best enhancement path is Option D:

- keep the Option C turn timeline,
- add a linked cumulative-growth chart,
- synchronize selection between chart and turn rows.

### What to Skip for First V1

- Do not ship Option A as the main timeline.
- Do not rely on Option B alone when Option C adds materially better
  per-turn attribution.
- Do not take on Option D's synchronized chart complexity in the first release.
