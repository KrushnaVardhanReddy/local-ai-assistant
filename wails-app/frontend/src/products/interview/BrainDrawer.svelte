<script lang="ts">
  import { wsState, sendChat } from "$lib/ws.svelte";
  import { renderMarkdown } from "$lib/markdownRenderer";
  import { onMount } from "svelte";
  import { apiFetch, getApiUrl } from "$lib/api";

  let { showHotkeys = false, onToggleHotkeys }: { showHotkeys?: boolean; onToggleHotkeys?: () => void } = $props();

  let responseEl: HTMLElement | undefined = $state();
  let renderedResponse = $state('');

  let cacheStats = $state<{ cached_pairs: number; estimated_tokens_saved: number } | null>(null);
  let clearingCache = $state(false);

  let starPrimed = $state(false);
  let starPrimedTimer: ReturnType<typeof setTimeout> | null = null;

  async function fetchCacheStats() {
    try {
      if ((window as any)?.go?.main?.App?.GetCacheStats) {
        cacheStats = await (window as any).go.main.App.GetCacheStats();
        return;
      }
      const apiUrl = getApiUrl();
      const res = await apiFetch(`${apiUrl}/api/cache/stats`);
      if (res.ok) {
        cacheStats = await res.json();
      } else {
        cacheStats = { cached_pairs: 0, estimated_tokens_saved: 0 };
      }
    } catch {
      cacheStats = { cached_pairs: 0, estimated_tokens_saved: 0 };
    }
  }

  async function clearCache() {
    clearingCache = true;
    try {
      if ((window as any)?.go?.main?.App?.ClearCache) {
        await (window as any).go.main.App.ClearCache();
      } else {
        const apiUrl = getApiUrl();
        await apiFetch(`${apiUrl}/api/cache/clear`, { method: "POST" });
      }
      await fetchCacheStats();
    } catch (e) {
      console.error("Failed to clear cache", e);
    } finally {
      clearingCache = false;
    }
  }

  function activateStarMethod() {
    sendChat("Format your next response using the STAR method (Situation, Task, Action, Result).");
    wsState.isThinking = true;
    starPrimed = true;
    if (starPrimedTimer) clearTimeout(starPrimedTimer);
    starPrimedTimer = setTimeout(() => {
      starPrimed = false;
    }, 15000);
  }

  function catchMeUp() {
    sendChat("Please catch me up on the current context of the interview or conversation.");
    wsState.isThinking = true;
  }

  onMount(() => {
    fetchCacheStats();
  });

  let renderTimer: any = null;

  // Keep response rendered and scrolled to bottom
  $effect(() => {
    const current = wsState.response;
    if (renderTimer) clearTimeout(renderTimer);
    renderTimer = setTimeout(async () => {
      if (current) {
        try {
          renderedResponse = await renderMarkdown(current);
        } catch (e) {
          console.error("renderMarkdown error:", e);
          renderedResponse = current;
        }
      } else {
        renderedResponse = '';
      }

      if (responseEl) {
        setTimeout(() => {
          if (responseEl) {
            responseEl.scrollTop = responseEl.scrollHeight;
          }
        }, 0);
      }
    }, 50);
  });

</script>

