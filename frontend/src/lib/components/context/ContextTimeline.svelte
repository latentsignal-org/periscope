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

  .turn-shell {
    border-top: 1px solid var(--border-default);
    padding-top: 12px;
  }

  .turn-shell.compaction-row {
    border-top: 4px solid #be123c;
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
    font-size: 12px;
    color: var(--text-secondary);
  }

  .ordinal-label {
    font-weight: 700;
  }

  .metric {
    display: grid;
    gap: 2px;
  }

  .turn-main {
    display: grid;
    gap: 8px;
  }

  .turn-topline,
  .category-summary {
    display: flex;
    gap: 10px;
    flex-wrap: wrap;
    align-items: center;
  }

  .turn-label {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
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

  .chevron {
    color: var(--text-secondary);
    transition: transform 0.15s ease;
    transform-origin: center;
    padding-top: 3px;
  }

  details[open] .chevron {
    transform: rotate(180deg);
  }

  .turn-children {
    margin-left: calc(64px + 112px + 120px + 24px);
    margin-top: 10px;
    display: grid;
    gap: 8px;
  }

  .entry-row {
    width: 100%;
    border: 1px solid var(--border-default);
    background: color-mix(in srgb, var(--bg-surface) 88%, #0f172a 12%);
    color: var(--text-primary);
    border-radius: 10px;
    padding: 10px 12px;
    display: grid;
    grid-template-columns: 88px minmax(0, 1fr) 52px;
    gap: 12px;
    align-items: start;
    text-align: left;
    cursor: pointer;
  }

  .entry-row:hover {
    background: color-mix(in srgb, var(--bg-surface) 80%, #0f172a 20%);
  }

  .entry-kind {
    text-transform: uppercase;
    letter-spacing: 0.04em;
    font-weight: 700;
  }

  .entry-content {
    display: grid;
    gap: 4px;
  }

  .entry-label {
    font-size: 13px;
    font-weight: 600;
  }

  .entry-preview {
    font-size: 14px;
    line-height: 1.35;
    color: var(--text-primary);
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
