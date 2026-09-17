<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { wsState, toggleMockMode } from "$lib/ws.svelte";
  import { getApiUrl, apiFetch } from "$lib/api";
  import { uiState } from "$lib/stores/uiState.svelte.ts";
  import { authState } from "$lib/auth.svelte";

  import StealthTitleBar from "$lib/components/StealthTitleBar.svelte";
  import IDEShell from "$lib/components/workspace/IDEShell.svelte";
  import CodeEditor from "$lib/components/workspace/CodeEditor.svelte";

  import LiveEarsDrawer from "./LiveEarsDrawer.svelte";
  import BrainDrawer from "./BrainDrawer.svelte";

  import SessionReport from "$lib/SessionReport.svelte";
  import AuthModal from "$lib/components/AuthModal.svelte";
  import HotkeysPanel from "$lib/components/HotkeysPanel.svelte";
  import Settings from "$lib/Settings.svelte";

  // Wails App methods
  let App: any;
  onMount(() => {
    App = (window as any).go?.main?.App;
  });

  // State
  let activeDrawer = $state('copilot'); // 'ears' or 'copilot'
  let clickthrough = $state(false);
  let showSessionReport = $state(false);
  let isAuthModalOpen = $state(false);
  let activeAction = $state('copilot');

  const isGated = $derived(authState.authMode === 'saas' && (!authState.user || (!authState.byok_pass_active && authState.remaining_sessions <= 0)));

  // Mock mode toggle wrapper
  function handleMockModeToggle() {
    const isNowEnabled = !wsState.isMockMode;
    toggleMockMode(isNowEnabled);
    if (isNowEnabled) {
      apiFetch(`${getApiUrl()}/ptt/start`, { method: 'POST' }).catch(console.error);
    } else {
      showSessionReport = true;
      apiFetch(`${getApiUrl()}/ptt/stop`, { method: 'POST' }).catch(console.error);
    }
  }

  // ActivityBar action handler
  function handleAction(action: string) {
    if (action === 'ears') {
      activeAction = 'ears';
      activeDrawer = 'ears';
    } else if (action === 'copilot') {
      activeAction = 'copilot';
      activeDrawer = 'copilot';
    } else if (action === 'mock') {
      handleMockModeToggle();
    } else if (action === 'keys') {
      uiState.hotkeysPanelOpen = !uiState.hotkeysPanelOpen;
    } else if (action === 'settings') {
      activeAction = activeAction === 'settings' ? activeDrawer : 'settings';
    } else {
      activeAction = action;
    }
  }

  async function handleClearContext() {
    try {
      if (App?.ClearState) {
        await App.ClearState();
      } else {
        await apiFetch(`${getApiUrl()}/history/clear`, { method: 'POST' });
      }
      wsState.transcript = "";
      wsState.response = "";
      wsState.isThinking = false;
      wsState.ragSources = [];
      wsState.pendingTranscripts = [];
      wsState.transcriptHistory = [];
    } catch (e) {
      console.error("Failed to clear context", e);
    }
  }

  async function handleSnip() {
    if (App?.CaptureScreen) {
      const b64 = await App.CaptureScreen();
      if (b64) {
        if (App.AnalyzeVision) {
          App.AnalyzeVision(b64);
        } else {
          apiFetch(`${getApiUrl()}/api/vision/analyze`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ image_base64: b64, type: 'snip' })
          });
        }
      }
    }
  }

  async function toggleStealthMode() {
    clickthrough = !clickthrough;
    if (App?.SetClickthrough) {
      await App.SetClickthrough(clickthrough);
    }
  }
</script>

