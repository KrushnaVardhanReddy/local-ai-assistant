<script lang="ts">
  import { page } from '$app/stores';

  let code = $derived($page.url.searchParams.get('code') || '');
  let loading = $state(false);
  let error = $state<string | null>(null);
  let sessionToken = $state<any>(null);

  async function startDemo() {
    if (!code) {
      error = "No referral code provided.";
      return;
    }

    loading = true;
    error = null;

    try {
      const response = await fetch('/api/demo', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ code })
      });

      if (!response.ok) {
        const data = await response.json().catch(() => ({}));
        throw new Error(data.error || 'Failed to start demo.');
      }

      sessionToken = await response.json();
    } catch (err: any) {
      error = err.message || 'An unexpected error occurred.';
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Free Demo - Local AI Assistant</title>
</svelte:head>

<main>
  <div class="demo-container">
    <h1>Welcome to your Free Demo</h1>

    {#if sessionToken}
      <div class="success-message">
        <h2>Your Demo Session is Ready!</h2>
        <p>You have a 15-minute free session.</p>
        <div class="token-box">
          <p><strong>Status:</strong> {sessionToken.status}</p>
          <p><strong>Expires:</strong> {new Date(sessionToken.expiresAt).toLocaleString()}</p>
          <p class="token-value"><strong>Token:</strong> <code>{sessionToken.token}</code></p>
        </div>
        <p class="instructions">Open the Local AI Assistant desktop app and use this token to authenticate your session.</p>
      </div>
    {:else}
      <p class="intro">
        You've been invited to try Local AI Assistant!
        Click the button below to generate your 15-minute free demo token.
      </p>

      {#if code}
        <div class="code-box">
          Referral Code: <strong>{code}</strong>
        </div>
      {/if}

      {#if error}
        <div class="error-message">
          {error}
        </div>
      {/if}

      <button class="btn btn-primary btn-large" onclick={startDemo} disabled={loading || !code}>
        {loading ? 'Starting Demo...' : 'Start My 15-Minute Free Demo'}
      </button>

      {#if !code}
        <p class="no-code-warning">A valid referral code is required to start a demo.</p>
      {/if}
    {/if}
  </div>
</main>

<style>
  main {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: calc(100vh - 80px);
    padding: 2rem;
    background-color: var(--background);
  }

  .demo-container {
    background-color: var(--surface);
    border: var(--border);
    border-radius: 12px;
    padding: 3rem;
    max-width: 600px;
    width: 100%;
    text-align: center;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }

  h1 {
    font-size: 2.25rem;
    margin-bottom: 1.5rem;
    color: var(--text);
  }

  .intro {
    font-size: 1.125rem;
    color: var(--muted);
    margin-bottom: 2rem;
    line-height: 1.6;
  }

  .code-box {
    background-color: var(--background);
    padding: 1rem;
    border-radius: 8px;
    margin-bottom: 2rem;
    font-size: 1.125rem;
    border: 1px dashed var(--muted);
  }

  .btn {
    padding: 0.75rem 1.5rem;
    border-radius: 6px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    border: none;
  }

  .btn-primary {
    background-color: var(--accent);
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background-color: #6d28d9;
    transform: translateY(-1px);
  }

  .btn-primary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .btn-large {
    font-size: 1.125rem;
    width: 100%;
    padding: 1rem;
  }

  .error-message {
    color: #ef4444;
    background-color: rgba(239, 68, 68, 0.1);
    padding: 1rem;
    border-radius: 8px;
    margin-bottom: 1.5rem;
  }

  .success-message {
    text-align: left;
  }

  .success-message h2 {
    color: #10b981;
    margin-bottom: 1rem;
  }

  .token-box {
    background-color: var(--background);
    padding: 1.5rem;
    border-radius: 8px;
    margin: 1.5rem 0;
    border: 1px solid var(--border-color, #e5e7eb);
  }

  .token-box p {
    margin-bottom: 0.5rem;
  }

  .token-box p:last-child {
    margin-bottom: 0;
  }

  .token-value code {
    background-color: rgba(124, 58, 237, 0.1);
    color: var(--accent);
    padding: 0.2rem 0.4rem;
    border-radius: 4px;
    word-break: break-all;
  }

  .instructions {
    color: var(--muted);
    font-size: 0.95rem;
    margin-top: 1.5rem;
  }

  .no-code-warning {
    color: var(--muted);
    margin-top: 1rem;
    font-size: 0.9rem;
  }
</style>
