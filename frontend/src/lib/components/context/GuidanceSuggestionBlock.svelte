<script lang="ts">
  import { copyToClipboard } from "../../utils/clipboard.js";

  interface Props {
    title: string;
    text?: string;
    provenance?: string;
    model?: string;
    hint?: string;
  }

  let {
    title,
    text = "",
    provenance = "",
    model = "",
    hint = "",
  }: Props = $props();

  let expanded = $state(false);
  let copied = $state(false);
  let copyTimer: ReturnType<typeof setTimeout> | undefined;

  const hasText = $derived(text.trim().length > 0);
  const modelLabel = $derived(formatModelLabel(model));
  const footerLabel = $derived(
    [provenance.trim(), modelLabel].filter(Boolean).join(" · "),
  );

  async function handleCopy() {
    if (!hasText) return;
    const ok = await copyToClipboard(text);
    if (!ok) return;
    copied = true;
    if (copyTimer) clearTimeout(copyTimer);
    copyTimer = setTimeout(() => {
      copied = false;
    }, 1500);
  }

  function formatModelLabel(value: string): string {
    return value.trim().replace(/-\d{8}$/, "");
  }
</script>

{#if hasText}
  <div class="suggestion-block">
    <div class="suggestion-header">
      <button
        type="button"
        class="toggle-btn"
        aria-expanded={expanded}
        onclick={() => (expanded = !expanded)}
      >
        <span class="chevron" class:expanded>{expanded ? "▾" : "▸"}</span>
        <span>{title}</span>
      </button>
      <button
        type="button"
        class="copy-btn"
        title={copied ? "Copied!" : `Copy ${title.toLowerCase()}`}
        aria-label={`Copy ${title.toLowerCase()}`}
        onclick={handleCopy}
      >
        {copied ? "Copied" : "Copy"}
      </button>
    </div>

    {#if expanded}
      <pre class="suggestion-body">{text}</pre>
      {#if footerLabel}
        <div class="suggestion-footer">{footerLabel}</div>
      {/if}
    {/if}
  </div>
{:else if hint}
  <div class="suggestion-hint">{hint}</div>
{/if}

<style>
  .suggestion-block {
    display: grid;
    gap: 8px;
    padding: 10px 12px;
    border: 1px solid var(--border-muted);
    border-radius: var(--radius-md);
    background: color-mix(in srgb, var(--bg-inset) 55%, var(--bg-surface));
  }

  .suggestion-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .toggle-btn,
  .copy-btn {
    border: none;
    background: transparent;
    color: inherit;
    cursor: pointer;
  }

  .toggle-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 0;
    font-size: 12px;
    font-weight: 700;
    color: var(--text-secondary);
  }

  .chevron {
    font-size: 12px;
    color: var(--text-muted);
  }

  .copy-btn {
    padding: 4px 8px;
    border-radius: var(--radius-sm, 4px);
    font-size: 11px;
    font-weight: 600;
    color: var(--text-muted);
  }

  .copy-btn:hover,
  .toggle-btn:hover {
    color: var(--text-primary);
  }

  .copy-btn:hover {
    background: var(--bg-surface-hover);
  }

  .suggestion-body {
    margin: 0;
    padding: 10px;
    border-radius: var(--radius-sm, 4px);
    background: color-mix(in srgb, var(--bg-surface) 70%, black 4%);
    color: var(--text-primary);
    font-family: var(--font-mono);
    font-size: 12px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .suggestion-footer,
  .suggestion-hint {
    font-size: 11px;
    color: var(--text-muted);
  }

  .suggestion-hint {
    margin-top: 2px;
  }
</style>