<div class="brain-drawer">
  <!-- Status & Actions Header -->
  <div class="p-3 border-b border-white/10 flex flex-col gap-2">
    <div class="flex justify-between items-center">
      <span class="text-sm font-bold text-primary flex items-center gap-1">
        <span class="material-symbols-outlined text-[16px]">auto_awesome</span>
        The Brain
      </span>
      <div class="flex gap-1">
        <button
          class="action-btn text-xs"
          onclick={activateStarMethod}
          title="Format next answer with STAR method"
          class:star-primed={starPrimed}
        >
          <span class="material-symbols-outlined text-[14px]">star</span>
          STAR
        </button>
        <button
          class="action-btn text-xs"
          onclick={catchMeUp}
          title="Catch me up"
        >
          <span class="material-symbols-outlined text-[14px]">history</span>
        </button>
        <button
          class="action-btn text-xs"
          class:active-view={showHotkeys}
          onclick={() => {
            if (onToggleHotkeys) {
              onToggleHotkeys();
            } else {
              showHotkeys = !showHotkeys;
            }
          }}
          title={showHotkeys ? "Show AI Response" : "Show Keyboard Shortcuts"}
          data-testid="brain-drawer-keys-toggle"
        >
          <span class="material-symbols-outlined text-[14px]">keyboard</span>
          Keys
        </button>
      </div>
    </div>

    {#if starPrimed}
      <div class="text-[10px] text-primary/80 uppercase tracking-wide font-bold animate-pulse text-right">
        STAR Method Primed
      </div>
    {/if}

    <!-- RAG Sources Indicator -->
    {#if wsState.ragSources && wsState.ragSources.length > 0}
      <div class="flex items-center gap-1 mt-1 text-[10px] text-on-surface-variant">
        <span class="material-symbols-outlined text-[12px]">library_books</span>
        <span>{wsState.ragSources.length} context item(s)</span>
      </div>
    {/if}
  </div>

  <!-- Content Area (Hotkeys or AI Response) -->
  <div class="flex-1 overflow-y-auto p-4 hide-scrollbar" bind:this={responseEl}>
    {#if showHotkeys}
      <div class="hotkeys-container" data-testid="brain-hotkeys-view">
        <div class="flex items-center justify-between pb-3 mb-3 border-b border-white/10">
          <h3 class="text-sm font-bold text-on-background flex items-center gap-2">
            <span class="material-symbols-outlined text-primary text-[18px]">keyboard</span>
            Hotkeys Cheat Sheet
          </h3>
          <span class="text-[11px] text-on-surface-variant">Press Keys to return</span>
        </div>
        <div class="hotkeys-list">
          <div class="hotkey-row">
            <span>Toggle Hotkeys Panel</span>
            <kbd>Ctrl+/</kbd>
          </div>
          <div class="hotkey-row">
            <span>Push to Talk</span>
            <kbd>Ctrl+Shift+Space</kbd>
          </div>
          <div class="hotkey-row">
            <span>Screenshot Vision</span>
            <kbd>Ctrl+Shift+S</kbd>
          </div>
          <div class="hotkey-row">
            <span>Send Transcript Chip 1-6</span>
            <kbd>Ctrl+Shift+1...6</kbd>
          </div>
          <div class="hotkey-row">
            <span>Scroll Answer Down / Up</span>
            <kbd>Ctrl+Shift+↓/↑</kbd>
          </div>
          <div class="hotkey-row">
            <span>Toggle Stealth Click-through</span>
            <kbd>Ctrl+Alt+M</kbd>
          </div>
          <div class="hotkey-row">
            <span>Session Report</span>
            <kbd>Ctrl+Shift+E</kbd>
          </div>
          <div class="hotkey-row">
            <span>Panic Clear / Hide</span>
            <kbd>Ctrl+Shift+X</kbd>
          </div>
        </div>
      </div>
    {:else if wsState.isThinking && !wsState.response}
      <div class="flex items-center gap-2 text-primary/60 text-sm">
        <span class="material-symbols-outlined animate-spin text-[16px]">progress_activity</span>
        Thinking...
      </div>
    {:else if wsState.response}
      <div class="response-content">
        {@html renderedResponse}
      </div>
    {:else}
      <div class="h-full flex flex-col items-center justify-center text-center text-on-surface-variant opacity-50 p-4">
        <span class="material-symbols-outlined text-[48px] mb-4">psychiatry</span>
        <p class="text-sm">The Brain is listening...</p>
      </div>
    {/if}
  </div>

  <!-- Cache Stats Footer -->
  <div class="p-3 border-t border-white/10 bg-surface/30">
    <div class="flex justify-between items-center text-xs text-on-surface-variant">
      <div class="flex items-center gap-2">
        <span class="material-symbols-outlined text-[14px]">memory</span>
        <span>
          {#if cacheStats}
            {cacheStats.cached_pairs} items
          {:else}
            ...
          {/if}
        </span>
      </div>
      <button
        class="hover:text-error transition-colors flex items-center gap-1"
        onclick={clearCache}
        disabled={clearingCache}
        title="Clear cache"
      >
        {#if clearingCache}
          <span class="material-symbols-outlined text-[14px] animate-spin">progress_activity</span>
        {:else}
          <span class="material-symbols-outlined text-[14px]">mop</span>
        {/if}
      </button>
    </div>
  </div>
</div>

<style>
  .brain-drawer {
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
    background: transparent;
  }

  .hide-scrollbar::-webkit-scrollbar {
    display: none;
  }
  .hide-scrollbar {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }

  .action-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 4px;
    color: var(--hud-text-gray, #aaaaaa);
    transition: all 0.2s;
  }

  .action-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: white;
  }

  .action-btn.active-view {
    background: rgba(74, 222, 128, 0.2);
    border-color: rgba(74, 222, 128, 0.5);
    color: #4ade80;
  }

  .action-btn.star-primed {
    background: rgba(74, 222, 128, 0.2);
    border-color: rgba(74, 222, 128, 0.5);
    color: #4ade80;
  }

  .hotkeys-container {
    padding: 0.5rem 0;
  }

  .hotkeys-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    padding: 0.75rem 1rem;
  }

  .hotkey-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.45rem 0;
    font-size: 0.82rem;
    color: rgba(255, 255, 255, 0.85);
    border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  }

  .hotkey-row:last-child {
    border-bottom: none;
  }

  .hotkey-row kbd {
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 4px;
    padding: 0.2rem 0.5rem;
    font-family: 'JetBrains Mono', 'Fira Code', monospace;
    font-size: 0.75rem;
    color: #4ade80;
  }

  .response-content {
      font-size: 0.95rem;
      line-height: 1.6;
      color: rgba(240, 240, 248, 0.95);
  }

  /* Shiki code block overrides — match our dark glass theme */
  .response-content :global(.code-block-wrapper) {
    position: relative;
    margin: 1rem 0;
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.12);
  }

  .response-content :global(.code-block-header) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.3rem 0.8rem;
    background: rgba(255, 255, 255, 0.06);
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    font-size: 0.7rem;
    font-family: 'JetBrains Mono', 'Fira Code', monospace;
    color: rgba(255, 255, 255, 0.45);
    letter-spacing: 0.05em;
    text-transform: uppercase;
  }

  .response-content :global(.code-copy-btn) {
    background: transparent;
    border: 1px solid rgba(255,255,255,0.15);
    color: rgba(255,255,255,0.5);
    border-radius: 4px;
    padding: 0.1rem 0.4rem;
    font-size: 0.65rem;
    cursor: pointer;
    transition: all 0.15s;
    font-family: inherit;
  }

  .response-content :global(.code-copy-btn:hover) {
    background: rgba(255,255,255,0.1);
    color: rgba(255,255,255,0.9);
    border-color: rgba(255,255,255,0.3);
  }

  .response-content :global(.shiki) {
    padding: 1rem;
    margin: 0;
    font-size: 0.85rem;
    line-height: 1.6;
    overflow-x: auto;
    border: none;
    background: rgba(8, 10, 18, 0.97) !important;
    font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
  }

  .response-content :global(.shiki code) {
    font-family: inherit;
    font-size: inherit;
    background: none !important;
    padding: 0;
    color: inherit;
  }

  .response-content :global(code:not(.shiki code)) {
    background: rgba(255, 255, 255, 0.08);
    padding: 0.15rem 0.4rem;
    border-radius: 4px;
    font-size: 0.85em;
    font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
    color: #7dd3fc;
  }

  .response-content :global(p) {
    margin: 0.5rem 0;
  }

  .response-content :global(ul),
  .response-content :global(ol) {
    padding-left: 1.5rem;
    margin: 0.5rem 0;
  }

  .response-content :global(li) {
    margin: 0.3rem 0;
  }

  .response-content :global(strong) {
    color: rgba(255, 255, 255, 0.95);
    font-weight: 600;
  }
</style>
