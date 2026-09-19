<script lang="ts">
  import { apiFetch } from "./api";
  import { authState, signOut } from "$lib/auth.svelte";
  import AuthModal from "$lib/components/AuthModal.svelte";
  import { reconnect } from "$lib/ws.svelte";
    import { onMount } from "svelte";
    import StealthTerminal from "$lib/StealthTerminal.svelte";
  import { getApiUrl, getWsUrl } from "$lib/api";


  const isCloudBuild = import.meta.env.VITE_BUILD_FLAVOR === 'cloud';


  let { embedded = false } = $props<{ embedded?: boolean }>();

  // New settings state
  let internalShowSettings = $state(false);
  let showSettings = $derived(embedded || internalShowSettings);
  let isAuthModalOpen = $state(false);
  let selectedProvider = $state(localStorage.getItem("custom_provider") || "auto");
  let customApiKey = $state(localStorage.getItem("custom_api_key") || "");
  let openRouterModel = $state(localStorage.getItem("openrouter_model") || "anthropic/claude-3.5-sonnet:beta");
  let isDevModeChecked = $state(false);
  let preferredLanguage = $state(localStorage.getItem("preferred_language") || "");
  let interviewLanguage = $state(localStorage.getItem("interview_language") || "auto");
  let currentLLMProvider = $state("auto");
  let localSttEngine = $state(localStorage.getItem("local_stt_engine") || "faster-whisper");
  let includeActiveDocContext = $state(true);
  let activeTab = $state('models');
  const tabs = [
    { id: 'models', label: 'AI Models' },
    { id: 'audio', label: 'Audio & Devices' },
    { id: 'context', label: 'Context & Prompts' },
    { id: 'general', label: 'General' },
    { id: 'account', label: 'Account' }
  ];

  // Audio state
  let audioDevices = $state<Array<{id: number, name: string, is_loopback_capable: boolean}>>([]);
  let selectedDeviceId = $state<number | null>(null);
  let isLoopbackEnabled = $state(false);

  let systemPrompts = [
    { label: "Stealth Interview (Concise)", value: "You are a stealth interview assistant. The user is in a live technical interview. You must provide EXTREMELY concise answers. Use a maximum of 3 short bullet points. NEVER write long paragraphs. If code is needed, provide only the core snippet." },
    { label: "Pair Programmer (Detailed)", value: "You are an expert pair programmer. Provide detailed, step-by-step code implementations with explanations." },
    { label: "General Chat (Default)", value: "You are a helpful AI assistant." }
  ];
  let selectedPrompt = $state(systemPrompts[0].value);
  // System Status state
  let systemStatus = $state<any>(null);
  onMount(async () => {
    try {
      if ((window as any).go?.main?.App?.GetSystemStatus) {
        systemStatus = await (window as any).go.main.App.GetSystemStatus();
        currentLLMProvider = systemStatus.llm_provider || 'auto';
      } else {
        const apiUrl = getApiUrl();
        const healthRes = await apiFetch(`${apiUrl}/health`);
        if (healthRes.ok) {
          const healthData = await healthRes.json();
          currentLLMProvider = healthData.llm_provider;
        }
        const statusRes = await apiFetch(`${apiUrl}/api/status`);
        if (statusRes.ok) {
          systemStatus = await statusRes.json();
        }
      }
    } catch (e) { console.error("Failed to load health or status", e); }
    try {
      const apiUrl = getApiUrl();
      const res = await apiFetch(`${apiUrl}/api/system_prompt`);
      if (res.ok) {
        const data = await res.json();
        selectedPrompt = data.prompt;
      }

      if (!isCloudBuild && (window as any).go?.main?.App?.GetAudioDevices) {
        audioDevices = await (window as any).go.main.App.GetAudioDevices();
      } else {
        const audioRes = await apiFetch(`${apiUrl}/api/audio/devices`);
        if (audioRes.ok) {
          audioDevices = await audioRes.json();
        }
      }
    } catch (e) {
      console.error("Failed to load initial settings", e);
    }

    try {
      const apiUrl = getApiUrl();
      const resLang = await apiFetch(`${apiUrl}/api/language`);
      if (resLang.ok) {
        const data = await resLang.json();
        if (data.language) {
          preferredLanguage = data.language;
        }
      }

      const resInterviewLang = await apiFetch(`${apiUrl}/api/interview_language`);
      if (resInterviewLang.ok) {
        const data = await resInterviewLang.json();
        if (data.language) {
          interviewLanguage = data.language;
        }
      }

      const resEngine = await apiFetch(`${apiUrl}/api/stt_engine`);
      if (resEngine.ok) {
        const data = await resEngine.json();
        if (data.engine) {
          localSttEngine = data.engine;
        }
      }
    } catch (e) {
      console.error("Failed to load language preference or STT engine", e);
    }

    try {
      if ((window as any).go?.main?.App?.GetIDEState) {
        const state = await (window as any).go.main.App.GetIDEState();
        includeActiveDocContext = state.includeActiveDocContext;
      }
    } catch (e) {
      console.error("Failed to get initial doc context state", e);
    }
  });

  async function toggleSettings() {
    internalShowSettings = !internalShowSettings;
    
    // Fetch audio devices when opened, ensuring Wails bindings are fully loaded
    if (showSettings && !isCloudBuild && (window as any).go?.main?.App?.GetAudioDevices) {
      try {
        audioDevices = await (window as any).go.main.App.GetAudioDevices();
      } catch (e) {
        console.error("Failed to load audio devices on open", e);
      }
    }
  }

  async function handleSaveSettings() {
    localStorage.setItem("preferred_language", preferredLanguage);
    localStorage.setItem("interview_language", interviewLanguage);
    localStorage.setItem("local_stt_engine", localSttEngine);
    localStorage.setItem("custom_provider", selectedProvider);
    localStorage.setItem("custom_api_key", customApiKey);
    localStorage.setItem("openrouter_model", openRouterModel);
    reconnect(getWsUrl());
    try {
      await (window as any).go.main.App.ToggleStealth({ enable: !isDevModeChecked });
    } catch (err) {
      console.error("Failed to toggle stealth", err);
    }

    try {
      if ((window as any).go?.main?.App?.SetIncludeActiveDocContext) {
        await (window as any).go.main.App.SetIncludeActiveDocContext(includeActiveDocContext);
      }
    } catch (e) {
      console.error("Failed to set doc context state", e);
    }

    try {
      const apiUrl = getApiUrl();
      await apiFetch(`${apiUrl}/api/language`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ language: preferredLanguage })
      });
      await apiFetch(`${apiUrl}/api/interview_language`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ language: interviewLanguage })
      });
      await apiFetch(`${apiUrl}/api/stt_engine`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ engine: localSttEngine })
      });
    } catch (e) {
      console.error("Failed to save language preference or STT engine", e);
    }

    try {
      const apiUrl = getApiUrl();
      await apiFetch(`${apiUrl}/api/system_prompt`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt: selectedPrompt })
      });

      if (!isCloudBuild && (window as any).go?.main?.App?.SetAudioDevice) {
        if (selectedDeviceId !== null) {
          await (window as any).go.main.App.SetAudioDevice(selectedDeviceId, isLoopbackEnabled);
          const { wsState } = await import("$lib/ws.svelte");
          wsState.isListening = true;
        }
      } else {
        await apiFetch(`${apiUrl}/api/audio/device`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ device_id: selectedDeviceId, is_loopback: isLoopbackEnabled })
        });
      }
    } catch (e) {
      console.error("Failed to save settings", e);
    }
  }



  async function handleSignOut() {
    await signOut();
  }

  async function handleBuySessions() {
    if (!authState.user) return;
    try {
      const res = await apiFetch('/create-checkout-session', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ userId: authState.user.id })
      });
      if (!res.ok) throw new Error('Failed to create checkout session');
      const data = await res.json();
      if (data.url) {
        window.location.href = data.url;
      }
    } catch (e) {
      console.error(e);
      alert('Error initiating checkout. Please try again.');
    }
  }
