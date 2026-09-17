<script lang="ts">
  import { wsState, sendChat, sendChip, dismissChip, clearAllChips } from "$lib/ws.svelte";
  import { onMount } from "svelte";

  let chatText = $state("");
  let transcriptContainer: HTMLElement;

  function handleChatSubmit() {
    if (chatText.trim()) {
      sendChat(chatText);
      wsState.isThinking = true;
      chatText = "";
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleChatSubmit();
    }
  }

  // Scroll to bottom when transcript history changes
  $effect(() => {
    // Read from wsState so the effect triggers
    const _h = wsState.transcriptHistory;
    const _t = wsState.transcript;
    if (transcriptContainer) {
      setTimeout(() => {
        if(transcriptContainer) {
          transcriptContainer.scrollTop = transcriptContainer.scrollHeight;
        }
      }, 0);
    }
  });
</script>

<div class="live-ears-drawer">
  <!-- Transcript History -->
  <div class="flex-1 overflow-y-auto p-4 hide-scrollbar flex flex-col gap-2" bind:this={transcriptContainer}>
    {#if wsState.transcriptHistory.length === 0 && !wsState.transcript}
      <div class="h-full flex flex-col items-center justify-center text-center text-on-surface-variant opacity-50 p-6">
        <span class="material-symbols-outlined text-[48px] mb-4">hearing_disabled</span>
        <p class="text-sm font-medium">No live transcript yet.</p>
        <p class="text-xs mt-2">Make sure your microphone is on or start the mock interview.</p>
      </div>
    {:else}
      {#each wsState.transcriptHistory as line}
        <div class="transcript-line p-3 rounded-lg bg-surface-variant/30 border border-white/5 text-sm">
           <span class="speaker-badge interviewer">Speaker</span>
           <p class="mt-1 text-on-surface/90">{line}</p>
        </div>
      {/each}
      {#if wsState.transcript}
        <div class="transcript-line p-3 rounded-lg bg-primary/10 border border-primary/20 text-sm animate-pulse">
           <span class="speaker-badge candidate">You</span>
           <p class="mt-1 text-primary/90">{wsState.transcript}</p>
        </div>
      {/if}
    {/if}
  </div>

  <!-- Pending Transcript Chips -->
  {#if wsState.pendingTranscripts.length > 0}
    <div class="p-3 border-t border-white/10 bg-surface/50 backdrop-blur-md flex flex-col gap-2">
      <div class="flex justify-between items-center px-1">
        <span class="text-[10px] uppercase font-bold tracking-wider text-on-surface-variant">Pending Responses</span>
        <button class="text-[10px] uppercase font-bold tracking-wider text-error/80 hover:text-error transition-colors" onclick={clearAllChips}>
          Clear All
        </button>
      </div>
      <div class="flex overflow-x-auto gap-2 pb-1 hide-scrollbar">
        {#each wsState.pendingTranscripts as chip (chip.id)}
          <div class="chip flex items-center bg-surface-variant/80 hover:bg-surface-variant border border-white/10 rounded-full pl-3 pr-1 py-1 whitespace-nowrap transition-all group">
            <button
              class="text-xs text-on-surface/90 truncate max-w-[150px] text-left hover:text-primary transition-colors"
              onclick={() => sendChip(chip)}
              title={chip.text}
              aria-label="Send suggestion: {chip.text}"
            >
              {chip.text}
            </button>
            <button
              class="ml-2 w-5 h-5 flex items-center justify-center rounded-full hover:bg-white/10 text-on-surface-variant hover:text-error transition-colors"
              onclick={() => dismissChip(chip.id)}
              aria-label="Dismiss suggestion: {chip.text}"
            >
              <span class="material-symbols-outlined text-[14px]">close</span>
            </button>
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Chat Input -->
  <div class="p-3 border-t border-white/10 bg-surface/80 backdrop-blur-md">
    <div class="relative">
      <textarea
        bind:value={chatText}
        onkeydown={handleKeydown}
        placeholder="Send custom prompt..."
        class="w-full bg-surface-variant/50 border border-white/10 rounded-lg py-2 pl-3 pr-10 text-sm text-on-surface focus:outline-none focus:border-primary/50 resize-none h-10 min-h-[40px] max-h-[120px]"
        rows="1"
      ></textarea>
      <button
        onclick={handleChatSubmit}
        disabled={!chatText.trim()}
        class="absolute right-2 top-1/2 -translate-y-1/2 w-6 h-6 flex items-center justify-center text-on-surface-variant hover:text-primary disabled:opacity-30 disabled:hover:text-on-surface-variant transition-colors"
      >
        <span class="material-symbols-outlined text-[18px]">send</span>
      </button>
    </div>
  </div>
</div>

<style>
  .live-ears-drawer {
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
    background: transparent;
  }

  .hide-scrollbar::-webkit-scrollbar {
    display: none;
  }
  .hide-scrollbar {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }

  .speaker-badge {
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    padding: 2px 6px;
    border-radius: 4px;
    display: inline-block;
  }

  .speaker-badge.interviewer {
    background: rgba(99, 179, 237, 0.15);
    color: #63b3ed;
    border: 1px solid rgba(99, 179, 237, 0.3);
  }

  .speaker-badge.candidate {
    background: rgba(74, 222, 128, 0.15);
    color: #4ade80;
    border: 1px solid rgba(74, 222, 128, 0.3);
  }
</style>
