<script lang="ts">
  import type { HeaderAction } from '$lib/types';
  import CacheManagerModal from './CacheManagerModal.svelte';

  let {
    response = '',
    isThinking = false,
    cacheCount = 0,
    ragSources = [],
    headerActions,
    onClearCache,
    onCopyAll,
  } = $props<{
    response?: string;
    isThinking?: boolean;
    cacheCount?: number;
    ragSources?: string[];
    headerActions?: HeaderAction[];
    onClearCache?: () => void;
    onCopyAll?: () => void;
  }>();

  let scrollEl: HTMLDivElement | undefined = $state();
  let userScrolledUp = $state(false);
  let renderedResponse = $state('');
  let copySuccess = $state(false);

  let showHistory = $state(false);
  let isCacheModalOpen = $state(false);
  let cacheItems = $state<any[]>([]);
  let localOverride = $state('');

  async function toggleHistory() {
    if (!showHistory) {
      try {
        const items = await (window as any).go.main.App.GetCacheItems();
        cacheItems = items || [];
      } catch (e) {
        console.error('Failed to fetch cache items:', e);
      }
    }
    showHistory = !showHistory;
  }

  // Clear localOverride when a new live response starts
  $effect(() => {
    if (isThinking) {
      localOverride = '';
    }
  });

  export function showLocalAnswer(answer: string) {
    localOverride = answer;
    showHistory = false;
  }

  export async function showCachedAnswerFor(questionText: string) {
    try {
      const items = await (window as any).go.main.App.GetCacheItems();
      const found = items.find((i: any) => (i.question || i.Question) === questionText);
      if (found) {
        localOverride = found.answer || found.Answer;
        showHistory = false;
      }
    } catch (e) {
      console.error('Failed to fetch cached answer:', e);
    }
  }

  // Dynamic markdown rendering
  let renderMarkdown: any;
  let markdownTimeout: ReturnType<typeof setTimeout>;

  $effect(() => {
    const current = localOverride || response;

    if (markdownTimeout) clearTimeout(markdownTimeout);

    markdownTimeout = setTimeout(async () => {
      if (current) {
        try {
          if (!renderMarkdown) {
            const m = await import('$lib/markdownRenderer');
            renderMarkdown = m.renderMarkdown;
          }
          // Append the cursor HTML before parsing so it sits inline inside the last paragraph
          let textToRender = current;
          if (isThinking) {
            textToRender += '<span class="streaming-cursor">▌</span>';
          }
          renderedResponse = await renderMarkdown(textToRender);
        } catch (e) {
          console.error("renderMarkdown error:", e);
          renderedResponse = current;
        }
      } else {
        renderedResponse = '';
      }
    }, 50);
  });

  // Inject copy buttons into <pre> elements after render
  $effect(() => {
    renderedResponse; // Dependency
    if (scrollEl) {
      setTimeout(() => {
        if (!scrollEl) return;
        const pres = scrollEl.querySelectorAll('pre');
        pres.forEach(pre => {
          // Avoid duplicate buttons
          if (!pre.querySelector('.copy-btn')) {
            const btn = document.createElement('button');
            btn.className = 'copy-btn text-[12px]';
            btn.textContent = 'Copy';
            btn.onclick = () => {
              const code = pre.querySelector('code')?.textContent || pre.textContent || '';
              // Remove "Copy" from the text if it was included in textContent
              const textToCopy = code.replace(/^Copy\n?/, '');
              navigator.clipboard.writeText(textToCopy).then(() => {
                btn.textContent = 'Copied!';
                setTimeout(() => {
                  btn.textContent = 'Copy';
                }, 1500);
              });
            };
            pre.style.position = 'relative';
            pre.appendChild(btn);
          }
        });
      }, 0);
    }
  });

  function onScroll() {
    if (!scrollEl) return;
    const atBottom = scrollEl.scrollHeight - scrollEl.scrollTop - scrollEl.clientHeight < 40;
    userScrolledUp = !atBottom;
  }

  $effect(() => {
    // Reactive dependency
    response;
    if (!userScrolledUp && scrollEl) {
      setTimeout(() => {
        if (scrollEl) {
          scrollEl.scrollTop = scrollEl.scrollHeight;
        }
      }, 0);
    }
  });

  function handleCopyAll() {
    if (response) {
      navigator.clipboard.writeText(response).then(() => {
        copySuccess = true;
        setTimeout(() => {
          copySuccess = false;
        }, 1500);
      });
      if (onCopyAll) onCopyAll();
    }
  }


  let exportStatus = $state<'idle' | 'saving' | 'done' | 'error'>('idle');

  async function handleExport() {
    exportStatus = 'saving';
    try {
      const path = await (window as any).go.main.App.ExportSession();
      if (!path) {
        exportStatus = 'idle'; // User cancelled
        return;
      }
      exportStatus = 'done';
      console.log('Session exported to:', path);
      setTimeout(() => exportStatus = 'idle', 3000);
    } catch (e) {
      console.error('Export failed:', e);
      exportStatus = 'error';
      setTimeout(() => exportStatus = 'idle', 3000);
    }
  }
