<script lang="ts">
  import type { ContextCapacity, ContextSummary } from "../../api/types.js";
  import type { Session } from "../../api/types.js";
  import { formatRelativeTime, formatTimestamp, formatTokenCount } from "../../utils/format.js";

  interface Props {
    summary: ContextSummary;
    capacity: ContextCapacity;
    session?: Session | null;
    warnings?: string[];
  }

  let { summary, capacity, session = null, warnings = [] }: Props = $props();

  let subtitle = $derived.by(() => {
    if (summary.visible_since_ordinal > 0) {
      return "After compaction";
    }
    return "Full conversation";
  });

  function percentLabel(value: number): string {
    if (!Number.isFinite(value)) return "—";
    return `${Math.round(value)}%`;
  }

  let remainingPercent = $derived.by(() => {
    if (!summary.remaining_known || capacity.max_tokens <= 0) return NaN;
    return (summary.remaining_tokens / capacity.max_tokens) * 100;
  });

  function provenanceTooltip(value: string): string {
    switch (value) {
      case "measured":
        return "Measured from actual token counts";
      case "inferred":
        return "Inferred from available data";
      case "estimated":
        return "Estimated based on heuristics";
      default:
        return "";
    }
  }
</script>

<section class="context-card">
  <div class="header">
    <div class="header-left">
      <div class="eyebrow">Context Summary</div>
      <h2>{subtitle}</h2>
      <div class="session-pills">
        <span class="pill">{session?.agent ?? capacity.agent ?? "Unknown agent"}</span>
        {#if capacity.model}
          <span class="pill">{capacity.model}</span>
        {/if}
        {#if summary.last_updated_at}
          <span class="pill" title={formatTimestamp(summary.last_updated_at)}>
            Updated {formatRelativeTime(summary.last_updated_at)}
          </span>
        {/if}
        {#if session?.ended_at}
          <span class="pill pill-dim">Ended</span>
        {:else}
          <span class="pill pill-live">Live</span>
        {/if}
      </div>
    </div>
    <div class="occupancy">
      <div class="tokens">
        {formatTokenCount(summary.tokens_in_use)}
        {#if capacity.max_tokens > 0}
          <span class="tokens-of">of {formatTokenCount(capacity.max_tokens)}</span>
          <span class="tokens-pct">({percentLabel(summary.percent_consumed)})</span>
        {/if}
      </div>
      {#if capacity.max_tokens <= 0}
        <div class="tokens-meta">capacity unknown</div>
      {/if}
    </div>
  </div>

  <div class="meter-shell">
    <div
      class="meter-fill"
      style={`width: ${Math.max(0, Math.min(100, summary.percent_consumed))}%`}
    ></div>
  </div>

  <div class="stats-grid">
    <div class="stat" title={provenanceTooltip(summary.tokens_provenance)}>
      <span class="label">Used</span>
      <strong>
        {formatTokenCount(summary.tokens_in_use)}
        {#if capacity.max_tokens > 0}
          <span class="stat-pct">({percentLabel(summary.percent_consumed)})</span>
        {/if}
      </strong>
    </div>
    <div class="stat" title={provenanceTooltip(capacity.provenance)}>
      <span class="label">Remaining</span>
      <strong>
        {#if summary.remaining_known}
          {formatTokenCount(summary.remaining_tokens)}
          <span class="stat-pct">({percentLabel(remainingPercent)})</span>
        {:else}
          —
        {/if}
      </strong>
    </div>
  </div>

  {#if warnings.length > 0}
    <div class="warnings">
      {#each warnings as warning}
        <div class="warning">{warning}</div>
      {/each}
    </div>
  {/if}
</section>

<style>
  .context-card {
    border: 1px solid var(--border-muted);
    background: var(--bg-surface);
    border-radius: var(--radius-md);
    padding: 12px;
    display: grid;
    gap: 12px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    align-items: end;
  }

  .header-left {
    display: grid;
    gap: 4px;
  }

  .eyebrow {
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--text-muted);
  }

  h2 {
    margin: 0;
    font-size: 13px;
    font-weight: 600;
    line-height: 1.3;
    color: var(--text-primary);
  }

  .session-pills {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 2px;
  }

  .pill {
    font-size: 10px;
    font-weight: 500;
    color: var(--text-secondary);
    background: var(--bg-inset);
    border: 1px solid var(--border-muted);
    border-radius: 999px;
    padding: 1px 8px;
    white-space: nowrap;
  }

  .pill-live {
    color: var(--accent-teal);
    border-color: var(--accent-teal);
  }

  .pill-dim {
    color: var(--text-muted);
  }

  .occupancy {
    text-align: right;
    flex-shrink: 0;
  }

  .tokens {
    font-size: 20px;
    font-weight: 700;
    line-height: 1;
    color: var(--text-primary);
  }

  .tokens-meta {
    font-size: 11px;
    color: var(--text-muted);
    margin-top: 4px;
  }

  .tokens-of,
  .tokens-pct {
    font-size: 13px;
    font-weight: 500;
    color: var(--text-muted);
  }

  .stat-pct {
    font-weight: 500;
    color: var(--text-muted);
    margin-left: 2px;
  }

  .meter-shell {
    height: 8px;
    background: var(--bg-inset);
    border-radius: 999px;
    overflow: hidden;
  }

  .meter-fill {
    height: 100%;
    background: linear-gradient(
      90deg,
      var(--accent-teal),
      var(--accent-amber),
      var(--accent-rose)
    );
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }

  .stat {
    padding: 10px;
    border-radius: var(--radius-sm);
    background: var(--bg-inset);
    border: 1px solid var(--border-muted);
    display: grid;
    gap: 2px;
  }

  .stat strong {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .label,
  .warning {
    font-size: 11px;
    color: var(--text-muted);
  }

  .warnings {
    display: grid;
    gap: 6px;
  }

  .warning {
    padding: 8px 10px;
    border-left: 3px solid var(--accent-amber);
    background: color-mix(in srgb, var(--bg-surface) 85%, var(--accent-amber) 15%);
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
  }

  @media (max-width: 900px) {
    .header {
      flex-direction: column;
      align-items: start;
    }

    .occupancy {
      text-align: left;
    }
  }
</style>
