<script lang="ts">
  let { data } = $props();
</script>

<svelte:head>
  <title>Dashboard - Local AI Assistant</title>
</svelte:head>

<main class="content">
  <header class="header">
    <h1>Welcome back, {data.user?.email || 'User'}</h1>
    <span class="badge pro">Pro Plan</span>
  </header>

  <div class="card">
    <h2>Quick Start</h2>
    <p>Connect your desktop app to the cloud dashboard.</p>

    <ol class="steps">
      <li>Open the Local AI Assistant desktop app.</li>
      <li>Go to <strong>Settings</strong> > <strong>Account</strong>.</li>
      <li>Click <strong>Sign In</strong> and authenticate with your account.</li>
      <li>Your API keys will now sync securely across all your devices.</li>
    </ol>
  </div>
</main>

<style>
  .content {
    flex-grow: 1;
    padding: 2rem 4rem;
    max-width: 1000px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 3rem;
  }

  .header h1 {
    font-size: 1.875rem;
    margin: 0;
  }

  .badge {
    padding: 0.25rem 0.75rem;
    border-radius: 99px;
    font-size: 0.875rem;
    font-weight: 600;
  }

  .badge.pro {
    background-color: rgba(124, 58, 237, 0.2);
    color: #a78bfa;
    border: 1px solid rgba(124, 58, 237, 0.5);
  }

  .card {
    background-color: var(--surface);
    border: var(--border);
    border-radius: 8px;
    padding: 2rem;
  }

  .card h2 {
    margin-top: 0;
    margin-bottom: 0.5rem;
  }

  .card p {
    color: var(--muted);
    margin-bottom: 2rem;
  }

  .steps {
    padding-left: 1.5rem;
    color: var(--text);
    line-height: 1.6;
  }

  .steps li {
    margin-bottom: 0.75rem;
  }
</style>
