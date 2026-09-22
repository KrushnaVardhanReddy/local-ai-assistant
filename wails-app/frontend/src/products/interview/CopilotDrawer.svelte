<script lang="ts">
  import { wsState, sendChat, sendChip, dismissChip, clearAllChips } from "$lib/ws.svelte";
  import { apiFetch, getApiUrl } from "$lib/api";
  import { onMount, onDestroy } from "svelte";

  const stealthMode = import.meta.env.VITE_STEALTH_MODE === 'true';

  let {
    showHotkeys = false,
    onToggleHotkeys
  } = $props<{
    showHotkeys?: boolean;
    onToggleHotkeys?: () => void;
  }>();

  let chatText = $state("");
  let responseEl: HTMLElement;

  let starPrimed = $state(false);
  let starPrimedTimer: ReturnType<typeof setTimeout>;

  let clearingCache = $state(false);
  let cacheStats = $state<{ cached_pairs: number; estimated_tokens_saved: number } | null>(null);

  let renderedResponse = $state('');

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

  function handleChatSubmit() {
    if (chatText.trim()) {
      sendChat(chatText);
      wsState.isThinking = true;
      chatText = "";
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleChatSubmit();
    }
  }

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

  onMount(() => {
    fetchCacheStats();
  });

  onDestroy(() => {
    if (starPrimedTimer) clearTimeout(starPrimedTimer);
  });

  // Dynamic markdown rendering - similar to BrainDrawer
  let renderMarkdown: any;
  let markdownTimeout: ReturnType<typeof setTimeout>;

  $effect(() => {
    const current = wsState.response;

    if (markdownTimeout) clearTimeout(markdownTimeout);

    markdownTimeout = setTimeout(async () => {
      if (current) {
        try {
          if (!renderMarkdown) {
            const m = await import('$lib/markdownRenderer');
            renderMarkdown = m.renderMarkdown;
          }
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

  // Scroll down transcript history on update
  $effect(() => {
    const _h = wsState.transcriptHistory;
    const _t = wsState.transcript;
    if (responseEl && !showHotkeys) {
      setTimeout(() => {
        if(responseEl) {
          responseEl.scrollTop = responseEl.scrollHeight;
        }
      }, 0);
    }
  });

</script>

<div class="copilot-drawer">
  <!-- Status & Actions Header -->
  <div class="p-3 border-b border-white/10 flex flex-col gap-2 bg-surface/50">
    <div class="flex justify-between items-center">
      <span class="text-sm font-bold text-primary flex items-center gap-1">
        <span class="material-symbols-outlined text-[16px]">auto_awesome</span>
        Copilot
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
        <button
          class="action-btn text-xs hover:text-error transition-colors ml-1"
          onclick={clearCache}
          disabled={clearingCache}
          title="Clear cache"
        >
          {#if clearingCache}
            <span class="material-symbols-outlined text-[14px] animate-spin">progress_activity</span>
          {:else}
            <span class="material-symbols-outlined text-[14px]">mop</span>
            <span class="text-[10px] ml-0.5">{cacheStats?.cached_pairs || 0}</span>
          {/if}
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

  <!-- Content Area (Unified Feed or Hotkeys) -->
  <div class="flex-1 overflow-y-auto p-4 hide-scrollbar flex flex-col gap-4" bind:this={responseEl}>
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
          {#if stealthMode}
            <div class="hotkey-row">
              <span>Toggle Stealth Click-through</span>
              <kbd>Ctrl+Alt+M</kbd>
            </div>
          {/if}
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
    {:else}
      <!-- Empty State -->
      {#if wsState.transcriptHistory.length === 0 && !wsState.transcript && !wsState.response && !wsState.isThinking}
        <div class="h-full flex flex-col items-center justify-center text-center text-on-surface-variant opacity-50 p-6">
          <span class="material-symbols-outlined text-[48px] mb-4">psychiatry</span>
          <p class="text-sm font-medium">The Copilot is listening...</p>
          <p class="text-xs mt-2">Make sure your microphone is on or start the mock interview.</p>
        </div>
      {:else}
        <!-- Unified Feed Items -->
        {#each wsState.transcriptHistory as line}
          <div class="transcript-line p-3 rounded-lg bg-surface-variant/30 border border-white/5 text-sm">
             <span class="speaker-badge interviewer">Speaker</span>
             <p class="mt-1 text-on-surface/90">{line}</p>
          </div>
        {/each}

        {#if wsState.transcript}
          <div class="transcript-line p-3 rounded-lg bg-primary/10 border border-primary/20 text-sm animate-pulse">
             <span class="speaker-badge candidate">You</span>
             <p class="mt-1 text-primary/90">{wsState.transcript}</p>
          </div>
        {/if}

        {#if wsState.isThinking && !wsState.response}
          <div class="flex items-center gap-2 text-primary/60 text-sm py-2">
            <span class="material-symbols-outlined animate-spin text-[16px]">progress_activity</span>
            Thinking...
          </div>
        {/if}

        {#if wsState.response}
          <div class="response-content bg-surface-variant/20 p-4 rounded-lg border border-white/5">
            {@html renderedResponse}
          </div>
        {/if}
      {/if}
    {/if}
  </div>

  <!-- Pending Suggestions -->
  {#if wsState.pendingTranscripts.length > 0 && !showHotkeys}
    <div class="p-3 border-t border-white/10 bg-surface/50 backdrop-blur-md flex flex-col gap-2">
      <div class="flex justify-between items-center px-1">
        <span class="text-[10px] uppercase font-bold tracking-wider text-on-surface-variant">Suggested Responses</span>
        <button class="text-[10px] uppercase font-bold tracking-wider text-error/80 hover:text-error transition-colors" onclick={clearAllChips}>
          Clear All
        </button>
      </div>
      <div class="flex overflow-x-auto gap-2 pb-1 hide-scrollbar pending-chips-container">
        {#each wsState.pendingTranscripts as chip (chip.id)}
          <div class="chip flex items-center bg-surface-variant/80 hover:bg-surface-variant border border-white/10 rounded-full pl-3 pr-1 py-1 whitespace-nowrap transition-all group">
            <button
              class="chip-btn text-xs text-on-surface/90 truncate max-w-[150px] text-left hover:text-primary transition-colors flex items-center gap-2"
              onclick={() => sendChip(chip)}
              title={chip.text}
              aria-label="Send suggestion: {chip.text}"
            >
              <span class="mock-chip-text">{chip.text}</span>
            </button>
            <button
              class="ml-2 w-5 h-5 flex items-center justify-center rounded-full hover:bg-white/10 text-on-surface-variant hover:text-error transition-colors"
              onclick={() => dismissChip(chip.id)}
              aria-label="Dismiss suggestion: {chip.text}"
            >
              <span class="material-symbols-outlined text-[14px]">close</span>
            </button>
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Chat Input -->
  {#if !showHotkeys}
    <div class="p-3 border-t border-white/10 bg-surface/80 backdrop-blur-md">
      <div class="relative">
        <textarea
          bind:value={chatText}
          onkeydown={handleKeydown}
          placeholder="Send custom prompt... (Shift+Enter for newline)"
          class="w-full bg-surface-variant/50 border border-white/10 rounded-lg py-2.5 pl-3 pr-10 text-sm text-on-surface focus:outline-none focus:border-primary/50 resize-y min-h-[84px] max-h-[220px] leading-relaxed"
          rows="3"
        ></textarea>
        <button
          onclick={handleChatSubmit}
          disabled={!chatText.trim()}
          title="Send prompt (Enter)"
          class="absolute right-2.5 bottom-3 w-7 h-7 flex items-center justify-center rounded-md bg-white/5 hover:bg-primary/20 text-on-surface-variant hover:text-primary disabled:opacity-30 disabled:hover:bg-transparent disabled:hover:text-on-surface-variant transition-all"
        >
          <span class="material-symbols-outlined text-[18px]">send</span>
        </button>
      </div>
    </div>
  {/if}
</div>

<style>
  .copilot-drawer {
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

  .action-btn:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.1);
    color: white;
  }

  .action-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
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

  .speaker-badge {
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    padding: 2px 6px;
    border-radius: 4px;
    display: inline-block;
  }

  .speaker-badge.interviewer {
    background: rgba(99, 179, 237, 0.15);
    color: #63b3ed;
    border: 1px solid rgba(99, 179, 237, 0.3);
  }

  .speaker-badge.candidate {
    background: rgba(74, 222, 128, 0.15);
    color: #4ade80;
    border: 1px solid rgba(74, 222, 128, 0.3);
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
