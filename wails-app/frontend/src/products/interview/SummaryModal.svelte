<script lang="ts">
  import { wsState } from "$lib/ws.svelte";

  export let onClose: () => void;

  import { onMount } from "svelte";

  let selectedTemplates: Record<string, boolean> = {
    granola: true,
    star: false,
    scorecard: false
  };

  type TemplateDef = {
    id: string;
    label: string;
    instruction?: string;
    exampleOutput?: string;
    isCustom?: boolean;
  };

  let baseTemplates: TemplateDef[] = [
    { id: "granola", label: "Granola Action Notes", instruction: "You are Granola. Create detailed action notes and a professional summary of this interview." },
    { id: "star", label: "STAR Method Breakdown", instruction: "Extract the key behaviors and responses from the candidate using the STAR (Situation, Task, Action, Result) method based on the interview transcript." },
    { id: "scorecard", label: "Technical Scorecard", instruction: "Create a technical scorecard evaluating the candidate based on the interview transcript. Provide ratings and justification." }
  ];

  let customTemplates: TemplateDef[] = [];
  $: templates = [...baseTemplates, ...customTemplates];

  let activeTab = "granola";
  let hasGenerated = false;

  onMount(() => {
    const saved = localStorage.getItem('barnowl_custom_templates');
    if (saved) {
      try {
        customTemplates = JSON.parse(saved);
        // Initialize selection state for loaded templates
        customTemplates.forEach(ct => {
          if (selectedTemplates[ct.id] === undefined) {
            selectedTemplates[ct.id] = false;
          }
        });
      } catch (e) {
        console.error("Failed to parse custom templates", e);
      }
    }
  });

  async function generateAll() {
    const ids = Object.keys(selectedTemplates).filter(id => selectedTemplates[id]);
    if (ids.length === 0) return;

    activeTab = ids[0];
    hasGenerated = true;

    const requests = ids.map(id => {
      const t = templates.find(temp => temp.id === id);
      let prompt = t?.instruction || "";
      if (t?.isCustom) {
        prompt = `Follow these exact instructions: ${t.instruction}`;
        if (t.exampleOutput && t.exampleOutput.trim() !== "") {
          prompt += `\n\nExample format: ${t.exampleOutput}`;
        }
      }
      return { id: id, prompt: prompt };
    });

    try {
      await (window as any).go.main.App.SummarizeSession(requests);
    } catch (e) {
      console.error("Failed to start summarization:", e);
    }
  }

  let showForm = false;
  let formName = "";
  let formInstruction = "";
  let formExample = "";

  function saveCustomTemplate() {
    if (!formName || !formInstruction) return;
    const newTemplate: TemplateDef = {
      id: "custom_" + Date.now(),
      label: formName,
      instruction: formInstruction,
      exampleOutput: formExample,
      isCustom: true
    };
    customTemplates = [...customTemplates, newTemplate];
    localStorage.setItem('barnowl_custom_templates', JSON.stringify(customTemplates));
    selectedTemplates[newTemplate.id] = true;
    showForm = false;
    formName = "";
    formInstruction = "";
    formExample = "";
  }

  function deleteCustomTemplate(id: string) {
    customTemplates = customTemplates.filter(t => t.id !== id);
    localStorage.setItem('barnowl_custom_templates', JSON.stringify(customTemplates));
    delete selectedTemplates[id];
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
              <div class="flex items-center justify-between group">
                <label class="flex items-center gap-3 cursor-pointer flex-1">
                  <input
                    type="checkbox"
                    bind:checked={selectedTemplates[t.id]}
                    class="w-4 h-4 rounded border-white/20 bg-black/40 text-primary focus:ring-primary/50"
                  />
                  <span class="text-sm text-white/80 group-hover:text-white transition-colors">{t.label}</span>
                </label>
                {#if t.isCustom}
                  <button
                    class="text-white/30 hover:text-red-400 opacity-0 group-hover:opacity-100 transition-all"
                    on:click={() => deleteCustomTemplate(t.id)}
                    title="Delete template"
                  >
                    <span class="material-symbols-outlined text-[16px]">delete</span>
                  </button>
                {/if}
              </div>
            {/each}
          </div>
        </div>

        <button
          class="flex items-center justify-center gap-2 py-2 px-3 border border-dashed border-white/20 text-white/60 text-sm rounded-lg hover:border-white/50 hover:text-white transition-all mt-4"
          on:click={() => (showForm = true)}
        >
          <span class="material-symbols-outlined text-[18px]">add</span>
          Add Custom Template
        </button>

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
        {#if showForm}
          <div class="flex-1 p-6 flex flex-col gap-4 text-white">
            <div class="flex justify-between items-center mb-2 border-b border-white/10 pb-4">
              <h3 class="text-lg font-bold">Create Custom Template</h3>
              <button class="text-white/50 hover:text-white" on:click={() => (showForm = false)}>Cancel</button>
            </div>

            <label class="flex flex-col gap-2">
              <span class="text-sm font-semibold text-white/70">Template Name</span>
              <input
                type="text"
                bind:value={formName}
                placeholder="e.g., Culture Fit Evaluation"
                class="bg-black/50 border border-white/10 rounded-lg px-4 py-2 text-sm focus:outline-none focus:border-primary"
              />
            </label>

            <label class="flex flex-col gap-2 flex-1">
              <span class="text-sm font-semibold text-white/70">Instructions (Prompt)</span>
              <textarea
                bind:value={formInstruction}
                placeholder="Detailed instructions for the LLM on how to summarize this session..."
                class="bg-black/50 border border-white/10 rounded-lg px-4 py-2 text-sm flex-1 resize-none focus:outline-none focus:border-primary"
              ></textarea>
            </label>

            <label class="flex flex-col gap-2 flex-1">
              <span class="text-sm font-semibold text-white/70">Example Output (Optional)</span>
              <textarea
                bind:value={formExample}
                placeholder="Provide an example of the desired format..."
                class="bg-black/50 border border-white/10 rounded-lg px-4 py-2 text-sm flex-1 resize-none focus:outline-none focus:border-primary"
              ></textarea>
            </label>

            <button
              class="mt-4 flex items-center justify-center gap-2 py-3 px-4 bg-primary text-black font-bold rounded-lg hover:bg-primary/90 transition-all disabled:opacity-50"
              on:click={saveCustomTemplate}
              disabled={!formName || !formInstruction}
            >
              <span class="material-symbols-outlined text-[20px]">save</span>
              Save Template
            </button>
          </div>
        {:else if hasGenerated}
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
