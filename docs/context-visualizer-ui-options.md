# Context Visualizer UI Layout Options

## Document Status

- Status: Design exploration
- Date: 2026-04-18
- Depends on:
  [`context-session-visualizer-spec.md`](./context-session-visualizer-spec.md),
  [`context-session-visualizer-mvp-plan.md`](./context-session-visualizer-mvp-plan.md)

## Purpose

This document presents four candidate UI layouts for the Context Session
Visualizer MVP. Each option shows how context summary, composition, timeline,
and turn-boundary markers can be arranged within the existing session detail
experience.

All options assume:

- The visualizer lives inside a session-level tab toggle (Transcript / Context).
- The center column of the existing three-column layout hosts the view.
- Turn boundaries are first-class visual elements.
- Compaction events, subagent spawns, and growth spikes are marked inline.
- Data is estimate-driven and labeled accordingly.

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

Health states use the existing grade color vocabulary:

| State    | Meaning                                         |
|----------|-------------------------------------------------|
| HEALTHY  | Low occupancy, modest growth                    |
| WATCH    | Moderate occupancy or recent spike              |
| DEGRADED | High occupancy, repeated spikes, tool dominance |
| CRITICAL | Near-threshold occupancy with sustained noise   |

---

## Option A: Full-Page Vertical Stack

### Concept

Three stacked sections fill the center column top to bottom: summary card,
composition bar, and a continuous area-style timeline chart. The timeline uses
turn indices on the x-axis and cumulative tokens on the y-axis.

### Strengths

- Simple top-to-bottom reading order.
- The area chart gives an immediate gestalt of session growth shape.
- Composition bar provides a quick "what is eating context" answer.
- Familiar dashboard layout; low learning curve.

### Weaknesses

- Turn boundaries are implicit (x-axis tick marks) rather than explicit rows.
- Hard to show per-turn event detail without a tooltip or hover layer.
- Long sessions compress the x-axis, making individual turns hard to inspect.
- No built-in place for turn-level annotations (spike reasons, tool names).

### Layout

