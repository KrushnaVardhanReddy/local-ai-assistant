<script lang="ts">
  type ActionDef = { id: string; icon: string; label: string; isActive?: boolean };

  let {
    activeAction = '',
    onAction,
    topActions = [
      { id: 'explorer', icon: 'folder', label: 'Folder' },
      { id: 'mock', icon: 'record_voice_over', label: 'Mock' }
    ],
    bottomActions = [
      { id: 'cache', icon: 'mop', label: 'Cache' },
      { id: 'settings', icon: 'settings', label: 'Settings' }
    ]
  } = $props<{
    activeAction?: string;
    onAction: (action: string) => void;
    topActions?: ActionDef[];
    bottomActions?: ActionDef[];
  }>();

  function handleAction(id: string) {
    onAction(id);
  }
</script>

<div class="activity-bar">
  <div class="actions top">
    {#each topActions as action}
      <button
        class="action-btn"
        class:active={action.isActive !== undefined ? action.isActive : activeAction === action.id}
        onclick={() => handleAction(action.id)}
        aria-label={action.label}
        data-testid="activity-bar-{action.id}"
      >
        <span class="material-symbols-outlined icon">{action.icon}</span>
        <span class="btn-label">{action.label}</span>
      </button>
    {/each}
  </div>

  <div class="actions bottom">
    {#each bottomActions as action}
      <button
        class="action-btn"
        class:active={action.isActive !== undefined ? action.isActive : activeAction === action.id}
        onclick={() => handleAction(action.id)}
        aria-label={action.label}
        data-testid="activity-bar-{action.id}"
      >
        <span class="material-symbols-outlined icon">{action.icon}</span>
        <span class="btn-label">{action.label}</span>
      </button>
    {/each}
  </div>
</div>

<style>
  .activity-bar {
    width: 54px;
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    background: rgba(15, 15, 15, 0.95);
    border-right: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    z-index: 100;
  }

  .actions {
    display: flex;
    flex-direction: column;
    align-items: center;
    width: 100%;
    padding: 8px 0;
    gap: 4px;
  }

  .action-btn {
    position: relative;
    width: 46px;
    height: 46px;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    border-radius: 8px;
    color: rgba(255, 255, 255, 0.5);
    cursor: pointer;
    transition: color 0.2s, background-color 0.2s;
    padding: 2px 0;
  }

  .action-btn:hover {
    color: rgba(255, 255, 255, 0.9);
    background-color: rgba(255, 255, 255, 0.08);
  }

  .action-btn.active {
    color: #4ade80;
    background-color: rgba(74, 222, 128, 0.1);
  }

  .action-btn.active::before {
    content: '';
    position: absolute;
    left: -4px;
    top: 6px;
    bottom: 6px;
    width: 3px;
    border-radius: 2px;
    background: #4ade80;
  }

  .icon {
    font-size: 19px;
    line-height: 1;
  }

  .btn-label {
    font-size: 8px;
    font-weight: 700;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    margin-top: 3px;
    line-height: 1;
  }
</style>
