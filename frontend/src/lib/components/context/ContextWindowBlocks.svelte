<script lang="ts">
  import type {
    ContextCapacity,
    ContextCompositionItem,
    ContextSummary,
    ContextTimelineTurn,
  } from "../../api/types.js";
  import { CATEGORY_COLORS, categoryLabel } from "./context-utils.js";
  import ContextCompositionPiePane from "./ContextCompositionPiePane.svelte";
  import ContextTurnSharesPane from "./ContextTurnSharesPane.svelte";
  import ContextWindowMapPane from "./ContextWindowMapPane.svelte";

  interface Props {
    summary: ContextSummary;
    capacity: ContextCapacity;
    timeline: ContextTimelineTurn[];
    composition: ContextCompositionItem[];
  }

  type Segment = {
    key: string;
    label: string;
    tokens: number;
    blocks: number;
    color: string;
    turn?: number;
    isFree?: boolean;
    meta?: string;
  };

  type BlockCell = {
    key: string;
    color: string;
    label: string;
    turn?: number;
    isFree?: boolean;
  };

  let { summary, capacity, timeline, composition }: Props = $props();

  const TOTAL_BLOCKS = 200;
  const BLOCK_COLUMNS = 20;
  const TURN_COLORS = [
    "var(--accent-blue)",
    "var(--accent-teal)",
    "var(--accent-indigo)",
    "var(--accent-amber)",
    "var(--accent-purple)",
    "var(--accent-orange)",
    "var(--accent-sky)",
    "var(--accent-rose)",
  ];

  function turnColor(index: number): string {
    return TURN_COLORS[index % TURN_COLORS.length];
  }

  function turnLabel(turn: ContextTimelineTurn): string {
    if (turn.markers?.includes("compaction")) {
      return `Turn ${turn.turn} · Compaction seed`;
    }
    return `Turn ${turn.turn} · Messages ${turn.start_ordinal}-${turn.end_ordinal}`;
  }

  function jumpToTurn(turn: number | undefined) {
    if (!turn) return;
    const el = document.getElementById(`context-turn-${turn}`);
    if (!(el instanceof HTMLDetailsElement)) return;
    el.open = true;
    el.scrollIntoView({ behavior: "smooth", block: "start" });
  }

  function polarToCartesian(
    centerX: number,
    centerY: number,
    radius: number,
    angleInDegrees: number,
  ) {
    const angleInRadians = ((angleInDegrees - 90) * Math.PI) / 180.0;
    return {
      x: centerX + radius * Math.cos(angleInRadians),
      y: centerY + radius * Math.sin(angleInRadians),
    };
  }

  function describeArc(
    centerX: number,
    centerY: number,
    radius: number,
    startAngle: number,
    endAngle: number,
  ) {
    const start = polarToCartesian(centerX, centerY, radius, endAngle);
    const end = polarToCartesian(centerX, centerY, radius, startAngle);
    const largeArcFlag = endAngle-startAngle <= 180 ? "0" : "1";

    return [
      "M",
      centerX,
      centerY,
      "L",
      start.x,
      start.y,
      "A",
      radius,
      radius,
      0,
      largeArcFlag,
      0,
      end.x,
      end.y,
      "Z",
    ].join(" ");
  }

  function allocateBlocks(values: number[], totalBlocks: number): number[] {
    if (totalBlocks <= 0 || values.length === 0) {
      return values.map(() => 0);
    }
    const safeValues = values.map((value) => Math.max(0, value));
    const sum = safeValues.reduce((acc, value) => acc + value, 0);
    if (sum <= 0) {
      return safeValues.map(() => 0);
    }

    const raw = safeValues.map((value) => (value / sum) * totalBlocks);
    const base = raw.map((value) => Math.floor(value));
    let remaining = totalBlocks - base.reduce((acc, value) => acc + value, 0);

    const order = raw
      .map((value, index) => ({
        index,
        remainder: value - Math.floor(value),
        weight: safeValues[index],
      }))
      .sort((a, b) => {
        if (b.remainder !== a.remainder) return b.remainder - a.remainder;
        return b.weight - a.weight;
      });

    for (const item of order) {
      if (remaining <= 0) break;
      base[item.index] += 1;
      remaining -= 1;
    }

    return base;
  }

  const visibleTurns = $derived(
    timeline.filter((turn) => turn.delta_tokens > 0),
  );

  const segments = $derived.by<Segment[]>(() => {
    if (capacity.max_tokens <= 0) {
      return [];
    }

    const cappedTokensInUse = Math.max(
      0,
      Math.min(summary.tokens_in_use, capacity.max_tokens),
    );
    const rawTurnTokens = visibleTurns.map((turn) =>
      Math.max(0, Math.min(turn.delta_tokens, capacity.max_tokens)),
    );
    const rawTurnTotal = rawTurnTokens.reduce((acc, value) => acc + value, 0);
    const turnTokens = rawTurnTokens.map((value) => {
      if (rawTurnTotal <= 0 || cappedTokensInUse <= 0) {
        return value;
      }
      return (value / rawTurnTotal) * cappedTokensInUse;
    });

    const parts = [...turnTokens];
    if (parts.length === 0 && cappedTokensInUse > 0) {
      parts.push(cappedTokensInUse);
    }
    parts.push(Math.max(0, capacity.max_tokens - cappedTokensInUse));

    const blocks = allocateBlocks(parts, TOTAL_BLOCKS);
    const turnSegments =
      visibleTurns.length > 0
        ? visibleTurns.map((turn, index) => ({
            key: `turn-${turn.turn}`,
            label: turnLabel(turn),
            tokens: Math.round(turnTokens[index] ?? turn.delta_tokens),
            blocks: blocks[index] ?? 0,
            color: turnColor(index),
            turn: turn.turn,
          }))
        : cappedTokensInUse > 0
          ? [{
              key: "visible-context",
              label: "Visible context",
              tokens: cappedTokensInUse,
              blocks: blocks[0] ?? 0,
              color: turnColor(0),
              turn: visibleTurns[0]?.turn,
            }]
          : [];

    const freeBlockCount = blocks[blocks.length - 1] ?? 0;
    if (freeBlockCount > 0) {
      turnSegments.push({
        key: "free-space",
        label: "Free space",
        tokens: Math.max(0, capacity.max_tokens - cappedTokensInUse),
        blocks: freeBlockCount,
        color: "var(--bg-inset)",
        isFree: true,
      });
    }

    return turnSegments.filter((segment) => segment.blocks > 0);
  });

  const blocks = $derived.by<BlockCell[]>(() => {
    const cells: BlockCell[] = [];
    for (const segment of segments) {
      for (let i = 0; i < segment.blocks; i += 1) {
        cells.push({
          key: `${segment.key}-${i}`,
          color: segment.color,
          label: segment.label,
          turn: segment.turn,
          isFree: segment.isFree,
        });
      }
    }
    return cells;
  });

  const unknownCapacity = $derived(capacity.max_tokens <= 0);
  const gridRows = $derived(Math.ceil(TOTAL_BLOCKS / BLOCK_COLUMNS));
  const pieSegments = $derived.by(() => {
    if (capacity.max_tokens <= 0) return [];
    let startAngle = 0;
    return segments.map((segment) => {
      const sweep = (segment.tokens / capacity.max_tokens) * 360;
      const endAngle = startAngle + sweep;
      const slice = {
        ...segment,
        startAngle,
        endAngle,
        path: describeArc(72, 72, 56, startAngle, endAngle),
      };
      startAngle = endAngle;
      return slice;
    }).filter((segment) => segment.tokens > 0);
  });

  const compositionSegments = $derived.by(() => {
    const items = composition.filter((item) => item.tokens > 0);
    let startAngle = 0;
    return items.map((item) => {
      const sweep = (item.percentage / 100) * 360;
      const endAngle = startAngle + sweep;
      const slice = {
        ...item,
        label: categoryLabel(item.category),
        color: CATEGORY_COLORS[item.category] ?? CATEGORY_COLORS.other,
        path: describeArc(72, 72, 56, startAngle, endAngle),
      };
      startAngle = endAngle;
      return slice;
    });
  });
