<script lang="ts">
  import type { CompactSignal } from "../../api/types/context.js";

  interface Props {
    signal: CompactSignal;
  }

  let { signal }: Props = $props();

  const confidenceColor: Record<string, string> = {
    high: "var(--accent-orange, #ea580c)",
    medium: "var(--accent-yellow, #e6a700)",
    low: "var(--text-muted)",
  };
</script>

<div class="compact-banner" style:border-left-color={confidenceColor[signal.confidence] ?? "var(--text-muted)"}>
  <div class="compact-header">
    <div class="compact-label">
      <span class="badge" class:high={signal.confidence === "high"} class:medium={signal.confidence === "medium"} class:low={signal.confidence === "low"}>
        Compact signal
      </span>
      <span class="confidence">{signal.confidence} confidence</span>
      <span class="score">score {signal.score}/100</span>
    </div>
    {#if signal.estimated_reclaimable > 0}
      <span class="reclaimable">
        ~{Math.round(signal.estimated_reclaimable / 1000)}k tokens reclaimable
      </span>
    {/if}
  </div>
  <ul class="reasons">
    {#each signal.reasons as reason}
      <li>{reason}</li>
    {/each}
  </ul>
  {#if signal.compact_focus && signal.compact_focus.length > 0}
    <div class="focus-section">
      <span class="focus-label">Focus areas:</span>
      {#each signal.compact_focus as focus}
        <span class="focus-tag">{focus}</span>
      {/each}
    </div>
  {/if}
</div>

<style>
  .compact-banner {
    background: var(--bg-surface);
    border: 1px solid var(--border-muted);
    border-left-width: 3px;
    border-radius: var(--radius-md);
    padding: 12px 14px;
    font-size: 12px;
    line-height: 1.5;
  }

  .compact-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .compact-label {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .badge {
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    padding: 2px 6px;
    border-radius: 3px;
  }

  .badge.high {
    background: color-mix(in srgb, var(--accent-orange, #ea580c) 20%, transparent);
    color: var(--accent-orange, #ea580c);
  }

  .badge.medium {
    background: color-mix(in srgb, var(--accent-yellow, #e6a700) 20%, transparent);
    color: var(--accent-yellow, #e6a700);
  }

  .badge.low {
    background: color-mix(in srgb, var(--text-muted) 15%, transparent);
    color: var(--text-muted);
  }

  .confidence {
    color: var(--text-secondary);
    font-weight: 500;
  }

  .score {
    color: var(--text-muted);
    font-variant-numeric: tabular-nums;
  }

  .reclaimable {
    color: var(--text-muted);
    font-variant-numeric: tabular-nums;
  }

  .reasons {
    margin: 8px 0 0;
    padding-left: 18px;
    color: var(--text-secondary);
  }

  .reasons li {
    margin-bottom: 2px;
  }

  .reasons li:last-child {
    margin-bottom: 0;
  }

  .focus-section {
    display: flex;
    align-items: center;
    gap: 6px;
    flex-wrap: wrap;
    margin-top: 8px;
    padding-top: 8px;
    border-top: 1px solid var(--border-muted);
  }

  .focus-label {
    color: var(--text-muted);
    font-size: 11px;
    font-weight: 500;
  }

  .focus-tag {
    font-size: 11px;
    padding: 1px 6px;
    border-radius: 3px;
    background: color-mix(in srgb, var(--accent-orange, #ea580c) 10%, transparent);
    color: var(--text-secondary);
  }
</style>
