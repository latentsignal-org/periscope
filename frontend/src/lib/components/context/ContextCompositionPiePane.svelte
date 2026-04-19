<script lang="ts">
  import { formatTokenCount } from "../../utils/format.js";

  interface Segment {
    category: string;
    label: string;
    color: string;
    percentage: number;
    path: string;
  }

  interface Props {
    tokensInUse: number;
    segments: Segment[];
  }

  let { tokensInUse, segments }: Props = $props();
</script>

<section class="subpanel">
  <div class="subheader">
    <div class="title">Composition</div>
    <div class="subtitle">What is consuming the visible window</div>
  </div>

  <div class="layout">
    <svg
      class="pie-chart"
      viewBox="0 0 144 144"
      role="img"
      aria-label="Composition distribution of visible context"
    >
      <circle cx="72" cy="72" r="56" fill="var(--bg-inset)" />
      {#each segments as segment (segment.category)}
        <path d={segment.path} fill={segment.color} />
      {/each}
      <circle cx="72" cy="72" r="28" fill="var(--bg-surface)" />
      <text x="72" y="68" text-anchor="middle" class="center-label">Window</text>
      <text x="72" y="84" text-anchor="middle" class="center-small">
        {formatTokenCount(tokensInUse)}
      </text>
    </svg>

    <div class="legend">
      {#each segments.slice(0, 6) as segment (segment.category)}
        <div class="legend-row">
          <div class="legend-name">
            <span class="swatch" style={`background:${segment.color};`}></span>
            <span>{segment.label}</span>
          </div>
          <div class="legend-value">{Math.round(segment.percentage)}%</div>
        </div>
      {/each}
      {#if segments.length > 6}
        <div class="more">+{segments.length - 6} more categories</div>
      {/if}
    </div>
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

  .layout {
    display: grid;
    gap: 16px;
    align-content: start;
  }

  .pie-chart {
    width: 180px;
    height: 180px;
    justify-self: center;
  }

  .center-label {
    font-size: 9px;
    fill: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .center-small {
    font-size: 11px;
    font-weight: 700;
    fill: var(--text-primary);
  }

  .legend {
    display: grid;
    gap: 6px;
    align-content: start;
  }

  .legend-row {
    display: flex;
    justify-content: space-between;
    gap: 10px;
    align-items: center;
    color: var(--text-muted);
    font-size: 11px;
  }

  .legend-name {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--text-primary);
    min-width: 0;
  }

  .legend-value {
    color: var(--text-muted);
    white-space: nowrap;
  }

  .more {
    font-size: 11px;
    color: var(--text-muted);
  }

  .swatch {
    width: 10px;
    height: 10px;
    border-radius: 999px;
    display: inline-block;
  }
</style>