</script>

<section class="panel">
  <div class="panel-header">
    <div class="header-main">
      <div class="eyebrow">Context Window</div>
      <h3>Visible context by turn and category</h3>
    </div>
  </div>

  {#if unknownCapacity}
    <div class="empty-state">
      This session does not have a known model context window, so the block map cannot be rendered.
    </div>
  {:else}
    <div class="viz-layout">
      <ContextWindowMapPane
        {summary}
        {capacity}
        {blocks}
        totalBlocks={TOTAL_BLOCKS}
        blockColumns={BLOCK_COLUMNS}
        {gridRows}
        onJumpToTurn={jumpToTurn}
      />
      <ContextTurnSharesPane
        segments={pieSegments}
        usedPercent={summary.percent_consumed}
        onJumpToTurn={jumpToTurn}
      />
      <ContextCompositionPiePane
        tokensInUse={summary.tokens_in_use}
        segments={compositionSegments}
      />
    </div>
  {/if}
</section>

<style>
  .panel {
    border: 1px solid var(--border-muted);
    background: var(--bg-surface);
    border-radius: var(--radius-md);
    padding: 12px;
    display: grid;
    gap: 12px;
  }

  .panel-header {
    display: grid;
    gap: 6px;
  }

  .header-main {
    display: grid;
    gap: 4px;
  }

  .eyebrow {
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--text-muted);
    margin-bottom: 4px;
  }

  h3 {
    margin: 0;
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .viz-layout {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 16px;
    align-items: start;
  }

  .empty-state {
    border: 1px dashed var(--border-muted);
    background: var(--bg-inset);
    border-radius: var(--radius-sm);
    padding: 16px;
    color: var(--text-secondary);
    font-size: 12px;
    line-height: 1.4;
  }

  @media (max-width: 900px) {
    .panel-header {
      gap: 6px;
    }

    .viz-layout {
      grid-template-columns: 1fr;
    }
  }
</style>
