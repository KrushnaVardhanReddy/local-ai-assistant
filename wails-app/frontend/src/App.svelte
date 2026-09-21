<script lang="ts">

  import { onMount, onDestroy } from "svelte";
  import { connect, disconnect, wsState } from "$lib/ws.svelte";
  import PresenterHUD from "./products/presenter/PresenterHUD.svelte";
  import InterviewHUD from "./products/interview/InterviewHUD.svelte";
  import { restoreSession, authState, initLicenseCheck, initAuthEventListeners } from "$lib/auth.svelte";
  import { uiState } from "$lib/stores/uiState.svelte.ts";
  import AuthModal from "$lib/components/AuthModal.svelte";
  import Titlebar from "$lib/components/Titlebar.svelte";

  import { WindowSetSize, WindowCenter, WindowSetAlwaysOnTop, WindowShow } from "../wailsjs/runtime/runtime";



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

  $effect(() => {
    if (!isGated && typeof window !== 'undefined') {
      try {
        WindowSetAlwaysOnTop(true);
        if ((window as any).go?.main?.App?.HideFromTaskbar) {
          (window as any).go.main.App.HideFromTaskbar();
        }
      } catch (err) {
        console.error("Failed to set window always on top", err);
      }
      // Auto-close the auth modal when the user is authenticated
      showAuthModal = false;
    }
  });

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

    // On Linux, the frameless+translucent Wails window gets hidden by the compositor
    // when it loses focus (Alt+Tab). Re-show it whenever it regains focus.
    // This is harmless on other platforms.
    const handleFocus = () => {
      try {
        WindowShow();
        // While gated (auth screen), keep the window in normal z-order (not always-on-top)
        // so the user can switch to their browser and log in. But we still need to
        // make it visible after focus returns.
        if (!isGated) {
          WindowSetAlwaysOnTop(true);
        }
      } catch (_) {}
    };
    window.addEventListener('focus', handleFocus);

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

    return () => {
      window.removeEventListener('focus', handleFocus);
    };
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

<svelte:head>
  {#if isGated}
    <!-- Force solid background during auth — the window has a transparent background colour
         for stealth mode, which causes the window to appear invisible on Alt+Tab on Linux 
         unless we explicitly paint a solid background here. -->
    <style>body { background: #1e1e1e !important; }</style>
  {:else}
    <style>body { background: transparent !important; }</style>
  {/if}
</svelte:head>

{#if isGated}
  <div class="fixed inset-0 z-[9999] flex flex-col bg-[#1e1e1e]">
    <Titlebar />
    <div class="flex-1 flex flex-col items-center justify-center gap-4">
      {#if authState.licenseStatus === 'expired'}
        <h1 class="text-3xl text-white font-semibold">🕐 Demo Expired</h1>
        <p class="text-gray-400 text-center max-w-sm">Your 15-minute free trial has ended. Enter a lifetime license to keep using BarnOwl AI.</p>
        <button class="bg-blue-600 hover:bg-blue-700 text-white font-bold py-3 px-8 rounded-lg shadow-lg transition-all" onclick={() => showAuthModal = true}>
          Enter License Key
        </button>
      {:else}
        <h1 class="text-3xl text-white font-semibold">Unlock BarnOwl AI</h1>
        <p class="text-gray-400 text-center max-w-sm">Start a free 15-minute demo or enter your lifetime license key.</p>
        <button class="bg-blue-600 hover:bg-blue-700 text-white font-bold py-3 px-8 rounded-lg shadow-lg transition-all" onclick={() => showAuthModal = true}>
          Unlock
        </button>
      {/if}
    </div>
  </div>
{/if}

<AuthModal bind:isOpen={showAuthModal} onClose={() => showAuthModal = false} />

{#if !isGated}
  <div class="app-shell" class:pointer-events-none={authState.productMode !== "presenter" && authState.productMode !== "interview"} class:pointer-events-auto={authState.productMode === "presenter" || authState.productMode === "interview"}>
    {#if authState.productMode === "presenter"}
      <PresenterHUD />
    {:else}
      <InterviewHUD />
    {/if}
  </div>
{/if}

<style>
  .app-shell {
    width: 100vw;
    height: 100vh;
    background: transparent;
  }
</style>
