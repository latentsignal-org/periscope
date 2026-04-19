<script lang="ts">
  import type {
    ContextCapacity,
    ContextSummary,
    ContextTimelineTurn,
  } from "../../api/types.js";
  import { formatTokenCount } from "../../utils/format.js";

  interface Props {
    summary: ContextSummary;
    capacity: ContextCapacity;
    timeline: ContextTimelineTurn[];
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

  let { summary, capacity, timeline }: Props = $props();

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

  function turnShare(turnTokens: number): number {
    if (capacity.max_tokens <= 0) return 0;
    return (turnTokens / capacity.max_tokens) * 100;
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
            meta: `${Math.round(turnShare(turnTokens[index] ?? turn.delta_tokens))}% of window`,
          }))
        : cappedTokensInUse > 0
          ? [{
              key: "visible-context",
              label: "Visible context",
              tokens: cappedTokensInUse,
              blocks: blocks[0] ?? 0,
              color: turnColor(0),
              turn: visibleTurns[0]?.turn,
              meta: `${Math.round(turnShare(cappedTokensInUse))}% of window`,
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
        meta: `${Math.round(
          Math.max(0, 100 - summary.percent_consumed),
        )}% of window`,
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
</script>

<section class="panel">
  <div class="panel-header">
    <div class="header-main">
      <div class="eyebrow">Window Map</div>
      <h3>Visible context accumulated across turns</h3>
    </div>
  </div>

  {#if unknownCapacity}
    <div class="empty-state">
      This session does not have a known model context window, so the block map cannot be rendered.
    </div>
  {:else}
    <div class="viz-layout">
      <div class="grid-shell">
        <div class="summary summary-inline">
          <strong>{formatTokenCount(summary.tokens_in_use)}</strong>
          <span>
            {#if capacity.max_tokens > 0}
              / {formatTokenCount(capacity.max_tokens)} tokens
            {:else}
              capacity unknown
            {/if}
          </span>
        </div>
        <div
          class="block-grid"
          style={`grid-template-columns: repeat(${BLOCK_COLUMNS}, 14px);`}
          aria-label={`Context window split into ${TOTAL_BLOCKS} equal blocks across ${gridRows} rows`}
        >
          {#each blocks as block (block.key)}
            <button
              type="button"
              class:free-block={block.isFree}
              class="block"
              title={block.label}
              style={`background:${block.color};`}
              onclick={() => jumpToTurn(block.turn)}
              disabled={block.isFree}
              aria-label={block.turn
                ? `${block.label}. Jump to turn ${block.turn}`
                : block.label}
            ></button>
          {/each}
        </div>
        <div class="grid-meta">
          <span>{TOTAL_BLOCKS} equal blocks</span>
          <span>Each block ≈ {formatTokenCount(Math.round(capacity.max_tokens / TOTAL_BLOCKS))}</span>
        </div>
      </div>

      <div class="pie-shell">
        <div class="pie-header">
          <div class="pie-title">Turn shares</div>
          <div class="pie-subtitle">Visible window split by turn</div>
        </div>

        <div class="pie-layout">
          <svg
            class="pie-chart"
            viewBox="0 0 144 144"
            role="img"
            aria-label="Turn distribution of visible context"
          >
            <circle
              cx="72"
              cy="72"
              r="56"
              fill="var(--bg-inset)"
            />
            {#each pieSegments as segment (segment.key)}
              <path
                d={segment.path}
                fill={segment.color}
                class="pie-slice"
                tabindex={segment.isFree ? undefined : 0}
                role={segment.isFree ? undefined : "button"}
                aria-label={segment.turn
                  ? `${segment.label}. Jump to turn ${segment.turn}`
                  : segment.label}
                onclick={() => jumpToTurn(segment.turn)}
                onkeydown={(event) => {
                  if (segment.isFree) return;
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    jumpToTurn(segment.turn);
                  }
                }}
              />
            {/each}
            <circle
              cx="72"
              cy="72"
              r="28"
              fill="var(--bg-surface)"
            />
            <text x="72" y="68" text-anchor="middle" class="pie-center-label">
              Used
            </text>
            <text x="72" y="84" text-anchor="middle" class="pie-center-value">
              {Math.round(summary.percent_consumed)}%
            </text>
          </svg>

          <div class="pie-meta">
            {#each segments.filter((segment) => !segment.isFree).slice(0, 6) as segment (segment.key)}
              <button
                type="button"
                class="pie-meta-row"
                onclick={() => jumpToTurn(segment.turn)}
              >
                <div class="pie-meta-name">
                  <span
                    class="swatch"
                    style={`background:${segment.color};`}
                  ></span>
                  <span>{segment.turn ? `T${segment.turn}` : segment.label}</span>
                </div>
                <div class="pie-meta-value">
                  {Math.round(turnShare(segment.tokens))}%
                </div>
              </button>
            {/each}
            {#if segments.length > 6}
              <div class="pie-more">+{segments.filter((segment) => !segment.isFree).length - 6} more turns</div>
            {/if}
            <div class="pie-meta-row free-row">
              <div class="pie-meta-name">
                <span
                  class="swatch free-swatch"
                  style={`background:var(--bg-inset);`}
                ></span>
                <span>Free</span>
              </div>
              <div class="pie-meta-value">
                {Math.max(0, 100 - Math.round(summary.percent_consumed))}%
              </div>
            </div>
          </div>
        </div>
      </div>
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

  .summary {
    display: grid;
    gap: 2px;
    color: var(--text-muted);
    font-size: 11px;
  }

  .summary-inline {
    justify-self: start;
    margin-bottom: 2px;
  }

  .summary strong {
    font-size: 16px;
    font-weight: 700;
    color: var(--text-primary);
  }

  .grid-shell {
    display: grid;
    gap: 8px;
  }

  .viz-layout {
    display: grid;
    grid-template-columns: max-content max-content;
    gap: 28px;
    align-items: start;
    justify-content: start;
  }

  .block-grid {
    display: grid;
    gap: 4px;
    justify-content: start;
  }

  .block {
    width: 14px;
    height: 14px;
    border-radius: 4px;
    border: 0;
    padding: 0;
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--bg-surface) 72%, transparent);
    cursor: pointer;
  }

  .free-block {
    box-shadow: inset 0 0 0 1px var(--border-muted);
    cursor: default;
  }

  .block:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow:
      inset 0 0 0 1px color-mix(in srgb, var(--bg-surface) 72%, transparent),
      0 0 0 1px color-mix(in srgb, var(--text-primary) 12%, transparent);
  }

  .block:focus-visible {
    outline: 2px solid var(--accent-blue);
    outline-offset: 2px;
  }

  .grid-meta {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
    font-size: 11px;
    color: var(--text-muted);
  }

  .pie-shell {
    display: grid;
    gap: 10px;
    align-content: start;
  }

  .pie-layout {
    display: grid;
    grid-template-columns: max-content max-content;
    gap: 16px;
    align-items: start;
  }

  .pie-header {
    display: grid;
    gap: 2px;
  }

  .pie-title {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .pie-subtitle {
    font-size: 11px;
    color: var(--text-muted);
  }

  .pie-chart {
    width: 180px;
    height: 180px;
    justify-self: center;
  }

  .pie-slice {
    cursor: pointer;
    transition: opacity 0.12s ease;
  }

  .pie-slice:hover {
    opacity: 0.86;
  }

  .pie-slice:focus-visible {
    outline: 2px solid var(--accent-blue);
    outline-offset: 2px;
  }

  .pie-center-label {
    font-size: 9px;
    fill: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .pie-center-value {
    font-size: 15px;
    font-weight: 700;
    fill: var(--text-primary);
  }

  .pie-meta {
    display: grid;
    gap: 6px;
    min-width: 140px;
    align-content: start;
  }

  .pie-meta-row {
    display: flex;
    justify-content: space-between;
    gap: 10px;
    align-items: center;
    color: var(--text-muted);
    font-size: 11px;
    border: 0;
    background: transparent;
    padding: 0;
    text-align: left;
  }

  .pie-meta-name {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--text-primary);
    min-width: 0;
  }

  .pie-meta-value {
    color: var(--text-muted);
    white-space: nowrap;
  }

  .pie-more {
    font-size: 11px;
    color: var(--text-muted);
  }

  .swatch {
    width: 10px;
    height: 10px;
    border-radius: 999px;
    display: inline-block;
  }

  .free-swatch {
    box-shadow: inset 0 0 0 1px var(--border-muted);
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

    .pie-layout {
      grid-template-columns: 1fr;
    }

    .summary {
      text-align: left;
    }
  }
</style>
