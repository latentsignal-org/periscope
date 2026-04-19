<script lang="ts">
  import type { ContextTimelineRow } from "../../api/types.js";
  import { formatTimestamp, formatTokenCount } from "../../utils/format.js";
  import { CATEGORY_COLORS, categoryLabel } from "./context-utils.js";

  interface Props {
    timeline: ContextTimelineRow[];
  }

  let { timeline }: Props = $props();

  function markerLabel(marker: string): string {
    switch (marker) {
      case "spike":
        return "▲ spike";
      case "compaction":
        return "↻ compaction";
      case "subagent":
        return "● subagent";
      default:
        return marker;
    }
  }
</script>

<section class="panel">
  <div>
    <div class="eyebrow">Timeline</div>
    <h3>Visible history, row by row</h3>
  </div>

  <div class="rows">
    {#each timeline as row (row.ordinal)}
      <div class:compaction-row={row.markers?.includes("compaction")} class="row-shell">
        <div class="row-main">
          <div class="ordinal">#{row.ordinal}</div>
          <div class="metric delta">
            <span class="metric-label">Delta</span>
            <strong>+{formatTokenCount(row.delta_tokens)}</strong>
            <span class="metric-meta">{row.delta_provenance}</span>
          </div>
          <div class="metric cumul">
            <span class="metric-label">Cumulative</span>
            <strong>{formatTokenCount(row.cumulative_tokens)}</strong>
            <span class="metric-meta">{row.cumulative_provenance}</span>
          </div>
          <div class="row-body">
            <div class="row-topline">
              <span class="row-label">{row.label}</span>
              {#if row.timestamp}
                <span class="row-time">{formatTimestamp(row.timestamp)}</span>
              {/if}
              {#each row.markers ?? [] as marker}
                <span class="marker">{markerLabel(marker)}</span>
              {/each}
            </div>
            <div class="bar" aria-hidden="true">
              {#each row.categories as category (category.category)}
                <div
                  class="bar-segment"
                  title={`${categoryLabel(category.category)} · ${formatTokenCount(category.tokens)} tokens`}
                  style={`flex:${Math.max(category.tokens, 1)};background:${CATEGORY_COLORS[category.category] ?? CATEGORY_COLORS.other}`}
                ></div>
              {/each}
            </div>
            <div class="category-summary">
              {#each row.categories.slice(0, 3) as category (category.category)}
                <span>
                  {categoryLabel(category.category)} {formatTokenCount(category.tokens)}
                </span>
              {/each}
            </div>
          </div>
        </div>
        {#each row.annotations ?? [] as annotation}
          <div class="annotation">{annotation}</div>
        {/each}
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
    gap: 16px;
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

  .rows {
    display: grid;
    gap: 12px;
  }

  .row-shell {
    display: grid;
    gap: 6px;
    padding-top: 12px;
    border-top: 1px solid var(--border-default);
  }

  .row-shell.compaction-row {
    border-top: 4px solid #be123c;
    padding-top: 10px;
  }

  .row-main {
    display: grid;
    grid-template-columns: 64px 112px 120px minmax(0, 1fr);
    gap: 12px;
    align-items: start;
  }

  .ordinal,
  .metric-label,
  .metric-meta,
  .row-time,
  .marker,
  .annotation,
  .category-summary {
    font-size: 12px;
    color: var(--text-secondary);
  }

  .ordinal {
    font-weight: 700;
  }

  .metric {
    display: grid;
    gap: 2px;
  }

  .row-body {
    display: grid;
    gap: 8px;
  }

  .row-topline,
  .category-summary {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
    align-items: center;
  }

  .row-label {
    font-size: 13px;
    font-weight: 600;
  }

  .bar {
    display: flex;
    min-height: 16px;
    overflow: hidden;
    border-radius: 8px;
    background: color-mix(in srgb, var(--bg-surface) 70%, #334155 30%);
  }

  .bar-segment {
    min-width: 2px;
  }

  .marker {
    border: 1px solid var(--border-default);
    border-radius: 999px;
    padding: 2px 6px;
  }

  .annotation {
    margin-left: calc(64px + 112px + 120px + 24px);
  }

  @media (max-width: 900px) {
    .row-main {
      grid-template-columns: 56px 1fr;
    }

    .cumul {
      grid-column: 2;
    }

    .row-body {
      grid-column: 1 / -1;
    }

    .annotation {
      margin-left: 0;
    }
  }
</style>
