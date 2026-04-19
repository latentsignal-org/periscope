<script lang="ts">
  interface Segment {
    key: string;
    label: string;
    tokens: number;
    color: string;
    turn?: number;
    isFree?: boolean;
    path: string;
  }

  interface Props {
    segments: Segment[];
    usedPercent: number;
    onJumpToTurn: (turn: number | undefined) => void;
  }

  let { segments, usedPercent, onJumpToTurn }: Props = $props();

  function turnShare(segmentTokens: number): number {
    const total = segments.reduce((sum, segment) => sum + segment.tokens, 0);
    if (total <= 0) return 0;
    return (segmentTokens / total) * 100;
  }
</script>

<section class="subpanel">
  <div class="subheader">
    <div class="title">Composition by turn</div>
    <div class="subtitle">Visible window split by turn</div>
  </div>

  <div class="layout">
    <svg
      class="pie-chart"
      viewBox="0 0 144 144"
      role="img"
      aria-label="Turn distribution of visible context"
    >
      <circle cx="72" cy="72" r="56" fill="var(--bg-inset)" />
      {#each segments as segment (segment.key)}
        <path
          d={segment.path}
          fill={segment.color}
          class="pie-slice"
          tabindex={segment.isFree ? undefined : 0}
          role={segment.isFree ? undefined : "button"}
          aria-label={segment.turn
            ? `${segment.label}. Jump to turn ${segment.turn}`
            : segment.label}
          onclick={() => onJumpToTurn(segment.turn)}
          onkeydown={(event) => {
            if (segment.isFree) return;
            if (event.key === "Enter" || event.key === " ") {
              event.preventDefault();
              onJumpToTurn(segment.turn);
            }
          }}
        />
      {/each}
      <circle cx="72" cy="72" r="28" fill="var(--bg-surface)" />
      <text x="72" y="68" text-anchor="middle" class="center-label">Used</text>
      <text x="72" y="84" text-anchor="middle" class="center-value">
        {Math.round(usedPercent)}%
      </text>
    </svg>

    <div class="legend">
      {#each segments.filter((segment) => !segment.isFree).slice(0, 6) as segment (segment.key)}
        <button
          type="button"
          class="legend-row"
          onclick={() => onJumpToTurn(segment.turn)}
        >
          <div class="legend-name">
            <span class="swatch" style={`background:${segment.color};`}></span>
            <span>{segment.turn ? `T${segment.turn}` : segment.label}</span>
          </div>
          <div class="legend-value">{Math.round(turnShare(segment.tokens))}%</div>
        </button>
      {/each}
      {#if segments.length > 6}
        <div class="more">+{segments.filter((segment) => !segment.isFree).length - 6} more turns</div>
      {/if}
      <div class="legend-row free-row">
        <div class="legend-name">
          <span class="swatch free-swatch" style={`background:var(--bg-inset);`}></span>
          <span>Free</span>
        </div>
        <div class="legend-value">{Math.max(0, 100 - Math.round(usedPercent))}%</div>
      </div>
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
    grid-template-columns: max-content minmax(140px, 1fr);
    gap: 18px;
    align-content: start;
    align-items: center;
  }

  .pie-chart {
    width: 220px;
    height: 220px;
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

  .center-label {
    font-size: 9px;
    fill: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .center-value {
    font-size: 15px;
    font-weight: 700;
    fill: var(--text-primary);
  }

  .legend {
    display: grid;
    gap: 6px;
    align-content: start;
    min-width: 140px;
  }

  .legend-row {
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

  .free-swatch {
    box-shadow: inset 0 0 0 1px var(--border-muted);
  }

  @media (max-width: 900px) {
    .layout {
      grid-template-columns: 1fr;
      gap: 16px;
    }
  }
</style>
