<script lang="ts">
  import { authState, signIn, signOut } from "$lib/auth.svelte";
  import { reconnect } from "$lib/ws.svelte";
  import { invoke } from "@tauri-apps/api/core";
  import { onMount } from "svelte";
  import * as pdfjsLib from 'pdfjs-dist';

  pdfjsLib.GlobalWorkerOptions.workerSrc = new URL('pdfjs-dist/build/pdf.worker.min.mjs', import.meta.url).href;

  let email = $state("");
  let password = $state("");
  let signInError = $state<string | null>(null);
  let isSigningIn = $state(false);

  // New settings state
  let showSettings = $state(false);
  let backendUrl = $state(localStorage.getItem("backend_url") || "127.0.0.1:8765");
  let isDevModeChecked = $state(false);
  let preferredLanguage = $state(localStorage.getItem("preferred_language") || "");

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

  // Resume state
  let resumeRawText = $state("");
  let resumeStatus = $state("");
  let resumeFilename = $state("");

  onMount(async () => {
    try {
      const apiUrl = backendUrl.startsWith('http') ? backendUrl : `http://${backendUrl}`;
      const res = await fetch(`${apiUrl}/api/system_prompt`);
      if (res.ok) {
        const data = await res.json();
        selectedPrompt = data.prompt;
      }

      const audioRes = await fetch(`${apiUrl}/api/audio/devices`);
      if (audioRes.ok) {
        audioDevices = await audioRes.json();
      }
    } catch (e) {
      console.error("Failed to load initial settings", e);
    }

    try {
      const apiUrl = backendUrl.startsWith('http') ? backendUrl : `http://${backendUrl}`;
      const resContext = await fetch(`${apiUrl}/api/resume/context`);
      if (resContext.ok) {
        const data = await resContext.json();
        if (data.context) {
          resumeStatus = "✅ Profile active (from previous session). Note: stored server-side in memory and resets on backend restart.";
        }
      }
    } catch(e) {
      console.error("Failed to check resume context", e);
    }

    try {
      const apiUrl = backendUrl.startsWith('http') ? backendUrl : `http://${backendUrl}`;
      const resLang = await fetch(`${apiUrl}/api/language`);
      if (resLang.ok) {
        const data = await resLang.json();
        if (data.language) {
          preferredLanguage = data.language;
        }
      }
    } catch (e) {
      console.error("Failed to load language preference", e);
    }
  });

  async function handleFileSelect(e: Event) {
    const target = e.target as HTMLInputElement;
    if (!target.files || target.files.length === 0) return;
    const file = target.files[0];
    resumeFilename = file.name;
    resumeStatus = `📄 ${file.name} — Ready to extract`;
    resumeRawText = "";

    try {
      if (file.name.toLowerCase().endsWith(".txt")) {
        const text = await file.text();
        resumeRawText = text;
      } else if (file.name.toLowerCase().endsWith(".pdf")) {
        const arrayBuffer = await file.arrayBuffer();
        const pdf = await pdfjsLib.getDocument(arrayBuffer).promise;
        let text = "";
        for (let i = 1; i <= pdf.numPages; i++) {
          const page = await pdf.getPage(i);
          const content = await page.getTextContent();
          const strings = content.items.map((item: any) => item.str);
          text += strings.join(" ") + " ";
        }
        resumeRawText = text;
      }
    } catch (error) {
      console.error("Error reading file", error);
      resumeStatus = "❌ Error reading file content.";
      resumeRawText = "";
    }
  }

  async function extractResume() {
    if (!resumeRawText) return;
    resumeStatus = "⏳ Extracting candidate profile...";
    try {
      const apiUrl = backendUrl.startsWith('http') ? backendUrl : `http://${backendUrl}`;
      const res = await fetch(`${apiUrl}/api/resume/extract`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ text: resumeRawText })
      });
      if (res.ok) {
        const data = await res.json();
        resumeStatus = "✅ Profile extracted and active";
        if (data.detected_language && !preferredLanguage) {
          preferredLanguage = data.detected_language;
          localStorage.setItem("preferred_language", preferredLanguage);
          resumeStatus = `✅ Profile extracted. Language auto-detected: ${preferredLanguage}`;
        }
      } else {
        resumeStatus = "❌ Extraction failed — check backend";
      }
    } catch (e) {
      console.error("Resume extraction failed", e);
      resumeStatus = "❌ Extraction failed — network error";
    }
  }

  function toggleSettings() {
    showSettings = !showSettings;
  }

  async function handleSaveSettings() {
    localStorage.setItem("backend_url", backendUrl);
    localStorage.setItem("preferred_language", preferredLanguage);
    reconnect(backendUrl);
    try {
      await invoke("toggle_stealth", { enable: !isDevModeChecked });
    } catch (err) {
      console.error("Failed to toggle stealth", err);
    }

    try {
      const apiUrl = backendUrl.startsWith('http') ? backendUrl : `http://${backendUrl}`;
      await fetch(`${apiUrl}/api/language`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ language: preferredLanguage })
      });
    } catch (e) {
      console.error("Failed to save language preference", e);
    }

    try {
      const apiUrl = backendUrl.startsWith('http') ? backendUrl : `http://${backendUrl}`;
      await fetch(`${apiUrl}/api/system_prompt`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ prompt: selectedPrompt })
      });

      await fetch(`${apiUrl}/api/audio/device`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ device_id: selectedDeviceId, is_loopback: isLoopbackEnabled })
      });
    } catch (e) {
      console.error("Failed to save settings", e);
    }
  }

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

