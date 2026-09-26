<script lang="ts">
  let { scorecard, onClose } = $props<{ scorecard: any; onClose: () => void }>();

  function scoreColor(score: number): string {
    if (score >= 8) return "text-green-400";
    if (score >= 5) return "text-yellow-400";
    return "text-red-400";
  }
</script>

<div
  class="modal-overlay pointer-events-auto"
  role="button"
  tabindex="0"
  aria-label="Close scorecard modal"
  onclick={onClose}
  onkeydown={(e) => e.key === 'Escape' && onClose()}
>
  <div
    class="modal-content scorecard-content"
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    onclick={(e) => e.stopPropagation()}
    onkeydown={(e) => e.stopPropagation()}
  >
    <button class="close-btn" onclick={onClose}>✖</button>

    <div class="auth-container">
      <h2>Session Scorecard</h2>

      <div class="overall-score-card">
        <div class="score-number {scoreColor(scorecard.overall_score)}">
          {scorecard.overall_score}<span class="score-denom">/10</span>
        </div>
        <p class="overall-summary">{scorecard.overall_summary}</p>
      </div>

      <div class="strengths-gaps-grid">
        <div class="sg-section">
          <h3>✅ Key Strengths</h3>
          <ul>
            {#each (scorecard.strengths || scorecard.key_strengths || []) as s}
              <li>{s}</li>
            {/each}
          </ul>
        </div>
        <div class="sg-section">
          <h3>⚠️ Critical Gaps</h3>
          <ul>
            {#each (scorecard.gaps || scorecard.critical_gaps || []) as g}
              <li>{g}</li>
            {/each}
          </ul>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10000;
  }

  .modal-content {
    position: relative;
    width: 95%;
    max-width: 600px;
    background: rgba(15, 15, 20, 0.97);
    border: 1px solid rgba(0, 123, 255, 0.2);
    box-shadow: 0 0 15px rgba(0, 123, 255, 0.1), 0 4px 12px rgba(0,0,0,0.5);
    border-radius: 12px;
    padding: 2rem;
    color: #fff;
    font-family: system-ui, -apple-system, sans-serif;
  }

  .scorecard-content {
    max-width: 700px;
  }

  .close-btn {
    position: absolute;
    top: 1rem;
    right: 1rem;
    background: none;
    border: none;
    color: #94a3b8;
    font-size: 1.2rem;
    cursor: pointer;
    z-index: 10;
    transition: color 0.2s;
  }

  .close-btn:hover {
    color: #e2e8f0;
  }

  .auth-container {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  h2 {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 600;
    text-align: center;
    color: #f1f5f9;
  }

  .overall-score-card {
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 12px;
    padding: 24px;
    text-align: center;
  }

  .score-number {
    font-size: 56px;
    font-weight: 700;
    line-height: 1;
    margin-bottom: 12px;
  }
  .score-denom { font-size: 24px; opacity: 0.5; }

  .overall-summary { color: rgba(255, 255, 255, 0.8); font-size: 14px; line-height: 1.6; margin: 0; }

  .strengths-gaps-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
  }

  .sg-section {
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.07);
    border-radius: 10px;
    padding: 16px;
  }

  .sg-section h3 { font-size: 14px; font-weight: 600; margin: 0 0 12px; color: rgba(255, 255, 255, 0.9); }
  .sg-section ul { margin: 0; padding-left: 20px; }
  .sg-section li { font-size: 13px; color: rgba(255, 255, 255, 0.7); margin-bottom: 6px; }

  .text-green-400 { color: #4ade80; }
  .text-yellow-400 { color: #facc15; }
  .text-red-400 { color: #f87171; }
</style>