</script>

<div class="answer-panel">
  <!-- Header -->
  <div class="panel-header">
    <div class="header-left">
      <span class="material-symbols-outlined text-[18px] text-primary">auto_awesome</span>
      <span class="title font-bold text-white">BarnOwl AI</span>
    </div>
    <div class="header-right">
      <button class="header-action-btn" onclick={handleExport}>
        <span class="material-symbols-outlined">
          {exportStatus === 'saving' ? 'hourglass_empty' : exportStatus === 'done' ? 'check' : 'download'}
        </span>
        <span class="action-label">Export</span>
      </button>
      <button class="header-action-btn" onclick={handleCopyAll}>
        <span class="material-symbols-outlined">
          {copySuccess ? 'check' : 'content_copy'}
        </span>
        <span class="action-label">Copy</span>
      </button>
      <button class="header-action-btn" onclick={toggleHistory} class:active={showHistory}>
        <span class="material-symbols-outlined">history</span>
        <span class="action-label">History</span>
      </button>
      <button class="header-action-btn relative" onclick={() => isCacheModalOpen = true}>
        <span class="material-symbols-outlined">mop</span>
        <span class="action-label">Cache</span>
        {#if cacheCount > 0}
          <span class="badge">{cacheCount}</span>
        {/if}
      </button>

      {#if headerActions}
        {#each headerActions as action}
          <button
            class="header-action-btn"
            class:active={action.active}
            onclick={action.onClick}
          >
            <span class="material-symbols-outlined">
              {action.active && action.activeIcon ? action.activeIcon : action.icon}
            </span>
            <span class="action-label">{action.label}</span>
          </button>
        {/each}
      {/if}
    </div>
  </div>

  <!-- Body -->
  <div class="panel-body" bind:this={scrollEl} onscroll={onScroll}>
    {#if showHistory}
      <div class="history-view">
        <h3 class="history-title">Cached Q&A History</h3>
        {#if cacheItems.length === 0}
          <p class="history-empty">No cached items found.</p>
        {:else}
          <div class="history-list">
            {#each cacheItems as item}
              <button class="history-item" onclick={() => {
                localOverride = item.answer || item.Answer;
                showHistory = false;
              }}>
                <div class="history-q">{item.question || item.Question}</div>
                <div class="history-meta">
                  <span>Tokens saved: ~{item.tokensSaved || item.TokensSaved || 250}</span>
                </div>
              </button>
            {/each}
          </div>
        {/if}
      </div>
    {:else}
      {#if !isThinking && !response && !localOverride}
        <!-- Empty State -->
        <div class="empty-state">
          <span class="material-symbols-outlined ghost-icon">auto_awesome</span>
          <p class="ghost-text">Waiting for a question…</p>
        </div>
      {:else if isThinking && !response && !localOverride}
        <!-- Thinking State -->
        <div class="thinking-state-container">
          <div class="thinking-card">
            <div class="dots-container">
              <span class="dot"></span>
              <span class="dot"></span>
              <span class="dot"></span>
            </div>
            <p>BarnOwl is thinking…</p>
          </div>
        </div>
      {:else}
        <!-- Response -->
        <div class="response-content">
          {@html renderedResponse}
        </div>
      {/if}
    {/if}
  </div>

  <!-- Footer -->
  {#if cacheCount > 0 || ragSources.length > 0}
    <div class="panel-footer">
      <div class="footer-left">
        {#if cacheCount > 0}
          <span class="pill">💾 {cacheCount} pairs cached</span>
        {/if}
      </div>
      <div class="footer-right">
        {#if ragSources.length > 0}
          <span class="pill">📎 {ragSources.length} context source{ragSources.length !== 1 ? 's' : ''}</span>
        {/if}
      </div>
    </div>
  {/if}
</div>

<CacheManagerModal
  isOpen={isCacheModalOpen}
  onClose={() => isCacheModalOpen = false}
  onCacheCleared={onClearCache}
/>

<style>
  .answer-panel {
    --answer-bg: rgba(18, 18, 18, 0.9);
    --answer-border: rgba(74, 222, 128, 0.25);
    --primary: #4ade80;
    --surface: rgba(20, 20, 20, 0.8);

    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    background: var(--answer-bg);
    border-top: 2px solid var(--answer-border);
  }

  /* Header */
  .panel-header {
    height: 44px;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    padding: 0 12px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .header-action-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    color: rgba(255, 255, 255, 0.6);
    background: transparent;
    border: none;
    cursor: pointer;
    transition: all 0.2s;
    padding: 2px 6px;
    min-width: 44px;
  }

  .header-action-btn .material-symbols-outlined {
    font-size: 16px;
    margin-bottom: 2px;
  }

  .header-action-btn:hover {
    color: rgba(255, 255, 255, 0.9);
    background: rgba(255, 255, 255, 0.1);
  }

  .header-action-btn.active {
    color: var(--primary);
    background: rgba(74, 222, 128, 0.1);
  }

  .action-label {
    font-size: 8px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.02em;
    line-height: 1;
  }

  .badge {
    position: absolute;
    top: -4px;
    right: -4px;
    background: var(--primary);
    color: black;
    font-size: 10px;
    font-weight: bold;
    border-radius: 50%;
    width: 14px;
    height: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  /* Body */
  .panel-body {
    flex-grow: 1;
    overflow-y: auto;
    padding: 16px 14px;
    color: rgba(255, 255, 255, 0.88);
    font-size: 14px;
    line-height: 1.65;
    user-select: text;
  }

  /* Empty State */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
  }

  .ghost-icon {
    font-size: 64px;
    opacity: 0.15;
    color: var(--primary);
    margin-bottom: 16px;
  }

  .ghost-text {
    opacity: 0.3;
  }

  /* Thinking State */
  .thinking-state-container {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
  }

  .thinking-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 24px;
    border-radius: 12px;
    background: linear-gradient(135deg, rgba(20, 40, 25, 0.6), rgba(15, 25, 18, 0.8));
    border: 1px solid rgba(74, 222, 128, 0.15);
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  }

  .dots-container {
    display: flex;
    gap: 6px;
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--primary);
    animation: thinking-bounce 1.4s infinite ease-in-out;
  }

  .dot:nth-child(1) { animation-delay: -0.32s; }
  .dot:nth-child(2) { animation-delay: -0.16s; }

  @keyframes thinking-bounce {
    0%, 80%, 100% { transform: scale(0); }
    40% { transform: scale(1); }
  }

  /* Markdown Rendering (Global within component) */
  .response-content {
    /* Required to enclose the content */
  }

  .response-content :global(h1),
  .response-content :global(h2),
  .response-content :global(h3) {
    color: white;
    margin-bottom: 8px;
  }

  .response-content :global(p) {
    margin-bottom: 12px;
  }

  .response-content :global(code:not(pre code)) {
    background: rgba(255, 255, 255, 0.08);
    border-radius: 4px;
    padding: 1px 5px;
    font-family: monospace;
    font-size: 12px;
  }

  .response-content :global(pre) {
    background: rgba(0, 0, 0, 0.4);
    border-radius: 8px;
    border: 1px solid rgba(255, 255, 255, 0.08);
    padding: 12px;
    overflow-x: auto;
    position: relative;
    margin-bottom: 12px;
  }

  .response-content :global(.code-block-wrapper) {
      position: relative;
  }

  /* Styling specific for markdown renderer's HTML output */
  .response-content :global(.code-block-header) {
      display: flex;
      justify-content: space-between;
      margin-bottom: 8px;
      font-size: 12px;
      color: rgba(255,255,255,0.5);
  }

  .response-content :global(pre .copy-btn),
  .response-content :global(pre .code-copy-btn) {
    position: absolute;
    top: 6px;
    right: 6px;
    height: 24px;
    background: transparent;
    border: none;
    color: rgba(255, 255, 255, 0.5);
    cursor: pointer;
    font-size: 12px;
    padding: 2px 6px;
    border-radius: 4px;
  }

  .response-content :global(pre .copy-btn:hover),
  .response-content :global(pre .code-copy-btn:hover) {
    color: var(--primary);
  }

  .response-content :global(ul),
  .response-content :global(ol) {
    margin-left: 20px;
    margin-bottom: 12px;
  }

  .response-content :global(blockquote) {
    border-left: 3px solid rgba(74, 222, 128, 0.4);
    padding-left: 12px;
    color: rgba(255, 255, 255, 0.6);
    margin-bottom: 12px;
  }

  .response-content :global(.streaming-cursor) {
    display: inline-block;
    animation: blink 1s step-end infinite;
    color: var(--primary);
  }

  @keyframes blink {
    0%, 100% { opacity: 1; }
    50% { opacity: 0; }
  }

  /* Footer */
  .panel-footer {
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 12px;
    border-top: 1px solid rgba(255, 255, 255, 0.06);
  }

  .pill {
    font-size: 10px;
    background: rgba(255, 255, 255, 0.06);
    border-radius: 12px;
    padding: 2px 8px;
    color: rgba(255, 255, 255, 0.7);
  }

  /* History View */
  .history-view {
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .history-title {
    font-size: 14px;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.9);
    margin-bottom: 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    padding-bottom: 8px;
  }

  .history-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .history-item {
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    padding: 12px;
    text-align: left;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .history-item:hover {
    background: rgba(255, 255, 255, 0.08);
    border-color: rgba(74, 222, 128, 0.3);
  }

  .history-q {
    font-size: 13px;
    color: rgba(255, 255, 255, 0.9);
    line-height: 1.4;
  }

  .history-meta {
    font-size: 11px;
    color: rgba(74, 222, 128, 0.7);
  }

  .history-empty {
    font-size: 13px;
    color: rgba(255, 255, 255, 0.5);
    font-style: italic;
  }
</style>