<div class="settings-wrapper">
  <button class="settings-toggle" onclick={toggleSettings} aria-label="Settings" data-testid="settings-btn">
    ⚙️
  </button>

  {#if showSettings}
  <div class="settings-panel">
  <div style="display: flex; justify-content: space-between; align-items: center;">
    <h2>Settings</h2>
  </div>

  <div class="config-section">
    <div class="input-group">
      <label for="backendUrl">Backend API URL</label>
      <input type="text" id="backendUrl" bind:value={backendUrl} placeholder="127.0.0.1:8765" data-testid="backend-url-input" />
    </div>

    <div class="checkbox-group">
      <label>
        <input type="checkbox" bind:checked={isDevModeChecked} data-testid="dev-mode-toggle" />
        Dev Mode: Disable Stealth (E2E Visibility)
      </label>
    </div>

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

    <div class="input-group">
      <label for="systemPrompt">Persona / System Prompt</label>
      <select id="systemPrompt" bind:value={selectedPrompt} class="custom-select">
        {#each systemPrompts as p}
          <option value={p.value}>{p.label}</option>
        {/each}
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

    <button class="btn-primary save-btn" onclick={handleSaveSettings} data-testid="settings-save-btn">Save & Reconnect</button>
  </div>

  <hr class="divider" />

  <div class="resume-section">
    <div class="section-label">Resume Context</div>
    <div class="input-group">
      <label for="resumeUpload">Upload Resume (PDF or TXT)</label>
      <input type="file" id="resumeUpload" accept=".pdf,.txt" onchange={handleFileSelect} />
    </div>
    {#if resumeStatus}
      <div class="status-indicator">{resumeStatus}</div>
    {/if}
    <button class="btn-primary extract-btn" disabled={!resumeRawText} onclick={extractResume}>
      Extract Profile
    </button>
  </div>

  <hr class="divider" />

  <h2>Account</h2>
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
  {/if}
</div>

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

  .settings-panel {
    position: absolute;
    bottom: calc(100% + 0.5rem);
    right: 0;
    background: rgba(30, 30, 30, 0.95);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 1.5rem;
    color: #fff;
    width: 320px;
    box-shadow: 0 4px 12px rgba(0,0,0,0.5);
    font-family: system-ui, -apple-system, sans-serif;
    z-index: 1000;
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

  .input-group input:focus, .custom-select:focus {
    outline: none;
    border-color: #007bff;
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

  .error-msg {
    color: #ff6b6b;
    background: rgba(255, 107, 107, 0.1);
    padding: 0.5rem;
    border-radius: 4px;
    font-size: 0.85rem;
  }

  .resume-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .section-label {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: #666;
  }

  .status-indicator {
    font-size: 0.85rem;
    color: #ddd;
    background: rgba(255, 255, 255, 0.05);
    padding: 0.5rem;
    border-radius: 4px;
    border: 1px solid #444;
  }

  .extract-btn {
    margin-top: 0.5rem;
  }
</style>
