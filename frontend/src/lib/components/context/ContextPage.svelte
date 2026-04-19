<script lang="ts">
  import { onDestroy } from "svelte";
  import {
    getSession,
    getSessionContext,
    getSessionContextTimeline,
    watchSession,
    enqueueSummarize,
  } from "../../api/client.js";
  import type {
    Session,
    SessionContextResponse,
    SessionContextTimelineResponse,
  } from "../../api/types.js";
  import { router } from "../../stores/router.svelte.js";
  import ContextSummaryCard from "./ContextSummaryCard.svelte";
  import ContextWindowBlocks from "./ContextWindowBlocks.svelte";
  import ContextTimeline from "./ContextTimeline.svelte";
  import RewindSignalBanner from "./RewindSignalBanner.svelte";
  import CompactSignalBanner from "./CompactSignalBanner.svelte";

  interface Props {
    sessionId: string;
    embedded?: boolean;
    session?: Session | null;
  }

  let { sessionId, embedded = false, session = null }: Props = $props();

  let summaryData: SessionContextResponse | null = $state(null);
  let timelineData: SessionContextTimelineResponse | null = $state(null);
  let sessionData: Session | null = $state(session);
  let loading = $state(true);
  let error = $state("");
  let summarizing = $state(false);
  let summarizeError = $state("");
  let loadVersion = 0;
  let watcher: EventSource | null = null;

  async function load() {
    if (!sessionId) return;
    const version = ++loadVersion;
    loading = true;
    error = "";
    try {
      const [summary, timeline, maybeSession] = await Promise.all([
        getSessionContext(sessionId),
        getSessionContextTimeline(sessionId),
        session ? Promise.resolve(session) : getSession(sessionId),
      ]);
      if (loadVersion !== version) return;
      summaryData = summary;
      timelineData = timeline;
      sessionData = maybeSession;
    } catch (err) {
      if (loadVersion !== version) return;
      error = err instanceof Error ? err.message : "Failed to load context view";
    } finally {
      if (loadVersion === version) {
        loading = false;
      }
    }
  }

  async function triggerSummarize() {
    summarizing = true;
    summarizeError = "";
    try {
      await enqueueSummarize(sessionId);
    } catch (err) {
      summarizeError = err instanceof Error ? err.message : "Failed to enqueue";
    } finally {
      summarizing = false;
    }
  }

  $effect(() => {
    sessionData = session;
  });

  $effect(() => {
    sessionId;
    load();
  });

  $effect(() => {
    watcher?.close();
    if (!sessionId) return;
    watcher = watchSession(sessionId, () => {
      load();
    });
  });

  onDestroy(() => {
    watcher?.close();
  });
</script>

