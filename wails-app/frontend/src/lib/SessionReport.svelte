<script lang="ts">
  import { apiFetch } from "./api";
  import { getApiUrl } from "$lib/api";

  let { onClose }: { onClose: () => void } = $props();

  let isLoading = $state(false);
  let error = $state<string | null>(null);
  let scorecard = $state<any>(null);
  let session = $state<any>(null);
  let showExportMenu = $state(false);
  let showCopiedToast = $state(false);

  let emailDraft = $state<string | null>(null);
  let isGeneratingEmail = $state(false);
  let emailError = $state<string | null>(null);

  let interviewerName = $state("");
  let companyName = $state("");
  let roleName = $state("");
  let showEmailMeta = $state(false);

  let showHistory = $state(false);
  let historyEntries = $state<any[]>([]);
  let isLoadingHistory = $state(false);

  async function loadHistory() {
    if (historyEntries.length > 0) return; // already loaded
    isLoadingHistory = true;
    try {
      const resp = await apiFetch(`${getApiUrl()}/session/history`);
      if (resp.ok) {
        historyEntries = await resp.json();
      }
    } catch (_) {
      // History is best-effort — silently fail
    } finally {
      isLoadingHistory = false;
    }
  }

  function buildSparkline(scores: number[]): string {
    if (scores.length < 2) return "";
    const W = 120, H = 32, pad = 4;
    const min = Math.min(...scores, 0);
    const max = Math.max(...scores, 10);
    const range = max - min || 1;
    const pts = scores.map((s, i) => {
      const x = pad + (i / (scores.length - 1)) * (W - pad * 2);
      const y = H - pad - ((s - min) / range) * (H - pad * 2);
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    });
    return `<svg width="${W}" height="${H}" viewBox="0 0 ${W} ${H}">`
      + `<polyline points="${pts.join(" ")}" fill="none" stroke="rgba(78,222,163,0.8)" stroke-width="1.5" stroke-linejoin="round"/>`
      + `<circle cx="${pts[pts.length-1].split(",")[0]}" cy="${pts[pts.length-1].split(",")[1]}" r="2.5" fill="rgb(78,222,163)"/>`
      + `</svg>`;
  }

  const VERDICT_COLORS: Record<string, string> = {
    Good: "text-green-400 bg-green-400/10 border-green-400/30",
    Partial: "text-yellow-400 bg-yellow-400/10 border-yellow-400/30",
    Incomplete: "text-orange-400 bg-orange-400/10 border-orange-400/30",
    "Off-topic": "text-red-400 bg-red-400/10 border-red-400/30",
  };

  async function loadReport() {
    isLoading = true;
    error = null;
    try {
      const resp = await apiFetch(`${getApiUrl()}/session/end`, { method: "POST" });
      if (!resp.ok) {
        const data = await resp.json();
        error = data.error || "Failed to generate report";
        return;
      }
      const data = await resp.json();
      session = data.session;
      scorecard = data.scorecard;
    } catch (e) {
      error = "Could not connect to backend";
    } finally {
      isLoading = false;
    }
  }

  function getReportMarkdown() {
    if (!scorecard) return "";
    const lines = [
      `# Interview Scorecard — ${new Date().toLocaleDateString()}`,
      ``,
      `**Overall Score:** ${scorecard.overall_score}/10`,
      ``,
      `**Summary:** ${scorecard.overall_summary}`,
      ``,
      `**Strengths:** ${scorecard.strengths?.join(", ") || "None"}`,
      `**Gaps:** ${scorecard.gaps?.join(", ") || "None"}`,
      ``,
      `## Turns Breakdown`,
      ...(scorecard.turns || []).map((t: any) =>
        `### Q${t.turn + 1}: ${t.verdict} (${t.score}/5)\n- **Good:** ${t.what_was_good}\n- **Missing:** ${t.what_was_missing || "—"}`
      ),
    ];
    return lines.join("\n");
  }

  function copyToClipboard() {
    const md = getReportMarkdown();
    if (!md) return;
    navigator.clipboard.writeText(md);
    showCopiedToast = true;
    setTimeout(() => showCopiedToast = false, 2000);
    showExportMenu = false;
  }

  function downloadMarkdown() {
    const md = getReportMarkdown();
    if (!md) return;
    const blob = new Blob([md], { type: "text/markdown" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "session-report.md";
    a.click();
    URL.revokeObjectURL(url);
    showExportMenu = false;
  }

  function draftEmail() {
    const md = getReportMarkdown();
    if (!md) return;
    const subject = encodeURIComponent("Session Report");
    const body = encodeURIComponent(md);
    window.location.href = `mailto:?subject=${subject}&body=${body}`;
    showExportMenu = false;
  }

  async function generateEmailDraft() {
    if (!session || !scorecard) return;
    isGeneratingEmail = true;
    emailError = null;
    try {
      const resp = await apiFetch(`${getApiUrl()}/session/email-draft`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          session,
          scorecard,
          interviewer_name: interviewerName,
          company_name: companyName,
          role_name: roleName,
        }),
      });
      if (!resp.ok) {
        const data = await resp.json();
        emailError = data.error || "Failed to generate email";
        return;
      }
      const data = await resp.json();
      emailDraft = data.email;
    } catch (e) {
      emailError = "Could not connect to backend";
    } finally {
      isGeneratingEmail = false;
    }
  }

  function copyEmail() {
    if (emailDraft) navigator.clipboard.writeText(emailDraft);
  }

  // Load report on mount
  $effect(() => {
    loadReport();
  });

  function scoreColor(score: number): string {
    if (score >= 8) return "text-green-400";
    if (score >= 5) return "text-yellow-400";
    return "text-red-400";
  }
