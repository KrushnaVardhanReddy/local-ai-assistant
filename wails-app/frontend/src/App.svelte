<script lang="ts">

  import { onMount, onDestroy } from "svelte";
  import { connect, disconnect, wsState } from "$lib/ws.svelte";
  import PresenterHUD from "./products/presenter/PresenterHUD.svelte";
  import InterviewHUD from "./products/interview/InterviewHUD.svelte";
  import { restoreSession, authState, initLicenseCheck, initAuthEventListeners } from "$lib/auth.svelte";
  import { uiState } from "$lib/stores/uiState.svelte.ts";
  import AuthModal from "$lib/components/AuthModal.svelte";

  import { WindowSetSize, WindowCenter } from "../wailsjs/runtime/runtime";



  let showAuthModal = $state(false);
  const isGated = $derived(
    authState.authMode !== "local" && (
      authState.productMode === "interview"
        ? (authState.licenseStatus !== "active" &&
           authState.licenseStatus !== "dev_allowed" &&
           authState.licenseStatus !== "demo")
        : (!authState.user || authState.paddleStatus !== "active")
    )
  );

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

    initAuthEventListeners();
    if (authState.productMode === "interview") {
      await restoreSession(); // Need session to check demo status
      await initLicenseCheck();
    } else if (authState.authMode === "saas") {
      await restoreSession();
    }

    setInterval(() => {
      if (authState.licenseStatus === "demo" && authState.demoExpiresAt) {
        if (new Date(authState.demoExpiresAt).getTime() <= Date.now()) {
          authState.licenseStatus = "expired";
        }
      }
    }, 60000);

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

{#if isGated}
  <div class="fixed inset-0 z-[9999] flex flex-col items-center justify-center bg-black/80 backdrop-blur-[20px]">
    <h1 class="text-3xl text-white font-semibold mb-6">
      {#if authState.licenseStatus === 'expired'}
        Demo Expired
      {:else}
        Unlock BarnOwl AI
      {/if}
    </h1>
    <button class="bg-blue-600 hover:bg-blue-700 text-white font-bold py-3 px-8 rounded-lg shadow-lg transition-all" onclick={() => showAuthModal = true}>
      Unlock
    </button>
  </div>
{/if}

<AuthModal bind:isOpen={showAuthModal} onClose={() => showAuthModal = false} />

<div class="app-shell" class:pointer-events-none={authState.productMode !== "presenter" && authState.productMode !== "interview"} class:pointer-events-auto={authState.productMode === "presenter" || authState.productMode === "interview"}>
  {#if authState.productMode === "presenter"}
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
