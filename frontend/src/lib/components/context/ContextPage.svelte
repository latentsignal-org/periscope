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
  import ContextCompositionChart from "./ContextCompositionChart.svelte";
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

<div class="context-page">
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
    <ContextCompositionChart composition={summaryData.composition} />
    <ContextTimeline timeline={timelineData.timeline} {sessionId} />
  {/if}
</div>

<style>
  .context-page {
    display: grid;
    gap: 18px;
    padding: 18px;
  }

  .context-page-header {
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

  h1 {
    margin: 0;
    font-size: 24px;
  }

  .actions {
    display: flex;
    gap: 10px;
  }

  .ghost-btn {
    border: 1px solid var(--border-default);
    background: var(--bg-surface);
    color: var(--text-primary);
    border-radius: 10px;
    padding: 9px 12px;
    cursor: pointer;
  }

  .empty {
    border: 1px dashed var(--border-default);
    background: var(--bg-surface);
    color: var(--text-secondary);
    border-radius: 14px;
    padding: 30px;
    text-align: center;
  }

  .empty.error {
    color: #dc2626;
  }

  @media (max-width: 900px) {
    .context-page-header {
      flex-direction: column;
      align-items: start;
    }
  }
</style>
