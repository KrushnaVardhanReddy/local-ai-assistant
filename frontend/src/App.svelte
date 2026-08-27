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

<div class="app-shell pointer-events-none">
  {#if showKnowledgeBase}
    <div
      class="modal-overlay pointer-events-auto"
      role="button"
      tabindex="0"
      aria-label="Close knowledge base"
      onclick={() => showKnowledgeBase = false}
      onkeydown={(e) => e.key === 'Escape' && (showKnowledgeBase = false)}
    >
      <div class="modal-content" role="dialog" aria-modal="true" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()}>
        <button class="close-btn" onclick={() => showKnowledgeBase = false}>✖</button>
        <KnowledgeBase />
      </div>
    </div>
  {/if}

  <Assistant />
  
  <div class="fixed bottom-6 right-8 z-[100] pointer-events-auto flex gap-4">
    <button class="text-on-surface-variant hover:text-primary transition-colors text-xl" onclick={() => showKnowledgeBase = !showKnowledgeBase} title="Knowledge Base">
      📚
    </button>
    <Settings />
  </div>
</div>

<style>
  .app-shell {
    width: 100vw;
    height: 100vh;
    background: transparent;
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
</style>