```
┌──────────────────────────────────────────────────────────────────────┐
│  Session: "Refactor auth middleware"       [Transcript] [Context]   │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─ Context Summary ──────────────────────────────────────────────┐  │
│  │  127.4k / 200.0k tokens (63%)           Health: WATCH         │  │
│  │  ████████████████████████████████░░░░░░░░░░░░░░░░░░            │  │
│  │  72.6k remaining    Est. confidence: high                     │  │
│  │  Reasons: Moderate occupancy, 2 recent growth spikes          │  │
│  └────────────────────────────────────────────────────────────────┘  │
│                                                                      │
│  ┌─ Context Composition ──────────────────────────────────────────┐  │
│  │  ┌────────────┬──────────┬────────┬──────┬─────┬──────┬─────┐ │  │
│  │  │ tool_result│ assistant│file_read│ user │think│t_call│other│ │  │
│  │  │   38.2%    │  22.1%   │ 18.4%  │ 9.7% │5.3% │ 4.1% │2.2%│ │  │
│  │  └────────────┴──────────┴────────┴──────┴─────┴──────┴─────┘ │  │
│  │                                                                │  │
│  │  tool_result ........ 48.7k tokens                            │  │
│  │  assistant .......... 28.2k tokens                            │  │
│  │  file_read .......... 23.4k tokens                            │  │
│  │  user ............... 12.4k tokens                            │  │
│  │  thinking ...........  6.8k tokens                            │  │
│  │  tool_call ..........  5.2k tokens                            │  │
│  │  other ..............  2.7k tokens                            │  │
│  └────────────────────────────────────────────────────────────────┘  │
│                                                                      │
│  ┌─ Context Timeline ────────────────────────────────────────────┐  │
│  │                                                                │  │
│  │  tokens                                                        │  │
│  │  127k ┤                                            ┌── 127.4k │  │
│  │       │                                       ╭────╯          │  │
│  │  100k ┤                                  ╭────╯               │  │
│  │       │                             ╭────╯                    │  │
│  │   75k ┤                        ╭────╯                         │  │
│  │       │                   ╭────╯  ▲ spike: +14k tool_result   │  │
│  │   50k ┤             ╭────╯                                    │  │
│  │       │        ╭────╯                                         │  │
│  │   25k ┤   ╭────╯                                              │  │
│  │       │╭──╯                                                    │  │
│  │    0k ┤╯                                                       │  │
│  │       ├──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬──┬── │  │
│  │        1    5   10   15   20   25   30   35   40   45   50     │  │
│  │                          turn index                            │  │
│  │                                                                │  │
│  │  markers:  ↻ = compaction   ● = subagent   ▲ = spike          │  │
│  └────────────────────────────────────────────────────────────────┘  │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

### When to choose

Best when the primary goal is a quick visual summary and the user does not need
to inspect individual turns. Works well for short-to-medium sessions where the
x-axis remains readable.

---

## Option B: Turn-by-Turn Table with Event Markers

### Concept

Every turn is a discrete row in a scrollable table. Each row shows the turn
index, token delta, cumulative total, dominant category, and any event markers.
Turn boundaries are explicit dashed lines. Compaction boundaries are heavy
double-line dividers that visually reset the flow. A horizontal bar on each row
gives a proportional sense of delta size.

### Strengths

- Turn boundaries are explicit and impossible to miss.
- Events (spikes, compactions, subagent spawns) are inline with the turn that
  caused them.
- Scales well for long sessions via scrolling.
- No charting library needed; pure table + CSS bars.
- Easy to add click-to-jump-to-transcript behavior per row.

### Weaknesses

- No gestalt view of overall growth shape; the user must scroll to build a
  mental picture.
- Repetitive for sessions with many low-delta turns.
- The horizontal bar is a rough proxy; lacks the visual richness of an area
  chart.

### Layout

```
┌─ Context Timeline (turn-by-turn) ──────────────────────────────────────┐
│                                                                         │
│  Turn  Delta    Cumul.   Category         Event           Bar           │
│  ────  ─────    ──────   ────────         ─────           ───           │
│                                                                         │
│    1    1.2k     1.2k    user                             ██            │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│    2    3.8k     5.0k    assistant                        █████         │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│    3    0.4k     5.4k    user                             █             │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│    4    8.2k    13.6k    file_read                        ██████████    │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│    5    2.1k    15.7k    assistant                        ███           │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│    6    0.3k    16.0k    user                             █             │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│    7   14.7k    30.7k    tool_result      ▲ SPIKE        ██████████████│
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│    8    1.8k    32.5k    assistant                        ██            │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│    9    0.5k    33.0k    user                             █             │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│   10   12.3k    45.3k    tool_result      ▲ SPIKE        ██████████████│
│                                                                         │
│   ═══════════════════════════ ↻ COMPACTION ════════════════════════════ │
│                                                                         │
│   11    4.2k     4.2k    summary                         ██████        │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│   12    0.6k     4.8k    user                             █             │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│   13    6.1k    10.9k    assistant        ● subagent     ████████      │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│   14    0.4k    11.3k    user                             █             │
│  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─  │
│   15    3.2k    14.5k    assistant                        ████          │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### When to choose

Best when turn-level detail matters more than overall shape. Good default for
power users who want to identify exactly which turn caused a problem. Works
well as the primary MVP view because it is simple to build and directly
supports click-to-jump-to-transcript.

---

## Option C: Compact Vertical Timeline with Stacked Category Bars

### Concept

