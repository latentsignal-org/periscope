<script lang="ts">
  import type {
    CompactSignal,
    SummaryCoverage,
  } from "../../api/types/context.js";
  import { formatTokenCount } from "../../utils/format.js";
  import GuidanceSuggestionBlock from "./GuidanceSuggestionBlock.svelte";

  interface Props {
    signal: CompactSignal;
    summaryCoverage?: SummaryCoverage;
  }

  let { signal, summaryCoverage }: Props = $props();

  let generatedHeadline = $derived.by(() => {
    const keep = signal.keep_items?.[0]?.trim();
    const drop = signal.drop_items?.[0]?.trim();
    if (keep && drop) return `Preserve ${keep}, drop ${drop}`;
    if (keep) return `Preserve ${keep}`;
    if (drop) return `Drop ${drop}`;
    return "";
  });
</script>

<section class="signal-card compact" class:high={signal.confidence === "high"} class:medium={signal.confidence === "medium"}>
  <div class="signal-icon">
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <polyline points="4 14 10 14 10 20" />
      <polyline points="20 10 14 10 14 4" />
      <line x1="14" y1="10" x2="21" y2="3" />
      <line x1="3" y1="21" x2="10" y2="14" />
    </svg>
  </div>

  <div class="signal-body">
    <div class="signal-top">
      <div class="signal-title-row">
        <h3>Compact recommended</h3>
        <div class="signal-badges">
          <span class="badge confidence">{signal.confidence}</span>
          <span class="badge score">{signal.score}/100</span>
        </div>
      </div>
      {#if generatedHeadline}
        <div class="focus-headline">{generatedHeadline}</div>
      {/if}
    </div>

    <ul class="reasons">
      {#each signal.reasons as reason}
        <li>{reason}</li>
      {/each}
    </ul>

    <GuidanceSuggestionBlock
      title="Suggested compact focus"
      text={signal.compact_focus_text}
      provenance={signal.focus_provenance}
      model={signal.focus_model}
      hint={summaryCoverage?.status === "idle"
        ? "Star to enable guidance text"
        : ""}
    />

    <div class="signal-footer">
      {#if signal.estimated_reclaimable > 0}
        <span class="metric">
          <strong>~{formatTokenCount(signal.estimated_reclaimable)}</strong> reclaimable
        </span>
      {/if}
      {#if signal.compact_focus && signal.compact_focus.length > 0}
        <span class="focus">
          Focus: {signal.compact_focus.join(", ")}
        </span>
      {/if}
    </div>
  </div>
</section>

<style>
  .signal-card {
    display: flex;
    gap: 14px;
    padding: 14px 16px;
    border-radius: var(--radius-lg);
    border: 1px solid var(--border-muted);
    background: var(--bg-surface);
    position: relative;
    overflow: hidden;
  }

  .signal-card::before {
    content: "";
    position: absolute;
    inset: 0;
    border-radius: inherit;
    pointer-events: none;
  }

  .signal-card.compact::before {
    background: linear-gradient(135deg,
      color-mix(in srgb, var(--accent-amber) 6%, transparent),
      color-mix(in srgb, var(--accent-orange) 3%, transparent));
  }

  .signal-card.high {
    border-color: color-mix(in srgb, var(--accent-orange) 40%, var(--border-muted));
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--accent-orange) 10%, transparent),
                var(--shadow-sm);
  }

  .signal-card.medium {
    border-color: color-mix(in srgb, var(--accent-amber) 35%, var(--border-muted));
  }

  .signal-icon {
    flex-shrink: 0;
    width: 40px;
    height: 40px;
    border-radius: var(--radius-md);
    display: grid;
    place-items: center;
  }

  .compact .signal-icon {
    background: color-mix(in srgb, var(--accent-amber) 14%, transparent);
    color: var(--accent-amber);
  }

  .signal-body {
    flex: 1;
    min-width: 0;
    display: grid;
    gap: 8px;
  }

  .signal-top {
    display: grid;
    gap: 4px;
  }

  .signal-title-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
  }

  h3 {
    margin: 0;
    font-size: 14px;
    font-weight: 700;
    color: var(--text-primary);
    letter-spacing: -0.01em;
  }

  .signal-badges {
    display: flex;
    gap: 5px;
    flex-shrink: 0;
  }

  .badge {
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    padding: 2px 7px;
    border-radius: 3px;
    line-height: 1.4;
  }

  .badge.confidence {
    background: color-mix(in srgb, var(--accent-amber) 14%, transparent);
    color: var(--accent-amber);
  }

  .high .badge.confidence {
    background: color-mix(in srgb, var(--accent-orange) 14%, transparent);
    color: var(--accent-orange);
  }

  .badge.score {
    background: var(--bg-inset);
    color: var(--text-muted);
    font-variant-numeric: tabular-nums;
  }

  .reasons {
    margin: 0;
    padding-left: 16px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--text-secondary);
  }

  .focus-headline {
    font-size: 12px;
    font-weight: 600;
    color: var(--accent-orange);
  }

  .reasons li {
    margin-bottom: 1px;
  }

  .signal-footer {
    display: flex;
    align-items: baseline;
    gap: 12px;
    flex-wrap: wrap;
    font-size: 11px;
    color: var(--text-muted);
    padding-top: 6px;
    border-top: 1px solid var(--border-muted);
  }

  .metric strong {
    font-weight: 700;
    color: var(--text-secondary);
    font-variant-numeric: tabular-nums;
  }

  .focus {
    font-style: italic;
  }
</style>
