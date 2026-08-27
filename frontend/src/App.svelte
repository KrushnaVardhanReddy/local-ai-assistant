<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { connect, disconnect, wsState } from "$lib/ws.svelte";
  import Assistant from "$lib/Assistant.svelte";
  import KnowledgeBase from "$lib/KnowledgeBase.svelte";
  import Settings from "$lib/Settings.svelte";
  import { restoreSession, authState } from "$lib/auth.svelte";

  let showKnowledgeBase = $state(false);

  onMount(async () => {
    if (authState.authMode === "saas") {
      await restoreSession();
    }
    connect();
  });

  onDestroy(() => {
    disconnect();
  });
</script>

<div class="app-shell">
  {#if showKnowledgeBase}
    <div class="modal-overlay" onclick={() => showKnowledgeBase = false}>
      <div class="modal-content" onclick={(e) => e.stopPropagation()}>
        <button class="close-btn" onclick={() => showKnowledgeBase = false}>✖</button>
        <KnowledgeBase />
      </div>
    </div>
  {/if}

  <div class="main-content">
    <div class="assistant-wrapper">
      <button class="kb-btn" onclick={() => showKnowledgeBase = !showKnowledgeBase} title="Knowledge Base">
        📚
      </button>
      <Assistant />
    </div>
  </div>
  <div class="settings-content">
    <Settings />
  </div>
</div>

<style>
  .app-shell {
    width: 100vw;
    height: 100vh;
    background: transparent;
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    justify-content: flex-start;
    padding: 1rem;
    position: relative;
    gap: 1rem;
  }

  .assistant-wrapper {
    position: relative;
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .kb-btn {
    position: absolute;
    top: 1rem;
    right: 1.5rem;
    background: none;
    border: none;
    font-size: 1.25rem;
    cursor: pointer;
    z-index: 10;
    transition: transform 0.2s;
  }

  .kb-btn:hover {
    transform: scale(1.1);
  }

  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .modal-content {
    position: relative;
    width: 90%;
    max-width: 600px;
    max-height: 90vh;
    overflow-y: auto;
    border-radius: 12px;
  }

  .close-btn {
    position: absolute;
    top: 1rem;
    right: 1rem;
    background: none;
    border: none;
    color: #94a3b8;
    font-size: 1.2rem;
    cursor: pointer;
    z-index: 10;
    transition: color 0.2s;
  }

  .close-btn:hover {
    color: #e2e8f0;
  }

  .main-content {
    display: flex;
    justify-content: flex-end;
  }
</style>
