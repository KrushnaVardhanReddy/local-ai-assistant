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

  function toggleSidebar() {
    onToggle(!isOpen);
  }
</script>

<div class="workspace-sidebar-container" class:open={isOpen}>
  <!-- Toggle Button -->
  <button
    class="toggle-btn"
    onclick={toggleSidebar}
    aria-label="Toggle Sidebar"
    title="Toggle Sidebar"
  >
    <span class="folder-icon">📁</span>
    {#if totalDocs > 0}
      <span class="badge">{totalDocs}</span>
    {/if}
  </button>

  <!-- Drawer Content -->
  <div class="drawer">
    <div class="drawer-header">
      <span class="title">Workspace</span>
      <div class="actions">
        <button class="action-btn" onclick={onOpenFile} title="Open File">📄+</button>
        <button class="action-btn" onclick={onOpenFolder} title="Open Folder">📁+</button>
      </div>
    </div>
    <div class="drawer-content">
      <FileTree {nodes} {activePath} {onSelect} />
    </div>
  </div>
</div>

<style>
  .workspace-sidebar-container {
    position: fixed;
    top: 60px; /* Adjust based on title bar / layout */
    left: 0;
    bottom: 0;
    display: flex;
    z-index: 1000;
    pointer-events: none; /* Let clicks pass through container */
  }

  /* Toggle Button */
  .toggle-btn {
    pointer-events: auto;
    position: absolute;
    left: 0;
    top: 20px;
    background: rgba(0, 0, 0, 0.6);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-left: none;
    border-top-right-radius: 8px;
    border-bottom-right-radius: 8px;
    padding: 8px 12px;
    cursor: pointer;
    display: flex;
    align-items: center;
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1), background-color 0.2s;
    box-shadow: 2px 0 8px rgba(0, 0, 0, 0.3);
  }

  .toggle-btn:hover {
    background: rgba(0, 0, 0, 0.8);
  }

  .workspace-sidebar-container.open .toggle-btn {
    transform: translateX(280px); /* Move button out when drawer is open */
  }

  .folder-icon {
    font-size: 1.2rem;
  }

  .badge {
    position: absolute;
    top: -6px;
    right: -6px;
    background: #007bff;
    color: white;
    font-size: 0.7rem;
    font-weight: bold;
    padding: 2px 6px;
    border-radius: 10px;
    min-width: 16px;
    text-align: center;
  }

  /* Drawer */
  .drawer {
    pointer-events: auto;
    width: 280px;
    height: 100%;
    background: rgba(15, 15, 15, 0.85);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-right: 1px solid rgba(255, 255, 255, 0.1);
    display: flex;
    flex-direction: column;
    transform: translateX(-100%);
    transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    box-shadow: 4px 0 16px rgba(0, 0, 0, 0.4);
  }

  .workspace-sidebar-container.open .drawer {
    transform: translateX(0);
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
    font-size: 1.1rem;
    padding: 4px;
    border-radius: 4px;
    transition: background-color 0.2s, color 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .action-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: var(--hud-text-white, #ffffff);
  }

  .drawer-content {
    flex-grow: 1;
    overflow: hidden; /* FileTree handles its own scroll */
  }
</style>
