<script lang="ts">
  import type { FileNode } from './types';
  import FileTreeNode from './FileTreeNode.svelte';

  let {
    node,
    activePath = '',
    onSelect,
    level = 0
  } = $props<{
    node: FileNode;
    activePath?: string;
    onSelect: (node: FileNode) => void;
    level?: number;
  }>();

  let isExpanded = $state(false);
  let isIndexed = $state(false);
  let isIndexing = $state(false);

  const isDir = $derived(Boolean(node.isDirectory ?? (node as any).isDir));

  $effect(() => {
    isExpanded = node.isExpanded ?? false;
    // Check if indexed via window.mockState or global paths array
    if ((window as any)._indexedPaths) {
      isIndexed = (window as any)._indexedPaths.includes(node.path);
    }
  });

  function getIcon(node: FileNode): string {
    const isDirectory = Boolean(node.isDirectory ?? (node as any).isDir);
    if (isDirectory) {
      return isExpanded ? '📂' : '📁';
    }
    const ext = node.name.split('.').pop()?.toLowerCase();
    switch (ext) {
      case 'md':
      case 'txt': return '📝';
      case 'go':
      case 'py':
      case 'ts':
      case 'js':
      case 'rs':
      case 'cpp': return '💻';
      case 'json':
      case 'yaml':
      case 'yml': return '⚙️';
      case 'pdf': return '📄';
      case 'pptx': return '📊';
      default: return '📄';
    }
  }

  function toggleExpand(e: MouseEvent | KeyboardEvent) {
    e.stopPropagation();
    isExpanded = !isExpanded;
    node.isExpanded = isExpanded;
  }

  async function handleCheckboxChange(e: Event) {
    e.stopPropagation();
    const checked = (e.target as HTMLInputElement).checked;
    isIndexing = true;
    try {
      if (checked) {
        if (isDir) {
          await (window as any).go?.main?.App?.IndexFolder(node.path);
        } else {
          await (window as any).go?.main?.App?.IndexFile(node.path);
        }
      } else {
        await (window as any).go?.main?.App?.RemoveIndexedPath(node.path);
      }
      
      // Refresh global paths
      if ((window as any).go?.main?.App?.GetIndexedPaths) {
        const paths = await (window as any).go.main.App.GetIndexedPaths();
        (window as any)._indexedPaths = paths || [];
        // Trigger a custom event or reactive update for all nodes
        window.dispatchEvent(new Event('indexed-paths-updated'));
      }
    } catch (err) {
      console.error("Failed to index path:", err);
      // Revert state on error
      isIndexed = !checked;
    } finally {
      isIndexing = false;
      isIndexed = checked;
    }
  }

  $effect(() => {
    const handleUpdate = () => {
      if ((window as any)._indexedPaths) {
        isIndexed = (window as any)._indexedPaths.includes(node.path);
      }
    };
    window.addEventListener('indexed-paths-updated', handleUpdate);
    return () => window.removeEventListener('indexed-paths-updated', handleUpdate);
  });

  function handleSelect(e: MouseEvent | KeyboardEvent) {
    if (isDir) {
      toggleExpand(e);
    } else {
      onSelect(node);
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleSelect(e);
    } else if (e.key === 'ArrowRight' && isDir && !isExpanded) {
      e.preventDefault();
      isExpanded = true;
      node.isExpanded = isExpanded;
    } else if (e.key === 'ArrowLeft' && isDir && isExpanded) {
      e.preventDefault();
      isExpanded = false;
      node.isExpanded = isExpanded;
    }
  }
</script>

<div class="tree-node" style="--indent: {level * 16}px">
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div
    class="node-content"
    class:active={activePath === node.path}
    role="treeitem"
    aria-expanded={isDir ? isExpanded : undefined}
    aria-selected={activePath === node.path}
    tabindex="0"
    onclick={handleSelect}
    onkeydown={handleKeyDown}
  >
    <div class="indent-spacer"></div>
    {#if isDir}
      <button
        type="button"
        class="expander"
        onclick={toggleExpand}
        tabindex="-1"
        aria-label={isExpanded ? 'Collapse folder' : 'Expand folder'}
      >
        <span class="chevron" class:expanded={isExpanded}>▶</span>
      </button>
    {:else}
      <div class="expander invisible"></div>
    {/if}
    <!-- Indexing checkbox -->
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="checkbox-wrapper" onclick={(e) => e.stopPropagation()}>
      <input type="checkbox" checked={isIndexed} onchange={handleCheckboxChange} disabled={isIndexing} />
      {#if isIndexing}
        <span class="spinner" aria-label="Indexing"></span>
      {/if}
    </div>
    <span class="icon relative">
      {getIcon(node)}
      {#if isIndexed && !isDir}
        <span class="indexed-badge" aria-label="Indexed">●</span>
      {/if}
    </span>
    <span class="name">{node.name}</span>
  </div>

  {#if isDir && isExpanded && node.children}
    <div class="children" role="group">
      {#each node.children as child (child.path)}
        <FileTreeNode
          node={child}
          {activePath}
          {onSelect}
          level={level + 1}
        />
      {/each}
    </div>
  {/if}
</div>

<style>
  .tree-node {
    display: flex;
    flex-direction: column;
    user-select: none;
    width: 100%;
  }

  .node-content {
    display: flex;
    align-items: center;
    padding: 4px 8px;
    cursor: pointer;
    border-radius: 4px;
    color: var(--hud-text-gray, #cccccc);
    transition: background-color 0.1s ease, color 0.1s ease;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .indent-spacer {
    width: var(--indent);
    flex-shrink: 0;
  }

  .node-content:hover {
    background-color: rgba(255, 255, 255, 0.05);
  }

  .node-content.active {
    background-color: rgba(255, 255, 255, 0.1);
    color: var(--hud-text-white, #ffffff);
    font-weight: 500;
  }

  .node-content:focus-visible {
    outline: 1px solid rgba(255, 255, 255, 0.3);
    outline-offset: -1px;
  }

  .expander {
    width: 16px;
    height: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    margin-right: 4px;
    cursor: pointer;
    background: transparent;
    border: none;
    padding: 0;
    color: inherit;
  }

  .expander.invisible {
    visibility: hidden;
  }

  .chevron {
    font-size: 0.6rem;
    transition: transform 0.2s ease;
    display: inline-block;
  }

  .chevron.expanded {
    transform: rotate(90deg);
  }

  .icon {
    margin-right: 6px;
    font-size: 0.9rem;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .name {
    font-size: 0.85rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .children {
    display: flex;
    flex-direction: column;
  }

  .checkbox-wrapper {
    display: flex;
    align-items: center;
    margin-right: 6px;
    position: relative;
  }

  .checkbox-wrapper input[type="checkbox"] {
    cursor: pointer;
    width: 14px;
    height: 14px;
    accent-color: #4ade80; /* Match typical hud green */
  }

  .spinner {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 12px;
    height: 12px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: #fff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    pointer-events: none;
  }

  @keyframes spin {
    to {
      transform: translate(-50%, -50%) rotate(360deg);
    }
  }
  
  .relative {
    position: relative;
  }
  
  .indexed-badge {
    position: absolute;
    bottom: -2px;
    right: -2px;
    color: #3b82f6; /* Blue dot */
    font-size: 0.5rem;
    line-height: 1;
    text-shadow: 0 0 2px rgba(0,0,0,0.8);
  }
</style>
