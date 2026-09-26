<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  function handleAgree() {
    dispatch('agreed');
  }
</script>

<div class="modal-overlay">
  <div class="modal-content">
    <div class="modal-header">
      <div class="modal-icon">⚖️</div>
      <h2>Welcome to BarnOwl AI</h2>
    </div>

    <div class="modal-body">
      <p class="intro-text">Before using the software, you must acknowledge and agree to the following conditions:</p>

      <div class="legal-section">
        <h3>1. Audio Recording Consent</h3>
        <p>You are solely responsible for complying with all local, state, and federal laws regarding audio recording and transcription. You must obtain explicit consent from all meeting participants where legally required.</p>
      </div>

      <div class="legal-section">
        <h3>2. Acceptable Use (No Exams)</h3>
        <p>BarnOwl AI is designed as a real-time meeting copilot, accessibility tool, and pair-programming assistant. <strong>It is strictly prohibited to use this software to gain an unfair advantage in exams</strong>, academic exams, or any technical assessments where outside assistance is disallowed.</p>
      </div>

      <p class="footer-text">By clicking below, you accept full legal responsibility for how you choose to use this software.</p>
    </div>

    <div class="modal-actions">
      <button class="btn-primary w-full" onclick={handleAgree}>
        I Understand and Agree
      </button>
    </div>
  </div>
</div>

<style>
  .modal-overlay {
    position: fixed;
    top: 0; left: 0; right: 0; bottom: 0;
    background: rgba(10, 10, 14, 0.95);
    backdrop-filter: blur(8px);
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .modal-content {
    background: #15151c;
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    width: 480px;
    max-width: 90vw;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
    overflow: hidden;
  }
  .modal-header {
    padding: 1.5rem 1.5rem 1rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }
  .modal-icon {
    font-size: 1.5rem;
    background: rgba(255, 255, 255, 0.05);
    padding: 0.5rem;
    border-radius: 8px;
  }
  .modal-header h2 {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 600;
    color: #e5e7eb;
  }
  .modal-body {
    padding: 1.5rem;
  }
  .intro-text {
    font-size: 0.85rem;
    color: #9ca3af;
    margin-bottom: 1.25rem;
  }
  .legal-section {
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.05);
    border-radius: 8px;
    padding: 1rem;
    margin-bottom: 1rem;
  }
  .legal-section h3 {
    margin: 0 0 0.5rem 0;
    font-size: 0.85rem;
    font-weight: 600;
    color: #f3f4f6;
  }
  .legal-section p {
    margin: 0;
    font-size: 0.78rem;
    color: #9ca3af;
    line-height: 1.5;
  }
  .legal-section strong {
    color: #f87171;
    font-weight: 600;
  }
  .footer-text {
    font-size: 0.75rem;
    color: #6b7280;
    text-align: center;
    margin-top: 1rem;
  }
  .modal-actions {
    padding: 1rem 1.5rem 1.5rem;
  }
  .btn-primary {
    background: #4ade80;
    color: #000;
    border: none;
    padding: 0.75rem 1.5rem;
    border-radius: 8px;
    font-weight: 600;
    font-size: 0.9rem;
    cursor: pointer;
    transition: background 0.2s;
  }
  .btn-primary:hover {
    background: #22c55e;
  }
  .w-full {
    width: 100%;
  }
</style>