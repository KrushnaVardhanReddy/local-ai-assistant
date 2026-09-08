<script lang="ts">
  import { wsState, sendChat, sendChip, dismissChip, clearAllChips } from "$lib/ws.svelte";
  import { onMount } from "svelte";
  import SessionReport from './SessionReport.svelte';
  import { authState } from "$lib/auth.svelte";

  let showSessionReport = $state(false);
  import { listen } from "@tauri-apps/api/event";
  import { invoke } from "@tauri-apps/api/core";
  import { getCurrentWindow } from "@tauri-apps/api/window";
  import { renderMarkdown } from "$lib/markdownRenderer";

  async function startDrag(e: MouseEvent) {
    // Only drag on left mouse button, skip if clicking a button/input
    if (e.button !== 0) return;
    const target = e.target as HTMLElement;
    if (target.closest('button, input, select, textarea, a')) return;
    try {
      await getCurrentWindow().startDragging();
    } catch (_) {}
  }

  async function hideWindow() {
    try {
      await getCurrentWindow().hide();
    } catch (_) {}
  }

  async function closeApp() {
    try {
      await invoke("quit_app");
    } catch (_) {}
  }

  let responseEl: HTMLElement | undefined = $state();
  let contentEl: HTMLElement | undefined = $state();

  let isBrowser = $state(false);
  let chatText = $state("");
  let renderedResponse = $state('');
  let renderTimer: ReturnType<typeof setTimeout> | null = null;

  onMount(() => {
    const unlistenScrollDown = listen("scroll-down", () => {
      contentEl?.scrollBy({ top: 100, behavior: 'smooth' });
    });
    const unlistenScrollUp = listen("scroll-up", () => {
      contentEl?.scrollBy({ top: -100, behavior: 'smooth' });
    });

    // Check if we are running in a regular browser instead of Tauri
    isBrowser = typeof window !== 'undefined' && typeof (window as any).__TAURI_INTERNALS__ === 'undefined';
    const apiUrl = isBrowser ? `${window.location.protocol}//${window.location.host}` : "http://127.0.0.1:8765";

    const unlisten = listen("trigger-vision", async () => {
      if (wsState.isAnalyzingScreen) return;
      if (wsState.plan === "demo" || wsState.plan === "payg") {
        wsState.error = "Vision features require a Monthly or Founding plan.";
        return;
      }

      wsState.isAnalyzingScreen = true;
      try {
        const base64Image = await invoke<string>("capture_screen");

        const headers: Record<string, string> = {
          "Content-Type": "application/json"
        };
        if (authState.accessToken) {
          headers["Authorization"] = `Bearer ${authState.accessToken}`;
        }

        const res = await fetch(`${apiUrl}/vision/analyze`, {
          method: "POST",
          headers,
          body: JSON.stringify({ image_base64: base64Image })
        });

        if (!res.ok) {
          const json = await res.json().catch(() => ({}));
          wsState.error = json.message || "Vision request failed";
        }

      } catch (e) {
        console.error("Failed to capture screen or send to backend:", e);
      } finally {
        wsState.isAnalyzingScreen = false;
      }
    });

    return () => {
      unlisten.then(f => f());
      unlistenScrollDown.then(f => f());
      unlistenScrollUp.then(f => f());
    };
  });

  $effect(() => {
    // Removed aggressive auto-scroll so the user can read from the top down
    // at their own pace without the text jumping away from them.
  });

  $effect(() => {
    const current = wsState.response;
    if (renderTimer) clearTimeout(renderTimer);
    renderTimer = setTimeout(async () => {
      if (current) {
        try {
          renderedResponse = await renderMarkdown(current);
        } catch (e) {
          console.error("[DEBUG UI] renderMarkdown error:", e);
          renderedResponse = "⚠️ Markdown render error: " + e;
        }
      } else {
        renderedResponse = '';
      }
    }, 50);
  });

  function clearError() {
    wsState.error = null;
  }

  function handleChatSubmit() {
    if (chatText.trim()) {
      sendChat(chatText);
      chatText = "";
    }
  }

  function handleChatKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleChatSubmit();
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.ctrlKey && e.shiftKey && e.key === 'E') {
      e.preventDefault();
      showSessionReport = !showSessionReport;
    }
  }

  function triggerVision() {
    if (!wsState.isAnalyzingScreen && !isBrowser) {
      if (wsState.plan === "demo" || wsState.plan === "payg") {
        wsState.error = "Vision features require a Monthly or Founding plan.";
        return;
      }

      // Dispatch a synthetic event that the backend/Tauri bridge will pick up
      // Or just invoke directly here if we are not relying on the global hotkey
      invoke<string>("capture_screen").then(async (base64Image) => {
        wsState.isAnalyzingScreen = true;
        const apiUrl = isBrowser ? `${window.location.protocol}//${window.location.host}` : "http://127.0.0.1:8765";
        try {
          const headers: Record<string, string> = {
            "Content-Type": "application/json"
          };
          if (authState.accessToken) {
            headers["Authorization"] = `Bearer ${authState.accessToken}`;
          }

          const res = await fetch(`${apiUrl}/vision/analyze`, {
            method: "POST",
            headers,
            body: JSON.stringify({ image_base64: base64Image })
          });

          if (!res.ok) {
            const json = await res.json().catch(() => ({}));
            wsState.error = json.message || "Vision request failed";
          }
        } finally {
          wsState.isAnalyzingScreen = false;
        }
      }).catch(console.error);
    }
  }

  function clearHistory() {
    wsState.transcript = "";
    wsState.response = "";
    wsState.ragSources = [];
    wsState.pendingTranscripts = [];
    const apiUrl = isBrowser ? `${window.location.protocol}//${window.location.host}` : "http://127.0.0.1:8765";
    fetch(`${apiUrl}/history/clear`, { method: 'POST' }).catch(console.error);
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="fixed inset-0 w-full h-full pointer-events-none flex flex-col z-50 p-container-padding gap-container-padding text-on-background antialiased font-body-md text-body-md select-none dark" id="dashboard-overlay">
  <!-- Top Toolbar -->
  <header class="toolbar glass-pill pointer-events-auto flex items-center justify-between px-6 h-toolbar-height rounded-full w-full max-w-7xl mx-auto shadow-2xl transition-all duration-300" onmousedown={startDrag}>
    <!-- Brand / Primary Action -->
    <div class="flex items-center gap-4 pointer-events-none">
      <span class="font-headline-md text-headline-md font-bold text-primary tracking-tight">Local AI</span>
      <!-- Live Indicator -->
      <div class="flex items-center gap-2 bg-white/5 rounded-full px-3 py-1 border border-white/5 {wsState.isListening ? '' : 'opacity-50'}">
        <div class="w-2 h-2 rounded-full {wsState.isListening ? 'bg-secondary animate-pulse shadow-[0_0_8px_rgba(78,222,163,0.6)]' : 'bg-gray-500'}"></div>
        <span class="font-label-caps text-label-caps {wsState.isListening ? 'text-secondary' : 'text-gray-400'} tracking-widest uppercase">
          {wsState.isListening ? (wsState.isPTTHeld ? 'PTT Active' : 'Live') : 'Mic Off'}
        </span>
      </div>
      {#if wsState.isConnected}
        <span class="font-label-caps text-label-caps text-secondary tracking-widest uppercase ml-2">● Connected</span>
      {:else}
        <span class="font-label-caps text-label-caps text-gray-400 tracking-widest uppercase ml-2">○ Reconnecting...</span>
      {/if}
      {#if wsState.error}
        <button class="font-label-caps text-label-caps text-red-400 tracking-widest uppercase ml-2 hover:underline pointer-events-auto" onclick={clearError}>
          ⚠ {wsState.error}
        </button>
      {/if}
    </div>

    <!-- Central Chat Input (Stealth) -->
    <div class="flex-1 max-w-xl mx-8">
      <div class="relative flex items-center w-full h-8 bg-white/5 rounded-lg border border-white/10 transition-colors focus-within:bg-white/10 focus-within:border-white/20">
        <span class="material-symbols-outlined text-[18px] text-on-surface-variant ml-3" data-icon="search" style="font-variation-settings: 'FILL' 0;">chat</span>
        <input 
          class="w-full bg-transparent border-none text-on-surface-variant font-body-sm text-body-sm focus:ring-0 placeholder-on-surface-variant/50 h-full px-3 outline-none pointer-events-auto" 
          placeholder="Silent chat (Helper Mode)..." 
          type="text"
          bind:value={chatText}
          onkeydown={handleChatKeydown}
        />
        <button class="font-mono-data text-mono-data text-on-surface-variant/40 hover:text-primary mr-3 text-[10px] pointer-events-auto" onclick={handleChatSubmit}>SEND</button>
      </div>
    </div>

    <!-- Trailing Actions -->
    <div class="flex items-center gap-2">
      <button
        aria-label="Screenshot"
        title={wsState.plan === 'demo' || wsState.plan === 'payg' ? 'Vision features require a Monthly or Founding plan.' : 'Screenshot'}
        class="w-8 h-8 flex items-center justify-center rounded-full transition-colors pointer-events-auto {(wsState.plan === 'demo' || wsState.plan === 'payg') ? 'opacity-50 cursor-not-allowed text-on-surface-variant' : 'hover:bg-white/10 text-on-surface-variant hover:text-primary ' + (wsState.isAnalyzingScreen ? 'text-primary animate-pulse' : '')}"
        onclick={triggerVision}
        disabled={wsState.plan === 'demo' || wsState.plan === 'payg'}
      >
        <span class="material-symbols-outlined text-[20px]" data-icon="screenshot_monitor">screenshot_monitor</span>
      </button>
      <button
        id="session-report-btn"
        class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-white/10 text-on-surface-variant hover:text-primary transition-colors pointer-events-auto"
        onclick={() => showSessionReport = true}
        title="Session Report (Ctrl+Shift+E)"
      >
        <span class="material-symbols-outlined text-[20px]">analytics</span>
      </button>
      <button aria-label="Clear Context" class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-white/10 text-on-surface-variant hover:text-primary transition-colors pointer-events-auto" onclick={clearHistory}>
        <span class="material-symbols-outlined text-[20px]" data-icon="mop">mop</span>
      </button>
      <!-- Separator -->
      <div class="w-px h-5 bg-white/10 mx-1"></div>
      <!-- Hide window (Ctrl+Shift+Space to restore) -->
      <button aria-label="Hide" title="Hide (Ctrl+Shift+Space to restore)" class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-white/10 text-on-surface-variant hover:text-yellow-400 transition-colors pointer-events-auto" onclick={hideWindow}>
        <span class="material-symbols-outlined text-[20px]" data-icon="visibility_off">visibility_off</span>
      </button>
      <!-- Close / Quit app -->
      <button aria-label="Close" title="Quit App" class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-red-500/20 text-on-surface-variant hover:text-red-400 transition-colors pointer-events-auto" onclick={closeApp}>
        <span class="material-symbols-outlined text-[20px]" data-icon="close">close</span>
      </button>
    </div>
  </header>

  <!-- Transcript Chip Bar -->
  {#if wsState.pendingTranscripts.length > 0}
    <div class="chip-bar pointer-events-auto flex items-center gap-2 px-6 py-2 w-full max-w-7xl mx-auto overflow-x-auto hide-scrollbar">
      {#each wsState.pendingTranscripts as chip (chip.id)}
        <div
          role="button"
          tabindex="0"
          class="chip-pill {chip.speaker === 'interviewer' ? 'chip-interviewer' : ''} group flex items-center gap-1.5 px-3 py-1.5 rounded-full
                 bg-white/8 border border-white/10 hover:bg-primary/20 hover:border-primary/40
                 transition-all duration-200 cursor-pointer whitespace-nowrap
                 animate-chip-in flex-shrink-0"
          onclick={() => sendChip(chip)}
          onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') sendChip(chip); }}
          title="Click to send: {chip.text}"
        >
          {#if chip.speaker === 'interviewer'}
            <span class="speaker-badge interviewer-badge">🎤</span>
          {:else if chip.speaker === 'candidate'}
            <span class="speaker-badge candidate-badge">👤</span>
          {/if}
          <span class="text-on-surface-variant text-[12px] font-body-sm max-w-[200px] truncate group-hover:text-primary transition-colors">
            {chip.text.length > 60 ? chip.text.slice(0, 60) + '…' : chip.text}
          </span>
          <button
            class="w-4 h-4 flex items-center justify-center rounded-full
                   text-on-surface-variant/40 hover:text-red-400 hover:bg-red-400/10
                   transition-colors text-[10px] ml-0.5 flex-shrink-0"
            onclick={(e) => { e.stopPropagation(); dismissChip(chip.id); }}
            aria-label="Dismiss"
          >✕</button>
        </div>
      {/each}

      <!-- Clear All button -->
      <button
        class="flex items-center gap-1 px-2.5 py-1 rounded-full
               text-on-surface-variant/50 hover:text-red-400 hover:bg-red-400/10
               border border-transparent hover:border-red-400/20
               transition-all duration-200 text-[11px] font-label-caps
               tracking-wider uppercase whitespace-nowrap flex-shrink-0"
        onclick={() => clearAllChips()}
      >
        Clear
        <span class="material-symbols-outlined text-[14px]">close</span>
      </button>
    </div>
  {/if}

  <!-- Main Content Grid -->
  <main class="flex-1 flex gap-container-padding w-full max-w-7xl mx-auto h-[calc(100vh-120px)] pb-6">
    <!-- Left Panel: Live Ears (Transcription) -->
    <aside class="live-ears-panel glass-panel pointer-events-auto w-1/3 rounded-[24px] flex flex-col overflow-hidden transition-transform duration-300 shadow-2xl">
      <!-- Panel Header -->
      <div class="flex items-center justify-between p-4 border-b border-white/5 bg-black/20" onmousedown={startDrag} style="cursor: grab;">
        <div class="flex items-center gap-2 pointer-events-none">
          <span class="material-symbols-outlined text-on-surface-variant text-[20px]" data-icon="hearing">hearing</span>
          <h2 class="font-label-caps text-label-caps text-on-surface-variant tracking-wider">Live Ears</h2>
        </div>
      </div>
      
      <!-- Panel Body (Live STT) -->
      <div class="flex-1 p-6 overflow-y-auto hide-scrollbar flex flex-col gap-4 font-body-sm text-body-sm text-on-surface-variant/80" bind:this={contentEl}>
        {#if !wsState.transcript}
           <div class="text-center text-on-surface-variant/50 mt-10">
             Waiting for speech...
           </div>
        {/if}

        {#if wsState.transcript}
          <div class="transcript-line border-l-secondary/30">
            <span class="block text-secondary/70 font-mono-data text-mono-data mb-1 text-[11px]">Interviewer</span>
            <p class="leading-relaxed text-on-background/90">{wsState.transcript}</p>
          </div>
        {/if}
      </div>
    </aside>

    <!-- Right Panel: The Brain (AI Insights) -->
    <section class="brain-panel glass-panel pointer-events-auto w-2/3 rounded-[24px] flex flex-col overflow-hidden shadow-2xl relative">
      <!-- Ambient Glow effect for the brain panel -->
      <div class="absolute inset-0 bg-gradient-to-br from-white/[0.02] to-transparent pointer-events-none"></div>
      
      <!-- Panel Header -->
      <div class="flex items-center justify-between p-4 border-b border-white/5 bg-black/20 z-10" onmousedown={startDrag} style="cursor: grab;">
        <div class="flex items-center gap-2 pointer-events-none">
          <span class="material-symbols-outlined text-primary text-[20px]" data-icon="memory">memory</span>
          <h2 class="font-label-caps text-label-caps text-primary tracking-wider text-glow">The Brain</h2>
        </div>
        {#if wsState.ragSources && wsState.ragSources.length > 0}
          <div class="text-[10px] text-secondary font-mono-data">
            🔍 Sources: {wsState.ragSources.join(' · ')}
          </div>
        {/if}
      </div>

      <!-- Panel Body (Insights & Suggestions) -->
      <div class="flex-1 p-6 overflow-y-auto hide-scrollbar z-10 flex flex-col gap-6" bind:this={responseEl}>
        {#if !wsState.response && !wsState.isThinking}
          <div class="text-center text-on-surface-variant/50 mt-20">
            Listening for questions to answer...
          </div>
        {/if}

        {#if wsState.response || wsState.isThinking}
          <div class="flex-1 flex flex-col gap-4">
            <div class="group flex items-start gap-4 p-4 rounded-xl hover:bg-white/5 transition-colors border border-transparent hover:border-white/5 cursor-default relative">
              <div class="ml-1 response-content text-on-surface-variant leading-relaxed prose prose-invert prose-sm max-w-none">
                {@html renderedResponse}
                {#if wsState.isThinking}
                  <span class="inline-block w-1.5 h-4 bg-primary align-middle animate-pulse ml-1"></span>
                {/if}
              </div>
            </div>
          </div>
        {/if}
      </div>

      <!-- Generating Indicator (Bottom pinned) -->
      {#if wsState.isThinking}
        <div class="absolute bottom-0 left-0 right-0 h-1 bg-white/5 z-20">
          <div class="h-full bg-primary/40 w-1/3 rounded-r-full relative overflow-hidden">
            <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/50 to-transparent -translate-x-full animate-[shimmer_2s_infinite]"></div>
          </div>
        </div>
      {/if}
    </section>
  </main>

  {#if showSessionReport}
    <div class="pointer-events-auto">
      <SessionReport onClose={() => showSessionReport = false} />
    </div>
  {/if}
</div>



{#if wsState.sessionExpired}
  <div class="fixed inset-0 z-[9999] pointer-events-auto flex items-center justify-center bg-black/80 backdrop-blur-md">
    <div class="glass-panel p-8 rounded-2xl max-w-md w-full text-center flex flex-col items-center gap-6 shadow-2xl border border-white/10 relative overflow-hidden">
      <!-- Background Glow -->
      <div class="absolute inset-0 bg-gradient-to-br from-red-500/10 to-transparent pointer-events-none"></div>

      <div class="w-16 h-16 rounded-full bg-red-500/20 flex items-center justify-center text-red-400 mb-2">
        <span class="material-symbols-outlined text-[32px]">hourglass_disabled</span>
      </div>

      <h2 class="text-xl font-headline-md text-on-background tracking-wide">
        Session ended — extend for $5 or upgrade to Monthly
      </h2>

      <p class="text-on-surface-variant text-body-sm font-body-sm">
        Your current session limit has been reached. Choose an option below to continue using the Local AI Assistant.
      </p>

      <div class="flex flex-col sm:flex-row gap-4 w-full mt-4">
        <a href="/pricing" target="_blank" rel="noopener noreferrer"
           class="flex-1 py-3 px-4 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 hover:border-white/20 transition-all duration-200 text-on-surface-variant font-label-caps text-label-caps tracking-wider uppercase text-center flex items-center justify-center gap-2">
          $5 Extension
        </a>
        <a href="/pricing" target="_blank" rel="noopener noreferrer"
           class="flex-1 py-3 px-4 rounded-xl bg-primary hover:bg-primary/90 text-background font-label-caps text-label-caps tracking-wider uppercase font-bold transition-all duration-200 text-center flex items-center justify-center gap-2 shadow-[0_0_20px_rgba(78,222,163,0.3)] hover:shadow-[0_0_30px_rgba(78,222,163,0.5)]">
          Upgrade to Monthly
        </a>
      </div>
    </div>
  </div>
{/if}

<style>
  .glass-panel {
      background: rgba(0, 0, 0, 0.15);
      backdrop-filter: blur(2px);
      -webkit-backdrop-filter: blur(2px);
      border: 1px solid rgba(255, 255, 255, 0.15);
      box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
  }
  .glass-pill {
      background: rgba(18, 18, 18, 0.95);
      backdrop-filter: blur(20px);
      -webkit-backdrop-filter: blur(20px);
      border: 1px solid rgba(255, 255, 255, 0.2);
      box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
  }
  .transcript-line {
      border-left: 1px solid rgba(255, 255, 255, 0.1);
      padding-left: 12px;
      margin-bottom: 16px;
  }
  .transcript-line p {
      text-shadow: 1px 1px 2px rgba(0,0,0,0.8), -1px -1px 2px rgba(0,0,0,0.8);
  }
  .hide-scrollbar::-webkit-scrollbar {
      display: none;
  }
  .hide-scrollbar {
      -ms-overflow-style: none;
      scrollbar-width: none;
  }
  .text-glow {
      text-shadow: 0 0 10px rgba(255, 255, 255, 0.3);
  }
  @keyframes shimmer {
      100% {
          transform: translateX(100%);
      }
  }
  .animate-\\[shimmer_2s_infinite\\] {
      animation: shimmer 2s infinite;
  }
  .response-content {
      font-size: 1rem;
      text-shadow: 1px 1px 2px rgba(0,0,0,0.8), -1px -1px 2px rgba(0,0,0,0.8);
  }
  
  /* Shiki code block overrides — match our dark glass theme */
  .response-content :global(.shiki) {
    text-shadow: none;
    border-radius: 8px;
    padding: 1rem;
    margin: 0.75rem 0;
    font-size: 0.85rem;
    line-height: 1.6;
    overflow-x: auto;
    border: 1px solid rgba(255, 255, 255, 0.08);
  }

  /* Copy button container for code blocks */
  .response-content :global(.code-block-wrapper) {
    position: relative;
  }

  /* Inline code styling */
  .response-content :global(code:not(.shiki code)) {
    background: rgba(255, 255, 255, 0.08);
    padding: 0.15rem 0.4rem;
    border-radius: 4px;
    font-size: 0.875em;
    font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace;
    color: #7dd3fc;
  }

  /* Paragraph spacing */
  .response-content :global(p) {
    margin: 0.5rem 0;
    line-height: 1.65;
  }

  /* Bullet / ordered lists */
  .response-content :global(ul),
  .response-content :global(ol) {
    padding-left: 1.5rem;
    margin: 0.5rem 0;
  }

  .response-content :global(li) {
    margin: 0.25rem 0;
    line-height: 1.55;
  }

  /* Bold text */
  .response-content :global(strong) {
    color: rgba(255, 255, 255, 0.95);
    font-weight: 600;
  }

  /* Blockquotes (tips, notes from LLM) */
  .response-content :global(blockquote) {
    border-left: 3px solid rgba(99, 179, 237, 0.5);
    margin: 0.75rem 0;
    padding: 0.5rem 1rem;
    color: rgba(255, 255, 255, 0.6);
    font-style: italic;
  }

  /* Headings inside responses */
  .response-content :global(h1),
  .response-content :global(h2),
  .response-content :global(h3) {
    color: rgba(255, 255, 255, 0.9);
    font-weight: 600;
    margin: 1rem 0 0.5rem 0;
    line-height: 1.3;
  }
  .response-content :global(h3) { font-size: 1rem; }
  .response-content :global(h2) { font-size: 1.1rem; }

  /* We remove default styles from the component since it's full screen now */

  /* Transcript chip bar */
  .chip-bar {
    flex-shrink: 0;
  }

  .chip-pill {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
  }

  .chip-pill:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3), 0 0 20px rgba(99, 179, 237, 0.1);
  }

  @keyframes chipIn {
    from {
      opacity: 0;
      transform: translateY(8px) scale(0.95);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  .animate-chip-in {
    animation: chipIn 0.25s ease-out;
  }

  .chip-interviewer {
    border-color: rgba(99, 179, 237, 0.3);
    background: rgba(99, 179, 237, 0.08);
  }
  .speaker-badge {
    font-size: 11px;
    flex-shrink: 0;
  }
</style>
