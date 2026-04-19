<script lang="ts">
  import type {
    ContextTimelineEntry,
    ContextTimelineTurn,
  } from "../../api/types.js";
  import { formatTimestamp, formatTokenCount } from "../../utils/format.js";
  import { CATEGORY_COLORS, categoryLabel } from "./context-utils.js";
  import { router } from "../../stores/router.svelte.js";

  interface Props {
    timeline: ContextTimelineTurn[];
    sessionId: string;
  }

  let { timeline, sessionId }: Props = $props();

  let reversed = $state(false);
  let displayTimeline = $derived(reversed ? [...timeline].reverse() : timeline);

  type GroupedItem =
    | { type: "single"; entry: ContextTimelineEntry; key: string }
    | { type: "tool_group"; entries: ContextTimelineEntry[]; key: string };

  function groupEntries(
    entries: ContextTimelineEntry[] | undefined,
  ): GroupedItem[] {
    if (!entries?.length) return [];
    const grouped: GroupedItem[] = [];
    let toolBuffer: ContextTimelineEntry[] = [];

    const flushTools = () => {
      if (toolBuffer.length === 0) return;
      grouped.push({
        type: "tool_group",
        entries: toolBuffer,
        key: `tools-${toolBuffer[0].ordinal}-${toolBuffer.length}`,
      });
      toolBuffer = [];
    };

    for (const entry of entries) {
      if (entry.kind === "tool_call") {
        toolBuffer.push(entry);
        continue;
      }
      // Skip assistant wrappers that only existed to carry tool
      // calls — without flushing, so adjacent tool calls still
      // group together across the empty wrapper.
      if (
        entry.kind === "assistant_message" &&
        !entry.preview?.trim()
      ) {
        continue;
      }
      flushTools();
      grouped.push({
        type: "single",
        entry,
        key: `${entry.kind}-${entry.ordinal}`,
      });
    }
    flushTools();
    return grouped;
  }

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

  function entryKindClass(kind: string): string {
    switch (kind) {
      case "user_message":
        return "entry-user";
      case "assistant_message":
        return "entry-assistant";
      case "tool_call":
        return "entry-tool";
      default:
        return "entry-other";
    }
  }

  function entryIcon(kind: string): string {
    switch (kind) {
      case "user_message":
        return "U";
      case "assistant_message":
        return "A";
      case "tool_call":
        return "T";
      default:
        return "•";
    }
  }
</script>

