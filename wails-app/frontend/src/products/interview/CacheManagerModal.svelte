<script lang="ts">
  let { isOpen, onClose, onCacheCleared } = $props<{
    isOpen: boolean;
    onClose: () => void;
    onCacheCleared?: () => void;
  }>();

  let items = $state<any[]>([]);
  let selectedIds = $state<Set<string>>(new Set());
  let isLoading = $state(false);

  async function fetchItems() {
    isLoading = true;
    try {
      if ((window as any).go?.main?.App?.GetCacheItems) {
        const fetched = await (window as any).go.main.App.GetCacheItems();
        items = fetched || [];
      }
    } catch (e) {
      console.error("Failed to fetch cache items:", e);
    } finally {
      isLoading = false;
    }
  }

  $effect(() => {
    if (isOpen) {
      fetchItems();
      selectedIds.clear();
    }
  });

  function toggleSelection(id: string) {
    if (selectedIds.has(id)) {
      selectedIds.delete(id);
    } else {
      selectedIds.add(id);
    }
    selectedIds = new Set(selectedIds);
  }

  async function handleDeleteSelected() {
    if (selectedIds.size === 0) return;
    try {
      if ((window as any).go?.main?.App?.DeleteCacheItems) {
        await (window as any).go.main.App.DeleteCacheItems(Array.from(selectedIds));
        await fetchItems();
        selectedIds.clear();
      }
    } catch (e) {
      console.error("Failed to delete cache items:", e);
    }
  }

  async function handleClearAll() {
    if (!confirm("Are you sure you want to clear the entire cache?")) return;
    try {
      if ((window as any).go?.main?.App?.ClearCache) {
        await (window as any).go.main.App.ClearCache();
        await fetchItems();
        if (onCacheCleared) onCacheCleared();
      }
    } catch (e) {
      console.error("Failed to clear cache:", e);
    }
  }
</script>

{#if isOpen}
  <div class="modal-overlay" onclick={onClose}>
    <div class="modal-content" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h2>Cache Manager</h2>
        <button class="close-btn" onclick={onClose}>
          <span class="material-symbols-outlined">close</span>
        </button>
      </div>

      <div class="modal-actions">
        <button class="btn btn-danger" onclick={handleDeleteSelected} disabled={selectedIds.size === 0}>
          Delete Selected ({selectedIds.size})
        </button>
        <button class="btn btn-danger" onclick={handleClearAll} disabled={items.length === 0}>
          Clear All
        </button>
      </div>

      <div class="modal-body">
        {#if isLoading}
          <p>Loading cache items...</p>
        {:else if items.length === 0}
          <p class="empty-state">No items in cache.</p>
        {:else}
          <table class="cache-table">
            <thead>
              <tr>
                <th><input type="checkbox" onchange={(e) => {
                  if (e.currentTarget.checked) {
                    selectedIds = new Set(items.map(i => i.id || i.ID));
                  } else {
                    selectedIds.clear();
                  }
                  selectedIds = new Set(selectedIds);
                }} checked={selectedIds.size === items.length && items.length > 0} /></th>
                <th>Question</th>
                <th>Answer</th>
              </tr>
            </thead>
            <tbody>
              {#each items as item}
                {@const id = item.id || item.ID}
                <tr class:selected={selectedIds.has(id)} onclick={() => toggleSelection(id)}>
                  <td><input type="checkbox" checked={selectedIds.has(id)} onclick={(e) => e.stopPropagation()} onchange={() => toggleSelection(id)} /></td>
                  <td class="text-cell">{item.question || item.Question}</td>
                  <td class="text-cell">{item.answer || item.Answer}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal-content {
    background: #1e1e1e;
    border: 1px solid rgba(74, 222, 128, 0.25);
    border-radius: 8px;
    width: 80%;
    max-width: 800px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    color: #fff;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .modal-header h2 {
    margin: 0;
    font-size: 1.25rem;
  }

  .close-btn {
    background: none;
    border: none;
    color: #fff;
    cursor: pointer;
    font-size: 1.5rem;
  }

  .modal-actions {
    display: flex;
    gap: 12px;
    padding: 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .btn {
    padding: 8px 16px;
    border-radius: 4px;
    border: none;
    cursor: pointer;
    font-weight: 500;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-danger {
    background: rgba(220, 38, 38, 0.8);
    color: white;
  }

  .btn-danger:hover:not(:disabled) {
    background: rgba(220, 38, 38, 1);
  }

  .modal-body {
    padding: 16px;
    overflow-y: auto;
    flex: 1;
  }

  .cache-table {
    width: 100%;
    border-collapse: collapse;
  }

  .cache-table th, .cache-table td {
    border: 1px solid rgba(255, 255, 255, 0.1);
    padding: 8px;
    text-align: left;
  }

  .cache-table th {
    background: rgba(255, 255, 255, 0.05);
  }

  .cache-table tr:hover {
    background: rgba(255, 255, 255, 0.02);
  }

  .cache-table tr.selected {
    background: rgba(74, 222, 128, 0.1);
  }

  .text-cell {
    max-width: 300px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .empty-state {
    text-align: center;
    color: rgba(255, 255, 255, 0.5);
    padding: 24px;
  }
</style>
