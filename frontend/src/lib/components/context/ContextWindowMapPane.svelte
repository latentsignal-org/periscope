<script lang="ts">
  import type { ContextCapacity, ContextSummary } from "../../api/types.js";
  import { formatTokenCount } from "../../utils/format.js";

  interface BlockCell {
    key: string;
    color: string;
    label: string;
    turn?: number;
    isFree?: boolean;
  }

  interface Props {
    summary: ContextSummary;
    capacity: ContextCapacity;
    blocks: BlockCell[];
    totalBlocks: number;
    blockColumns: number;
    gridRows: number;
    onJumpToTurn: (turn: number | undefined) => void;
  }

  let {
    summary,
    capacity,
    blocks,
    totalBlocks,
    blockColumns,
    gridRows,
    onJumpToTurn,
  }: Props = $props();
</script>

<section class="subpanel">
  <div class="summary">
    <strong>{formatTokenCount(summary.tokens_in_use)}</strong>
    <span>
      {#if capacity.max_tokens > 0}
        / {formatTokenCount(capacity.max_tokens)} tokens
      {:else}
        capacity unknown
      {/if}
    </span>
  </div>

  <div class="subheader">
    <div class="title">Window map</div>
    <div class="subtitle">Visible context accumulated across turns</div>
  </div>

  <div
    class="block-grid"
    style={`grid-template-columns: repeat(${blockColumns}, 14px);`}
    aria-label={`Context window split into ${totalBlocks} equal blocks across ${gridRows} rows`}
  >
    {#each blocks as block (block.key)}
      <button
        type="button"
        class:free-block={block.isFree}
        class="block"
        title={block.label}
        style={`background:${block.color};`}
        onclick={() => onJumpToTurn(block.turn)}
        disabled={block.isFree}
        aria-label={block.turn
          ? `${block.label}. Jump to turn ${block.turn}`
          : block.label}
      ></button>
    {/each}
  </div>

  <div class="meta">
    <span>{totalBlocks} equal blocks</span>
    <span>Each block ≈ {formatTokenCount(Math.round(capacity.max_tokens / totalBlocks))}</span>
  </div>
</section>

<style>
  .subpanel {
    border: 1px solid var(--border-muted);
    background: var(--bg-surface);
    border-radius: var(--radius-sm);
    padding: 12px;
    display: grid;
    gap: 10px;
    align-content: start;
    min-height: 100%;
  }

  .summary {
    display: grid;
    gap: 2px;
    color: var(--text-muted);
    font-size: 11px;
  }

  .summary strong {
    font-size: 16px;
    font-weight: 700;
    color: var(--text-primary);
  }

  .subheader {
    display: grid;
    gap: 2px;
  }

  .title {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .subtitle {
    font-size: 11px;
    color: var(--text-muted);
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

  .meta {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
    font-size: 11px;
    color: var(--text-muted);
  }
</style>
