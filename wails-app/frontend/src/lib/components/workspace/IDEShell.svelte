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
    children,
    rightDrawer,
    settingsPanel,
    onFileOpen,
    onOpenFolder,
    onTabSelect,
    onTabClose
  } = $props<{
    workspaceTree?: FileNode[];
    openTabs?: WorkspaceTab[];
    activePath?: string;
    totalDocs?: number;
    status?: IDEStatus;
    children?: import('svelte').Snippet;
    rightDrawer?: import('svelte').Snippet;
    settingsPanel?: import('svelte').Snippet;
    onFileOpen?: () => void;
    onOpenFolder?: () => void;
    onTabSelect?: (path: string) => void;
    onTabClose?: (path: string) => void;
  }>();

  // State
  let isExplorerOpen = $state(false);
  let isBottomPanelOpen = $state(true);
  let isCopilotOpen = $state(false);
  let activeAction = $state('');

  function handleAction(action: string) {
    if (action === 'explorer') {
      isExplorerOpen = !isExplorerOpen;
      activeAction = isExplorerOpen ? 'explorer' : '';
    } else if (action === 'copilot') {
      isCopilotOpen = !isCopilotOpen;
      activeAction = isCopilotOpen ? 'copilot' : '';
    } else {
      // Handle other actions (search, settings) by setting active state
      // or opening respective modals
      activeAction = activeAction === action ? '' : action;
    }
  }

  function handleSidebarToggle(isOpen: boolean) {
    isExplorerOpen = isOpen;
    activeAction = isOpen ? 'explorer' : (activeAction === 'explorer' ? '' : activeAction);
  }

  function handleNodeSelect(node: FileNode) {
    if (!node.isDirectory && onTabSelect) {
      onTabSelect(node.path);
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
    <ActivityBar {activeAction} onAction={handleAction} />

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
      <div class="right-drawer">
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
      <div class="settings-panel fixed inset-y-0 right-0 z-50">
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
    height: 100vh;
    width: 100vw;
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

  .right-drawer {
    width: 320px;
    background: rgba(15, 15, 15, 0.9);
    border-left: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    display: flex;
    flex-direction: column;
    z-index: 50;
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
