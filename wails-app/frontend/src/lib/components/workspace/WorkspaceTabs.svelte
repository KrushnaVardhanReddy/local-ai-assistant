<script lang="ts">
  import type { WorkspaceTab } from './types';

  let {
    tabs,
    activePath = '',
    onTabSelect,
    onTabClose
  } = $props<{
    tabs: WorkspaceTab[];
    activePath?: string;
    onTabSelect: (path: string) => void;
    onTabClose: (path: string) => void;
  }>();

  function getIcon(name: string): string {
    const ext = name.split('.').pop()?.toLowerCase();
    switch (ext) {
      case 'md':
      case 'txt': return '📝';
      case 'pdf': return '📄';
      case 'pptx': return '📊';
      default: return '📄';
    }
  }

  function handleSelect(path: string, e: MouseEvent | KeyboardEvent) {
    if (e.type === 'keydown') {
      const kbEvent = e as KeyboardEvent;
      if (kbEvent.key !== 'Enter' && kbEvent.key !== ' ') return;
      kbEvent.preventDefault();
    }
    onTabSelect(path);
  }

  function handleClose(path: string, e: MouseEvent | KeyboardEvent) {
    e.stopPropagation();
    if (e.type === 'keydown') {
      const kbEvent = e as KeyboardEvent;
      if (kbEvent.key !== 'Enter' && kbEvent.key !== ' ') return;
      kbEvent.preventDefault();
    }
    onTabClose(path);
  }
</script>

<div class="workspace-tabs-container">
  {#if tabs.length === 0}
    <div class="empty-tabs">No open documents</div>
  {:else}
    <div class="tabs-scroll-area">
      {#each tabs as tab (tab.path)}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <div
          class="tab"
          class:active={activePath === tab.path}
          role="tab"
          aria-selected={activePath === tab.path}
          tabindex="0"
          onclick={(e) => handleSelect(tab.path, e)}
          onkeydown={(e) => handleSelect(tab.path, e)}
          title={tab.path}
        >
          <span class="icon">{getIcon(tab.name)}</span>
          <span class="name">{tab.name}</span>

          <div
            class="close-btn"
            role="button"
            aria-label="Close tab"
            tabindex="0"
            onclick={(e) => handleClose(tab.path, e)}
            onkeydown={(e) => handleClose(tab.path, e)}
          >
            ×
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .workspace-tabs-container {
    display: flex;
    width: 100%;
    height: 36px;
    background: rgba(0, 0, 0, 0.4);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    overflow: hidden;
  }

  .empty-tabs {
    display: flex;
    align-items: center;
    padding: 0 16px;
    font-size: 0.8rem;
    color: rgba(255, 255, 255, 0.3);
    height: 100%;
    font-style: italic;
  }

  .tabs-scroll-area {
    display: flex;
    flex-wrap: nowrap;
    overflow-x: auto;
    overflow-y: hidden;
    height: 100%;
    width: 100%;
    scrollbar-width: none; /* Firefox */
  }

  .tabs-scroll-area::-webkit-scrollbar {
    display: none; /* Chrome, Safari, Edge */
  }

  .tab {
    display: flex;
    align-items: center;
    height: 100%;
    padding: 0 12px;
    min-width: 100px;
    max-width: 200px;
    background: transparent;
    color: rgba(255, 255, 255, 0.5);
    border-right: 1px solid rgba(255, 255, 255, 0.05);
    cursor: pointer;
    user-select: none;
    transition: background-color 0.2s, color 0.2s;
    position: relative;
  }

  .tab:hover {
    background: rgba(255, 255, 255, 0.05);
    color: rgba(255, 255, 255, 0.8);
  }

  .tab.active {
    background: rgba(255, 255, 255, 0.1);
    color: #ffffff;
  }

  .tab.active::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 2px;
    background: #007bff; /* Accent color */
    border-top-left-radius: 2px;
    border-top-right-radius: 2px;
  }

  .tab:focus-visible {
    outline: 2px solid rgba(255, 255, 255, 0.3);
    outline-offset: -2px;
  }

  .icon {
    font-size: 0.9rem;
    margin-right: 6px;
    flex-shrink: 0;
  }

  .name {
    font-size: 0.85rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    flex-grow: 1;
    margin-right: 8px;
  }

  .close-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    border-radius: 4px;
    font-size: 1.1rem;
    line-height: 1;
    opacity: 0;
    transition: opacity 0.2s, background-color 0.2s;
    flex-shrink: 0;
  }

  .tab:hover .close-btn,
  .tab.active .close-btn,
  .close-btn:focus-visible {
    opacity: 1;
  }

  .close-btn:hover,
  .close-btn:focus-visible {
    background-color: rgba(255, 255, 255, 0.2);
    color: #ff4444;
  }
</style>
