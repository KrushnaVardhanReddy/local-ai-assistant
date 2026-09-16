<script lang="ts">
  let {
    activeAction = '',
    onAction
  } = $props<{
    activeAction?: string;
    onAction: (action: string) => void;
  }>();

  const topActions = [
    { id: 'explorer', icon: '📁', label: 'Explorer (Ctrl+B)' },
    { id: 'search', icon: '🔍', label: 'Search' },
    { id: 'copilot', icon: '✨', label: 'Audience Copilot' }
  ];

  const bottomActions = [
    { id: 'settings', icon: '⚙️', label: 'Settings' }
  ];

  function handleAction(id: string) {
    onAction(id);
  }
</script>

<div class="activity-bar">
  <div class="actions top">
    {#each topActions as action}
      <button
        class="action-btn"
        class:active={activeAction === action.id}
        onclick={() => handleAction(action.id)}
        title={action.label}
        aria-label={action.label}
      >
        <span class="icon">{action.icon}</span>
      </button>
    {/each}
  </div>

  <div class="actions bottom">
    {#each bottomActions as action}
      <button
        class="action-btn"
        class:active={activeAction === action.id}
        onclick={() => handleAction(action.id)}
        title={action.label}
        aria-label={action.label}
      >
        <span class="icon">{action.icon}</span>
      </button>
    {/each}
  </div>
</div>

<style>
  .activity-bar {
    width: 48px;
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
  }

  .action-btn {
    position: relative;
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    color: rgba(255, 255, 255, 0.4);
    cursor: pointer;
    transition: color 0.2s;
  }

  .action-btn:hover {
    color: rgba(255, 255, 255, 0.8);
  }

  .action-btn.active {
    color: #ffffff;
  }

  .action-btn.active::before {
    content: '';
    position: absolute;
    left: 0;
    top: 0;
    bottom: 0;
    width: 2px;
    background: #ffffff; /* Accent white */
  }

  .icon {
    font-size: 1.4rem;
  }
</style>
