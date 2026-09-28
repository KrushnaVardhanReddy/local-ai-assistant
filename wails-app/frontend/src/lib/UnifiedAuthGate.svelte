<script lang="ts">
  import { activateLicense, checkEnterpriseDomain, claimEnterpriseSeat,
           supabase, authState } from "$lib/auth.svelte";
  import { BrowserOpenURL } from "../../wailsjs/runtime/runtime";

  // Panel states
  type Panel = "none" | "license" | "demo" | "enterprise";
  let activePanel = $state<Panel>("none");

  // License key panel
  let licenseKey = $state("");
  let licenseError = $state("");
  let licenseLoading = $state(false);

  // Demo panel — just a button, OAuth handled by Wails event
  let demoLoading = $state(false);

  // Enterprise panel
  let enterpriseEmail = $state(import.meta.env.VITE_DEV_ENTERPRISE_EMAIL ?? "");
  let enterpriseError = $state("");
  let enterpriseLoading = $state(false);
  let enterpriseOrg = $state<{ id: string; name: string } | null>(null);
  let enterpriseSSOStarted = $state(false);

  const isDevBypass = import.meta.env.VITE_DEV_ENTERPRISE_EMAIL &&
                      import.meta.env.VITE_DEV_ENTERPRISE_EMAIL !== "undefined" &&
                      !import.meta.env.PROD;

  async function handleLicenseSubmit() {
    licenseError = "";
    licenseLoading = true;
    try {
      await activateLicense(licenseKey.trim());
      // authState.licenseStatus will become "active" — App.svelte re-renders
    } catch (err: any) {
      licenseError = err.message ?? "Invalid license key. Please try again.";
    } finally {
      licenseLoading = false;
    }
  }

  async function handleStartDemo() {
    if (!supabase) return;
    demoLoading = true;
    try {
      const { data, error } = await supabase.auth.signInWithOAuth({
        provider: "google",
        options: {
          redirectTo: "barnowl://auth-callback",
          skipBrowserRedirect: true,
        },
      });
      if (error) throw error;
      if (data?.url) {
        // Open the OAuth URL in the user's default browser
        // The Wails `on_auth_complete` event handles the callback
        BrowserOpenURL(data.url);
      }
    } catch (err: any) {
      enterpriseError = err.message;
    } finally {
      demoLoading = false;
    }
  }

  async function handleEnterpriseCheck() {
    enterpriseError = "";
    enterpriseLoading = true;
    enterpriseOrg = null;
    try {
      const result = await checkEnterpriseDomain(enterpriseEmail.trim());
      if (!result.found) {
        enterpriseError = result.error ?? "No enterprise license found for this domain.";
        return;
      }
      enterpriseOrg = result.org!;
    } finally {
      enterpriseLoading = false;
    }
  }

  async function handleEnterpriseSSO() {
    if (!enterpriseOrg || !supabase) return;
    enterpriseSSOStarted = true;
    const { data, error } = await supabase.auth.signInWithOAuth({
      provider: "google",
      options: {
        redirectTo: "barnowl://auth-callback",
        skipBrowserRedirect: true,
        queryParams: { hd: enterpriseEmail.split("@")[1] }, // restrict to org domain
      },
    });
    if (error) { enterpriseError = error.message; enterpriseSSOStarted = false; return; }
    if (data?.url) BrowserOpenURL(data.url);

    // After on_auth_complete fires (handled in auth.svelte.ts initAuthEventListeners),
    // call claimEnterpriseSeat here via a one-time $effect watcher:
    const unwatch = $effect.root(() => {
      $effect(() => {
        if (authState.user && enterpriseOrg && authState.licenseStatus !== "enterprise") {
          claimEnterpriseSeat(enterpriseOrg.id).catch((err: any) => {
            enterpriseError = err.message;
            enterpriseSSOStarted = false;
          });
          unwatch(); // cleanup after first call
        }
      });
    });
  }
</script>

