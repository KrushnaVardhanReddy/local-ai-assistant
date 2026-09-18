<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import type { FileNode, WorkspaceTab, IDEStatus } from './types';
  import ActivityBar from './ActivityBar.svelte';
  import WorkspaceSidebar from './WorkspaceSidebar.svelte';
  import WorkspaceTabs from './WorkspaceTabs.svelte';
  import StatusBar from './StatusBar.svelte';

  let {
    workspaceTree = [],
    openTabs = [],
    activePath = '',
    totalDocs = 0,
    status = { left: {}, right: {} },
    activeAction = '',
    onAction,
    children,
    rightDrawer,
    settingsPanel,
    onFileOpen,
    onOpenFolder,
    onTabSelect,
    onTabClose,
    onSelect
  } = $props<{
    workspaceTree?: FileNode[];
    openTabs?: WorkspaceTab[];
    activePath?: string;
    totalDocs?: number;
    status?: IDEStatus;
    activeAction?: string;
    onAction?: (action: string) => void;
    children?: import('svelte').Snippet;
    rightDrawer?: import('svelte').Snippet;
    settingsPanel?: import('svelte').Snippet;
    onFileOpen?: () => void;
    onOpenFolder?: () => void;
    onTabSelect?: (path: string) => void;
    onTabClose?: (path: string) => void;
    onSelect?: (node: FileNode) => void;
  }>();

  // State
  let internalExplorerOpen = $state(false);
  let isBottomPanelOpen = $state(true);
  let internalIsCopilotOpen = $state(false);
  let internalActiveAction = $state('');

  // Resizable drawer width state (default 420px, min 260px, max 80vw)
  const savedWidth = typeof window !== 'undefined' ? parseInt(localStorage.getItem('barnowl_copilot_drawer_width') || '420', 10) : 420;
  let drawerWidth = $state(Number.isNaN(savedWidth) ? 420 : Math.max(260, Math.min(savedWidth, 1200)));
  let isResizing = $state(false);

  function startResize(e: MouseEvent) {
    e.preventDefault();
    isResizing = true;
    const startX = e.clientX;
    const startWidth = drawerWidth;

    const onMouseMove = (moveEvent: MouseEvent) => {
      const deltaX = startX - moveEvent.clientX; // Moving left increases drawer width
      const maxWidth = window.innerWidth * 0.75;
      const newWidth = Math.max(260, Math.min(startWidth + deltaX, maxWidth));
      drawerWidth = newWidth;
    };

    const onMouseUp = () => {
      isResizing = false;
      window.removeEventListener('mousemove', onMouseMove);
      window.removeEventListener('mouseup', onMouseUp);
      try {
        localStorage.setItem('barnowl_copilot_drawer_width', String(Math.round(drawerWidth)));
      } catch (err) {
        console.error('Failed to persist drawer width', err);
      }
    };

    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);
  }

  let currentActiveAction = $derived(onAction ? activeAction : internalActiveAction);
  let isExplorerOpen = $derived(currentActiveAction === 'explorer' || (currentActiveAction === '' && internalExplorerOpen));
  let isCopilotOpen = $derived(onAction ? (activeAction === 'copilot' || activeAction === 'keys') : internalIsCopilotOpen);

  function handleAction(action: string) {
    if (onAction) {
      onAction(action);
      return;
    }

    if (action === 'explorer') {
      internalExplorerOpen = !isExplorerOpen;
      internalActiveAction = internalExplorerOpen ? 'explorer' : '';
    } else if (action === 'copilot') {
      internalIsCopilotOpen = !internalIsCopilotOpen;
      internalActiveAction = internalIsCopilotOpen ? action : '';
    } else {
      internalActiveAction = internalActiveAction === action ? '' : action;
    }
  }

  function handleSidebarToggle(isOpen: boolean) {
    if (onAction) {
      if (isOpen && currentActiveAction !== 'explorer') {
        onAction('explorer');
      } else if (!isOpen && currentActiveAction === 'explorer') {
        onAction('explorer');
      }
    } else {
      internalExplorerOpen = isOpen;
      internalActiveAction = isOpen ? 'explorer' : (internalActiveAction === 'explorer' ? '' : internalActiveAction);
    }
  }

  function handleNodeSelect(node: FileNode) {
    if (!node.isDirectory) {
      if (onSelect) {
        onSelect(node);
      } else if (onTabSelect) {
        onTabSelect(node.path);
      }
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    // Ctrl+B: Toggle Explorer Sidebar
    if (e.ctrlKey && e.key.toLowerCase() === 'b') {
      e.preventDefault();
      handleAction('explorer');
    }
    // Ctrl+J: Toggle bottom panel / status bar
    else if (e.ctrlKey && e.key.toLowerCase() === 'j') {
      e.preventDefault();
      isBottomPanelOpen = !isBottomPanelOpen;
    }
    // Ctrl+\: Toggle Copilot side-drawer
    else if (e.ctrlKey && e.key === '\\') {
      e.preventDefault();
      handleAction('copilot');
    }
    // Ctrl+P: Quick file switch modal (placeholder)
    else if (e.ctrlKey && e.key.toLowerCase() === 'p') {
      e.preventDefault();
      // Emitting an event or calling a prop function could be added here
      console.log('Ctrl+P triggered (Quick file switch)');
    }
  }

  onMount(() => {
    window.addEventListener('keydown', handleKeydown);
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleKeydown);
  });

