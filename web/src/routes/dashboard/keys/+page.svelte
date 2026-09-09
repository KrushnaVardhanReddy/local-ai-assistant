<script lang="ts">
  let providers = $state([
    { name: 'OpenAI', icon: 'O', key: 'sk-...abcd' },
    { name: 'Groq', icon: 'G', key: 'gsk-...efgh' },
    { name: 'Gemini', icon: 'G', key: null },
    { name: 'Anthropic', icon: 'A', key: null },
    { name: 'OpenRouter', icon: 'O', key: null }
  ]);

  async function testConnection(provider: string) {
    console.log(`Testing connection for ${provider}`);
    try {
      const response = await fetch(`/api/keys/${provider}/test`, {
        method: 'POST',
      });
      if (response.ok) {
        console.log(`Successfully connected to ${provider}`);
      } else {
        console.error(`Failed to connect to ${provider}`);
      }
    } catch (e) {
      console.error(`Error testing connection for ${provider}:`, e);
    }
  }
</script>

<svelte:head>
  <title>API Keys - Local AI Assistant</title>
</svelte:head>

<main class="content">
  <header class="header">
    <h1>API Keys</h1>
    <p class="subtitle">Manage your cloud provider keys. Keys are encrypted before being saved.</p>
  </header>

  <div class="keys-list">
    {#each providers as provider}
      <div class="key-row">
        <div class="provider-info">
          <div class="provider-icon">{provider.icon}</div>
          <div>
            <div class="provider-name">{provider.name}</div>
            <div class="key-status">
              {#if provider.key}
                <span class="status-indicator active"></span> Saved ({provider.key})
              {:else}
                <span class="status-indicator inactive"></span> Not configured
              {/if}
            </div>
          </div>
        </div>

        <div class="actions">
          {#if provider.key}
            <button class="btn btn-outline" onclick={() => testConnection(provider.name)}>Test</button>
            <button class="btn btn-outline">Update</button>
            <button class="btn btn-danger">Delete</button>
          {:else}
            <button class="btn btn-primary">Add Key</button>
          {/if}
        </div>
      </div>
    {/each}
  </div>
</main>

<style>
  .content {
    flex-grow: 1;
    padding: 2rem 4rem;
    max-width: 1000px;
  }

  .header {
    margin-bottom: 2rem;
  }

  .header h1 {
    font-size: 1.875rem;
    margin: 0 0 0.5rem 0;
  }

  .subtitle {
    color: var(--muted);
    margin: 0;
  }

  .keys-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .key-row {
    background-color: var(--surface);
    border: var(--border);
    border-radius: 8px;
    padding: 1.5rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .provider-info {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .provider-icon {
    width: 40px;
    height: 40px;
    background-color: rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    font-size: 1.25rem;
    color: var(--text);
  }

  .provider-name {
    font-weight: 600;
    font-size: 1.1rem;
    margin-bottom: 0.25rem;
  }

  .key-status {
    color: var(--muted);
    font-size: 0.875rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-family: monospace;
  }

  .status-indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    display: inline-block;
  }

  .status-indicator.active {
    background-color: #10b981;
  }

  .status-indicator.inactive {
    background-color: #64748b;
  }

  .actions {
    display: flex;
    gap: 0.5rem;
  }

  .btn {
    padding: 0.5rem 1rem;
    border-radius: 4px;
    font-weight: 500;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s;
    border: none;
  }

  .btn-primary {
    background-color: var(--accent);
    color: white;
  }

  .btn-primary:hover {
    background-color: #6d28d9;
  }

  .btn-outline {
    background-color: transparent;
    border: 1px solid var(--muted);
    color: var(--text);
  }

  .btn-outline:hover {
    border-color: var(--text);
  }

  .btn-danger {
    background-color: transparent;
    border: 1px solid rgba(239, 68, 68, 0.5);
    color: #ef4444;
  }

  .btn-danger:hover {
    background-color: rgba(239, 68, 68, 0.1);
  }
</style>