<div class="context-page" class:embedded>
  {#if !embedded}
    <div class="context-page-header">
      <div>
        <div class="eyebrow">Periscope</div>
        <h1>{sessionData?.display_name ?? sessionData?.project ?? sessionId}</h1>
      </div>
      <div class="actions">
        <button class="ghost-btn" onclick={() => router.navigate("sessions")}>
          Sessions
        </button>
        <button class="ghost-btn" onclick={() => router.navigateToSession(sessionId)}>
          Transcript
        </button>
      </div>
    </div>
  {/if}

  {#if loading}
    <div class="empty">Loading context…</div>
  {:else if error}
    <div class="empty error">{error}</div>
  {:else if summaryData && timelineData}
    {#if summaryData.rewind_signal || summaryData.compact_signal || summaryData.summary_coverage}
      <div class="signals-group">
        <div class="signals-header">
          <div class="signals-eyebrow">Context Guidance</div>
          {#if summaryData.summary_coverage}
            {@const sc = summaryData.summary_coverage}
            <span
              class="coverage-pill"
              class:disabled={sc.status === "disabled"}
              class:idle={sc.status === "idle"}
              class:pending={sc.status === "pending"}
              class:complete={sc.status === "complete"}
                  title={sc.status === "disabled"
                ? "Set ANTHROPIC_API_KEY to enable turn summaries"
                : sc.status === "idle"
                  ? "Star this session to generate turn summaries"
                  : `${sc.summarised_turns} of ${sc.total_turns} turns summarized`}
            >
              {#if sc.status === "disabled"}
                Summaries · disabled
              {:else if sc.status === "idle"}
                Star to summarize
              {:else if sc.status === "pending"}
                Summarizing · {sc.summarised_turns}/{sc.total_turns}
              {:else}
                Summaries · {sc.total_turns}/{sc.total_turns}
              {/if}
            </span>
            {#if sc.status === "idle" || sc.status === "pending"}
              <button
                class="summarize-btn"
                onclick={triggerSummarize}
                disabled={summarizing}
                title="Enqueue this session for turn summarization"
              >
                {summarizing ? "Queued…" : "Generate summaries"}
              </button>
            {/if}
            {#if summarizeError}
              <span class="summarize-error">{summarizeError}</span>
            {/if}
          {/if}
        </div>
        {#if summaryData.rewind_signal}
          <RewindSignalBanner
            signal={summaryData.rewind_signal}
            summaryCoverage={summaryData.summary_coverage}
          />
        {/if}
        {#if summaryData.compact_signal}
          <CompactSignalBanner
            signal={summaryData.compact_signal}
            summaryCoverage={summaryData.summary_coverage}
          />
        {/if}
      </div>
    {/if}
    <ContextSummaryCard
      summary={summaryData.summary}
      capacity={summaryData.capacity}
      session={sessionData}
      warnings={summaryData.warnings ?? []}
    />
    <ContextWindowBlocks
      summary={summaryData.summary}
      capacity={summaryData.capacity}
      timeline={timelineData.timeline}
      composition={summaryData.composition}
    />
    <ContextTimeline timeline={timelineData.timeline} {sessionId} />
  {/if}
</div>

<style>
  .context-page {
    display: grid;
    gap: 16px;
    padding: 16px;
  }

  .context-page.embedded {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    align-content: start;
  }

  .context-page-header {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    align-items: end;
  }

  .eyebrow {
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--text-muted);
    margin-bottom: 4px;
  }

  h1 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .actions {
    display: flex;
    gap: 8px;
  }

  .ghost-btn {
    border: 1px solid var(--border-muted);
    background: var(--bg-surface);
    color: var(--text-secondary);
    border-radius: var(--radius-sm);
    padding: 4px 10px;
    height: 24px;
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.1s, color 0.1s;
  }

  .ghost-btn:hover {
    background: var(--bg-surface-hover);
    color: var(--text-primary);
  }

  .signals-group {
    display: grid;
    gap: 10px;
  }

  .signals-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .signals-eyebrow {
    font-size: 10px;
    font-weight: 600;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--text-muted);
  }

  .coverage-pill {
    font-size: 10px;
    font-weight: 500;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    padding: 2px 8px;
    border-radius: 999px;
    border: 1px solid var(--border-muted);
    background: var(--bg-surface);
    color: var(--text-muted);
    white-space: nowrap;
  }

  .coverage-pill.pending {
    color: var(--text-primary);
    border-color: var(--accent-blue, #3b82f6);
  }

  .coverage-pill.complete {
    color: var(--accent-green, #10b981);
    border-color: var(--accent-green, #10b981);
  }

  .coverage-pill.idle {
    color: var(--text-secondary);
  }

  .summarize-btn {
    font-size: 11px;
    font-weight: 500;
    padding: 2px 10px;
    border-radius: 999px;
    border: 1px solid var(--accent-teal);
    color: var(--accent-teal);
    background: transparent;
    cursor: pointer;
    white-space: nowrap;
  }

  .summarize-btn:hover:not(:disabled) {
    background: color-mix(in srgb, var(--bg-surface) 85%, var(--accent-teal) 15%);
  }

  .summarize-btn:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .summarize-error {
    font-size: 11px;
    color: var(--accent-rose);
  }

  .empty {
    border: 1px dashed var(--border-muted);
    background: var(--bg-surface);
    color: var(--text-secondary);
    border-radius: var(--radius-md);
    padding: 24px;
    text-align: center;
    font-size: 13px;
  }

  .empty.error {
    color: var(--accent-red);
  }

  @media (max-width: 900px) {
    .context-page-header {
      flex-direction: column;
      align-items: start;
    }
  }
</style>
