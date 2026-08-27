<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { connect, disconnect, wsState } from "$lib/ws.svelte";

  onMount(() => {
    connect();
    return () => disconnect();
  });
</script>

<h1 style="color:white;font-family:Inter,sans-serif;padding:2rem;">
  Local AI Assistant
</h1>

<div style="color:white;font-family:Inter,sans-serif;padding:2rem;">
  <p>Connected: {wsState.isConnected}</p>
  <p>Thinking: {wsState.isThinking}</p>
  <p>Error: {wsState.error}</p>
  <p>Transcript: {wsState.transcript}</p>
  <p>Response: {wsState.response}</p>
</div>