<script lang="ts">
  import type { FileNode } from './types';
  import FileTreeNode from './FileTreeNode.svelte';

  let {
    nodes,
    activePath = '',
    onSelect
  } = $props<{
    nodes: FileNode[];
    activePath?: string;
    onSelect: (node: FileNode) => void;
  }>();

  let searchQuery = $state('');

  function filterNodes(nodesToFilter: FileNode[], query: string): FileNode[] {
    if (!query) return nodesToFilter;
    const lowerQuery = query.toLowerCase();

    return nodesToFilter.reduce<FileNode[]>((acc, node) => {
      const isDir = Boolean(node.isDirectory ?? (node as any).isDir);
      if (isDir && node.children) {
        const filteredChildren = filterNodes(node.children, query);
        if (filteredChildren.length > 0) {
          acc.push({ ...node, children: filteredChildren, isExpanded: true });
        } else if (node.name.toLowerCase().includes(lowerQuery)) {
          acc.push(node);
        }
      } else if (node.name.toLowerCase().includes(lowerQuery)) {
        acc.push(node);
      }
      return acc;
    }, []);
  }

  let filteredNodes = $derived(filterNodes(nodes, searchQuery));
</script>

<div class="file-tree-container">
  <div class="search-bar">
    <input
      type="text"
      placeholder="Filter documents..."
      bind:value={searchQuery}
      aria-label="Filter documents"
    />
  </div>

  <div class="tree-list" role="tree">
    {#if filteredNodes.length === 0}
      <div class="no-results">No documents found.</div>
    {:else}
      {#each filteredNodes as node (node.path)}
        <FileTreeNode {node} {activePath} {onSelect} />
      {/each}
    {/if}
  </div>
</div>

<style>
  .file-tree-container {
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
    background: transparent;
    color: var(--hud-text-gray, #cccccc);
  }

  .search-bar {
    padding: 8px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .search-bar input {
    width: 100%;
    padding: 6px 8px;
    border-radius: 4px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    background: rgba(0, 0, 0, 0.2);
    color: var(--hud-text-white, #ffffff);
    font-size: 0.85rem;
    transition: border-color 0.2s, background-color 0.2s;
  }

  .search-bar input:focus {
    outline: none;
    border-color: rgba(255, 255, 255, 0.3);
    background: rgba(0, 0, 0, 0.4);
  }

  .search-bar input::placeholder {
    color: rgba(255, 255, 255, 0.3);
  }

  .tree-list {
    flex-grow: 1;
    overflow-y: auto;
    padding: 8px 4px;
  }

  .tree-list::-webkit-scrollbar {
    width: 8px;
  }

  .tree-list::-webkit-scrollbar-track {
    background: transparent;
  }

  .tree-list::-webkit-scrollbar-thumb {
    background-color: rgba(255, 255, 255, 0.1);
    border-radius: 4px;
  }

  .tree-list::-webkit-scrollbar-thumb:hover {
    background-color: rgba(255, 255, 255, 0.2);
  }

  .no-results {
    padding: 8px 16px;
    font-size: 0.85rem;
    color: rgba(255, 255, 255, 0.4);
    text-align: center;
  }
</style>
