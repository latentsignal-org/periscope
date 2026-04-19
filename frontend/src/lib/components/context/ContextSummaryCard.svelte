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

  function percentLabel(value: number): string {
    if (!Number.isFinite(value)) return "—";
    return `${Math.round(value)}%`;
  }

  function provenanceLabel(value: string): string {
    switch (value) {
      case "measured":
        return "Measured";
      case "inferred":
        return "Inferred";
      case "estimated":
        return "Estimated";
      default:
        return "Unknown";
    }
  }
</script>

<section class="context-card">
  <div class="header">
    <div>
      <div class="eyebrow">Context Summary</div>
      <h2>Visible context after latest compaction</h2>
    </div>
    <div class="occupancy">
      <div class="tokens">{formatTokenCount(summary.tokens_in_use)}</div>
      <div class="tokens-meta">
        {#if capacity.max_tokens > 0}
          of {formatTokenCount(capacity.max_tokens)} tokens
        {:else}
          capacity unknown
        {/if}
      </div>
    </div>
  </div>

  <div class="meter-shell">
    <div
      class="meter-fill"
      style={`width: ${Math.max(0, Math.min(100, summary.percent_consumed))}%`}
    ></div>
  </div>

  <div class="stats-grid">
    <div class="stat">
      <span class="label">Used</span>
      <strong>{percentLabel(summary.percent_consumed)}</strong>
      <span class="meta">{provenanceLabel(summary.tokens_provenance)}</span>
    </div>
    <div class="stat">
      <span class="label">Remaining</span>
      <strong>
        {#if summary.remaining_known}
          {formatTokenCount(summary.remaining_tokens)}
        {:else}
          —
        {/if}
      </strong>
      <span class="meta">{provenanceLabel(capacity.provenance)}</span>
    </div>
    <div class="stat">
      <span class="label">Window</span>
      <strong>
        {#if capacity.max_tokens > 0}
          {formatTokenCount(capacity.max_tokens)}
        {:else}
          Unknown
        {/if}
      </strong>
      <span class="meta">{provenanceLabel(capacity.provenance)}</span>
    </div>
    <div class="stat">
      <span class="label">Granularity</span>
      <strong>{summary.row_granularity}</strong>
      <span class="meta">Visible since row {summary.visible_since_ordinal}</span>
    </div>
  </div>

  <div class="session-meta">
    <span>{session?.agent ?? capacity.agent ?? "Unknown agent"}</span>
    {#if capacity.model}
      <span>{capacity.model}</span>
    {/if}
    {#if summary.last_updated_at}
      <span title={formatTimestamp(summary.last_updated_at)}>
        Updated {formatRelativeTime(summary.last_updated_at)}
      </span>
    {/if}
    {#if session?.ended_at}
      <span>Historical session</span>
    {:else}
      <span>Live session</span>
    {/if}
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
    border: 1px solid var(--border-default);
    background: var(--bg-surface);
    border-radius: 14px;
    padding: 18px;
    display: grid;
    gap: 14px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    align-items: end;
  }

  .eyebrow {
    font-size: 11px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--text-secondary);
    margin-bottom: 6px;
  }

  h2 {
    margin: 0;
    font-size: 20px;
    line-height: 1.2;
  }

  .occupancy {
    text-align: right;
  }

  .tokens {
    font-size: 28px;
    font-weight: 700;
    line-height: 1;
  }

  .tokens-meta {
    font-size: 12px;
    color: var(--text-secondary);
    margin-top: 4px;
  }

  .meter-shell {
    height: 12px;
    background: color-mix(in srgb, var(--bg-surface) 65%, #334155 35%);
    border-radius: 999px;
    overflow: hidden;
  }

  .meter-fill {
    height: 100%;
    background: linear-gradient(90deg, #0f766e, #ea580c);
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 12px;
  }

  .stat {
    padding: 12px;
    border-radius: 12px;
    background: color-mix(in srgb, var(--bg-surface) 88%, #0f172a 12%);
    border: 1px solid var(--border-default);
    display: grid;
    gap: 4px;
  }

  .label,
  .meta,
  .session-meta,
  .warning {
    font-size: 12px;
    color: var(--text-secondary);
  }

  .session-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }

  .warnings {
    display: grid;
    gap: 8px;
  }

  .warning {
    padding: 10px 12px;
    border-left: 3px solid #f59e0b;
    background: color-mix(in srgb, var(--bg-surface) 85%, #f59e0b 15%);
    border-radius: 8px;
  }

  @media (max-width: 900px) {
    .stats-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .header {
      flex-direction: column;
      align-items: start;
    }

    .occupancy {
      text-align: left;
    }
  }
</style>
