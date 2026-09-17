<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { connect, disconnect, wsState } from "$lib/ws.svelte";
  import PresenterHUD from "./products/presenter/PresenterHUD.svelte";
  import InterviewHUD from "./products/interview/InterviewHUD.svelte";
  import { restoreSession, authState } from "$lib/auth.svelte";
  import { uiState } from "$lib/stores/uiState.svelte.ts";

  import { WindowSetSize, WindowCenter } from "../wailsjs/runtime/runtime";

  const product = import.meta.env.VITE_PRODUCT || "interview";

  onMount(async () => {
    // Dynamically size window based on screen width, clamped between 1024 and 1440
    if (typeof window !== 'undefined' && window.screen) {
      const screenWidth = window.screen.availWidth;
      let targetWidth = Math.floor(screenWidth * 0.9);
      if (targetWidth < 1024) targetWidth = 1024;
      if (targetWidth > 1440) targetWidth = 1440;
      
      WindowSetSize(targetWidth, 768);
      setTimeout(WindowCenter, 100);
    }

    if (authState.authMode === "saas") {
      await restoreSession();
    }
    connect();
  });

  onDestroy(() => {
    disconnect();
  });

  function handleKeydown(e: KeyboardEvent) {
    // Check for Ctrl+/ or Cmd+/
    if ((e.ctrlKey || e.metaKey) && e.key === '/') {
      e.preventDefault();
      uiState.hotkeysPanelOpen = !uiState.hotkeysPanelOpen;
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="app-shell" class:pointer-events-none={product !== "presenter" && product !== "interview"} class:pointer-events-auto={product === "presenter" || product === "interview"}>
  {#if product === "presenter"}
    <PresenterHUD />
  {:else}
    <InterviewHUD />
  {/if}
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
