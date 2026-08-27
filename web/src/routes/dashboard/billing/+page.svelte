<script lang="ts">
  import { page } from '$app/stores';
  let { data } = $props();

  let plan = $derived(data.plan);
  let showSuccess = $derived($page.url.searchParams.get('success') === 'true');

  async function upgradeToPro() {
      const res = await fetch('/api/billing/checkout', { method: 'POST' });
      if (res.ok) {
          const { url } = await res.json();
          window.location.href = url;
      }
  }

  async function manageSubscription() {
      const res = await fetch('/api/billing/portal', { method: 'POST' });
      if (res.ok) {
          const { url } = await res.json();
          window.location.href = url;
      }
  }
</script>

<main class="content">
  <header class="header">
    <h1>Billing & Subscription</h1>
    {#if plan === 'pro'}
        <span class="badge pro">Pro</span>
    {:else}
        <span class="badge free">Free</span>
    {/if}
  </header>

  {#if showSuccess}
    <div class="banner success">
        🎉 Welcome to Pro!
    </div>
  {/if}

  <div class="card">
    {#if plan === 'pro'}
        <h2>Pro Plan Active</h2>
        <p>Thank you for subscribing to Pro.</p>
        {#if data.nextBillingDate}
            <p>Next billing date: {data.nextBillingDate}</p>
        {/if}
        <button onclick={manageSubscription} class="btn">Manage Subscription</button>
    {:else}
        <h2>Free Plan</h2>
        <p>Upgrade to Pro for advanced features.</p>
        <button onclick={upgradeToPro} class="btn btn-primary">Upgrade to Pro — $12/mo</button>
    {/if}
  </div>
</main>

<style>
  .content { flex-grow: 1; padding: 2rem 4rem; max-width: 1000px; }
  .header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 3rem; }
  .badge { padding: 0.25rem 0.75rem; border-radius: 99px; font-size: 0.875rem; font-weight: 600; }
  .badge.pro { background-color: rgba(124, 58, 237, 0.2); color: #a78bfa; border: 1px solid rgba(124, 58, 237, 0.5); }
  .badge.free { background-color: rgba(156, 163, 175, 0.2); color: #9ca3af; border: 1px solid rgba(156, 163, 175, 0.5); }
  .card { background-color: var(--surface); border: var(--border); border-radius: 8px; padding: 2rem; }
  .btn { padding: 0.75rem 1.5rem; border-radius: 6px; cursor: pointer; border: none; font-weight: 500; }
  .btn-primary { background-color: #7c3aed; color: white; }
  .banner.success { padding: 1rem; background-color: rgba(16, 185, 129, 0.1); color: #10b981; border: 1px solid #10b981; border-radius: 6px; margin-bottom: 2rem; }
</style>