</script>

<div class="ide-shell">
  <div class="main-layout">
    <ActivityBar activeAction={currentActiveAction} onAction={handleAction} />

    <WorkspaceSidebar
      nodes={workspaceTree}
      {activePath}
      {totalDocs}
      isOpen={isExplorerOpen}
      onToggle={handleSidebarToggle}
      onSelect={handleNodeSelect}
      onOpenFile={() => onFileOpen?.()}
      onOpenFolder={() => onOpenFolder?.()}
    />

    <div class="center-editor-pane">
      {#if openTabs.length > 0}
        <WorkspaceTabs
          tabs={openTabs}
          {activePath}
          onTabSelect={(p) => onTabSelect?.(p)}
          onTabClose={(p) => onTabClose?.(p)}
        />
      {/if}

      <div class="editor-content">
        {#if children}
          {@render children()}
        {/if}
      </div>
    </div>

    {#if isCopilotOpen}
      <!-- Resizer divider handle -->
      <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
      <div
        class="resizer-handle"
        class:resizing={isResizing}
        onmousedown={startResize}
        role="separator"
        aria-orientation="vertical"
        aria-valuenow={drawerWidth}
        title="Drag to resize Brain / Copilot drawer"
      >
        <div class="resizer-indicator"></div>
      </div>

      <div class="right-drawer" style="width: {drawerWidth}px;">
        {#if rightDrawer}
          {@render rightDrawer()}
        {:else}
          <div class="placeholder-drawer">
            <span>Copilot Drawer</span>
          </div>
        {/if}
      </div>
    {/if}

    {#if activeAction === 'settings'}
      <div class="settings-panel absolute inset-y-0 right-0 z-50">
        {#if settingsPanel}
          {@render settingsPanel()}
        {/if}
      </div>
    {/if}
  </div>

  {#if isBottomPanelOpen}
    <StatusBar {status} />
  {/if}
</div>

<style>
  .ide-shell {
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
    flex-grow: 1;
    overflow: hidden;
    background-color: var(--hud-bg, #0a0a0a);
    color: var(--hud-text-white, #ffffff);
  }

  .main-layout {
    display: flex;
    flex-grow: 1;
    overflow: hidden;
    position: relative;
  }

  .center-editor-pane {
    display: flex;
    flex-direction: column;
    flex-grow: 1;
    min-width: 0;
  }

  .editor-content {
    flex-grow: 1;
    position: relative;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  /* Resizer handle between Center Editor and Right Drawer */
  .resizer-handle {
    width: 6px;
    cursor: col-resize;
    position: relative;
    background: transparent;
    z-index: 60;
    transition: background-color 0.15s ease;
    flex-shrink: 0;
    user-select: none;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .resizer-handle:hover,
  .resizer-handle.resizing {
    background: rgba(74, 222, 128, 0.4); /* Primary accent tint on drag/hover */
  }

  .resizer-indicator {
    width: 2px;
    height: 32px;
    border-radius: 2px;
    background: rgba(255, 255, 255, 0.15);
    transition: background-color 0.15s ease;
  }

  .resizer-handle:hover .resizer-indicator,
  .resizer-handle.resizing .resizer-indicator {
    background: #4ade80; /* Glowing primary pill */
    height: 48px;
  }

  .right-drawer {
    min-width: 260px;
    max-width: 80vw;
    background: rgba(15, 15, 15, 0.9);
    border-left: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    display: flex;
    flex-direction: column;
    z-index: 50;
    flex-shrink: 0;
  }

  .settings-panel {
    width: 400px;
    background: rgba(15, 15, 15, 0.95);
    border-left: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    box-shadow: -4px 0 16px rgba(0,0,0,0.5);
    display: flex;
    flex-direction: column;
    padding: 24px;
    overflow-y: auto;
  }

  .placeholder-drawer {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: rgba(255, 255, 255, 0.4);
    font-style: italic;
  }
</style>
