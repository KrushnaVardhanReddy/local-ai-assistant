<script lang="ts">
  import { signIn, signUp } from "$lib/auth.svelte";
  import { onMount } from "svelte";

  let { isOpen = $bindable(), onClose } = $props<{ isOpen: boolean; onClose: () => void }>();

  let mode: "signin" | "signup" = $state("signin");
  let email = $state("");
  let password = $state("");
  let isLoading = $state(false);
  let errorMsg = $state<string | null>(null);
  let successMsg = $state<string | null>(null);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!email || !password) {
      errorMsg = "Please enter email and password.";
      return;
    }

    isLoading = true;
    errorMsg = null;
    successMsg = null;

    try {
      if (mode === "signin") {
        const { error } = await signIn(email, password);
        if (error) {
          errorMsg = error;
        } else {
          onClose();
        }
      } else {
        const { error } = await signUp(email, password);
        if (error) {
          errorMsg = error;
        } else {
          successMsg = "Registration successful! You can now sign in.";
          mode = "signin";
        }
      }
    } catch (err: any) {
      errorMsg = err.message || "An unexpected error occurred.";
    } finally {
      isLoading = false;
    }
  }

  function toggleMode() {
    mode = mode === "signin" ? "signup" : "signin";
    errorMsg = null;
    successMsg = null;
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
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
    >
      <button class="close-btn" onclick={onClose}>✖</button>
      <div class="auth-container">
        <h2>{mode === 'signin' ? 'Sign In' : 'Create Account'}</h2>

        <form onsubmit={handleSubmit} class="auth-form">
          {#if errorMsg}
            <div class="error-msg">{errorMsg}</div>
          {/if}
          {#if successMsg}
            <div class="success-msg">{successMsg}</div>
          {/if}

          <div class="input-group">
            <label for="auth-email">Email</label>
            <input
              type="email"
              id="auth-email"
              bind:value={email}
              placeholder="you@example.com"
              required
            />
          </div>

          <div class="input-group">
            <label for="auth-password">Password</label>
            <input
              type="password"
              id="auth-password"
              bind:value={password}
              placeholder="••••••••"
              required
            />
          </div>

          <button type="submit" class="btn-primary" disabled={isLoading}>
            {#if isLoading}
              {mode === 'signin' ? 'Signing In...' : 'Creating Account...'}
            {:else}
              {mode === 'signin' ? 'Sign In' : 'Create Account'}
            {/if}
          </button>
        </form>

        <div class="toggle-mode">
          <p>
            {mode === 'signin' ? "Don't have an account?" : 'Already have an account?'}
            <button class="btn-link" onclick={toggleMode}>
              {mode === 'signin' ? 'Sign Up' : 'Sign In'}
            </button>
          </p>
        </div>
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
    width: 90%;
    max-width: 400px;
    background: #1e1e1e;
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    padding: 2rem;
    box-shadow: 0 4px 12px rgba(0,0,0,0.5);
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
    font-size: 1.5rem;
    font-weight: 600;
    text-align: center;
  }

  .auth-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .input-group {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .input-group label {
    font-size: 0.8rem;
    color: #aaa;
  }

  .input-group input {
    padding: 0.75rem;
    border-radius: 4px;
    border: 1px solid #444;
    background: #222;
    color: white;
    font-size: 1rem;
  }

  .input-group input:focus {
    outline: none;
    border-color: #007bff;
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
    margin-top: 0.5rem;
  }

  .btn-primary:hover:not(:disabled) {
    background: #0056b3;
  }

  .btn-primary:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .error-msg {
    color: #ff6b6b;
    background: rgba(255, 107, 107, 0.1);
    padding: 0.5rem;
    border-radius: 4px;
    font-size: 0.85rem;
  }

  .success-msg {
    color: #28a745;
    background: rgba(40, 167, 69, 0.1);
    padding: 0.5rem;
    border-radius: 4px;
    font-size: 0.85rem;
  }

  .toggle-mode {
    text-align: center;
    font-size: 0.9rem;
    color: #aaa;
  }

  .btn-link {
    background: none;
    border: none;
    color: #007bff;
    cursor: pointer;
    font-size: 0.9rem;
    padding: 0;
    margin-left: 0.25rem;
    text-decoration: underline;
  }

  .btn-link:hover {
    color: #0056b3;
  }
</style>
