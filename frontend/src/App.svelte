<script lang="ts">
  import { onMount } from "svelte";
  import { connect, disconnect } from "$lib/ws.svelte";
  import Assistant from "$lib/Assistant.svelte";

  onMount(() => {
    connect();
    return () => disconnect();
  });
</script>

<Assistant />