<div class="interview-hud" class:clickthrough-mode={clickthrough}>
  <StealthTitleBar showTitle={false} />

  <!-- Header ToolBar -->
  <div class="header-toolbar">
    <div class="branding">
      <span class="font-bold text-primary tracking-wider">BarnOwl AI</span>
      {#if wsState.isMockMode}
        <div class="badge mock-badge"><span class="w-2 h-2 rounded-full bg-red-500 animate-pulse inline-block mr-1"></span>Mock</div>
      {:else if wsState.isListening}
        <div class="badge live-badge"><span class="w-2 h-2 rounded-full bg-primary animate-pulse inline-block mr-1"></span>Live</div>
      {:else}
        <div class="badge off-badge"><span class="w-2 h-2 rounded-full bg-red-500 inline-block mr-1"></span>Mic Off</div>
      {/if}
      {#if wsState.isPTTHeld}
        <div class="badge ptt-badge text-primary animate-pulse border-primary"><span class="material-symbols-outlined text-[12px] mr-1">mic</span>PTT</div>
      {/if}
    </div>

    <div class="toolbar-actions" style="--wails-draggable: no-drag">
      <button class="tool-btn" onclick={handleSnip} title="Vision Snip">
        <span class="material-symbols-outlined">screenshot_monitor</span>
        <span class="tool-label">Snip</span>
      </button>
      <button class="tool-btn" onclick={() => showSessionReport = true} title="Session Report">
        <span class="material-symbols-outlined">analytics</span>
        <span class="tool-label">Report</span>
      </button>
      <button class="tool-btn" onclick={handleClearContext} title="Clear Context">
        <span class="material-symbols-outlined">mop</span>
        <span class="tool-label">Clear</span>
      </button>
      <button class="tool-btn" class:text-primary={clickthrough} onclick={toggleStealthMode} title="Stealth Mode (Clickthrough)">
        <span class="material-symbols-outlined">{clickthrough ? 'mouse' : 'back_hand'}</span>
        <span class="tool-label">Stealth</span>
      </button>
    </div>
  </div>

  <div class="main-content-area">
    <IDEShell
      {activeAction}
      onAction={handleAction}
    >
      <!-- Center Slot: Code Editor -->
      <div class="h-full w-full flex flex-col bg-surface/50 relative">
         <CodeEditor
           content=""
           language="markdown"
           readonly={false}
         />
         {#if !activeAction || activeAction === 'explorer'}
           <!-- Empty state overlay to suggest opening files -->
           <div class="absolute inset-0 pointer-events-none flex items-center justify-center">
             <div class="text-center opacity-30">
               <span class="material-symbols-outlined text-6xl mb-4 block">description</span>
               <p class="text-lg">Open a note or cheat sheet</p>
               <p class="text-sm mt-2">Use the Explorer to load files</p>
             </div>
           </div>
         {/if}
      </div>

      <!-- Right Drawer Slot -->
      {#snippet rightDrawer()}
        {#if activeDrawer === 'ears'}
          <LiveEarsDrawer />
        {:else}
          <BrainDrawer />
        {/if}
      {/snippet}

      <!-- Settings Panel Slot -->
      {#snippet settingsPanel()}
        <div class="h-full flex flex-col">
          <div class="flex justify-between items-center mb-4">
            <h2 class="text-lg font-bold text-on-surface">Settings</h2>
            <button class="text-on-surface-variant hover:text-error transition-colors" onclick={() => handleAction('copilot')}>
              <span class="material-symbols-outlined">close</span>
            </button>
          </div>
          <div class="flex-1 overflow-y-auto hide-scrollbar">
            <Settings />
          </div>
        </div>
      {/snippet}
    </IDEShell>
  </div>

  <!-- Modals -->
  {#if showSessionReport}
    <div class="fixed inset-0 z-[200] bg-black/80 backdrop-blur-md flex items-center justify-center p-8 pointer-events-auto">
      <div class="w-full max-w-4xl max-h-[90vh] overflow-y-auto bg-surface/90 border border-white/10 rounded-2xl shadow-2xl relative hide-scrollbar">
        <button class="absolute top-4 right-4 text-on-surface-variant hover:text-white" onclick={() => showSessionReport = false}>
          <span class="material-symbols-outlined text-2xl">close</span>
        </button>
        <div class="p-6">
          <SessionReport />
        </div>
      </div>
    </div>
  {/if}

  {#if uiState.hotkeysPanelOpen}
    <div class="fixed inset-0 z-[200] bg-black/60 backdrop-blur-sm flex items-center justify-center p-8 pointer-events-auto" onclick={() => uiState.hotkeysPanelOpen = false}>
      <div class="w-full max-w-2xl bg-surface/95 border border-white/10 rounded-xl shadow-2xl" onclick={(e) => e.stopPropagation()}>
        <HotkeysPanel />
      </div>
    </div>
  {/if}

  {#if isGated}
    <AuthModal bind:isOpen={isAuthModalOpen} />
  {/if}

  {#if wsState.sessionExpired}
    <div class="fixed inset-0 z-[9999] pointer-events-auto flex items-center justify-center bg-black/80 backdrop-blur-md">
      <div class="bg-surface/90 p-8 rounded-2xl max-w-md w-full text-center flex flex-col items-center gap-6 shadow-2xl border border-error/20 relative overflow-hidden">
        <div class="absolute inset-0 bg-gradient-to-br from-error/10 to-transparent pointer-events-none"></div>
        <div class="w-16 h-16 rounded-full bg-error/20 flex items-center justify-center text-error mb-2">
          <span class="material-symbols-outlined text-[32px]">hourglass_disabled</span>
        </div>
        <h2 class="text-xl font-bold text-on-surface tracking-wide">Session Ended</h2>
        <p class="text-on-surface-variant text-sm">Your session limit has been reached. Please upgrade to continue.</p>
        <div class="flex gap-4 w-full mt-4">
          <a href="/pricing" target="_blank" rel="noopener noreferrer" class="flex-1 py-3 px-4 rounded-xl bg-primary hover:bg-primary/90 text-background font-bold uppercase text-xs tracking-wider transition-all duration-200 shadow-[0_0_20px_rgba(74,222,128,0.3)]">Upgrade</a>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  .interview-hud {
    display: flex;
    flex-direction: column;
    height: 100vh;
    width: 100vw;
    background: var(--hud-bg, #0a0a0a);
    color: var(--hud-text-white, #ffffff);
    overflow: hidden;
    position: relative;
    pointer-events: auto;
  }

  .interview-hud.clickthrough-mode {
    background: rgba(10, 10, 14, 0.35) !important;
    backdrop-filter: blur(6px) !important;
    pointer-events: none;
  }

  .interview-hud.clickthrough-mode :global(.tool-btn),
  .interview-hud.clickthrough-mode :global(.action-btn) {
    pointer-events: auto;
  }

  .header-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 16px;
    height: 48px;
    background: rgba(15, 15, 15, 0.8);
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    backdrop-filter: blur(10px);
    z-index: 10;
  }

  .branding {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .badge {
    display: flex;
    align-items: center;
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    padding: 2px 6px;
    border-radius: 4px;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .live-badge { background: rgba(74, 222, 128, 0.1); color: #4ade80; border-color: rgba(74, 222, 128, 0.3); }
  .off-badge { background: rgba(239, 68, 68, 0.1); color: #ef4444; border-color: rgba(239, 68, 68, 0.3); }
  .mock-badge { background: rgba(245, 158, 11, 0.1); color: #f59e0b; border-color: rgba(245, 158, 11, 0.3); }
  .ptt-badge { background: rgba(74, 222, 128, 0.15); border-color: rgba(74, 222, 128, 0.5); }

  .toolbar-actions {
    display: flex;
    gap: 8px;
  }

  .tool-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    color: rgba(255, 255, 255, 0.5);
    cursor: pointer;
    transition: all 0.2s;
    padding: 4px 8px;
    border-radius: 6px;
  }

  .tool-btn:hover {
    color: white;
    background: rgba(255, 255, 255, 0.08);
  }

  .tool-btn .material-symbols-outlined {
    font-size: 18px;
    margin-bottom: 2px;
  }

  .tool-label {
    font-size: 9px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.02em;
  }

  .main-content-area {
    flex-grow: 1;
    position: relative;
    overflow: hidden;
  }

  .hide-scrollbar::-webkit-scrollbar {
    display: none;
  }
  .hide-scrollbar {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }
</style>
