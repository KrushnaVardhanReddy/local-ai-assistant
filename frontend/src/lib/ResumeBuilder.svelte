<script lang="ts">
  import { onMount } from "svelte";
  import { renderMarkdown } from "$lib/markdownRenderer";

  let editorText = $state("");
  let renderedMarkdown = $state("");
  let isTailoring = $state(false);
  let statusMessage = $state("");

  // Re-render when editorText changes
  $effect(() => {
    let active = true;
    if (editorText) {
      renderMarkdown(editorText).then((html) => {
        if (active) renderedMarkdown = html;
      });
    } else {
      renderedMarkdown = "<div class='text-on-surface-variant/50 p-4 text-center'>Markdown preview will appear here...</div>";
    }
    return () => { active = false; };
  });

  async function handleTailor() {
    isTailoring = true;
    statusMessage = "Tailoring resume to job description...";

    try {
      // Get the backend URL from localStorage (saved by Settings.svelte) or default
      const storedBackendUrl = localStorage.getItem("backend_url") || "127.0.0.1:8765";
      const apiUrl = storedBackendUrl.startsWith('http') ? storedBackendUrl : `http://${storedBackendUrl}`;

      // Fetch the context and job description
      const [resContext, resJob] = await Promise.all([
        fetch(`${apiUrl}/api/resume/context`),
        fetch(`${apiUrl}/config/job-description`).catch(() => ({ ok: false }))
      ]);

      let baseResume = "";
      let jobDescription = "";

      if (resContext.ok) {
        const contextData = await resContext.json();
        baseResume = contextData.context || "";
      }
      if (resJob.ok) {
        const jobData = await resJob.json();
        jobDescription = jobData.text || "";
      }

      const res = await fetch(`${apiUrl}/resume/tailor`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          base_resume: baseResume,
          job_description: jobDescription
        })
      });

      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.detail || "Failed to tailor resume");
      }

      const data = await res.json();
      editorText = data.resume || "";
      statusMessage = "Resume tailored successfully!";

      setTimeout(() => {
        if (statusMessage === "Resume tailored successfully!") {
          statusMessage = "";
        }
      }, 3000);
    } catch (e: any) {
      statusMessage = `Error: ${e.message}`;
    } finally {
      isTailoring = false;
    }
  }

  function handleExport() {
    const blob = new Blob([editorText], { type: "text/markdown" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "tailored_resume.md";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }
</script>

<div class="resume-builder-container glass-panel pointer-events-auto h-full flex flex-col overflow-hidden w-full max-w-7xl mx-auto rounded-[24px] shadow-2xl relative">
  <!-- Header -->
  <div class="flex items-center justify-between p-4 border-b border-white/5 bg-black/20 shrink-0">
    <div class="flex items-center gap-3">
      <span class="material-symbols-outlined text-primary text-[24px]" data-icon="description">description</span>
      <h2 class="font-headline-md text-headline-md text-primary tracking-wide">Resume Builder</h2>
    </div>

    <div class="flex items-center gap-4">
      {#if statusMessage}
        <span class="text-sm {statusMessage.startsWith('Error') ? 'text-red-400' : 'text-green-400'} animate-pulse">{statusMessage}</span>
      {/if}

      <button
        class="flex items-center gap-2 px-4 py-2 rounded-full bg-primary/20 text-primary hover:bg-primary/30 transition-colors font-label-caps text-label-caps tracking-wider {isTailoring ? 'opacity-50 cursor-not-allowed' : ''}"
        onclick={handleTailor}
        disabled={isTailoring}
      >
        <span class="material-symbols-outlined text-[18px]">auto_awesome</span>
        {isTailoring ? 'TAILORING...' : 'TAILOR TO JOB'}
      </button>

      <button
        class="flex items-center gap-2 px-4 py-2 rounded-full bg-white/5 hover:bg-white/10 text-on-surface-variant transition-colors font-label-caps text-label-caps tracking-wider border border-white/10 {!editorText ? 'opacity-50 cursor-not-allowed' : ''}"
        onclick={handleExport}
        disabled={!editorText}
      >
        <span class="material-symbols-outlined text-[18px]">download</span>
        EXPORT
      </button>
    </div>
  </div>

  <!-- Workspace: Split View -->
  <div class="flex-1 flex overflow-hidden">
    <!-- Editor Pane -->
    <div class="w-1/2 flex flex-col border-r border-white/5 relative">
      <div class="absolute top-2 right-4 bg-surface-variant/80 px-2 py-1 rounded text-xs font-mono text-on-surface-variant/50 pointer-events-none z-10">MARKDOWN</div>
      <textarea
        class="flex-1 w-full p-6 bg-transparent border-none text-on-surface-variant font-mono-data text-mono-data focus:ring-0 outline-none resize-none"
        placeholder="Markdown editor. Click 'Tailor to Job' to generate..."
        bind:value={editorText}
        spellcheck="false"
      ></textarea>
    </div>

    <!-- Preview Pane -->
    <div class="w-1/2 flex flex-col bg-black/10 relative">
      <div class="absolute top-2 right-4 bg-surface-variant/80 px-2 py-1 rounded text-xs font-mono text-on-surface-variant/50 pointer-events-none z-10">PREVIEW</div>
      <div class="flex-1 overflow-y-auto p-8 markdown-preview">
        {@html renderedMarkdown}
      </div>
    </div>
  </div>
</div>

<style>
  .resume-builder-container {
    backdrop-filter: blur(20px);
    background: rgba(17, 24, 39, 0.7);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  /* Basic markdown preview styles (similar to what might be in global CSS but scoped here just in case) */
  :global(.markdown-preview) {
    color: #cbd5e1;
    font-size: 0.95rem;
    line-height: 1.6;
  }
  :global(.markdown-preview h1) {
    font-size: 1.8rem;
    font-weight: 700;
    color: #f8fafc;
    margin-bottom: 1rem;
    margin-top: 1.5rem;
    border-bottom: 1px solid rgba(255,255,255,0.1);
    padding-bottom: 0.5rem;
  }
  :global(.markdown-preview h2) {
    font-size: 1.4rem;
    font-weight: 600;
    color: #e2e8f0;
    margin-bottom: 0.75rem;
    margin-top: 1.5rem;
  }
  :global(.markdown-preview h3) {
    font-size: 1.1rem;
    font-weight: 600;
    color: #cbd5e1;
    margin-bottom: 0.5rem;
    margin-top: 1rem;
  }
  :global(.markdown-preview p) {
    margin-bottom: 1rem;
  }
  :global(.markdown-preview ul) {
    list-style-type: disc;
    margin-left: 1.5rem;
    margin-bottom: 1rem;
  }
  :global(.markdown-preview ol) {
    list-style-type: decimal;
    margin-left: 1.5rem;
    margin-bottom: 1rem;
  }
  :global(.markdown-preview li) {
    margin-bottom: 0.25rem;
  }
  :global(.markdown-preview strong) {
    color: #f8fafc;
    font-weight: 600;
  }
  :global(.markdown-preview em) {
    font-style: italic;
  }
</style>
