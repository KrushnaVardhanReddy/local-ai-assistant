<script lang="ts">
  import type { IDEStatus } from './types';

  let {
    status = { left: {}, right: {} }
  } = $props<{
    status?: IDEStatus;
  }>();

  const stealthMode = import.meta.env.VITE_STEALTH_MODE === 'true';
</script>

<div class="status-bar">
  <div class="left-section">
    {#if status.left.workspaceName}
      <div class="status-item branch" title="Active Branch / Workspace">
        <span>{status.left.workspaceName}</span>
      </div>
    {/if}
    {#if status.left.sttStatus}
      <div class="status-item stt" title="Speech-to-Text Status">
        <span>{status.left.sttStatus}</span>
      </div>
    {/if}
    {#if status.left.docStats}
      <div class="status-item stats" title="Document Stats">
        <span>{status.left.docStats}</span>
      </div>
    {/if}
  </div>

  <div class="right-section">
    {#if status.right.scrollerStatus}
      <div class="status-item scroller" title="Scroller Status">
        <span>{status.right.scrollerStatus}</span>
      </div>
    {/if}
    {#if status.right.ragStatus}
      <div class="status-item rag" title="RAG Index Status">
        <span>{status.right.ragStatus}</span>
      </div>
    {/if}
    {#if stealthMode && status.right.stealthStatus}
      <div class="status-item stealth" title="Window Stealth Indicator">
        <span>{status.right.stealthStatus}</span>
      </div>
    {/if}
  </div>
</div>

<style>
  .status-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    height: 22px;
    background: rgba(0, 122, 204, 0.7); /* VS Code blue with transparency */
    color: #ffffff;
    font-size: 0.75rem;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
    padding: 0 8px;
    z-index: 100;
    backdrop-filter: blur(5px);
    -webkit-backdrop-filter: blur(5px);
  }

  .left-section,
  .right-section {
    display: flex;
    align-items: center;
    gap: 12px;
    height: 100%;
  }

  .status-item {
    display: flex;
    align-items: center;
    height: 100%;
    padding: 0 4px;
    cursor: default;
    opacity: 0.9;
    transition: background-color 0.1s, opacity 0.1s;
  }

  .status-item:hover {
    background-color: rgba(255, 255, 255, 0.1);
    opacity: 1;
  }
</style>
