<script lang="ts">
  import { wsState } from "$lib/ws.svelte";

  export let onClose: () => void;

  let selectedTemplates: Record<string, boolean> = {
    granola: true,
    star: false,
    scorecard: false
  };

  let templates = [
    { id: "granola", label: "Granola Action Notes" },
    { id: "star", label: "STAR Method Breakdown" },
    { id: "scorecard", label: "Technical Scorecard" }
  ];

  let activeTab = "granola";
  let hasGenerated = false;

  async function generateAll() {
    const ids = Object.keys(selectedTemplates).filter(id => selectedTemplates[id]);
    if (ids.length === 0) return;

    activeTab = ids[0];
    hasGenerated = true;

    try {
      await (window as any).go.main.App.SummarizeSession(ids);
    } catch (e) {
      console.error("Failed to start summarization:", e);
    }
  }

  function handleCopy() {
    const text = wsState.summaryResults[activeTab];
    if (text) {
      navigator.clipboard.writeText(text);
    }
  }
</script>

<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm pointer-events-auto">
  <div class="modal-surface w-[800px] h-[600px] rounded-xl flex flex-col overflow-hidden">
    <div class="modal-header px-6 py-4 flex justify-between items-center border-b border-white/10">
      <h2 class="text-xl font-bold text-white flex items-center gap-2">
        <span class="material-symbols-outlined text-primary">auto_awesome</span>
        Summarize Session
      </h2>
      <button class="text-white/50 hover:text-white transition-colors" on:click={onClose}>
        <span class="material-symbols-outlined">close</span>
      </button>
    </div>

    <div class="modal-body flex-1 flex overflow-hidden">
      <!-- Left sidebar: Options -->
      <div class="w-[250px] p-6 border-r border-white/10 flex flex-col gap-6 bg-black/20">
        <div>
          <h3 class="text-sm font-semibold text-white/70 mb-4 uppercase tracking-wider">Select Templates</h3>
          <div class="flex flex-col gap-3">
            {#each templates as t}
              <label class="flex items-center gap-3 cursor-pointer group">
                <input
                  type="checkbox"
                  bind:checked={selectedTemplates[t.id]}
                  class="w-4 h-4 rounded border-white/20 bg-black/40 text-primary focus:ring-primary/50"
                />
                <span class="text-sm text-white/80 group-hover:text-white transition-colors">{t.label}</span>
              </label>
            {/each}
          </div>
        </div>

        <button
          class="mt-auto flex items-center justify-center gap-2 py-3 px-4 bg-primary text-black font-bold rounded-lg hover:bg-primary/90 transition-all hover:scale-[1.02] active:scale-[0.98] disabled:opacity-50 disabled:pointer-events-none"
          on:click={generateAll}
          disabled={Object.values(selectedTemplates).every(v => !v) || Object.keys(selectedTemplates).some(k => wsState.isSummarizing[k])}
        >
          <span class="material-symbols-outlined">generating_tokens</span>
          Generate
        </button>
      </div>

      <!-- Right content: Results -->
      <div class="flex-1 flex flex-col bg-[#0a0a0a]">
        {#if hasGenerated}
          <div class="flex border-b border-white/10 px-2 pt-2 gap-2 bg-[#141414]">
            {#each templates as t}
              {#if selectedTemplates[t.id] || wsState.summaryResults[t.id]}
                <button
                  class="px-4 py-2 text-sm font-medium rounded-t-lg transition-colors flex items-center gap-2 {activeTab === t.id ? 'bg-[#2a2a2a] text-white border-b-2 border-primary' : 'text-white/50 hover:text-white hover:bg-white/5'}"
                  on:click={() => (activeTab = t.id)}
                >
                  {t.label}
                  {#if wsState.isSummarizing[t.id]}
                    <span class="material-symbols-outlined text-[14px] animate-spin">progress_activity</span>
                  {/if}
                </button>
              {/if}
            {/each}
          </div>

          <div class="flex-1 p-6 overflow-y-auto custom-scrollbar relative text-white/90">
            {#if wsState.summaryResults[activeTab]}
              <button
                class="absolute top-4 right-4 text-white/50 hover:text-primary transition-colors flex items-center gap-1 text-xs"
                on:click={handleCopy}
              >
                <span class="material-symbols-outlined text-[16px]">content_copy</span>
                Copy
              </button>
              <div class="prose prose-invert prose-sm max-w-none prose-p:leading-relaxed prose-pre:bg-black/50 prose-pre:border prose-pre:border-white/10 whitespace-pre-wrap">
                {wsState.summaryResults[activeTab]}
                {#if wsState.isSummarizing[activeTab]}
                  <span class="inline-block w-2 h-4 bg-primary animate-pulse ml-1"></span>
                {/if}
              </div>
            {:else if wsState.isSummarizing[activeTab]}
              <div class="flex flex-col items-center justify-center h-full gap-4 opacity-50">
                <span class="material-symbols-outlined text-4xl animate-spin text-primary">progress_activity</span>
                <p>Generating summary...</p>
              </div>
            {/if}
          </div>
        {:else}
          <div class="flex-1 flex flex-col items-center justify-center opacity-30 text-center px-10">
            <span class="material-symbols-outlined text-6xl mb-4">document_scanner</span>
            <h3 class="text-xl font-bold mb-2">No Summaries Yet</h3>
            <p class="text-sm">Select templates on the left and click Generate to start summarizing your session.</p>
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>

<style>
  .modal-surface {
    background: rgba(18, 18, 18, 0.95);
    border: 1px solid rgba(255, 255, 255, 0.1);
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
  }
  .custom-scrollbar::-webkit-scrollbar {
    width: 8px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 4px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: rgba(255, 255, 255, 0.2);
  }
</style>
