<script lang="ts">
  import { onDestroy } from "svelte";
  import {
    getSession,
    getSessionContext,
    getSessionContextTimeline,
    watchSession,
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
        <div class="eyebrow">Periscope V1</div>
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
