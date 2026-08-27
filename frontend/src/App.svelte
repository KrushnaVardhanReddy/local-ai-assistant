<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { connect, disconnect, wsState } from "$lib/ws.svelte";
  import Assistant from "$lib/Assistant.svelte";
  import Settings from "$lib/Settings.svelte";
  import { restoreSession, authState } from "$lib/auth.svelte";

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
  <div class="main-content">
    <Assistant />
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
    gap: 1rem;
  }

  .main-content {
    display: flex;
    justify-content: flex-end;
  }
  .settings-content {
    position: absolute;
    top: 1rem;
    right: 1rem;
    z-index: 1000;
  }
</style>
