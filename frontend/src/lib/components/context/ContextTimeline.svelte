<script lang="ts">
  import type { ContextTimelineTurn } from "../../api/types.js";
  import { formatTimestamp, formatTokenCount } from "../../utils/format.js";
  import { CATEGORY_COLORS, categoryLabel } from "./context-utils.js";
  import { router } from "../../stores/router.svelte.js";

  interface Props {
    timeline: ContextTimelineTurn[];
    sessionId: string;
  }

  let { timeline, sessionId }: Props = $props();

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

  function jumpToTranscript(ordinal: number) {
    router.navigateToSession(sessionId, { msg: String(ordinal) });
  }

  function entryKindLabel(kind: string): string {
    switch (kind) {
      case "user_message":
        return "User";
      case "assistant_message":
        return "Assistant";
      case "tool_call":
        return "Tool";
      default:
        return kind;
    }
  }
</script>

<section class="panel">
  <div>
    <div class="eyebrow">Timeline</div>
    <h3>Visible history, row by row</h3>
  </div>

  <div class="rows">
    {#each timeline as turn (turn.turn)}
      <details class:compaction-row={turn.markers?.includes("compaction")} class="turn-shell">
        <summary class="turn-summary">
          <div class="turn-summary-grid">
            <div class="ordinal">
              <span class="ordinal-label">T{turn.turn}</span>
            </div>
            <div class="metric">
              <span class="metric-label">Delta</span>
              <strong>+{formatTokenCount(turn.delta_tokens)}</strong>
              <span class="metric-meta">{turn.delta_provenance}</span>
            </div>
            <div class="metric">
              <span class="metric-label">Cumulative</span>
              <strong>{formatTokenCount(turn.cumulative_tokens)}</strong>
              <span class="metric-meta">{turn.cumulative_provenance}</span>
            </div>
            <div class="turn-main">
              <div class="turn-topline">
                <span class="turn-label">
                  {#if turn.markers?.includes("compaction")}
                    {turn.label || "Compaction seed"}
                  {:else}
                    Messages {turn.start_ordinal}-{turn.end_ordinal}
                  {/if}
                </span>
                {#if turn.timestamp}
                  <span class="row-time">{formatTimestamp(turn.timestamp)}</span>
                {/if}
                {#each turn.markers ?? [] as marker}
                  <span class="marker">{markerLabel(marker)}</span>
                {/each}
              </div>
              <div class="bar" aria-hidden="true">
                {#each turn.categories as category (category.category)}
                  <div
                    class="bar-segment"
                    title={`${categoryLabel(category.category)} · ${formatTokenCount(category.tokens)} tokens`}
                    style={`flex:${Math.max(category.tokens, 1)};background:${CATEGORY_COLORS[category.category] ?? CATEGORY_COLORS.other}`}
                  ></div>
                {/each}
              </div>
              <div class="category-summary">
                {#each turn.categories.slice(0, 3) as category (category.category)}
                  <span>{categoryLabel(category.category)} {formatTokenCount(category.tokens)}</span>
                {/each}
              </div>
            </div>
            <div class="chevron" aria-hidden="true">▾</div>
          </div>
        </summary>

        <div class="turn-children">
          {#each turn.entries ?? [] as entry, i (`${entry.kind}-${entry.ordinal}-${i}`)}
            <button
              type="button"
              class="entry-row"
              onclick={() => jumpToTranscript(entry.ordinal)}
            >
              <div class="entry-kind">{entryKindLabel(entry.kind)}</div>
              <div class="entry-content">
                <div class="entry-label">{entry.label}</div>
                {#if entry.preview}
                  <div class="entry-preview">{entry.preview}</div>
                {/if}
              </div>
              <div class="entry-ordinal">#{entry.ordinal}</div>
            </button>
          {/each}

          {#each turn.annotations ?? [] as annotation}
            <div class="annotation">{annotation}</div>
          {/each}
        </div>
      </details>
    {/each}
  </div>
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

  .rows {
    display: grid;
    gap: 10px;
  }

  .turn-shell {
    border-top: 1px solid var(--border-muted);
    padding-top: 10px;
  }

  .turn-shell.compaction-row {
    border-top: 2px solid var(--accent-rose);
    padding-top: 10px;
  }

  .turn-summary {
    list-style: none;
    cursor: pointer;
  }

  .turn-summary::-webkit-details-marker {
    display: none;
  }

  .turn-summary-grid {
    display: grid;
    grid-template-columns: 64px 112px 120px minmax(0, 1fr) 24px;
    gap: 12px;
    align-items: start;
  }

  .ordinal,
  .metric-label,
  .metric-meta,
  .row-time,
  .marker,
  .category-summary,
  .entry-kind,
  .entry-ordinal,
  .annotation {
    font-size: 11px;
    color: var(--text-muted);
  }

  .ordinal-label {
    font-weight: 700;
    color: var(--text-secondary);
  }

  .metric {
    display: grid;
    gap: 2px;
  }

  .metric strong {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .turn-main {
    display: grid;
    gap: 6px;
  }

  .turn-topline,
  .category-summary {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
    align-items: center;
  }

  .turn-label {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .bar {
    display: flex;
    min-height: 12px;
    overflow: hidden;
    border-radius: var(--radius-sm);
    background: var(--bg-inset);
  }

  .bar-segment {
    min-width: 2px;
  }

  .marker {
    border: 1px solid var(--border-muted);
    border-radius: 999px;
    padding: 1px 6px;
    color: var(--text-secondary);
  }

  .chevron {
    color: var(--text-muted);
    transition: transform 0.15s ease;
    transform-origin: center;
    padding-top: 3px;
  }

  details[open] .chevron {
    transform: rotate(180deg);
  }

  .turn-children {
    margin-left: calc(64px + 112px + 120px + 24px);
    margin-top: 8px;
    display: grid;
    gap: 6px;
  }

  .entry-row {
    width: 100%;
    border: 1px solid var(--border-muted);
    background: var(--bg-inset);
    color: var(--text-primary);
    border-radius: var(--radius-sm);
    padding: 8px 10px;
    display: grid;
    grid-template-columns: 88px minmax(0, 1fr) 52px;
    gap: 12px;
    align-items: start;
    text-align: left;
    cursor: pointer;
    transition: background 0.1s, border-color 0.1s;
  }

  .entry-row:hover {
    background: var(--bg-surface-hover);
    border-color: var(--border-default);
  }

  .entry-kind {
    text-transform: uppercase;
    letter-spacing: 0.04em;
    font-weight: 700;
    color: var(--text-secondary);
  }

  .entry-content {
    display: grid;
    gap: 2px;
  }

  .entry-label {
    font-size: 12px;
    font-weight: 600;
  }

  .entry-preview {
    font-size: 12px;
    line-height: 1.4;
    color: var(--text-secondary);
  }

  .entry-ordinal {
    text-align: right;
  }

  .annotation {
    padding-left: 12px;
  }

  @media (max-width: 900px) {
    .turn-summary-grid {
      grid-template-columns: 56px 1fr 24px;
    }

    .turn-main {
      grid-column: 1 / span 3;
    }

    .turn-children {
      margin-left: 0;
    }

    .entry-row {
      grid-template-columns: 1fr;
    }

    .entry-ordinal {
      text-align: left;
    }
  }
</style>
