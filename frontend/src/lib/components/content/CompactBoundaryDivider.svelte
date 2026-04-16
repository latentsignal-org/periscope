<script lang="ts">
  import type { Message } from "../../api/types.js";
  import { formatTimestamp } from "../../utils/format.js";

  interface Props {
    message: Message;
  }

  let { message }: Props = $props();

  let preview = $derived.by(() => {
    const text = (message.content ?? "").trim();
    if (text.length === 0) return "";
    const firstLine = text.split("\n")[0]!;
    return firstLine.length > 140
      ? firstLine.slice(0, 140) + "…"
      : firstLine;
  });
</script>

<div class="boundary" title="Context window compacted at this point">
  <span class="boundary-line"></span>
  <span class="boundary-label">
    <span class="boundary-icon" aria-hidden="true">↻</span>
    Context compacted
    {#if message.timestamp}
      <span class="boundary-time">
        &middot; {formatTimestamp(message.timestamp)}
      </span>
    {/if}
  </span>
  <span class="boundary-line"></span>
</div>
{#if preview}
  <div class="boundary-preview">{preview}</div>
{/if}

<style>
  .boundary {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px 0;
    color: var(--accent-amber);
    font-size: 11px;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    font-weight: 600;
  }
  .boundary-line {
    flex: 1;
    height: 1px;
    background: color-mix(
      in srgb, var(--accent-amber) 35%, transparent
    );
  }
  .boundary-label {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
  }
  .boundary-icon {
    font-size: 13px;
    line-height: 1;
  }
  .boundary-time {
    color: var(--text-muted);
    font-weight: 500;
    text-transform: none;
    letter-spacing: 0;
  }
  .boundary-preview {
    margin: 4px 12px 6px;
    padding: 6px 10px;
    background: color-mix(
      in srgb, var(--accent-amber) 8%, transparent
    );
    border-left: 2px solid
      color-mix(in srgb, var(--accent-amber) 50%, transparent);
    color: var(--text-secondary);
    font-size: 12px;
    line-height: 1.5;
    border-radius: 0 var(--radius-sm, 4px)
      var(--radius-sm, 4px) 0;
  }
</style>
