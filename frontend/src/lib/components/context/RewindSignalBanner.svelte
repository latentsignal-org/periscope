<script lang="ts">
  import type {
    RewindSignal,
    SummaryCoverage,
  } from "../../api/types/context.js";
  import { formatTokenCount } from "../../utils/format.js";
  import GuidanceSuggestionBlock from "./GuidanceSuggestionBlock.svelte";

  interface Props {
    signal: RewindSignal;
    summaryCoverage?: SummaryCoverage;
  }

  let { signal, summaryCoverage }: Props = $props();
</script>

<section class="signal-card rewind" class:high={signal.confidence === "high"} class:medium={signal.confidence === "medium"}>
  <div class="signal-icon">
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <polyline points="1 4 1 10 7 10" />
      <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10" />
    </svg>
  </div>

  <div class="signal-body">
    <div class="signal-top">
      <div class="signal-title-row">
        <h3>Rewind recommended</h3>
        <div class="signal-badges">
          <span class="badge confidence">{signal.confidence}</span>
          <span class="badge score">{signal.score}/100</span>
        </div>
      </div>

      {#if signal.rewind_to_turn}
        <div class="target">
          <span class="target-label">Rewind to turn {signal.rewind_to_turn}</span>
          {#if signal.tangent_label}
            <span class="target-meta">&mdash; drop the {signal.tangent_label}</span>
          {:else if signal.bad_stretch_from && signal.bad_stretch_to}
            <span class="target-meta">
              {#if signal.bad_stretch_from === signal.bad_stretch_to}
                &mdash; drop turn {signal.bad_stretch_to}
              {:else}
                &mdash; drop turns {signal.bad_stretch_from}&ndash;{signal.bad_stretch_to}
              {/if}
            </span>
          {/if}
        </div>
      {/if}
    </div>

    <ul class="reasons">
      {#each signal.reasons as reason}
        <li>{reason}</li>
      {/each}
    </ul>

    <GuidanceSuggestionBlock
      title="Suggested rewind prompt"
      text={signal.rewind_reprompt_text}
      provenance={signal.reprompt_provenance}
      model={signal.reprompt_model}
      hint={summaryCoverage?.status === "idle"
        ? "Star to enable guidance text"
        : ""}
    />

    <div class="signal-footer">
      {#if signal.tokens_recoverable > 0}
        <span class="metric">
          <strong>{formatTokenCount(signal.tokens_recoverable)}</strong> recoverable
        </span>
      {/if}
      {#if signal.rewind_to_reason}
        <span class="detail">{signal.rewind_to_reason}</span>
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

  .signal-card.rewind::before {
    background: linear-gradient(135deg,
      color-mix(in srgb, var(--accent-red) 6%, transparent),
      color-mix(in srgb, var(--accent-orange) 3%, transparent));
  }

  .signal-card.high {
    border-color: color-mix(in srgb, var(--accent-red) 40%, var(--border-muted));
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--accent-red) 10%, transparent),
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

  .rewind .signal-icon {
    background: color-mix(in srgb, var(--accent-red) 12%, transparent);
    color: var(--accent-red);
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
    background: color-mix(in srgb, var(--accent-red) 14%, transparent);
    color: var(--accent-red);
  }

  .medium .badge.confidence {
    background: color-mix(in srgb, var(--accent-amber) 14%, transparent);
    color: var(--accent-amber);
  }

  .badge.score {
    background: var(--bg-inset);
    color: var(--text-muted);
    font-variant-numeric: tabular-nums;
  }

  .target {
    display: flex;
    align-items: baseline;
    gap: 6px;
    flex-wrap: wrap;
  }

  .target-label {
    font-size: 13px;
    font-weight: 600;
    color: var(--accent-blue);
  }

  .target-meta {
    font-size: 12px;
    color: var(--text-muted);
  }

  .reasons {
    margin: 0;
    padding-left: 16px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--text-secondary);
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

  .detail {
    font-style: italic;
  }
</style>
