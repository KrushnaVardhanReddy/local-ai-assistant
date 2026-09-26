<script lang="ts">
  import { authStore } from "./authStore.svelte.ts";
  import Titlebar from "./components/Titlebar.svelte";

  let email = $state("");
  let isLoading = $state(false);
  let errorMsg = $state<string | null>(null);

  async function handleSSO() {
    isLoading = true;
    errorMsg = null;

    // Simulate brief loading state
    await new Promise(resolve => setTimeout(resolve, 800));

    const result = authStore.login(email);
    if (!result.success) {
      errorMsg = result.error || "Authentication failed.";
    }

    isLoading = false;
  }
</script>

<div class="fixed inset-0 z-[10000] flex flex-col bg-[#0f0f14]">
  <Titlebar />
  <div class="flex-1 flex flex-col items-center justify-center p-4">
    <div class="modal-card">
      <h1 class="title">BarnOwl AI Enterprise Login</h1>
      <p class="subtitle">Please authenticate with your corporate SSO to continue.</p>

      <form onsubmit={(e) => { e.preventDefault(); handleSSO(); }} class="auth-form">
        {#if errorMsg}
          <div class="error-banner">{errorMsg}</div>
        {/if}

        <div class="input-group">
          <label for="work-email">Work Email</label>
          <input
            id="work-email"
            type="email"
            bind:value={email}
            placeholder="name@company.com"
            required
            disabled={isLoading}
          />
        </div>

        <button type="submit" class="btn-sso" disabled={isLoading || !email}>
          {#if isLoading}
            Authenticating...
          {:else}
            Continue with SSO
          {/if}
        </button>
      </form>
    </div>
  </div>
</div>

<style>
  .modal-card {
    width: 100%;
    max-width: 480px;
    background: rgba(20, 20, 28, 0.95);
    border: 1px solid rgba(255, 255, 255, 0.1);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
    border-radius: 12px;
    padding: 2.5rem 2rem;
    color: #f1f5f9;
  }

  .title {
    margin: 0 0 0.5rem 0;
    font-size: 1.5rem;
    font-weight: 600;
    text-align: center;
    color: #ffffff;
  }

  .subtitle {
    font-size: 0.9rem;
    color: #94a3b8;
    text-align: center;
    margin: 0 0 2rem 0;
  }

  .auth-form {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
  }

  .error-banner {
    color: #ff6b6b;
    background: rgba(255, 107, 107, 0.1);
    padding: 0.75rem;
    border-radius: 6px;
    font-size: 0.9rem;
    text-align: center;
    border: 1px solid rgba(255, 107, 107, 0.2);
  }

  .input-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .input-group label {
    font-size: 0.85rem;
    color: #cbd5e1;
    font-weight: 500;
  }

  .input-group input {
    padding: 0.75rem 1rem;
    border-radius: 6px;
    border: 1px solid #334155;
    background: #1e293b;
    color: white;
    font-size: 1rem;
    transition: all 0.2s;
  }

  .input-group input:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.25);
  }

  .input-group input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .btn-sso {
    background: #3b82f6;
    color: white;
    font-weight: 600;
    width: 100%;
    padding: 0.875rem;
    border-radius: 6px;
    border: none;
    cursor: pointer;
    font-size: 1rem;
    transition: background 0.2s;
    margin-top: 0.5rem;
  }

  .btn-sso:hover:not(:disabled) {
    background: #2563eb;
  }

  .btn-sso:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }
</style>