<section class="panel">
  <div class="panel-header">
    <div>
      <div class="eyebrow">Timeline</div>
      <h3>Visible history, row by row</h3>
    </div>
    <button
      type="button"
      class="sort-btn"
      onclick={() => (reversed = !reversed)}
      title={reversed ? "Showing newest first — click to show oldest first" : "Showing oldest first — click to show newest first"}
      aria-label={reversed ? "Sort oldest first" : "Sort newest first"}
    >
      <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor" aria-hidden="true">
        <path d="M4 2a.5.5 0 0 1 .5.5v9.793l2.146-2.147a.5.5 0 0 1 .708.708l-3 3a.5.5 0 0 1-.708 0l-3-3a.5.5 0 0 1 .708-.708L3.5 12.293V2.5A.5.5 0 0 1 4 2zm8 0a.5.5 0 0 1 .354.146l3 3a.5.5 0 0 1-.708.708L12.5 3.707V13.5a.5.5 0 0 1-1 0V3.707l-2.146 2.147a.5.5 0 1 1-.708-.708l3-3A.5.5 0 0 1 12 2z"/>
      </svg>
      {reversed ? "Newest first" : "Oldest first"}
    </button>
  </div>

  <div class="rows">
    {#each displayTimeline as turn (turn.turn)}
      <details
        id={`context-turn-${turn.turn}`}
        class:compaction-row={turn.markers?.includes("compaction")}
        class="turn-shell"
      >
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
            <div class="chevron" aria-hidden="true">
              <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                <path d="M1.646 4.646a.5.5 0 0 1 .708 0L8 10.293l5.646-5.647a.5.5 0 0 1 .708.708l-6 6a.5.5 0 0 1-.708 0l-6-6a.5.5 0 0 1 0-.708z"/>
              </svg>
            </div>
          </div>
        </summary>

        <div class="turn-children">
          {#each groupEntries(turn.entries) as item (item.key)}
            {#if item.type === "single"}
              <button
                type="button"
                class={`entry-row ${entryKindClass(item.entry.kind)}`}
                onclick={() => jumpToTranscript(item.entry.ordinal)}
              >
                <div class="entry-header">
                  <span class="entry-icon">{entryIcon(item.entry.kind)}</span>
                  <span class="entry-label">{entryKindLabel(item.entry.kind)}</span>
                  <span class="entry-ordinal">#{item.entry.ordinal}</span>
                </div>
                {#if item.entry.preview}
                  <div class="entry-preview">{item.entry.preview}</div>
                {/if}
              </button>
            {:else}
              <div class="tool-group">
                <div class="tool-group-header">
                  <svg
                    class="gear-icon"
                    width="12" height="12" viewBox="0 0 16 16"
                    fill="var(--accent-amber)"
                    aria-hidden="true"
                  >
                    <path d="M8 4.754a3.246 3.246 0 100 6.492 3.246 3.246 0 000-6.492zM5.754 8a2.246 2.246 0 114.492 0 2.246 2.246 0 01-4.492 0z"/>
                    <path d="M9.796 1.343c-.527-1.79-3.065-1.79-3.592 0l-.094.319a.873.873 0 01-1.255.52l-.292-.16c-1.64-.892-3.433.902-2.54 2.541l.159.292a.873.873 0 01-.52 1.255l-.319.094c-1.79.527-1.79 3.065 0 3.592l.319.094a.873.873 0 01.52 1.255l-.16.292c-.892 1.64.901 3.434 2.541 2.54l.292-.159a.873.873 0 011.255.52l.094.319c.527 1.79 3.065 1.79 3.592 0l.094-.319a.873.873 0 011.255-.52l.292.16c1.64.893 3.434-.902 2.54-2.541l-.159-.292a.873.873 0 01.52-1.255l.319-.094c1.79-.527 1.79-3.065 0-3.592l-.319-.094a.873.873 0 01-.52-1.255l.16-.292c.893-1.64-.902-3.433-2.541-2.54l-.292.159a.873.873 0 01-1.255-.52l-.094-.319z"/>
                  </svg>
                  <span class="group-label">
                    {item.entries.length === 1
                      ? "1 tool call"
                      : `${item.entries.length} tool calls`}
                  </span>
                </div>
                <div class="tool-group-body">
                  {#each item.entries as tool, i (`${tool.ordinal}-${i}`)}
                    <button
                      type="button"
                      class="tool-row"
                      class:has-output={!!tool.output_preview}
                      onclick={() => jumpToTranscript(tool.ordinal)}
                    >
                      <div class="tool-line tool-input-line">
                        <span class="tool-chevron" aria-hidden="true">$</span>
                        <span class="tool-name">{tool.label}</span>
                        {#if tool.preview}
                          <span class="tool-snippet">{tool.preview}</span>
                        {/if}
                      </div>
                      {#if tool.output_preview}
                        <div class="tool-line tool-output-line">
                          <span class="tool-chevron" aria-hidden="true">↳</span>
                          <span class="tool-snippet">{tool.output_preview}</span>
                        </div>
                      {/if}
                    </button>
                  {/each}
                </div>
              </div>
            {/if}
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

  .panel-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 8px;
  }

  .sort-btn {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 4px 10px;
    border: 1px solid var(--border-muted);
    border-radius: var(--radius-md);
    background: var(--bg-inset);
    color: var(--text-secondary);
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    white-space: nowrap;
    transition: background 0.1s, color 0.1s;
    flex-shrink: 0;
  }

  .sort-btn:hover {
    background: var(--bg-surface-hover);
    color: var(--text-primary);
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
    padding-top: 2px;
    display: flex;
    align-items: center;
    justify-content: center;
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
    border: none;
    border-left: 4px solid var(--border-default);
    background: var(--bg-inset);
    color: var(--text-primary);
    border-radius: 0 var(--radius-md) var(--radius-md) 0;
    padding: 10px 14px;
    display: grid;
    gap: 6px;
    text-align: left;
    cursor: pointer;
    transition: filter 0.12s ease;
  }

  .entry-row:hover {
    filter: brightness(0.97);
  }

  :global(.dark) .entry-row:hover {
    filter: brightness(1.15);
  }

  .entry-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .entry-icon {
    width: 20px;
    height: 20px;
    border-radius: 50%;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    font-size: 10px;
    font-weight: 700;
    color: white;
    background: var(--text-muted);
    flex-shrink: 0;
    line-height: 1;
  }

  .entry-label {
    font-size: 12px;
    font-weight: 600;
    letter-spacing: 0.01em;
    color: var(--text-secondary);
  }

  .entry-tool-name {
    font-family: var(--font-mono);
    font-size: 11px;
    color: var(--text-secondary);
    background: var(--bg-surface);
    border: 1px solid var(--border-muted);
    border-radius: var(--radius-sm);
    padding: 1px 6px;
  }

  .entry-ordinal {
    margin-left: auto;
    text-align: right;
  }

  .entry-preview {
    font-size: 12px;
    line-height: 1.45;
    color: var(--text-primary);
    padding-left: 28px;
  }

  .entry-user {
    background: var(--user-bg);
    border-left-color: var(--accent-blue);
  }
  .entry-user .entry-icon { background: var(--accent-blue); }
  .entry-user .entry-label { color: var(--accent-blue); }

  .entry-assistant {
    background: var(--assistant-bg);
    border-left-color: var(--accent-purple);
  }
  .entry-assistant .entry-icon { background: var(--accent-purple); }
  .entry-assistant .entry-label { color: var(--accent-purple); }

  .entry-tool {
    background: var(--tool-bg);
    border-left-color: var(--accent-amber);
  }
  .entry-tool .entry-icon { background: var(--accent-amber); }
  .entry-tool .entry-label { color: var(--accent-amber); }

  .tool-group {
    border-left: 4px solid var(--accent-amber);
    background: var(--tool-bg);
    border-radius: 0 var(--radius-md) var(--radius-md) 0;
    padding: 10px 14px;
    display: grid;
    gap: 6px;
  }

  .tool-group-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .gear-icon {
    flex-shrink: 0;
  }

  .group-label {
    font-size: 12px;
    font-weight: 600;
    color: var(--accent-amber);
  }

  .tool-group-body {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding-left: 20px;
  }

  .tool-row {
    width: 100%;
    border: none;
    background: transparent;
    color: var(--text-primary);
    padding: 4px 6px;
    border-radius: var(--radius-sm);
    display: grid;
    gap: 2px;
    text-align: left;
    cursor: pointer;
    font-family: var(--font-mono);
    font-size: 12px;
    line-height: 1.5;
    transition: background 0.1s;
  }

  .tool-row:hover {
    background: var(--bg-surface-hover);
  }

  .tool-line {
    display: flex;
    align-items: baseline;
    gap: 8px;
    min-width: 0;
  }

  .tool-chevron {
    color: var(--text-muted);
    flex-shrink: 0;
    width: 12px;
    text-align: center;
  }

  .tool-name {
    color: var(--text-secondary);
    font-weight: 600;
    flex-shrink: 0;
  }

  .tool-snippet {
    color: var(--text-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    min-width: 0;
    flex: 1;
  }

  .tool-output-line .tool-snippet {
    color: var(--text-secondary);
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
  }
</style>
