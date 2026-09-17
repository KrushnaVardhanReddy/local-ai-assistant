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

  $effect(() => {
    isExpanded = node.isExpanded ?? false;
  });

  function getIcon(node: FileNode): string {
    if (node.isDirectory) {
      return isExpanded ? '📂' : '📁';
    }
    const ext = node.name.split('.').pop()?.toLowerCase();
    switch (ext) {
      case 'md':
      case 'txt': return '📝';
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

  function handleSelect(e: MouseEvent | KeyboardEvent) {
    if (node.isDirectory) {
      toggleExpand(e);
    } else {
      onSelect(node);
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleSelect(e);
    } else if (e.key === 'ArrowRight' && node.isDirectory && !isExpanded) {
      e.preventDefault();
      isExpanded = true;
    } else if (e.key === 'ArrowLeft' && node.isDirectory && isExpanded) {
      e.preventDefault();
      isExpanded = false;
    }
  }
</script>

<div class="tree-node" style="--indent: {level * 16}px">
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div
    class="node-content"
    class:active={activePath === node.path}
    role="treeitem"
    aria-expanded={node.isDirectory ? isExpanded : undefined}
    aria-selected={activePath === node.path}
    tabindex="0"
    onclick={handleSelect}
    onkeydown={handleKeyDown}
  >
    <div class="indent-spacer"></div>
    {#if node.isDirectory}
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
    <span class="icon">{getIcon(node)}</span>
    <span class="name">{node.name}</span>
  </div>

  {#if node.isDirectory && isExpanded && node.children}
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
</style>
