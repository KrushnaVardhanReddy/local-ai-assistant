<script lang="ts">
  import { page } from '$app/stores';

  let { data, children } = $props();
</script>

<div class="dashboard-layout">
  <aside class="sidebar">
    <div class="logo">✦ BarnOwl</div>
    <nav class="nav-menu">
      <a href="/dashboard" class="nav-item {($page.url.pathname as string) === '/dashboard' ? 'active' : ''}">Overview</a>
      <a href="/dashboard/keys" class="nav-item {($page.url.pathname as string) === '/dashboard/keys' ? 'active' : ''}">API Keys</a>
      <a href="/dashboard/billing" class="nav-item {($page.url.pathname as string) === '/dashboard/billing' ? 'active' : ''}">Billing</a>
      <a href="/dashboard/settings" class="nav-item {($page.url.pathname as string) === '/dashboard/settings' ? 'active' : ''}">Settings</a>
      {#if data.is_org_admin}
        <a href="/dashboard/admin" class="nav-item {($page.url.pathname as string) === '/dashboard/admin' ? 'active' : ''}">Admin</a>
      {/if}
    </nav>
  </aside>

  {@render children()}
</div>

<style>
  .dashboard-layout {
    display: flex;
    min-height: 100vh;
  }

  .sidebar {
    width: 250px;
    background-color: var(--surface);
    border-right: var(--border);
    padding: 1.5rem;
    position: sticky;
    top: 0;
    height: 100vh;
  }

  .logo {
    font-weight: 700;
    font-size: 1.25rem;
    color: var(--text);
    margin-bottom: 2rem;
  }

  .nav-menu {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .nav-item {
    padding: 0.75rem 1rem;
    border-radius: 4px;
    color: var(--muted);
    text-decoration: none;
    transition: all 0.2s;
  }

  .nav-item:hover {
    background-color: rgba(255, 255, 255, 0.05);
    color: var(--text);
  }

  .nav-item.active {
    background-color: rgba(124, 58, 237, 0.1);
    color: var(--accent);
    font-weight: 500;
  }
</style>