</script>

<div class="report-overlay" role="dialog" aria-modal="true">
  <div class="report-panel">
    <!-- Header -->
    <div class="report-header">
      <div class="report-title">
        <span class="material-symbols-outlined text-primary">analytics</span>
        <h2>Session Report</h2>
        {#if session}
          <span class="session-meta">{session.turn_count} questions · {Math.round(session.session_duration_s / 60)}m</span>
        {/if}
      </div>
      <div class="header-actions">
        {#if scorecard}
          <div class="relative">
            <button class="action-btn flex items-center gap-2 px-3 py-1.5" onclick={() => showExportMenu = !showExportMenu} aria-label="Export Report">
              <span class="material-symbols-outlined text-[18px]">ios_share</span>
              <span class="text-sm font-medium">Export</span>
            </button>
            {#if showExportMenu}
              <div class="fixed inset-0 z-40" aria-label="Close export menu" role="button" tabindex="0" onclick={() => showExportMenu = false} onkeydown={(e) => e.key === 'Escape' && (showExportMenu = false)}></div>
              <div class="absolute right-0 top-full mt-2 w-48 bg-[#1a1a2e]/95 backdrop-blur-md border border-white/10 rounded-xl shadow-2xl overflow-hidden z-50 flex flex-col">
                <button class="px-4 py-3 text-left text-sm text-white/80 hover:bg-white/10 hover:text-white transition-colors flex items-center gap-2" onclick={copyToClipboard}>
                  <span class="material-symbols-outlined text-[16px]">content_copy</span>
                  Copy to Clipboard
                </button>
                <button class="px-4 py-3 text-left text-sm text-white/80 hover:bg-white/10 hover:text-white transition-colors flex items-center gap-2" onclick={downloadMarkdown}>
                  <span class="material-symbols-outlined text-[16px]">download</span>
                  Download as Markdown
                </button>
                <button class="px-4 py-3 text-left text-sm text-white/80 hover:bg-white/10 hover:text-white transition-colors flex items-center gap-2" onclick={() => window.print()}>
                  <span class="material-symbols-outlined text-[16px]">picture_as_pdf</span>
                  Download as PDF
                </button>
                <button class="px-4 py-3 text-left text-sm text-white/80 hover:bg-white/10 hover:text-white transition-colors flex items-center gap-2" onclick={draftEmail}>
                  <span class="material-symbols-outlined text-[16px]">mail</span>
                  Draft as Email
                </button>
              </div>
            {/if}
            {#if showCopiedToast}
              <div class="absolute -top-10 left-1/2 -translate-x-1/2 bg-green-500/20 border border-green-500/30 text-green-400 text-xs px-3 py-1.5 rounded-lg whitespace-nowrap z-50 shadow-lg">
                Copied to clipboard!
              </div>
            {/if}
          </div>
        {/if}
        <button class="action-btn close-btn" onclick={onClose} aria-label="Close">
          <span class="material-symbols-outlined">close</span>
        </button>
      </div>
    </div>

    <!-- Body -->
    <div class="report-body">
      {#if isLoading}
        <div class="loading-state">
          <div class="spinner"></div>
          <p>Generating scorecard...</p>
        </div>
      {:else if error}
        <div class="error-state">
          <span class="material-symbols-outlined">error</span>
          <p>{error}</p>
          <button class="retry-btn" onclick={loadReport}>Retry</button>
        </div>
      {:else if scorecard}
        <!-- Overall Score -->
        <div class="overall-score-card">
          <div class="score-number {scoreColor(scorecard.overall_score)}">
            {scorecard.overall_score}<span class="score-denom">/10</span>
          </div>
          <p class="overall-summary">{scorecard.overall_summary}</p>
        </div>

        <!-- Strengths & Gaps -->
        <div class="strengths-gaps-grid">
          <div class="sg-section">
            <h3>✅ Strengths</h3>
            <ul>
              {#each scorecard.strengths || [] as s}
                <li>{s}</li>
              {/each}
            </ul>
          </div>
          <div class="sg-section">
            <h3>⚠️ Gaps</h3>
            <ul>
              {#each scorecard.gaps || [] as g}
                <li>{g}</li>
              {/each}
            </ul>
          </div>
        </div>

        <!-- Per-Turn Breakdown -->
        <div class="turns-list">
          {#each scorecard.turns || [] as turn}
            <div class="turn-card">
              <div class="turn-header">
                <span class="turn-label">Q{turn.turn + 1}</span>
                <span class="verdict-badge {VERDICT_COLORS[turn.verdict] || ''}">{turn.verdict}</span>
                <span class="turn-score">{turn.score}/5</span>
              </div>
              {#if session?.turns?.[turn.turn]}
                <p class="turn-transcript">"{session.turns[turn.turn].transcript}"</p>
              {/if}
              <p class="turn-good">👍 {turn.what_was_good}</p>
              {#if turn.what_was_missing}
                <p class="turn-missing">⚠️ {turn.what_was_missing}</p>
              {/if}
              {#if turn.suggested_addition}
                <p class="turn-suggestion">💡 Should have said: <em>{turn.suggested_addition}</em></p>
              {/if}
            </div>
          {/each}
        </div>

        <!-- Email Draft Section -->
        <div class="mt-6 border-t border-white/10 pt-6">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-sm font-semibold text-on-background/90 flex items-center gap-2">
              <span class="material-symbols-outlined text-[16px] text-secondary">mail</span>
              Follow-Up Email Draft
            </h3>
            <button
              class="text-[10px] text-on-surface-variant/50 hover:text-primary transition-colors"
              onclick={() => showEmailMeta = !showEmailMeta}
            >
              {showEmailMeta ? 'Hide options ▲' : 'Add details ▼'}
            </button>
          </div>

          {#if showEmailMeta}
            <div class="grid grid-cols-3 gap-2 mb-3">
              <input
                type="text"
                placeholder="Interviewer name"
                bind:value={interviewerName}
                class="bg-white/5 border border-white/10 rounded-lg px-3 py-2
                       text-on-background text-xs focus:outline-none
                       focus:border-primary/40 transition-colors"
              />
              <input
                type="text"
                placeholder="Company name"
                bind:value={companyName}
                class="bg-white/5 border border-white/10 rounded-lg px-3 py-2
                       text-on-background text-xs focus:outline-none
                       focus:border-primary/40 transition-colors"
              />
              <input
                type="text"
                placeholder="Role / position"
                bind:value={roleName}
                class="bg-white/5 border border-white/10 rounded-lg px-3 py-2
                       text-on-background text-xs focus:outline-none
                       focus:border-primary/40 transition-colors"
              />
            </div>
          {/if}

          {#if !emailDraft}
            <button
              class="w-full py-2.5 px-4 rounded-xl bg-secondary/10 border border-secondary/20
                     hover:bg-secondary/20 hover:border-secondary/40 text-secondary
                     text-sm font-medium transition-all duration-200 flex items-center
                     justify-center gap-2 disabled:opacity-40 disabled:cursor-not-allowed"
              onclick={generateEmailDraft}
              disabled={isGeneratingEmail}
            >
              {#if isGeneratingEmail}
                <span class="material-symbols-outlined text-[16px] animate-spin">progress_activity</span>
                Generating...
              {:else}
                <span class="material-symbols-outlined text-[16px]">auto_awesome</span>
                Draft Follow-Up Email
              {/if}
            </button>
          {/if}

          {#if emailError}
            <p class="text-red-400 text-xs mt-2">{emailError}</p>
          {/if}

          {#if emailDraft}
            <div class="relative mt-3">
              <textarea
                bind:value={emailDraft}
                class="w-full h-52 bg-white/5 border border-white/10 rounded-xl
                       p-4 text-on-background/90 text-sm font-mono leading-relaxed
                       resize-none focus:outline-none focus:border-primary/40
                       focus:bg-white/8 transition-colors"
                spellcheck="true"
              ></textarea>
              <div class="flex gap-2 mt-2">
                <button
                  class="flex-1 py-2 rounded-xl bg-white/5 hover:bg-white/10 border
                         border-white/10 hover:border-white/20 text-on-surface-variant
                         hover:text-on-background text-xs font-medium transition-all
                         flex items-center justify-center gap-1.5"
                  onclick={copyEmail}
                >
                  <span class="material-symbols-outlined text-[14px]">content_copy</span>
                  Copy Email
                </button>
                <button
                  class="py-2 px-4 rounded-xl bg-white/5 hover:bg-white/10 border
                         border-white/10 text-on-surface-variant/60 text-xs
                         transition-all"
                  onclick={() => { emailDraft = null; }}
                >
                  Regenerate
                </button>
              </div>
            </div>
          {/if}
        </div>

        <!-- Past Sessions / Trend View -->
        <div class="mt-6 border-t border-white/10 pt-6">
          <button
            class="w-full flex items-center justify-between text-left group"
            onclick={() => { showHistory = !showHistory; if (showHistory) loadHistory(); }}
          >
            <h3 class="text-sm font-semibold text-on-background/90 flex items-center gap-2">
              <span class="material-symbols-outlined text-[16px] text-primary">trending_up</span>
              Past Sessions
              {#if historyEntries.length > 0}
                <span class="text-[10px] text-on-surface-variant/50 font-normal">
                  ({historyEntries.length} sessions)
                </span>
              {/if}
            </h3>
            <span class="material-symbols-outlined text-[18px] text-on-surface-variant/40
                         group-hover:text-primary transition-colors">
              {showHistory ? 'expand_less' : 'expand_more'}
            </span>
          </button>

          {#if showHistory}
            <div class="mt-4">
              {#if isLoadingHistory}
                <div class="flex items-center gap-2 text-on-surface-variant/50 text-xs py-4">
                  <span class="material-symbols-outlined text-[14px] animate-spin">progress_activity</span>
                  Loading history...
                </div>
              {:else if historyEntries.length === 0}
                <p class="text-on-surface-variant/40 text-xs py-4 text-center">
                  No past sessions yet. Complete more interviews to see your trend.
                </p>
              {:else}
                <!-- Sparkline trend -->
                <div class="flex items-center gap-3 mb-4 p-3 bg-white/3 rounded-xl border border-white/5">
                  <div>
                    {@html buildSparkline(historyEntries.slice(0, 20).map(e => e.overall_score).reverse())}
                  </div>
                  <div>
                    <p class="text-[10px] text-on-surface-variant/50 uppercase tracking-wider">Score Trend</p>
                    <p class="text-sm font-semibold text-on-background">
                      {(historyEntries.slice(0,5).reduce((a,e) => a + e.overall_score, 0) / Math.min(historyEntries.length, 5)).toFixed(1)}/10
                      <span class="text-[10px] font-normal text-on-surface-variant/50">avg (last 5)</span>
                    </p>
                  </div>
                </div>

                <!-- Session list -->
                <div class="flex flex-col gap-2 max-h-64 overflow-y-auto hide-scrollbar">
                  {#each historyEntries.slice(0, 15) as entry, i}
                    <div class="p-3 rounded-xl bg-white/3 border border-white/5
                                hover:bg-white/5 transition-colors">
                      <div class="flex items-center justify-between mb-1">
                        <span class="text-[10px] text-on-surface-variant/50">
                          {new Date(entry.date_iso).toLocaleDateString(undefined,
                            { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                        </span>
                        <span class="text-sm font-bold {entry.overall_score >= 8 ? 'text-green-400' : entry.overall_score >= 6 ? 'text-yellow-400' : 'text-orange-400'}">
                          {entry.overall_score}/10
                        </span>
                      </div>
                      {#if entry.strengths?.length}
                        <p class="text-[11px] text-green-400/80 truncate">
                          ✓ {entry.strengths[0]}
                        </p>
                      {/if}
                      {#if entry.gaps?.length}
                        <p class="text-[11px] text-orange-400/70 truncate">
                          ✗ {entry.gaps[0]}
                        </p>
                      {/if}
                      <p class="text-[10px] text-on-surface-variant/40 mt-1">
                        {entry.turn_count} questions · {Math.round(entry.duration_s / 60)}min
                      </p>
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          {/if}
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .report-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .report-panel {
    background: rgba(15, 15, 25, 0.95);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 16px;
    width: 90%;
    max-width: 800px;
    max-height: 85vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 25px 60px rgba(0, 0, 0, 0.5);
  }

  .report-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px 24px 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    flex-shrink: 0;
  }

  .report-title {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .report-title h2 {
    font-size: 18px;
    font-weight: 600;
    color: #e2e8f0;
    margin: 0;
  }

  .session-meta {
    font-size: 12px;
    color: rgba(255,255,255,0.4);
    background: rgba(255,255,255,0.06);
    padding: 2px 8px;
    border-radius: 999px;
  }

  .header-actions { display: flex; gap: 8px; }

  .action-btn {
    background: rgba(255,255,255,0.06);
    border: 1px solid rgba(255,255,255,0.1);
    border-radius: 8px;
    color: rgba(255,255,255,0.6);
    cursor: pointer;
    padding: 6px;
    display: flex;
    transition: all 0.2s;
  }
  .action-btn:hover { background: rgba(255,255,255,0.12); color: white; }

  .report-body { overflow-y: auto; padding: 24px; flex: 1; }

  .loading-state, .error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 60px 0;
    color: rgba(255,255,255,0.5);
  }

  .spinner {
    width: 32px; height: 32px;
    border: 3px solid rgba(255,255,255,0.1);
    border-top-color: #63b3ed;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }

  .overall-score-card {
    background: rgba(255,255,255,0.04);
    border: 1px solid rgba(255,255,255,0.08);
    border-radius: 12px;
    padding: 24px;
    text-align: center;
    margin-bottom: 20px;
  }

  .score-number {
    font-size: 56px;
    font-weight: 700;
    line-height: 1;
    margin-bottom: 12px;
  }
  .score-denom { font-size: 24px; opacity: 0.5; }

  .overall-summary { color: rgba(255,255,255,0.7); font-size: 14px; line-height: 1.6; margin: 0; }

  .strengths-gaps-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
    margin-bottom: 20px;
  }

  .sg-section {
    background: rgba(255,255,255,0.03);
    border: 1px solid rgba(255,255,255,0.07);
    border-radius: 10px;
    padding: 16px;
  }
  .sg-section h3 { font-size: 13px; font-weight: 600; margin: 0 0 10px; color: rgba(255,255,255,0.8); }
  .sg-section ul { margin: 0; padding-left: 16px; }
  .sg-section li { font-size: 13px; color: rgba(255,255,255,0.6); margin-bottom: 4px; }

  .turns-list { display: flex; flex-direction: column; gap: 12px; }

  .turn-card {
    background: rgba(255,255,255,0.03);
    border: 1px solid rgba(255,255,255,0.07);
    border-radius: 10px;
    padding: 16px;
  }

  .turn-header { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; }
  .turn-label { font-size: 13px; font-weight: 600; color: rgba(255,255,255,0.5); min-width: 28px; }
  .verdict-badge { font-size: 11px; padding: 2px 8px; border-radius: 999px; border: 1px solid; font-weight: 600; }
  .turn-score { font-size: 13px; color: rgba(255,255,255,0.4); margin-left: auto; }

  .turn-transcript { font-size: 12px; color: rgba(255,255,255,0.4); font-style: italic; margin: 0 0 8px; }
  .turn-good { font-size: 13px; color: rgba(255,255,255,0.7); margin: 4px 0; }
  .turn-missing { font-size: 13px; color: rgba(251, 191, 36, 0.8); margin: 4px 0; }
  .turn-suggestion { font-size: 13px; color: rgba(147, 197, 253, 0.8); margin: 4px 0; }
  .turn-suggestion em { font-style: normal; }

  .retry-btn {
    background: rgba(99, 179, 237, 0.15);
    border: 1px solid rgba(99, 179, 237, 0.3);
    color: #63b3ed;
    border-radius: 8px;
    padding: 8px 20px;
    cursor: pointer;
    font-size: 13px;
  }
  .hide-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }
  .hide-scrollbar::-webkit-scrollbar { display: none; }
</style>