A hybrid of Option A and Option B. Each turn is a row (like B), but the bar
for each turn is a stacked horizontal bar showing the category breakdown within
that turn (like A's composition chart, but per-turn). Annotations appear as
indented sub-rows beneath the turn they belong to. Compaction boundaries are
heavy dividers that visually reset the cumulative counter.

### Strengths

- Turn boundaries are first-class; each row is a turn.
- Per-turn category breakdown is visible without clicking or hovering.
- Events (spikes, subagent spawns) appear inline with context annotations.
- Compaction resets are visually dramatic and easy to spot while scrolling.
- The stacked bar answers "what happened in this turn" at a glance.
- No charting library needed; CSS segments within a flex row.

### Weaknesses

- Slightly more complex to render than Option B's single-color bars.
- Annotation sub-rows add vertical height; very long sessions need more
  scrolling.
- Category colors in small stacked segments can be hard to distinguish at
  narrow widths.

### Layout

```
┌─ Context Timeline ───────────────────────────────────────────────────┐
│                                                                       │
│  63% full   127.4k / 200.0k         WATCH   2 spikes, 1 compaction  │
│  ████████████████████████████████░░░░░░░░░░░░░░░░░░░░                │
│                                                                       │
│  Turn │ Delta  │ Cumul. │ Composition (stacked)                      │
│  ─────┼────────┼────────┼────────────────────────────────────────── │
│     1 │  +1.2k │   1.2k │ ▓▓                               user    │
│     2 │  +4.1k │   5.3k │ ░░░░▓▓▓                         asst+thk │
│     3 │  +0.3k │   5.6k │ ▓                                user    │
│     4 │  +9.8k │  15.4k │ ████████░░                       file+tc │
│     5 │  +3.2k │  18.6k │ ░░░▓                             asst+thk│
│     6 │  +0.5k │  19.1k │ ▓                                user    │
│     7 │ +14.7k │  33.8k │ ████████████████▲                tool_res│
│       │        │        │ ▲ Grep returned 847 matches               │
│     8 │  +2.4k │  36.2k │ ░░░                              asst    │
│     9 │  +0.4k │  36.6k │ ▓                                user    │
│    10 │ +12.3k │  48.9k │ ███████████████▲                 tool_res│
│       │        │        │ ▲ Read 6 files, 3 large                   │
│       │        │        │                                            │
│  ─────┴────────┴────────┴── ↻ COMPACTED ── cumulative resets ───── │
│       │        │        │                                            │
│    11 │  +4.2k │   4.2k │ ░░░░░                            summary │
│    12 │  +0.6k │   4.8k │ ▓                                user    │
│    13 │  +7.3k │  12.1k │ ░░░░████●                        asst+sub│
│       │        │        │ ● Spawned subagent: "verify tests"        │
│    14 │  +0.4k │  12.5k │ ▓                                user    │
│    15 │  +3.2k │  15.7k │ ░░░░                             asst    │
│                                                                       │
│  Legend: ▓ user  ░ assistant/thinking  █ tool/file  ● subagent       │
│          ▲ spike (>2.5x median)       ↻ compaction boundary          │
│                                                                       │
└───────────────────────────────────────────────────────────────────────┘
```

### When to choose

Best all-around MVP option. Combines the turn-level detail of Option B with
category-level insight that usually requires a separate composition chart.
Particularly strong for sessions where the user wants to see both the growth
pattern and the composition pattern in a single scrollable view.

---

## Option D: Dual-Panel Chart + Turn Detail

### Concept

The center column is split horizontally. The left panel shows an area chart of
cumulative context growth with inline markers. The right panel shows a
scrollable turn-by-turn detail list. Clicking a point on the chart highlights
and scrolls to the corresponding turn in the detail list. Clicking a turn in
the detail list highlights the corresponding point on the chart.

### Strengths

- Best of both worlds: gestalt shape (left) and turn-level detail (right).
- Cross-panel interaction (click chart point to jump to turn) creates a
  powerful inspection tool.
- The chart provides instant "is this session healthy" signal.
- The turn list provides drill-down capability.
- Compaction boundaries appear in both panels, reinforcing the event.

### Weaknesses

- More complex to build; requires a chart library or custom SVG plus
  synchronized scroll/highlight state.
- Horizontal split reduces the width available for each panel.
- May feel cramped in narrow viewports or when the sidebar is expanded.
- Higher implementation cost than Options B or C for MVP.

### Layout

```
┌─ Context X-Ray ──────────────────────────────────────────────────────┐
│                                                                      │
│  127.4k / 200.0k (63%)    Health: WATCH    ↻ 1 compaction            │
│                                                                      │
│  ┌─ Growth Chart ─────────────┐  ┌─ Turn Detail ──────────────────┐  │
│  │                            │  │                                │  │
│  │ 50k┤          ▲             │  │  T7   +14.7k  tool_res ▲ SPIKE│  │
│  │    │         / \            │  │  ── Grep "handleAuth" 847 hits │  │
│  │    │        /   \  ▲        │  │                                 │  │
│  │ 40k┤       /     \/  \      │  │  T8   +2.4k   assistant       │  │
│  │    │      /            \    │  │  ── "Based on the grep..."    │  │
│  │ 30k┤    /               \   │  │                                 │  │
│  │    │   /                 \  │  │  T9   +0.4k   user            │  │
│  │ 20k┤  /         ↻        \ │  │  ── "Focus on auth.go only"   │  │
│  │    │ /        compacted    \│  │                                 │  │
│  │ 10k┤/                      │  │  T10  +12.3k  tool_res ▲ SPIKE│  │
│  │    │                        │  │  ── Read 6 files (auth: 4.8k) │  │
│  │  0k┤───────────────────────│  │                                 │  │
│  │     1  5  10 15 20 25 30   │  │  ═══ ↻ COMPACTION ═══════════  │  │
│  │                             │  │                                 │  │
│  │  [click point to jump -->] │  │  T11  +4.2k   summary          │  │
│  │                             │  │  ── Compacted to 4.2k          │  │
│  │                             │  │                                 │  │
│  │                             │  │  T12  +0.6k   user            │  │
│  │                             │  │  ── "Now implement the fix"   │  │
│  └─────────────────────────────┘  └─────────────────────────────────┘  │
│                                                                       │
└───────────────────────────────────────────────────────────────────────┘
```

### When to choose

Best for v1 or later when the MVP is proven and users want richer interaction.
The dual-panel approach is the natural evolution path: start with Option C for
MVP, then promote the chart into a persistent left panel in v1 while the turn
list moves to the right.

---

## Comparison Matrix

| Criterion                     | A: Full Stack | B: Turn Table | C: Stacked Bars | D: Dual Panel |
|-------------------------------|---------------|---------------|-----------------|---------------|
| Turn boundaries visible       | Implicit      | Explicit      | Explicit        | Explicit      |
| Per-turn category breakdown   | No            | No            | Yes             | No (right)    |
| Overall growth shape          | Yes (chart)   | No            | Partial (bars)  | Yes (chart)   |
| Event markers inline          | Tooltip only  | Yes           | Yes             | Yes           |
| Compaction boundary prominent | Moderate      | High          | High            | High          |
| Long session scalability      | Degrades      | Scrolls well  | Scrolls well    | Scrolls well  |
| Implementation complexity     | Medium        | Low           | Low-Medium      | High          |
| Charting library required     | Yes           | No            | No              | Yes           |
| Click-to-transcript ready     | Hard          | Easy          | Easy            | Easy          |
| MVP suitability               | Moderate      | High          | High            | Low           |

## Recommendation

### MVP: Option C (Compact Vertical Timeline with Stacked Category Bars)

Option C is the strongest MVP choice because:

1. Turn boundaries are first-class, which is the primary design requirement.
2. Per-turn stacked bars eliminate the need for a separate composition chart
   component, reducing the number of components to build.
3. Compaction boundaries use a heavy divider with cumulative reset, making them
   visually unmistakable while scrolling.
4. Event annotations (spike reasons, subagent names) are inline sub-rows rather
   than tooltips, so they work without hover interaction.
5. The summary header at the top provides the gestalt health check without
   needing a separate chart.
6. No charting library is required. The bars are CSS flex segments with
   percentage widths.
7. The layout mirrors the mental model from the Anthropic session management
   guide: every turn is a branching point.
8. Click-to-jump-to-transcript is straightforward since each row maps to a turn
   index.

### v1: Evolve toward Option D

Once the MVP is validated, add a persistent growth chart as a left panel
(Option D) while keeping the turn list on the right. This gives users both
the gestalt shape and the turn-level detail, with cross-panel interaction
for navigation.

### What to skip

Option A is too shallow for power users; the chart alone does not answer
turn-level questions. Option B is viable but strictly weaker than Option C,
which adds per-turn category breakdown at minimal extra cost.