</script>

<div class="settings-wrapper" class:embedded>
  {#if !embedded}
    <button class="settings-toggle" onclick={toggleSettings} aria-label="Settings" data-testid="settings-btn">
      ⚙️
    </button>
  {/if}

  {#if showSettings || embedded}
  <div class="settings-panel glass-panel flex flex-col pointer-events-auto" class:embedded>

    <!-- Header with Tabs -->
    <div class="flex-none flex flex-col border-b border-white/10">
      <div class="flex items-center justify-between p-4 pb-2">
        <h2 class="text-xl font-headline-md text-on-background tracking-wide m-0">Settings</h2>
        {#if !embedded}
        <button
            class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-white/10 text-on-surface-variant hover:text-red-400 transition-colors"
            onclick={() => internalShowSettings = false}
        >
            <span class="material-symbols-outlined text-[20px]">close</span>
        </button>
        {/if}
      </div>

      <div class="flex flex-wrap gap-y-1 px-2 scrollbar-hide">
        {#each tabs as tab}
          <button
            class="px-3 py-2 text-xs font-medium border-b-2 whitespace-nowrap transition-colors {activeTab === tab.id ? 'border-primary text-primary' : 'border-transparent text-on-surface-variant hover:text-on-background hover:border-white/20'}"
            onclick={() => activeTab = tab.id}
          >
            {tab.label}
          </button>
        {/each}
      </div>
    </div>

    <!-- Scrollable Content -->
    <div class="flex-1 overflow-y-auto p-5 pb-20 flex flex-col gap-6">

      {#if activeTab === 'models'}
        <div class="config-section">
          <div class="section-label">System Status</div>
          {#if systemStatus}
            <div class="status-grid">
              <div class="status-item">
                <span class="status-key">LLM Provider</span>
                <span class="status-value badge">{systemStatus.llm_provider}</span>
              </div>
              <div class="status-item">
                <span class="status-key">LLM Model</span>
                <span class="status-value badge">{systemStatus.llm_model}</span>
              </div>
              <div class="status-item">
                <span class="status-key">STT Provider</span>
                <span class="status-value badge">{systemStatus.stt_provider}</span>
              </div>
              <div class="status-item">
                <span class="status-key">STT Model</span>
                <span class="status-value badge">{systemStatus.stt_model}</span>
              </div>
              <div class="status-item">
                <span class="status-key">Local STT Engine</span>
                <span class="status-value badge">{systemStatus.local_stt_engine}</span>
              </div>
            </div>
          {:else}
            <div class="status-indicator">Loading status...</div>
          {/if}
        </div>

        <div class="config-section">
          {#if currentLLMProvider === "gemini"}
            <div class="gemini-live-badge">
              <span class="material-symbols-outlined">bolt</span>
              <div>
                <strong>Gemini Live Mode</strong>
                <p>Audio is processed directly by Gemini — no separate STT provider needed.</p>
              </div>
            </div>
          {/if}

          <div class="input-group">
            <label for="selectedProvider">Custom LLM Provider (Lifetime Tier Only)</label>
            <select id="selectedProvider" bind:value={selectedProvider} class="custom-select" disabled={authState.plan !== 'lifetime'}>
              <option value="auto">Auto (Default)</option>
              <option value="openai">OpenAI</option>
              <option value="anthropic">Anthropic</option>
              <option value="gemini">Gemini</option>
              <option value="groq">Groq</option>
              <option value="openrouter">OpenRouter</option>
            </select>
          </div>

          {#if selectedProvider !== 'auto'}
            <div class="input-group">
              <label for="customApiKey">Custom API Key</label>
              <input
                type="password"
                id="customApiKey"
                bind:value={customApiKey}
                placeholder="sk-..."
                disabled={authState.plan !== 'lifetime'}
              />
              <span class="text-xs text-gray-500 mt-1">Key is stored locally and never saved to our database.</span>
            </div>
          {/if}

          {#if selectedProvider === 'openrouter'}
            <div class="input-group">
              <label for="openRouterModel">OpenRouter Model ID</label>
              <input
                type="text"
                id="openRouterModel"
                bind:value={openRouterModel}
                placeholder="anthropic/claude-3.5-sonnet:beta"
                disabled={authState.plan !== 'lifetime'}
              />
            </div>
          {/if}

          {#if authState.plan !== 'lifetime'}
            <span class="text-xs text-red-400 mt-1">Custom models require a Lifetime Subscription.</span>
          {/if}

          {#if !isCloudBuild}
            <div class="input-group">
              <label for="localSttEngine">Local STT Engine</label>
              <select id="localSttEngine" bind:value={localSttEngine} class="custom-select">
                <option value="faster-whisper">Faster-Whisper (Universal)</option>
                <option value="parakeet">Nvidia Parakeet (RTX GPUs only)</option>
              </select>
            </div>
          {/if}
        </div>
      {/if}

      {#if activeTab === 'audio'}
        <div class="config-section">
          {#if !isCloudBuild}
            <div class="input-group">
              <label for="audioDevice">Audio Input Device</label>
              <select id="audioDevice" bind:value={selectedDeviceId} class="custom-select">
                <option value={null}>Default Microphone</option>
                {#each audioDevices as dev}
                  <option value={dev.id}>{dev.name}</option>
                {/each}
              </select>
            </div>

            {#if audioDevices.find(d => d.id === selectedDeviceId)?.is_loopback_capable}
              <div class="checkbox-group">
                <label>
                  <input type="checkbox" bind:checked={isLoopbackEnabled} />
                  Enable System Audio Loopback (Windows WASAPI only)
                </label>
              </div>
            {/if}
          {:else}
             <p class="text-sm text-gray-400">Audio devices are managed by your browser in Cloud mode.</p>
          {/if}
        </div>
      {/if}

      {#if activeTab === 'context'}
        <div class="config-section">
          <div class="input-group">
            <label for="systemPrompt">Persona / System Prompt</label>
            <select id="systemPrompt" bind:value={selectedPrompt} class="custom-select">
              {#each systemPrompts as p}
                <option value={p.value}>{p.label}</option>
              {/each}
            </select>
          </div>

          <div class="input-group">
            <label for="interviewLanguage">Interview Language (STT & LLM Override)</label>
            <select id="interviewLanguage" bind:value={interviewLanguage} class="custom-select">
              <option value="auto">Auto-Detect</option>
              <option value="en">English (en)</option>
              <option value="es">Spanish (es)</option>
              <option value="fr">French (fr)</option>
              <option value="de">German (de)</option>
              <option value="hi">Hindi (hi)</option>
              <option value="zh">Mandarin (zh)</option>
            </select>
          </div>

          <div class="input-group">
            <label for="preferredLanguage">Code Language Preference</label>
            <select id="preferredLanguage" bind:value={preferredLanguage} class="custom-select">
              <option value="">Auto (Let LLM Decide)</option>
              <option value="Python">Python</option>
              <option value="JavaScript">JavaScript</option>
              <option value="TypeScript">TypeScript</option>
              <option value="Java">Java</option>
              <option value="C++">C++</option>
              <option value="C#">C#</option>
              <option value="Go">Go</option>
              <option value="Rust">Rust</option>
              <option value="Ruby">Ruby</option>
              <option value="Swift">Swift</option>
              <option value="Kotlin">Kotlin</option>
              <option value="SQL">SQL</option>
              <option value="PHP">PHP</option>
              <option value="Scala">Scala</option>
            </select>
            <span style="font-size: 0.72rem; color: #666; margin-top: 2px;">
              Auto-filled from resume if detected
            </span>
          </div>
        </div>
      {/if}

      {#if activeTab === 'general'}
        <div class="config-section">
          {#if !isCloudBuild}
            <div style="margin-bottom: 1rem;">
              <h3 style="font-size: 0.9rem; margin: 0 0 0.5rem 0; color: #ddd;">Embedded Stealth Terminal</h3>
              <StealthTerminal />
            </div>
            <hr class="divider" style="margin-top: 0;" />
          {/if}

          <div class="checkbox-group">
            <label>
              <input type="checkbox" bind:checked={isDevModeChecked} data-testid="dev-mode-toggle" />
              Dev Mode: Disable Stealth (E2E Visibility)
            </label>
          </div>

          <div class="checkbox-group">
            <label>
              <input type="checkbox" bind:checked={includeActiveDocContext} />
              Include Active File as Context
            </label>
          </div>
        </div>
      {/if}

      {#if activeTab === 'account'}
        <div class="content">
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
                  <button class="btn-primary" style="background-color: #28a745; margin-bottom: 0.5rem;" onclick={handleBuySessions}>Buy 5 Interviews for $10</button>
                  <a href="https://example.com/dashboard" target="_blank" rel="noopener noreferrer" class="btn-link">Open Dashboard</a>
                  <button class="btn-secondary" onclick={handleSignOut}>Sign Out</button>
                </div>
              </div>
            {:else}
              <div class="signin-form">
                <p class="form-title">Sign In to Sync</p>
                <button class="btn-primary" onclick={() => isAuthModalOpen = true}>Sign In / Register</button>
              </div>
            {/if}
          {/if}
        </div>
      {/if}

    </div>

    <!-- Footer Save Button (Hidden on Account Tab) -->
    {#if activeTab !== 'account'}
      <div class="flex-none p-4 border-t border-white/10 bg-surface-variant/50">
         <button class="btn-primary save-btn w-full !mt-0" onclick={handleSaveSettings} data-testid="settings-save-btn">Save & Reconnect</button>
      </div>
    {/if}
  </div>
  {/if}
</div>

<AuthModal bind:isOpen={isAuthModalOpen} onClose={() => isAuthModalOpen = false} />

<style>
  .settings-wrapper {
    position: relative;
  }

  .settings-toggle {
    background: none;
    border: none;
    font-size: 1.25rem;
    cursor: pointer;
    color: #fff;
    opacity: 0.7;
    transition: opacity 0.2s;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
  }

  .settings-toggle:hover {
    opacity: 1;
  }


  .glass-panel {
      background: rgba(10, 10, 14, 0.88);
      backdrop-filter: blur(24px) saturate(1.4);
      -webkit-backdrop-filter: blur(24px) saturate(1.4);
      border: 1px solid rgba(255, 255, 255, 0.10);
      box-shadow: 0 24px 48px rgba(0, 0, 0, 0.7), inset 0 1px 0 rgba(255,255,255,0.06);
  }

  /* Hide scrollbar for tab list */
  .scrollbar-hide::-webkit-scrollbar {
      display: none;
  }
  .scrollbar-hide {
      -ms-overflow-style: none;
      scrollbar-width: none;
  }

  .settings-panel {
    position: absolute;
    bottom: calc(100% + 0.5rem);
    right: 0;


    border-radius: 8px;
    padding: 0;
    color: #fff;
    width: 320px;
    max-height: 80vh;
    overflow-y: auto;

    font-family: system-ui, -apple-system, sans-serif;
    z-index: 1000;
  }

  .settings-wrapper.embedded {
    width: 100%;
    height: 100%;
  }

  .settings-panel.embedded {
    position: static;
    width: 100%;
    max-height: none;
    background: transparent;
    border: none;
    border-radius: 0;
    box-shadow: none;
  }

  .hotkeys-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    padding: 0.75rem 1rem;
  }

  .hotkey-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.35rem 0;
    font-size: 0.8rem;
    color: rgba(255, 255, 255, 0.8);
    border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  }

  .hotkey-row:last-child {
    border-bottom: none;
  }

  .hotkey-row kbd {
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 4px;
    padding: 0.15rem 0.4rem;
    font-family: monospace;
    font-size: 0.75rem;
    color: #4ade80;
  }

  .config-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .checkbox-group label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.85rem;
    color: #ddd;
    cursor: pointer;
  }

  .save-btn {
    margin-top: 0.5rem;
  }

  .divider {
    border: 0;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    margin: 1.5rem 0;
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

  .input-group input:focus, .input-group textarea:focus, .custom-select:focus {
    outline: none;
    border-color: #007bff;
  }

  .input-group textarea {
    padding: 0.5rem;
    border-radius: 4px;
    border: 1px solid #444;
    background: #222;
    color: white;
    resize: vertical;
    font-family: inherit;
  }

  .custom-select {
    padding: 0.5rem;
    border-radius: 4px;
    border: 1px solid #444;
    background: #222;
    color: white;
    font-size: 0.85rem;
    cursor: pointer;
  }
  .section-label {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: #666;
  }

  .status-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 0.5rem;
    background: rgba(0, 0, 0, 0.2);
    padding: 1rem;
    border-radius: 8px;
    border: 1px solid rgba(255, 255, 255, 0.05);
  }

  .status-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.85rem;
  }

  .status-key {
    color: #aaa;
  }

  .status-value {
    color: #fff;
    font-weight: 500;
  }

  .gemini-live-badge {

    display: flex;
    align-items: flex-start;
    gap: 10px;
    background: rgba(99, 179, 237, 0.08);
    border: 1px solid rgba(99, 179, 237, 0.2);
    border-radius: 10px;
    padding: 12px 16px;
    color: rgba(255,255,255,0.7);
    font-size: 13px;
    margin-bottom: 1rem;
  }
  .gemini-live-badge strong { color: #63b3ed; display: block; margin-bottom: 2px; }
  .gemini-live-badge p { margin: 0; opacity: 0.7; font-size: 12px; }

</style>
