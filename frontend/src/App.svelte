<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { connect, disconnect, wsState } from "$lib/ws.svelte";
  import Assistant from "$lib/Assistant.svelte";

  onMount(() => {
    connect();
  });

  onDestroy(() => {
    disconnect();
  });
</script>

<div class="app-shell">
  <Assistant />
</div>

<style>
  .app-shell {
    width: 100vw;
    height: 100vh;
    background: transparent;
    display: flex;
    align-items: flex-start;
    justify-content: flex-end;
    padding: 1rem;
  }
</style>