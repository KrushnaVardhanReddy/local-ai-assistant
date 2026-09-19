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
    // Check for Ctrl+M or Cmd+M to toggle mic
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'm') {
      e.preventDefault();
      if ((window as any).go?.main?.App?.ToggleMic) {
        (window as any).go.main.App.ToggleMic().then((newState: boolean) => {
          wsState.isListening = newState;
        });
      }
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
</style>
