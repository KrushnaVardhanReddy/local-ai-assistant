<script lang="ts">
  import type { FileNode } from './types';
  import FileTree from './FileTree.svelte';

  let {
    nodes,
    activePath = '',
    totalDocs = 0,
    isOpen = false,
    onToggle,
    onSelect,
    onOpenFile,
    onOpenFolder
  } = $props<{
    nodes: FileNode[];
    activePath?: string;
    totalDocs?: number;
    isOpen?: boolean;
    onToggle: (isOpen: boolean) => void;
    onSelect: (node: FileNode) => void;
    onOpenFile: () => void;
    onOpenFolder: () => void;
  }>();

  $effect(() => {
    // Fetch initially indexed paths
    if ((window as any).go?.main?.App?.GetIndexedPaths && !(window as any)._indexedPaths) {
      (window as any).go.main.App.GetIndexedPaths().then((paths: string[]) => {
        (window as any)._indexedPaths = paths || [];
        window.dispatchEvent(new Event('indexed-paths-updated'));
      });
    }
  });
</script>

<div class="workspace-sidebar-container" class:open={isOpen}>
  <!-- Drawer Content -->
  <div class="drawer">
    <div class="drawer-header">
      <span class="title">Workspace</span>
      <div class="actions">
        <button class="action-btn" onclick={() => onOpenFile()} aria-label="Open File">
          <span class="material-symbols-outlined icon">note_add</span>
          <span class="btn-label">FILE</span>
        </button>
        <button class="action-btn" onclick={() => onOpenFolder()} aria-label="Open Folder">
          <span class="material-symbols-outlined icon">create_new_folder</span>
          <span class="btn-label">FOLDER</span>
        </button>
      </div>
    </div>
    <div class="drawer-content">
      <FileTree {nodes} {activePath} {onSelect} />
    </div>
  </div>
</div>

<style>
  .workspace-sidebar-container {
    width: 0;
    height: 100%;
    overflow: hidden;
    transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    background: rgba(18, 18, 18, 0.95);
    border-right: none;
    flex-shrink: 0;
    position: relative;
    z-index: 10;
  }

  .workspace-sidebar-container.open {
    width: 240px;
    min-width: 240px;
    border-right: 1px solid rgba(255, 255, 255, 0.08);
  }

  /* Drawer */
  .drawer {
    width: 240px;
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .drawer-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .title {
    color: var(--hud-text-white, #ffffff);
    font-size: 0.9rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .actions {
    display: flex;
    gap: 8px;
  }

  .action-btn {
    background: transparent;
    border: none;
    color: var(--hud-text-gray, #cccccc);
    cursor: pointer;
    padding: 3px 6px;
    border-radius: 6px;
    transition: background-color 0.2s, color 0.2s;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }

  .action-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: var(--hud-text-white, #ffffff);
  }

  .action-btn .icon {
    font-size: 16px;
    line-height: 1;
  }

  .action-btn .btn-label {
    font-size: 8px;
    font-weight: 700;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    margin-top: 2px;
    line-height: 1;
  }

  .drawer-content {
    flex-grow: 1;
    overflow: hidden; /* FileTree handles its own scroll */
  }
</style>
