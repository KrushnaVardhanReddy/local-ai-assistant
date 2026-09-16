<script lang="ts">
  import { onMount } from 'svelte';

  let provider = 'local';
  let groqApiKey = '';
  let opacity = 0.85;

  onMount(() => {
    // Load saved settings if any
    const savedKey = localStorage.getItem('GROQ_API_KEY');
    if (savedKey) {
      groqApiKey = savedKey;
      provider = 'cloud';
    }
    const savedOpacity = localStorage.getItem('PRESENTER_OPACITY');
    if (savedOpacity) {
      opacity = parseFloat(savedOpacity);
    }
  });

  function handleKeyChange() {
    if (groqApiKey) {
      localStorage.setItem('GROQ_API_KEY', groqApiKey);
    } else {
      localStorage.removeItem('GROQ_API_KEY');
    }
  }

  function handleOpacityChange() {
    localStorage.setItem('PRESENTER_OPACITY', opacity.toString());
    // In a full implementation, this might call a Wails runtime method to set native window opacity
    // For now we can dispatch a custom event or let a wrapper component handle updating the HUD css var
    document.documentElement.style.setProperty('--hud-bg', `rgba(10, 10, 10, ${opacity})`);
  }

  async function handleClearCache() {
    try {
      // Stub for clearing presenter_cache.db
      // Since we didn't expose ClearCache method on PresenterApp specifically yet,
      // this is just a UI scaffold.
      alert('Cache cleared (stub)');
    } catch (err) {
      console.error('Failed to clear cache:', err);
    }
  }
</script>

<div class="presenter-settings">
  <h2>Presenter Settings</h2>

  <div class="setting-group">
    <label>STT Provider:</label>
    <div class="radio-group">
      <label>
        <input type="radio" bind:group={provider} value="local" />
        Local (Whisper)
      </label>
      <label>
        <input type="radio" bind:group={provider} value="cloud" />
        Cloud (Groq)
      </label>
    </div>
  </div>

  {#if provider === 'cloud'}
    <div class="setting-group">
      <label for="groq-key">Groq API Key:</label>
      <input
        id="groq-key"
        type="password"
        bind:value={groqApiKey}
        on:input={handleKeyChange}
        placeholder="gsk_..."
      />
    </div>
  {/if}

  <div class="setting-group">
    <label for="opacity-slider">HUD Opacity: {opacity}</label>
    <input
      id="opacity-slider"
      type="range"
      min="0.5"
      max="1.0"
      step="0.05"
      bind:value={opacity}
      on:input={handleOpacityChange}
    />
  </div>

  <div class="setting-group actions">
    <button on:click={handleClearCache}>Clear Cache</button>
  </div>
</div>

<style>
  .presenter-settings {
    font-family: system-ui, -apple-system, sans-serif;
    padding: 24px;
    background: #f9f9f9;
    border-radius: 8px;
    max-width: 400px;
    color: #333;
  }

  h2 {
    margin-top: 0;
    margin-bottom: 24px;
    font-size: 1.5rem;
  }

  .setting-group {
    margin-bottom: 16px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  label {
    font-weight: 500;
  }

  .radio-group {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .radio-group label {
    font-weight: normal;
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
  }

  input[type="password"] {
    padding: 8px;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-family: inherit;
  }

  input[type="range"] {
    width: 100%;
  }

  button {
    padding: 8px 16px;
    background-color: #dc3545;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
  }

  button:hover {
    background-color: #c82333;
  }

  .actions {
    margin-top: 24px;
    align-items: flex-start;
  }
</style>