<div class="gate-overlay">
  <div class="gate-card">
    <div class="gate-logo">🦉</div>
    <h1 class="gate-title">BarnOwl AI</h1>
    <p class="gate-subtitle">Choose how to access</p>

    <!-- License Key Panel -->
    <div class="panel" class:active={activePanel === 'license'}>
      <button class="panel-header" onclick={() => activePanel = activePanel === 'license' ? 'none' : 'license'}>
        🔑 Enter License Key
        <span class="panel-badge">Lifetime Access</span>
      </button>
      {#if activePanel === 'license'}
        <div class="panel-body">
          <input bind:value={licenseKey} placeholder="XXXX-XXXX-XXXX-XXXX"
                 class="gate-input" onkeydown={(e) => e.key === 'Enter' && handleLicenseSubmit()} />
          {#if licenseError}<p class="gate-error">{licenseError}</p>{/if}
          <button onclick={handleLicenseSubmit} disabled={licenseLoading || !licenseKey.trim()} class="gate-btn gate-btn-primary">
            {licenseLoading ? 'Activating...' : 'Activate License'}
          </button>
        </div>
      {/if}
    </div>

    <!-- Demo Panel -->
    <div class="panel" class:active={activePanel === 'demo'}>
      <button class="panel-header" onclick={() => activePanel = activePanel === 'demo' ? 'none' : 'demo'}>
        ⏱️ Start Free Demo
        <span class="panel-badge panel-badge-gray">15 minutes</span>
      </button>
      {#if activePanel === 'demo'}
        <div class="panel-body">
          <p class="gate-hint">Sign in with Google to start a one-time 15-minute free trial. No credit card required.</p>
          <button onclick={handleStartDemo} disabled={demoLoading} class="gate-btn gate-btn-google">
            {demoLoading ? 'Opening browser...' : '🔐 Continue with Google'}
          </button>
          {#if demoLoading}
            <p class="gate-hint">Complete sign-in in your browser. This window will update automatically.</p>
          {/if}
        </div>
      {/if}
    </div>

    <!-- Enterprise SSO Panel -->
    <div class="panel" class:active={activePanel === 'enterprise'}>
      <button class="panel-header" onclick={() => activePanel = activePanel === 'enterprise' ? 'none' : 'enterprise'}>
        🏢 Company / Team Login
        <span class="panel-badge panel-badge-blue">Enterprise SSO</span>
      </button>
      {#if activePanel === 'enterprise'}
        <div class="panel-body">
          {#if isDevBypass}
            <div class="dev-badge">⚠️ DEV BYPASS ACTIVE — enterprise checks skipped</div>
          {/if}
          {#if !enterpriseOrg}
            <input bind:value={enterpriseEmail} placeholder="you@company.com" type="email"
                   class="gate-input" />
            {#if enterpriseError}<p class="gate-error">{enterpriseError}</p>{/if}
            <button onclick={handleEnterpriseCheck} disabled={enterpriseLoading || !enterpriseEmail.trim()} class="gate-btn gate-btn-primary">
              {enterpriseLoading ? 'Checking...' : 'Check Domain →'}
            </button>
          {:else}
            <div class="org-card">
              <p class="org-name">🏢 {enterpriseOrg.name}</p>
              <p class="gate-hint">License found. Sign in with your company Google account to claim a seat.</p>
            </div>
            {#if enterpriseError}<p class="gate-error">{enterpriseError}</p>{/if}
            <button onclick={handleEnterpriseSSO} disabled={enterpriseSSOStarted} class="gate-btn gate-btn-google">
              {enterpriseSSOStarted ? 'Waiting for browser sign-in...' : '🔐 Sign in with SSO'}
            </button>
            <button class="gate-btn gate-btn-ghost" onclick={() => { enterpriseOrg = null; enterpriseError = ''; }}>
              ← Use a different email
            </button>
          {/if}
        </div>
      {/if}
    </div>

  </div>
</div>

<style>
  .gate-overlay {
    position: fixed; inset: 0; z-index: 9999;
    background: #1a1a2e;
    display: flex; align-items: center; justify-content: center;
  }
  .gate-card {
    width: 420px; padding: 2.5rem;
    background: rgba(255,255,255,0.04);
    border: 1px solid rgba(255,255,255,0.08);
    border-radius: 20px;
    box-shadow: 0 24px 64px rgba(0,0,0,0.6);
    backdrop-filter: blur(20px);
    display: flex; flex-direction: column; gap: 0.5rem;
  }
  .gate-logo { font-size: 3rem; text-align: center; }
  .gate-title { font-size: 1.75rem; font-weight: 700; text-align: center; color: #fff; margin: 0; }
  .gate-subtitle { color: #888; text-align: center; font-size: 0.875rem; margin: 0 0 0.5rem; }
  .panel { border: 1px solid rgba(255,255,255,0.07); border-radius: 12px; overflow: hidden; }
  .panel.active { border-color: rgba(99,102,241,0.4); }
  .panel-header {
    width: 100%; background: rgba(255,255,255,0.03); border: none;
    padding: 0.85rem 1rem; display: flex; align-items: center; justify-content: space-between;
    color: #e5e7eb; font-size: 0.9rem; font-weight: 600; cursor: pointer;
    transition: background 0.2s;
  }
  .panel-header:hover { background: rgba(255,255,255,0.07); }
  .panel-badge {
    font-size: 0.7rem; padding: 2px 8px; border-radius: 999px;
    background: rgba(99,102,241,0.2); color: #a5b4fc; font-weight: 500;
  }
  .panel-badge-gray { background: rgba(255,255,255,0.08); color: #9ca3af; }
  .panel-badge-blue { background: rgba(59,130,246,0.2); color: #93c5fd; }
  .panel-body { padding: 1rem; display: flex; flex-direction: column; gap: 0.75rem; }
  .gate-input {
    width: 100%; padding: 0.65rem 0.85rem; border-radius: 8px;
    background: rgba(0,0,0,0.3); border: 1px solid rgba(255,255,255,0.12);
    color: #fff; font-size: 0.9rem; outline: none; box-sizing: border-box;
    transition: border-color 0.2s;
  }
  .gate-input:focus { border-color: rgba(99,102,241,0.6); }
  .gate-btn {
    width: 100%; padding: 0.75rem; border-radius: 8px; border: none;
    font-size: 0.9rem; font-weight: 600; cursor: pointer; transition: all 0.2s;
  }
  .gate-btn:disabled { opacity: 0.4; cursor: not-allowed; }
  .gate-btn-primary { background: #6366f1; color: #fff; }
  .gate-btn-primary:hover:not(:disabled) { background: #5254cc; }
  .gate-btn-google { background: rgba(255,255,255,0.1); color: #e5e7eb; border: 1px solid rgba(255,255,255,0.15); }
  .gate-btn-google:hover:not(:disabled) { background: rgba(255,255,255,0.16); }
  .gate-btn-ghost { background: transparent; color: #6b7280; font-size: 0.8rem; padding: 0.4rem; }
  .gate-btn-ghost:hover { color: #9ca3af; }
  .gate-error { color: #f87171; font-size: 0.8rem; margin: 0; }
  .gate-hint { color: #6b7280; font-size: 0.8rem; margin: 0; line-height: 1.5; }
  .org-card { background: rgba(99,102,241,0.08); border: 1px solid rgba(99,102,241,0.2); border-radius: 8px; padding: 0.75rem; }
  .org-name { color: #c7d2fe; font-weight: 600; margin: 0 0 0.25rem; }
  .dev-badge {
    background: rgba(251,191,36,0.12); border: 1px solid rgba(251,191,36,0.3);
    color: #fbbf24; font-size: 0.75rem; padding: 0.4rem 0.75rem; border-radius: 6px;
  }
</style>
