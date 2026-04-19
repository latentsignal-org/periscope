<script lang="ts">
  import type { ContextCompositionItem } from "../../api/types.js";
  import { formatTokenCount } from "../../utils/format.js";
  import { CATEGORY_COLORS, categoryLabel } from "./context-utils.js";

  interface Props {
    composition: ContextCompositionItem[];
  }

  let { composition }: Props = $props();

  const visibleItems = $derived(
    composition.filter((item) => item.tokens > 0),
  );
</script>

<section class="panel">
  <div class="panel-header">
    <div>
      <div class="eyebrow">Composition</div>
      <h3>What is consuming the visible window</h3>
    </div>
  </div>

  <div class="stacked-bar" aria-hidden="true">
    {#each visibleItems as item (item.category)}
      <div
        class="segment"
        title={`${categoryLabel(item.category)} · ${formatTokenCount(item.tokens)} tokens`}
        style={`width:${Math.max(item.percentage, 0.5)}%;background:${CATEGORY_COLORS[item.category] ?? CATEGORY_COLORS.other}`}
      ></div>
    {/each}
  </div>

  <div class="legend">
    {#each visibleItems as item (item.category)}
      <div class="legend-row">
        <div class="legend-name">
          <span
            class="swatch"
            style={`background:${CATEGORY_COLORS[item.category] ?? CATEGORY_COLORS.other}`}
          ></span>
          <span>{categoryLabel(item.category)}</span>
        </div>
        <div class="legend-metrics">
          <strong>{formatTokenCount(item.tokens)}</strong>
          <span>{Math.round(item.percentage)}%</span>
          <span>{item.provenance}</span>
        </div>
      </div>
    {/each}
  </div>
</section>

<style>
  .panel {
    border: 1px solid var(--border-default);
    background: var(--bg-surface);
    border-radius: 14px;
    padding: 18px;
    display: grid;
    gap: 14px;
  }

  .eyebrow {
    font-size: 11px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--text-secondary);
    margin-bottom: 6px;
  }

  h3 {
    margin: 0;
    font-size: 18px;
  }

  .stacked-bar {
    display: flex;
    height: 18px;
    overflow: hidden;
    border-radius: 999px;
    background: color-mix(in srgb, var(--bg-surface) 70%, #334155 30%);
  }

  .segment {
    min-width: 2px;
  }

  .legend {
    display: grid;
    gap: 10px;
  }

  .legend-row {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    align-items: center;
    font-size: 13px;
  }

  .legend-name,
  .legend-metrics {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
  }

  .legend-metrics {
    color: var(--text-secondary);
  }

  .swatch {
    width: 10px;
    height: 10px;
    border-radius: 999px;
    display: inline-block;
  }
</style>
