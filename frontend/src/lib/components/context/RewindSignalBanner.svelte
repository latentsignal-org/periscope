<script lang="ts">
  import type { RewindSignal } from "../../api/types/context.js";

  interface Props {
    signal: RewindSignal;
  }

  let { signal }: Props = $props();

  const confidenceColor: Record<string, string> = {
    high: "var(--accent-red)",
    medium: "var(--accent-yellow, #e6a700)",
    low: "var(--text-muted)",
  };
</script>

<div class="rewind-banner" style:border-left-color={confidenceColor[signal.confidence] ?? "var(--text-muted)"}>
  <div class="rewind-header">
    <div class="rewind-label">
      <span class="badge" class:high={signal.confidence === "high"} class:medium={signal.confidence === "medium"} class:low={signal.confidence === "low"}>
        Rewind signal
      </span>
      <span class="confidence">{signal.confidence} confidence</span>
      <span class="score">score {signal.score}/100</span>
    </div>
    {#if signal.tokens_recoverable > 0}
      <span class="recoverable">
        ~{Math.round(signal.tokens_recoverable / 1000)}k tokens recoverable
      </span>
    {/if}
  </div>
  {#if signal.rewind_to_turn}
    <div class="rewind-target">
      <span class="target-arrow">&#x21B6;</span>
      <span class="target-text">
        Rewind to turn {signal.rewind_to_turn}
      </span>
      {#if signal.bad_stretch_from && signal.bad_stretch_to}
        <span class="bad-stretch">
          {#if signal.bad_stretch_from === signal.bad_stretch_to}
            — turn {signal.bad_stretch_to} is problematic
          {:else}
            — turns {signal.bad_stretch_from}–{signal.bad_stretch_to} are problematic
          {/if}
        </span>
      {/if}
    </div>
  {/if}
  <ul class="reasons">
    {#each signal.reasons as reason}
      <li>{reason}</li>
    {/each}
    {#if signal.rewind_to_reason}
      <li class="target-reason">{signal.rewind_to_reason}</li>
    {/if}
  </ul>
</div>

<style>
  .rewind-banner {
    background: var(--bg-surface);
    border: 1px solid var(--border-muted);
    border-left-width: 3px;
    border-radius: var(--radius-md);
    padding: 12px 14px;
    font-size: 12px;
    line-height: 1.5;
  }

  .rewind-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }

  .rewind-label {
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
    color: var(--text-primary);
  }

  .badge.high {
    background: color-mix(in srgb, var(--accent-red) 20%, transparent);
    color: var(--accent-red);
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

  .recoverable {
    color: var(--text-muted);
    font-variant-numeric: tabular-nums;
  }

  .rewind-target {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: 10px;
    padding: 6px 10px;
    background: color-mix(in srgb, var(--accent-blue, #3b82f6) 8%, transparent);
    border-radius: var(--radius-sm);
    font-size: 12px;
  }

  .target-arrow {
    font-size: 14px;
    color: var(--accent-blue, #3b82f6);
  }

  .target-text {
    font-weight: 600;
    color: var(--text-primary);
  }

  .bad-stretch {
    color: var(--text-muted);
  }

  .reasons {
    margin: 8px 0 0;
    padding-left: 18px;
    color: var(--text-secondary);
  }

  .target-reason {
    color: var(--text-muted);
    font-style: italic;
  }

  .reasons li {
    margin-bottom: 2px;
  }

  .reasons li:last-child {
    margin-bottom: 0;
  }
</style>
