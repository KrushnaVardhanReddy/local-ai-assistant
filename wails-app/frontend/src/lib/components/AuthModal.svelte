<script lang="ts">
  import { authState, activateLicense } from "$lib/auth.svelte";
  import { onMount } from "svelte";
  import { BrowserOpenURL } from "../../wailsjs/runtime/runtime";

  let { isOpen = $bindable(), onClose } = $props<{ isOpen: boolean; onClose: () => void }>();

  let licenseKey = $state("");
  let isLoading = $state(false);
  let errorMsg = $state<string | null>(null);
  let successMsg = $state<string | null>(null);

  async function handleActivateLicense(e: Event) {
    e.preventDefault();
    if (!licenseKey) {
      errorMsg = "Please enter a valid license key.";
      return;
    }

    isLoading = true;
    errorMsg = null;
    successMsg = null;

    try {
      await activateLicense(licenseKey);
      successMsg = "License activated successfully!";
      setTimeout(() => {
        onClose();
      }, 1500);
    } catch (err: any) {
      errorMsg = err.message || "Failed to activate license.";
    } finally {
      isLoading = false;
    }
  }

  async function handleGoogleOAuth() {
    isLoading = true;
    errorMsg = null;
    try {
      await (window as any).go.main.App.StartOAuthFlow("google");
      // The on_auth_complete listener in auth.svelte.ts handles the rest
      // We must reset isLoading here so the button isn't permanently stuck
      // if the auth finishes but the user's demo is expired.
      isLoading = false;
    } catch (err: any) {
      errorMsg = err.message || "Failed to start Google OAuth flow.";
      isLoading = false;
    }
  }

</script>

{#if isOpen}
  <div
    class="modal-overlay pointer-events-auto"
    role="button"
    tabindex="0"
    aria-label="Close auth modal"
    onclick={onClose}
    onkeydown={(e) => e.key === 'Escape' && onClose()}
  >
    <div
      class="modal-content"
      role="dialog"
      aria-modal="true"
      tabindex="-1"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
    >
      <button class="close-btn" onclick={onClose}>✖</button>

      <div class="auth-container">
        {#if errorMsg}
          <div class="error-msg">{errorMsg}</div>
        {/if}
        {#if authState.licenseStatus === "expired"}
          <div class="error-msg">Your 15-minute free demo has expired. Please enter a lifetime license to continue using BarnOwl AI.</div>
        {/if}
        {#if successMsg}
          <div class="success-msg">{successMsg}</div>
        {/if}

        {#if authState.productMode === "interview"}
          <!-- MODE A: BarnOwl AI -->
          <div class="split-layout">
            <div class="panel demo-panel">
              <h2>Start 15-Min Free Demo</h2>
              <p class="subtext">Try all features, no credit card required.</p>
              <button class="btn-google" onclick={handleGoogleOAuth} disabled={isLoading || authState.licenseStatus === "expired"}>
                {#if isLoading}
                  Loading...
                {:else if authState.licenseStatus === "expired"}
                  Demo Expired
                {:else}
                  Continue with Google
                {/if}
              </button>
            </div>

            <div class="panel license-panel">
              <h2>Enter Lifetime License</h2>
              <p class="subtext">Already purchased? Enter your key to unlock forever.</p>

              <form onsubmit={handleActivateLicense} class="auth-form">
                <div class="input-group">
                  <input
                    type="text"
                    bind:value={licenseKey}
                    placeholder="XXXX-XXXX-XXXX-XXXX"
                    required
                  />
                </div>
                <button type="submit" class="btn-primary" disabled={isLoading || !licenseKey}>
                  {#if isLoading}
                    Activating...
                  {:else}
                    Activate License
                  {/if}
                </button>
              </form>
            </div>
          </div>

          <div class="footer-link">
            <a href="#" onclick={(e) => { e.preventDefault(); BrowserOpenURL("https://store.parakeet.app"); }}>Buy a Lifetime License</a>
          </div>

        {:else}
          <!-- MODE B: SaaS Products -->
          <h2>Sign in to {authState.productMode.charAt(0).toUpperCase() + authState.productMode.slice(1)}</h2>
          <div class="saas-panel">
            <button class="btn-google saas-btn" onclick={handleGoogleOAuth} disabled={isLoading}>
              {#if isLoading}
                Signing In...
              {:else}
                Continue with Google
              {/if}
            </button>
          </div>
        {/if}

      </div>
    </div>
  </div>
{/if}

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
    font-size: 1.3rem;
    font-weight: 600;
    text-align: center;
    color: #f1f5f9;
  }

  .subtext {
    font-size: 0.85rem;
    color: #94a3b8;
    text-align: center;
    margin: 0.5rem 0 1rem 0;
  }

  .split-layout {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  @media (min-width: 500px) {
    .split-layout {
      flex-direction: row;
      gap: 1.5rem;
    }
  }

  .panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    background: rgba(255, 255, 255, 0.03);
    border-radius: 8px;
    padding: 1.5rem;
    border: 1px solid rgba(255, 255, 255, 0.05);
  }

  .saas-panel {
    display: flex;
    justify-content: center;
    padding: 2rem 0;
  }

  .auth-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-top: auto;
  }

  .input-group {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .input-group input {
    padding: 0.75rem;
    border-radius: 4px;
    border: 1px solid #444;
    background: #222;
    color: white;
    font-size: 0.9rem;
    text-align: center;
    letter-spacing: 1px;
  }

  .input-group input:focus {
    outline: none;
    border-color: #007bff;
    box-shadow: 0 0 0 2px rgba(0,123,255,0.25);
  }

  .btn-primary {
    background: #007bff;
    color: white;
    font-weight: 600;
    width: 100%;
    padding: 0.75rem;
    border-radius: 4px;
    border: none;
    cursor: pointer;
    font-size: 1rem;
    transition: background 0.2s;
  }

  .btn-primary:hover:not(:disabled) {
    background: #0056b3;
  }

  .btn-primary:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .btn-google {
    background: white;
    color: #333;
    font-weight: 600;
    width: 100%;
    padding: 0.75rem;
    border-radius: 4px;
    border: none;
    cursor: pointer;
    font-size: 1rem;
    transition: background 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: auto;
  }

  .btn-google:hover:not(:disabled) {
    background: #f1f5f9;
  }

  .btn-google:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .saas-btn {
    max-width: 300px;
  }

  .error-msg {
    color: #ff6b6b;
    background: rgba(255, 107, 107, 0.1);
    padding: 0.75rem;
    border-radius: 4px;
    font-size: 0.9rem;
    text-align: center;
  }

  .success-msg {
    color: #28a745;
    background: rgba(40, 167, 69, 0.1);
    padding: 0.75rem;
    border-radius: 4px;
    font-size: 0.9rem;
    text-align: center;
  }

  .footer-link {
    text-align: center;
    margin-top: 0.5rem;
  }

  .footer-link a {
    color: #94a3b8;
    font-size: 0.9rem;
    text-decoration: underline;
    transition: color 0.2s;
  }

  .footer-link a:hover {
    color: #e2e8f0;
  }
</style>
