<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { wsState, toggleMockMode } from "$lib/ws.svelte";
  import { getApiUrl, apiFetch } from "$lib/api";
  import { uiState } from "$lib/stores/uiState.svelte.ts";
  import { authState } from "$lib/auth.svelte";

  import StealthTitleBar from "$lib/components/StealthTitleBar.svelte";
  import CodeEditor from "$lib/components/workspace/CodeEditor.svelte";
  import ActivityBar from "$lib/components/workspace/ActivityBar.svelte";
  import WorkspaceSidebar from "$lib/components/workspace/WorkspaceSidebar.svelte";
  import WorkspaceTabs from "$lib/components/workspace/WorkspaceTabs.svelte";
  import ConvPanel from "./ConvPanel.svelte";
  import AnswerPanel from "./AnswerPanel.svelte";

  import SessionReport from "$lib/SessionReport.svelte";
  import SummaryModal from "./SummaryModal.svelte";
  import AuthModal from "$lib/components/AuthModal.svelte";
  import Settings from "$lib/Settings.svelte";
  import type { FileNode } from "$lib/components/workspace/types";
  import type { HeaderAction } from "$lib/types";

  // Wails App methods
  let App: any;
  let statePollInterval: any;
  onMount(() => {
    App = (window as any).go?.main?.App;
    statePollInterval = setInterval(refreshIDEState, 1000);
  });

  onDestroy(() => {
    if (statePollInterval) clearInterval(statePollInterval);
  });

  // State
  let showConvHotkeys = $state(false);

  const convHeaderActions: HeaderAction[] = [
    {
      icon: 'summarize',
      label: 'Summarize',
      onClick: () => { showSummaryModal = true; },
    },
    {
      icon: 'star',
      label: 'STAR',
      onClick: handleStarMethod,
    },
    {
      icon: 'history',
      label: 'Catch Up',
      onClick: handleCatchMeUp,
    },
    {
      icon: 'keyboard',
      label: 'Hotkeys',
      onClick: () => { showConvHotkeys = !showConvHotkeys; },
      get active() { return showConvHotkeys; },
    },

    {
      icon: 'last_page',
      label: 'Collapse',
      onClick: () => { isConvPanelCollapsed = true; },
    }
  ];
  let cacheCount = $state(0);
  let clickthrough = $state(false);
  let showSessionReport = $state(false);
  let showSummaryModal = $state(false);
  let isAuthModalOpen = $state(false);
  let activeAction = $state<string | null>(null);

  let answerPanelRef = $state<any>();

  let includeActiveDocContext = $state(true);
  let activeDocumentName = $state('');
  let workspaceTree = $state<FileNode[]>([]);
  let openTabs = $state<any[]>([]);
  let activeDocumentPath = $state('');
  let activeDocumentContent = $state('');
  let expandedMap = new Map<string, boolean>();

  async function refreshIDEState() {
    if (App?.GetIDEState) {
      const state = await App.GetIDEState();
      if (state) {
        includeActiveDocContext = state.includeActiveDocContext;
        activeDocumentName = state.activeDocumentName || '';

        const recordExpanded = (nodes: FileNode[]) => {
          for (const n of nodes) {
            if (n.isDirectory && n.isExpanded !== undefined) {
              expandedMap.set(n.path, n.isExpanded);
            }
            if (n.children) recordExpanded(n.children);
          }
        };
        recordExpanded(workspaceTree);

        const mapTree = (nodes: any[]): FileNode[] => {
          if (!nodes) return [];
          return nodes.map((n: any) => ({
            ...n,
            isDirectory: n.isDir,
            isExpanded: expandedMap.has(n.path) ? expandedMap.get(n.path) : (n.isExpanded ?? false),
            children: mapTree(n.children)
          }));
        };

        workspaceTree = mapTree(state.workspaceTree);

        const rawDocs = state.openDocuments || [];
        openTabs = rawDocs.map((doc: any) => ({
          name: doc.name,
          path: doc.path
        }));
        activeDocumentPath = state.activeDocumentPath || '';
        activeDocumentContent = state.activeDocumentContent || '';
      }
    }
    try {
      if (App?.GetCacheStats) {
        const stats = await App.GetCacheStats();
        if (stats && stats.count !== undefined) cacheCount = stats.count;
      } else {
        const stats = await apiFetch(`${getApiUrl()}/api/cache/stats`, { method: 'GET' });
        if (stats?.count !== undefined) cacheCount = stats.count;
      }
    } catch { /* ignore if endpoint not available */ }
  }

  const isGated = $derived(authState.authMode === 'saas' && (!authState.user || (!authState.byok_pass_active && authState.remaining_sessions <= 0)));

  async function handleSelectNode(node: any) {
    if (App?.OpenFile && node && node.path && !node.isDirectory) {
      await App.OpenFile(node.path);
      await refreshIDEState();
    }
  }

  async function handleTabSelect(path: string) {
    if (App?.SetActiveDocument) {
      await App.SetActiveDocument(path);
      await refreshIDEState();
    }
  }

  async function handleTabClose(path: string) {
    if (App?.CloseDocument) {
      await App.CloseDocument(path);
      await refreshIDEState();
    }
  }

  async function handleOpenFile() {
    if (App?.PromptOpenFile) {
      await App.PromptOpenFile();
      await refreshIDEState();
    }
  }

  async function handleOpenFolder() {
    if (App?.PromptOpenDirectory) {
      await App.PromptOpenDirectory();
      await refreshIDEState();
    }
  }

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
    if (action === 'explorer') {
      // Toggle IDE file tree + CodeEditor panel
      activeAction = activeAction === 'explorer' ? '' : 'explorer';
    } else if (action === 'mock') {
      toggleMockMode(!wsState.isMockMode);
    } else if (action === 'settings') {
      activeAction = activeAction === 'settings' ? '' : 'settings';
    } else {
      activeAction = action;
    }
  }


  const savedConvWidth = typeof window !== 'undefined'
    ? parseInt(localStorage.getItem('barnowl_conv_panel_width') || '380', 10)
    : 380;
  let convPanelWidth = $state(Math.max(280, Math.min(savedConvWidth, window.innerWidth * 0.6)));
  let isConvPanelCollapsed = $state(false);
  let isPanelResizing = $state(false);

  function startPanelResize(e: MouseEvent) {
    e.preventDefault();
    isPanelResizing = true;
    const startX = e.clientX;
    const startWidth = convPanelWidth;

    const onMouseMove = (ev: MouseEvent) => {
      const maxWidth = window.innerWidth * 0.6;
      convPanelWidth = Math.max(280, Math.min(startWidth + (ev.clientX - startX), maxWidth));
    };

    const onMouseUp = () => {
      isPanelResizing = false;
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseup', onMouseUp);
      try { localStorage.setItem('barnowl_conv_panel_width', String(Math.round(convPanelWidth))); } catch {}
    };

    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);
  }

  function sendChat(text: string) {
    if (!text.trim()) return;
    wsState.isThinking = true;
    if (App?.SendChat) {
      App.SendChat(text);
    } else {
      apiFetch(`${getApiUrl()}/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: text })
      });
    }
  }

  function sendChip(chip: { id: string; text: string }) {
    sendChat(chip.text);
    wsState.pendingTranscripts = wsState.pendingTranscripts.filter(
      (c: any) => c.id !== chip.id
    );
  }

  function dismissChip(chipId: string) {
    wsState.pendingTranscripts = wsState.pendingTranscripts.filter((c: any) => c.id !== chipId);
  }

  function clearAllChips() {
    wsState.pendingTranscripts = [];
  }

  let starPrimed = $state(false);
  let starPrimedTimer: ReturnType<typeof setTimeout>;

  function handleStarMethod() {
    sendChat("Format your next response using the STAR method (Situation, Task, Action, Result).");
    wsState.isThinking = true;
    starPrimed = true;
    if (starPrimedTimer) clearTimeout(starPrimedTimer);
    starPrimedTimer = setTimeout(() => {
      starPrimed = false;
    }, 15000);
  }

  function handleCatchMeUp() {
    sendChat("Please catch me up on the current context of the interview or conversation.");
    wsState.isThinking = true;
  }

  function handleCopyAll() {
    if (wsState.response) {
      navigator.clipboard.writeText(wsState.response).catch(console.error);
    }
  }

  function handleClearCache() {
    apiFetch(`${getApiUrl()}/api/cache/clear`, { method: 'POST' }).then(() => {
        cacheCount = 0;
    }).catch(e => console.error("Failed to clear cache:", e));
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
          App.AnalyzeVision(b64, "Analyze this technical interview screen and provide key hints, solution or code concisely.");
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
      {#if activeDocumentName && includeActiveDocContext}
        <div class="badge border-primary/30 text-primary/80" style="background: rgba(74, 222, 128, 0.05); text-transform: none;">
          📄 Context: {activeDocumentName}
        </div>
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
    <!-- ActivityBar always visible -->
    <ActivityBar {activeAction} onAction={handleAction} />

    <!-- IDE Mode: File tree sidebar (only when explorer is active) -->
    {#if activeAction === 'explorer'}
      <WorkspaceSidebar
        nodes={workspaceTree}
        activePath={activeDocumentPath}
        totalDocs={openTabs.length}
        isOpen={true}
        onToggle={() => handleAction('explorer')}
        onSelect={handleSelectNode}
        onOpenFile={handleOpenFile}
        onOpenFolder={handleOpenFolder}
      />
      <!-- CodeEditor panel (only in IDE mode) -->
      <div class="ide-editor-pane">
        <WorkspaceTabs
          tabs={openTabs}
          activePath={activeDocumentPath}
          onTabSelect={handleTabSelect}
          onTabClose={handleTabClose}
        />
        <div class="ide-editor-content">
          <CodeEditor
            content={activeDocumentContent}
            language="markdown"
            readonly={false}
          />
          {#if !activeDocumentContent}
            <div class="editor-empty-state">
              <span class="material-symbols-outlined">description</span>
              <p>Open a note or cheat sheet</p>
            </div>
          {/if}
        </div>
      </div>
    {/if}

    {#if !isConvPanelCollapsed}
    <!-- LEFT: Conversation Panel -->
    <div class="conv-panel-wrapper" style="width: {convPanelWidth}px;">
      <ConvPanel
        transcriptHistory={wsState.transcriptHistory}
        pendingTranscripts={wsState.pendingTranscripts}
        ragSources={wsState.ragSources}
        isListening={wsState.isListening}
        isMockMode={wsState.isMockMode}
        isPTTHeld={wsState.isPTTHeld}
        isThinking={wsState.isThinking}
        {includeActiveDocContext}
        {activeDocumentName}
        headerActions={convHeaderActions}
        showHotkeys={showConvHotkeys}
        onSendChat={sendChat}
        onSendChip={sendChip}
        onDismissChip={dismissChip}
        onClearChips={clearAllChips}
        onSelectTranscript={(text: string, answer?: string) => {
          if (answerPanelRef) {
            if (answer && answerPanelRef.showLocalAnswer) {
              answerPanelRef.showLocalAnswer(answer);
            } else if (answerPanelRef.showCachedAnswerFor) {
              answerPanelRef.showCachedAnswerFor(text);
            }
          }
        }}
        onToggleMic={async () => {
          const newState = await (window as any).go.main.App.ToggleMic();
          wsState.isListening = newState;
        }}
      />
    </div>

    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <!-- Resizable divider between ConvPanel and AnswerPanel -->
    <div
      class="panel-resizer"
      class:resizing={isPanelResizing}
      onmousedown={startPanelResize}
      role="separator"
      aria-orientation="vertical"
      title="Drag to resize panels"
    >
      <div class="resizer-indicator"></div>
    </div>
    {/if}
    {#if isConvPanelCollapsed}
      <button
        class="absolute left-0 top-1/2 -translate-y-1/2 z-50 bg-surface border border-white/10 rounded-r-md p-1 hover:bg-white/10 transition-colors shadow-lg"
        onclick={() => { isConvPanelCollapsed = false; }}
        title="Expand Live Session"
      >
        <span class="material-symbols-outlined text-white/70 hover:text-white">keyboard_double_arrow_right</span>
      </button>
    {/if}

    <!-- RIGHT: Answer Panel (always visible) -->
    <div class="answer-panel-wrapper">
      <AnswerPanel
        bind:this={answerPanelRef}
        response={wsState.response}
        isThinking={wsState.isThinking}
        {cacheCount}
        ragSources={wsState.ragSources}
        onClearCache={handleClearCache}
        onCopyAll={handleCopyAll}
      />
    </div>

    <!-- Settings overlay (absolute positioned, same as before) -->
    {#if activeAction === 'settings'}
      <div class="settings-overlay">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-lg font-bold">Settings</h2>
          <button onclick={() => handleAction('settings')}>
            <span class="material-symbols-outlined">close</span>
          </button>
        </div>
        <div class="flex-1 overflow-y-auto hide-scrollbar">
          <Settings embedded={true} />
        </div>
      </div>
    {/if}
  </div>

  <!-- Modals -->

  {#if showSummaryModal}
    <SummaryModal onClose={() => showSummaryModal = false} />
  {/if}

  {#if showSessionReport}
    <SessionReport onClose={() => showSessionReport = false} />
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
    display: flex;
    flex-direction: row;
    overflow: hidden;
    position: relative;
  }

  .conv-panel-wrapper {
    flex-shrink: 0;
    min-width: 280px;
    max-width: 60vw;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .answer-panel-wrapper {
    flex-grow: 1;
    min-width: 300px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  /* Reuse same resizer style as IDEShell */
  .panel-resizer {
    width: 6px;
    cursor: col-resize;
    background: transparent;
    z-index: 60;
    transition: background-color 0.15s ease;
    flex-shrink: 0;
    user-select: none;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .panel-resizer:hover,
  .panel-resizer.resizing {
    background: rgba(74, 222, 128, 0.3);
  }

  .resizer-indicator {
    width: 2px;
    height: 32px;
    border-radius: 2px;
    background: rgba(255, 255, 255, 0.12);
    transition: all 0.15s ease;
  }

  .panel-resizer:hover .resizer-indicator,
  .panel-resizer.resizing .resizer-indicator {
    background: #4ade80;
    height: 48px;
  }

  /* IDE mode panes */
  .ide-editor-pane {
    display: flex;
    flex-direction: column;
    width: 320px;
    flex-shrink: 0;
    border-right: 1px solid rgba(255,255,255,0.06);
    overflow: hidden;
  }

  .ide-editor-content {
    flex-grow: 1;
    position: relative;
    overflow: hidden;
  }

  .editor-empty-state {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    opacity: 0.25;
    pointer-events: none;
  }

  .settings-overlay {
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    width: 400px;
    background: rgba(15, 15, 15, 0.97);
    border-left: 1px solid rgba(255,255,255,0.1);
    backdrop-filter: blur(12px);
    z-index: 50;
    display: flex;
    flex-direction: column;
    padding: 24px;
    overflow-y: auto;
  }

  .hide-scrollbar::-webkit-scrollbar {
    display: none;
  }
  .hide-scrollbar {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }
</style>
