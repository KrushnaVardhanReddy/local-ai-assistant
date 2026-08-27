<script lang="ts">
  import { authState, signIn, signOut } from "$lib/auth.svelte";

  let email = $state("");
  let password = $state("");
  let signInError = $state<string | null>(null);
  let isSigningIn = $state(false);

  async function handleSignIn(e: Event) {
    e.preventDefault();
    if (!email || !password) {
      signInError = "Please enter email and password.";
      return;
    }

    isSigningIn = true;
    signInError = null;
    const { error } = await signIn(email, password);
    if (error) {
      signInError = error;
    } else {
      email = "";
      password = "";
    }
    isSigningIn = false;
  }

  async function handleSignOut() {
    await signOut();
  }
</script>

<div class="settings-panel">
  <div style="display: flex; justify-content: space-between; align-items: center;">
    <h2>Account</h2>
    <button data-testid="settings-btn" aria-label="Settings" style="background: none; border: none; color: white; cursor: pointer; font-size: 1.2rem;">⚙</button>
  </div>
  <div class="content">
    <div style="margin-bottom: 1rem;">
      <label style="display: flex; align-items: center; gap: 0.5rem; font-size: 0.9rem;">
        <input type="checkbox" data-testid="dev-mode-toggle" />
        Developer Mode
      </label>
      <div style="margin-top: 0.5rem;">
        <input type="text" data-testid="backend-url-input" placeholder="Backend URL" style="width: 100%; padding: 0.5rem; border-radius: 4px; border: 1px solid #444; background: #222; color: white;" />
      </div>
    </div>
    {#if authState.authMode === "local"}
      <p class="local-mode-text">Running in local mode &mdash; no account needed.</p>
    {:else}
      {#if authState.user}
        <div class="account-info">
          <p class="email"><strong>{authState.user.email}</strong></p>
          <div class="badges">
            <span class="badge plan-badge">SaaS User</span>
          </div>
          <div class="actions">
            <a href="https://example.com/dashboard" target="_blank" rel="noopener noreferrer" class="btn-link">Open Dashboard</a>
            <button class="btn-secondary" onclick={handleSignOut}>Sign Out</button>
          </div>
        </div>
      {:else}
        <form class="signin-form" onsubmit={handleSignIn}>
          <p class="form-title">Sign In to Sync</p>
          {#if signInError}
            <div class="error-msg">{signInError}</div>
          {/if}
          <div class="input-group">
            <label for="email">Email</label>
            <input type="email" id="email" bind:value={email} placeholder="you@example.com" required />
          </div>
          <div class="input-group">
            <label for="password">Password</label>
            <input type="password" id="password" bind:value={password} placeholder="••••••••" required />
          </div>
          <button type="submit" class="btn-primary" disabled={isSigningIn}>
            {isSigningIn ? 'Signing In...' : 'Sign In'}
          </button>
        </form>
      {/if}
    {/if}
  </div>
</div>

<style>
  .settings-panel {
    background: rgba(30, 30, 30, 0.95);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 1.5rem;
    color: #fff;
    width: 320px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.5);
    font-family: system-ui, -apple-system, sans-serif;
  }

  h2 {
    margin: 0 0 1rem 0;
    font-size: 1.25rem;
    font-weight: 600;
  }

  .content {
    font-size: 0.9rem;
  }

  .local-mode-text {
    color: #aaa;
    margin: 0;
  }

  .account-info {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .email {
    margin: 0;
    font-size: 1rem;
  }

  .badges {
    display: flex;
  }

  .badge {
    background: #007bff;
    color: white;
    padding: 0.2rem 0.5rem;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .actions {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-top: 0.5rem;
  }

  .btn-link, .btn-secondary, .btn-primary {
    display: inline-block;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    text-align: center;
    text-decoration: none;
    font-size: 0.9rem;
    cursor: pointer;
    border: none;
    transition: background 0.2s;
  }

  .btn-link {
    background: transparent;
    color: #007bff;
    border: 1px solid #007bff;
  }

  .btn-link:hover {
    background: rgba(0, 123, 255, 0.1);
  }

  .btn-secondary {
    background: #444;
    color: #fff;
  }

  .btn-secondary:hover {
    background: #555;
  }

  .btn-primary {
    background: #007bff;
    color: white;
    font-weight: 600;
    width: 100%;
  }

  .btn-primary:hover:not(:disabled) {
    background: #0056b3;
  }

  .btn-primary:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .signin-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .form-title {
    margin: 0;
    font-weight: 500;
    color: #ddd;
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
    padding: 0.5rem;
    border-radius: 4px;
    border: 1px solid #444;
    background: #222;
    color: white;
  }

  .input-group input:focus {
    outline: none;
    border-color: #007bff;
  }

  .error-msg {
    color: #ff6b6b;
    background: rgba(255, 107, 107, 0.1);
    padding: 0.5rem;
    border-radius: 4px;
    font-size: 0.85rem;
  }
</style>
