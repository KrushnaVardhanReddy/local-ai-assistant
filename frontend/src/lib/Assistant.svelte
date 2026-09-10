<script lang="ts">
  import { wsState, sendChat, sendChip, dismissChip, clearAllChips, toggleMockMode } from "$lib/ws.svelte";
  import { onMount } from "svelte";
  import SessionReport from './SessionReport.svelte';
  import ResumeBuilder from './ResumeBuilder.svelte';
  import { authState } from "$lib/auth.svelte";
  import { uiState } from "$lib/stores/uiState.svelte.ts";
  import HotkeysPanel from "$lib/components/HotkeysPanel.svelte";

  let showSessionReport = $state(false);
  let currentView = $state<'interview' | 'resume'>('interview');
  let starPrimed = $state(false);
  let starPrimedTimer: ReturnType<typeof setTimeout> | null = null;

  let clearingCache = $state(false);
  let cacheStats = $state<{ cached_pairs: number; estimated_tokens_saved: number } | null>(null);

  // Pre-warm state
  let showPrewarmModal = $state(false);
  let prewarmJobDescription = $state("");
  let prewarmingCache = $state(false);

  function handleMockModeToggle() {
    const isNowEnabled = !wsState.isMockMode;
    toggleMockMode(isNowEnabled);
    if (isNowEnabled) {
      // Start microphone for hands-free mock interview
      fetch('http://127.0.0.1:8765/ptt/start', { method: 'POST' }).catch(console.error);
    } else {
      showSessionReport = true;
      // Stop microphone
      fetch('http://127.0.0.1:8765/ptt/stop', { method: 'POST' }).catch(console.error);
    }
  }
  let liveEarsCollapsed = $state(false);
  let brainCollapsed = $state(false);
  let clickthrough = $state(false);
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

  async function fetchCacheStats() {
    try {
      const configuredUrl = localStorage.getItem('backend_url') || '127.0.0.1:8765';
      const apiUrl = configuredUrl.startsWith('http') ? configuredUrl : `http://${configuredUrl}`;
      const res = await fetch(`${apiUrl}/api/cache/stats`);
      if (res.ok) {
        cacheStats = await res.json();
      } else {
        cacheStats = null;
      }
    } catch {
      cacheStats = null;
    }
  }


  async function prewarmCache() {
    prewarmingCache = true;
    try {
      const storedBackendUrl = localStorage.getItem("backend_url") || "127.0.0.1:8765";
      const apiUrl = storedBackendUrl.startsWith('http') ? storedBackendUrl : `http://${storedBackendUrl}`;

      const resContext = await fetch(`${apiUrl}/api/resume/context`);
      let baseResume = "";
      if (resContext.ok) {
        const contextData = await resContext.json();
        baseResume = contextData.context || "";
      }

      const res = await fetch(`${apiUrl}/api/cache/prewarm`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          resume_text: baseResume,
          job_description: prewarmJobDescription
        })
      });

      if (!res.ok) {
        throw new Error("Failed to prewarm cache");
      }

      const data = await res.json();
      showPrewarmModal = false;
      alert(`Cache Warmed with ${data.stored_count || 50} Questions`);
      await fetchCacheStats();
    } catch (e: any) {
      alert(`Error pre-warming cache: ${e.message}`);
    } finally {
      prewarmingCache = false;
    }
  }

  async function clearCache() {

    const confirmed = window.confirm(
      `Are you sure you want to clear all ${cacheStats?.cached_pairs ?? 0} cached Q&A pairs? This cannot be undone.`
    );
    if (!confirmed) return;
    clearingCache = true;
    try {
      const configuredUrl = localStorage.getItem('backend_url') || '127.0.0.1:8765';
      const apiUrl = configuredUrl.startsWith('http') ? configuredUrl : `http://${configuredUrl}`;
      await fetch(`${apiUrl}/api/cache`, { method: "DELETE" });
      await fetchCacheStats();
    } finally {
      clearingCache = false;
    }
  }

  onMount(() => {
    fetchCacheStats();

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

    const unlistenClickthrough = listen("toggle-clickthrough", () => {
      toggleClickthrough();
    });

    const unlistenTranscript1 = listen("hotkey_transcript_1", () => {
      if (wsState.pendingTranscripts.length > 0) sendChip(wsState.pendingTranscripts[0]);
    });
    const unlistenTranscript2 = listen("hotkey_transcript_2", () => {
      if (wsState.pendingTranscripts.length > 1) sendChip(wsState.pendingTranscripts[1]);
    });
    const unlistenTranscript3 = listen("hotkey_transcript_3", () => {
      if (wsState.pendingTranscripts.length > 2) sendChip(wsState.pendingTranscripts[2]);
    });
    const unlistenTranscript4 = listen("hotkey_transcript_4", () => {
      if (wsState.pendingTranscripts.length > 3) sendChip(wsState.pendingTranscripts[3]);
    });
    const unlistenTranscript5 = listen("hotkey_transcript_5", () => {
      if (wsState.pendingTranscripts.length > 4) sendChip(wsState.pendingTranscripts[4]);
    });
    const unlistenTranscript6 = listen("hotkey_transcript_6", () => {
      if (wsState.pendingTranscripts.length > 5) sendChip(wsState.pendingTranscripts[5]);
    });

    return () => {
      unlisten.then(f => f());
      unlistenScrollDown.then(f => f());
      unlistenScrollUp.then(f => f());
      unlistenClickthrough.then(f => f());
      unlistenTranscript1.then(f => f());
      unlistenTranscript2.then(f => f());
      unlistenTranscript3.then(f => f());
      unlistenTranscript4.then(f => f());
      unlistenTranscript5.then(f => f());
      unlistenTranscript6.then(f => f());
      if (starPrimedTimer) clearTimeout(starPrimedTimer);
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
    if (e.ctrlKey && e.shiftKey && e.key === 'L') {
      e.preventDefault();
      liveEarsCollapsed = !liveEarsCollapsed;
    }
    if (e.ctrlKey && e.shiftKey && e.key === 'B') {
      e.preventDefault();
      brainCollapsed = !brainCollapsed;
    }
  }

  async function toggleClickthrough() {
    clickthrough = !clickthrough;
    if (!isBrowser) {
      try {
        await invoke('set_clickthrough', { enable: clickthrough });
      } catch (e) {
        console.error('set_clickthrough failed:', e);
        clickthrough = !clickthrough; // revert on error
      }
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
    wsState.transcriptHistory = [];
    const apiUrl = isBrowser ? `${window.location.protocol}//${window.location.host}` : "http://127.0.0.1:8765";
    fetch(`${apiUrl}/history/clear`, { method: 'POST' }).catch(console.error);
  }

  function triggerCatchMeUp() {
    if (wsState.isThinking) return;
    const history = wsState.transcriptHistory;
    if (!history || history.length === 0) {
      wsState.error = "No conversation history yet — start speaking first.";
      setTimeout(() => { wsState.error = null; }, 3000);
      return;
    }
    const historyText = history
      .map((t, i) => `${i + 1}. "${t}"`)
      .join("\n");
    sendChat(
      `Here is the conversation transcript so far:\n${historyText}\n\n` +
      `Give me a concise 3-5 bullet point summary of the key topics, ` +
      `questions, and decisions discussed. Be brief and actionable.`
    );
  }

  function triggerStarPreset() {
    // Send the STAR primer as a silent chat message
    sendChat(
      "For your next response only, structure your answer using the STAR " +
      "method with bold section headers: **Situation** → **Task** → " +
      "**Action** → **Result**. Keep each section concise (2-3 sentences). " +
      "After this response, return to your normal answering style."
    );
    // Activate visual primed state for 3s
    starPrimed = true;
    if (starPrimedTimer) clearTimeout(starPrimedTimer);
    starPrimedTimer = setTimeout(() => {
      starPrimed = false;
    }, 3000);
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="fixed inset-0 w-full h-full pointer-events-none flex flex-col z-50 p-container-padding gap-container-padding text-on-background antialiased font-body-md text-body-md select-none dark" id="dashboard-overlay">
  <!-- Top Toolbar -->
  <header class="toolbar glass-pill {clickthrough ? 'clickthrough-mode' : ''} pointer-events-auto flex items-center justify-between px-6 h-toolbar-height rounded-full w-full max-w-7xl mx-auto shadow-2xl transition-all duration-300" onmousedown={startDrag}>
    <!-- Brand / Primary Action -->
    <div class="flex items-center gap-4 pointer-events-none">
      <span class="font-headline-md text-headline-md font-bold text-primary tracking-tight">BarnOwl</span>
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



    <!-- Trailing Actions -->
    <div class="flex items-center gap-2">
      <button
        id="mock-mode-btn"
        class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-white/10 text-on-surface-variant transition-colors pointer-events-auto {wsState.isMockMode ? 'text-green-400 animate-pulse bg-green-400/10' : 'hover:text-primary'}"
        onclick={handleMockModeToggle}
        title={wsState.isMockMode ? "Stop Mock Interview" : "Start Mock Interview Practice"}
      >
        <span class="material-symbols-outlined text-[20px]">record_voice_over</span>
      </button>
      <button
        aria-label="Screenshot"
        title={wsState.plan === 'demo' || wsState.plan === 'payg' ? 'Vision features require a Monthly or Founding plan.' : 'Screenshot'}
        class="w-8 h-8 flex items-center justify-center rounded-full transition-colors pointer-events-auto {(wsState.plan === 'demo' || wsState.plan === 'payg') ? 'opacity-50 cursor-not-allowed text-on-surface-variant' : 'hover:bg-white/10 text-on-surface-variant hover:text-primary ' + (wsState.isAnalyzingScreen ? 'text-primary animate-pulse' : '')}"
        onclick={triggerVision}
        disabled={wsState.plan === 'demo' || wsState.plan === 'payg'}
      >
        <span class="material-symbols-outlined text-[20px]" data-icon="screenshot_monitor">screenshot_monitor</span>
      </button>
      <!-- STAR Method Preset -->
      <button
        aria-label="STAR Method Preset"
        title="Prime next answer with STAR format (Situation → Task → Action → Result)"
        class="flex items-center gap-1 px-2.5 py-1 rounded-full
               transition-colors pointer-events-auto text-[11px] font-bold
               tracking-wider border
               {starPrimed
                 ? 'text-yellow-300 bg-yellow-400/15 border-yellow-400/40 animate-pulse'
                 : 'text-on-surface-variant hover:text-yellow-300 hover:bg-yellow-400/10 border-transparent hover:border-yellow-400/20'}"
        onclick={triggerStarPreset}
      >
        STAR
      </button>
      <button
        aria-label="Hotkeys Cheatsheet"
        title="Hotkeys Cheatsheet"
        class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-white/10 text-on-surface-variant hover:text-primary transition-colors pointer-events-auto"
        onclick={() => uiState.hotkeysPanelOpen = true}
      >
        <span class="material-symbols-outlined text-[20px]" data-icon="keyboard">keyboard</span>
      </button>
      <button
        id="session-report-btn"
        class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-white/10 text-on-surface-variant hover:text-primary transition-colors pointer-events-auto"
        onclick={() => showSessionReport = true}
        title="Session Report (Ctrl+Shift+E)"
      >
        <span class="material-symbols-outlined text-[20px]">analytics</span>
      </button>

      <button
        aria-label="Toggle Resume Builder"
        title={currentView === 'interview' ? 'Switch to Resume Builder' : 'Switch to Live Interview'}
        class="w-8 h-8 flex items-center justify-center rounded-full transition-colors pointer-events-auto {currentView === 'resume' ? 'bg-primary/20 text-primary ring-1 ring-primary/40' : 'hover:bg-white/10 text-on-surface-variant hover:text-primary'}"
        onclick={() => currentView = currentView === 'interview' ? 'resume' : 'interview'}
      >
        <span class="material-symbols-outlined text-[20px]">{currentView === 'resume' ? 'edit_document' : 'description'}</span>
      </button>

      <button aria-label="Clear Context" class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-white/10 text-on-surface-variant hover:text-primary transition-colors pointer-events-auto" onclick={clearHistory}>
        <span class="material-symbols-outlined text-[20px]" data-icon="mop">mop</span>
      </button>
      <!-- Click-through toggle -->
      <button
        aria-label="Toggle Click-Through"
        title="Click-Through Mode (Ctrl+Shift+M) — lets you click apps behind the overlay"
        class="w-8 h-8 flex items-center justify-center rounded-full transition-colors pointer-events-auto
               {clickthrough ? 'bg-primary/20 text-primary ring-1 ring-primary/40' : 'hover:bg-white/10 text-on-surface-variant hover:text-primary'}"
        onclick={toggleClickthrough}
      >
        <span class="material-symbols-outlined text-[20px]">{clickthrough ? 'mouse' : 'back_hand'}</span>
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
    {#if currentView === 'resume'}
      <ResumeBuilder />
    {:else}
      {#if wsState.isMockMode}
    <section class="glass-panel pointer-events-auto rounded-[24px] flex flex-col justify-center items-center w-full shadow-2xl relative p-12 overflow-hidden border border-green-500/30">
      <div class="absolute inset-0 bg-gradient-to-br from-green-500/5 to-transparent pointer-events-none"></div>
      <div class="flex items-center gap-3 mb-6 z-10">
        <span class="material-symbols-outlined text-green-400 text-4xl animate-pulse">record_voice_over</span>
        <h2 class="text-3xl font-headline-md text-green-400 tracking-wider">Mock Interview Mode</h2>
      </div>

      <div class="z-10 text-center max-w-3xl flex-1 flex flex-col justify-center w-full">
        {#if wsState.isThinking && !wsState.response}
          <div class="text-on-surface-variant/70 text-xl animate-pulse">
            The interviewer is thinking...
          </div>
        {:else if wsState.response || wsState.isThinking}
          <div class="text-2xl text-on-background/90 leading-relaxed font-medium mb-8">
            {wsState.response}
          </div>
        {:else}
          <div class="text-on-surface-variant/70 text-xl">
            Listening to your answer...
          </div>
        {/if}

        {#if wsState.transcript}
          <div class="mt-8 p-6 rounded-2xl bg-white/5 border border-white/10 text-left">
            <span class="block text-secondary/70 font-mono-data text-mono-data mb-2 text-sm uppercase tracking-widest">You said:</span>
            <p class="text-lg text-on-surface-variant leading-relaxed">{wsState.transcript}</p>
          </div>
        {/if}
      </div>

      <!-- Generating Indicator -->
      {#if wsState.isThinking}
        <div class="absolute bottom-0 left-0 right-0 h-1 bg-white/5 z-20">
          <div class="h-full bg-green-400/40 w-1/3 rounded-r-full relative overflow-hidden">
            <div class="absolute inset-0 bg-gradient-to-r from-transparent via-white/50 to-transparent -translate-x-full animate-[shimmer_2s_infinite]"></div>
          </div>
        </div>
      {/if}
    </section>
    {:else}
    <!-- Left Panel: Live Ears (Transcription) -->
    <aside
      class="live-ears-panel glass-panel {clickthrough ? 'clickthrough-mode' : ''} pointer-events-auto rounded-[24px] flex flex-col overflow-hidden shadow-2xl
             transition-all duration-300 ease-in-out
             {liveEarsCollapsed ? 'w-0 opacity-0 p-0 min-w-0 border-0' : brainCollapsed ? 'w-full' : 'w-1/3'}"
      style="{liveEarsCollapsed ? 'pointer-events:none;' : ''}"
    >
      <!-- Panel Header -->
      <div class="flex items-center justify-between p-4 border-b border-white/5 bg-black/20 flex-shrink-0" onmousedown={startDrag} style="cursor: grab;">
        <div class="flex items-center gap-2">
          <span class="material-symbols-outlined text-on-surface-variant text-[20px]" data-icon="hearing">hearing</span>
          <h2 class="font-label-caps text-label-caps text-on-surface-variant tracking-wider">Live Ears</h2>
        </div>
        <button
          class="w-7 h-7 flex items-center justify-center rounded-full hover:bg-white/10 text-on-surface-variant/60 hover:text-primary transition-colors pointer-events-auto flex-shrink-0"
          onclick={() => liveEarsCollapsed = !liveEarsCollapsed}
          title="Collapse Live Ears (Ctrl+Shift+L)"
          aria-label="Collapse Live Ears"
        >
          <span class="material-symbols-outlined text-[18px]">chevron_left</span>
        </button>
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

      <!-- Chat Input (Moved from Header) -->
      <div class="p-5 border-t border-white/5 bg-black/20 flex-shrink-0 h-[25%] min-h-[160px] flex flex-col">
        <div class="relative flex w-full h-full bg-white/5 rounded-2xl border border-white/10 transition-colors focus-within:bg-white/10 focus-within:border-white/20 shadow-inner">
          <span class="absolute top-4 left-4 material-symbols-outlined text-[24px] text-on-surface-variant pointer-events-none" style="font-variation-settings: 'FILL' 0;">chat</span>
          <textarea 
            class="w-full h-full bg-transparent border-none text-on-background text-lg focus:ring-0 placeholder-on-surface-variant/50 outline-none pointer-events-auto resize-none hide-scrollbar pt-4 pl-12 pr-4 pb-14" 
            placeholder="Type message to assistant (Press Enter to send)..." 
            bind:value={chatText}
            onkeydown={handleChatKeydown}
          ></textarea>
          <button class="absolute bottom-3 right-3 font-mono-data text-mono-data text-primary bg-primary/10 hover:bg-primary/20 px-4 py-2 rounded-xl text-[12px] font-bold tracking-widest pointer-events-auto transition-all" onclick={handleChatSubmit}>SEND</button>
        </div>
      </div>
    </aside>

    <!-- Expand Live Ears button (shown when collapsed) -->
    {#if liveEarsCollapsed}
      <button
        class="self-center flex-shrink-0 w-8 h-16 glass-panel rounded-xl flex items-center justify-center pointer-events-auto
               text-on-surface-variant/50 hover:text-primary hover:bg-white/10 transition-all duration-200 shadow-xl border border-white/5"
        onclick={() => liveEarsCollapsed = false}
        title="Expand Live Ears (Ctrl+Shift+L)"
        aria-label="Expand Live Ears"
      >
        <span class="material-symbols-outlined text-[18px]">chevron_right</span>
      </button>
    {/if}

    <!-- Right Panel: The Brain (AI Insights) -->
    <section
      class="brain-panel glass-panel {clickthrough ? 'clickthrough-mode' : ''} pointer-events-auto rounded-[24px] flex flex-col overflow-hidden shadow-2xl relative
             transition-all duration-300 ease-in-out
             {brainCollapsed ? 'w-0 opacity-0 p-0 min-w-0 border-0' : liveEarsCollapsed ? 'w-full' : 'w-2/3'}"
      style="{brainCollapsed ? 'pointer-events:none;' : ''}"
    >
      <!-- Ambient Glow effect for the brain panel -->
      <div class="absolute inset-0 bg-gradient-to-br from-white/[0.02] to-transparent pointer-events-none"></div>

      <!-- Panel Header -->
      <div class="flex items-center justify-between p-4 border-b border-white/5 bg-black/20 z-10 flex-shrink-0" onmousedown={startDrag} style="cursor: grab;">
        <div class="flex items-center gap-2">
          <span class="material-symbols-outlined text-primary text-[20px]" data-icon="memory">memory</span>
          <h2 class="font-label-caps text-label-caps text-primary tracking-wider text-glow">The Brain</h2>
        </div>
        <div class="flex items-center gap-2">
          {#if wsState.ragSources && wsState.ragSources.length > 0}
            <div class="text-[10px] text-secondary font-mono-data">
              🔍 Sources: {wsState.ragSources.join(' · ')}
            </div>
          {/if}
          <!-- Catch Me Up -->
          <button
            class="w-7 h-7 flex items-center justify-center rounded-full
                   hover:bg-white/10 transition-colors pointer-events-auto
                   {wsState.isThinking || wsState.transcriptHistory.length === 0
                     ? 'text-on-surface-variant/30 cursor-not-allowed'
                     : 'text-on-surface-variant/60 hover:text-secondary'}"
            title="Catch me up — summarize conversation so far"
            aria-label="Catch Me Up"
            onclick={triggerCatchMeUp}
            disabled={wsState.isThinking || wsState.transcriptHistory.length === 0}
          >
            <span class="material-symbols-outlined text-[18px]">history</span>
          </button>
          <button
            class="w-7 h-7 flex items-center justify-center rounded-full hover:bg-white/10 text-on-surface-variant/60 hover:text-primary transition-colors pointer-events-auto"
            onclick={() => brainCollapsed = !brainCollapsed}
            title="Collapse Brain Panel (Ctrl+Shift+B)"
            aria-label="Collapse Brain Panel"
          >
            <span class="material-symbols-outlined text-[18px]">chevron_right</span>
          </button>
        </div>
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
              <div class="ml-1 response-content leading-relaxed prose prose-invert max-w-none">
                {@html renderedResponse}
                {#if wsState.isThinking}
                  <span class="inline-block w-1.5 h-4 bg-primary align-middle animate-pulse ml-1"></span>
                {/if}
              </div>
            </div>
          </div>
        {/if}

        <!-- Local Cache Section -->
        <div class="settings-section mt-4 pt-4 border-t border-white/5">
          <h3 class="settings-section-title text-on-surface-variant font-label-caps text-label-caps tracking-wider flex items-center gap-2 mb-3">
            <span class="material-symbols-outlined text-[16px]">database</span>
            Local Q&A Cache
          </h3>
          {#if cacheStats}
            <div class="cache-stats-row">
              <span class="cache-stat">
                <strong>{cacheStats.cached_pairs}</strong> answers cached
              </span>
              <span class="cache-stat muted">
                ~{cacheStats.estimated_tokens_saved.toLocaleString()} tokens saved
              </span>
            </div>
          {:else}
            <p class="muted-text text-on-surface-variant/50 text-[12px] mb-3">Cache unavailable (SmolLM2 not enabled)</p>
          {/if}
          <button
            class="btn-danger"
            onclick={clearCache}
            disabled={clearingCache || !cacheStats || cacheStats.cached_pairs === 0}
          >
            {clearingCache ? "Clearing..." : "Clear Cache"}
          </button>
        </div>
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

    <!-- Expand Brain button (shown when collapsed) -->
    {#if brainCollapsed}
      <button
        class="self-center flex-shrink-0 w-8 h-16 glass-panel rounded-xl flex items-center justify-center pointer-events-auto
               text-on-surface-variant/50 hover:text-primary hover:bg-white/10 transition-all duration-200 shadow-xl border border-white/5"
        onclick={() => brainCollapsed = false}
        title="Expand Brain Panel (Ctrl+Shift+B)"
        aria-label="Expand Brain Panel"
      >
        <span class="material-symbols-outlined text-[18px]">chevron_left</span>
      </button>
    {/if}
    {/if}
    {/if}
    <HotkeysPanel />
  </main>

  {#if showSessionReport}
    <div class="pointer-events-auto">
      <SessionReport onClose={() => showSessionReport = false} />
    </div>
  {/if}
</div>



{#if showPrewarmModal}
  <div class="fixed inset-0 z-[9999] pointer-events-auto flex items-center justify-center bg-black/80 backdrop-blur-md">
    <div class="glass-panel p-8 rounded-2xl max-w-lg w-full flex flex-col gap-4 shadow-2xl border border-white/10 relative">
      <div class="flex justify-between items-center mb-2">
        <h2 class="text-xl font-headline-md text-on-background tracking-wide">Pre-Warm Cache</h2>
        <button class="text-on-surface-variant hover:text-white" onclick={() => showPrewarmModal = false}>
          <span class="material-symbols-outlined">close</span>
        </button>
      </div>
      <p class="text-on-surface-variant text-sm mb-2">Paste the Job Description to proactively cache the 50 most likely interview questions and perfect answers based on your stored resume.</p>
      <textarea
        class="w-full h-40 p-4 bg-black/40 border border-white/10 rounded-xl text-on-surface focus:outline-none focus:border-primary font-mono text-sm resize-none"
        placeholder="Paste Job Description here..."
        bind:value={prewarmJobDescription}
      ></textarea>

      <div class="flex justify-end gap-3 mt-4">
        <button
          class="px-4 py-2 rounded-lg text-sm bg-white/10 hover:bg-white/20 text-on-background transition-colors"
          onclick={() => showPrewarmModal = false}
          disabled={prewarmingCache}
        >
          Cancel
        </button>
        <button
          class="px-6 py-2 rounded-lg text-sm bg-primary text-background font-bold hover:bg-primary/90 transition-colors shadow-[0_0_15px_rgba(78,222,163,0.3)] flex items-center gap-2 disabled:opacity-50"
          onclick={prewarmCache}
          disabled={prewarmingCache || !prewarmJobDescription.trim()}
        >
          {#if prewarmingCache}
            <span class="material-symbols-outlined animate-spin text-[18px]">progress_activity</span>
            Pre-warming...
          {:else}
            <span class="material-symbols-outlined text-[18px]">bolt</span>
            Pre-Warm
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}
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
        Your current session limit has been reached. Choose an option below to continue using the BarnOwl.
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
      background: rgba(10, 10, 14, 0.88);
      backdrop-filter: blur(24px) saturate(1.4);
      -webkit-backdrop-filter: blur(24px) saturate(1.4);
      border: 1px solid rgba(255, 255, 255, 0.10);
      box-shadow: 0 24px 48px rgba(0, 0, 0, 0.7), inset 0 1px 0 rgba(255,255,255,0.06);
      transition: background 0.3s ease, backdrop-filter 0.3s ease, border-color 0.3s ease, box-shadow 0.3s ease;
  }

  /* Click-through mode: ultra-transparent so you can read/edit behind the overlay */
  .glass-panel.clickthrough-mode,
  .glass-pill.clickthrough-mode {
      background: rgba(0, 0, 0, 0.08) !important;
      backdrop-filter: blur(3px) !important;
      -webkit-backdrop-filter: blur(3px) !important;
      border-color: rgba(255, 255, 255, 0.06) !important;
      box-shadow: none !important;
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
      text-shadow: 0 1px 3px rgba(0,0,0,0.9);
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
      font-size: 1.05rem;
      line-height: 1.75;
      color: rgba(240, 240, 248, 0.97);
      text-shadow: 0 1px 3px rgba(0,0,0,0.9);
  }
  
  /* Shiki code block overrides — match our dark glass theme */
  .response-content :global(.code-block-wrapper) {
    position: relative;
    margin: 1rem 0;
    border-radius: 10px;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.12);
    box-shadow: 0 4px 16px rgba(0,0,0,0.5);
  }

  .response-content :global(.code-block-header) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.4rem 0.9rem;
    background: rgba(255, 255, 255, 0.06);
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    font-size: 0.72rem;
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
    padding: 0.1rem 0.5rem;
    font-size: 0.68rem;
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
    text-shadow: none;
    border-radius: 0;
    padding: 1.1rem 1.2rem;
    margin: 0;
    font-size: 0.88rem;
    line-height: 1.7;
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
    margin: 0.6rem 0;
    line-height: 1.75;
    color: rgba(235, 235, 245, 0.95);
  }

  /* Bullet / ordered lists */
  .response-content :global(ul),
  .response-content :global(ol) {
    padding-left: 1.5rem;
    margin: 0.5rem 0;
  }

  .response-content :global(li) {
    margin: 0.3rem 0;
    line-height: 1.7;
    color: rgba(225, 225, 240, 0.92);
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

  .cache-stats-row {
    display: flex;
    gap: 16px;
    align-items: center;
    margin-bottom: 10px;
  }
  .cache-stat {
    font-size: 13px;
    color: rgba(255,255,255,0.75);
  }
  .cache-stat strong {
    color: #fff;
    font-size: 16px;
  }
  .btn-danger {
    background: rgba(239, 68, 68, 0.15);
    border: 1px solid rgba(239, 68, 68, 0.4);
    color: #f87171;
    padding: 8px 16px;
    border-radius: 8px;
    font-size: 13px;
    cursor: pointer;
    transition: background 0.2s;
  }
  .btn-danger:hover:not(:disabled) {
    background: rgba(239, 68, 68, 0.3);
  }
  .btn-danger:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }
</style>
