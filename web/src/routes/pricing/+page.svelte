<script lang="ts">
  import { env } from '$env/dynamic/public';
  import { page } from '$app/stores';
</script>

<svelte:head>
  <title>Pricing - BarnOwl</title>
</svelte:head>

<nav class="navbar">
  <a href="/" class="logo">✦ BarnOwl</a>
  <div class="nav-links">
    <a href="/download" class="btn btn-secondary">Download</a>
    <a href="/login" class="btn btn-primary">Sign In</a>
  </div>
</nav>

<main>
  <div class="header">
    <h1>Simple, transparent pricing</h1>
    <p>Run locally for free, or upgrade for cloud sync and advanced features.</p>
  </div>

  <div class="pricing-grid">
    <!-- Demo Tier -->
    <div class="pricing-card">
      <h2>Demo</h2>
      <div class="price">Free</div>
      <p class="description">No account needed, uses referral link.</p>

      <ul class="feature-list">
        <li>✓ 3 questions</li>
        <li>✓ or 15 min limit</li>
      </ul>

      <a href="/demo" class="btn btn-secondary btn-full">Try Demo</a>
    </div>

    <!-- Pay-as-you-go Tier -->
    <div class="pricing-card">
      <h2>Pay-as-you-go</h2>
      <div class="price">$5<span>/session</span></div>
      <p class="description">Perfect for occasional use.</p>

      <ul class="feature-list">
        <li>✓ 1 session = 90 min</li>
        <li>✓ Session token expires after 90 min</li>
      </ul>

      <a href="/checkout/pay-as-you-go" class="btn btn-secondary btn-full">Buy Session</a>
    </div>

    <!-- Monthly Tier -->
    <div class="pricing-card pro">
      <div class="popular-badge">Recommended</div>
      <h2>Monthly</h2>
      <div class="price">$19<span>/mo</span></div>
      <p class="description">For regular users.</p>

      <ul class="feature-list">
        <li>✓ Unlimited sessions</li>
        <li>✓ Cancel anytime</li>
      </ul>

      <a href="/checkout/monthly" class="btn btn-primary btn-full">Subscribe</a>
    </div>

    <!-- Founding Member Tier -->
    <div class="pricing-card">
      <h2>Founding Member</h2>
      <div class="price">$49<span>/one-time</span></div>
      <p class="description">Early adopter deal.</p>

      <ul class="feature-list">
        <li>✓ Unlimited sessions</li>
        <li>✓ Valid while supported</li>
      </ul>

      <a href="/checkout/founding-member" class="btn btn-secondary btn-full">Become a Founder</a>
    </div>
  </div>
</main>

<style>
  .navbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem 2rem;
    border-bottom: var(--border);
  }

  .logo {
    font-weight: 700;
    font-size: 1.25rem;
    color: var(--text);
    text-decoration: none;
  }

  .nav-links {
    display: flex;
    gap: 1rem;
    align-items: center;
  }

  .btn {
    padding: 0.5rem 1rem;
    border-radius: 4px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    text-align: center;
    text-decoration: none;
    display: inline-block;
  }

  .btn-primary {
    background-color: var(--accent);
    color: white;
    border: none;
  }

  .btn-primary:hover {
    background-color: #6d28d9;
  }

  .btn-secondary {
    background: transparent;
    border: 1px solid var(--muted);
    color: var(--text);
  }

  .btn-secondary:hover {
    border-color: var(--text);
  }

  .btn-full {
    width: 100%;
    padding: 0.75rem;
    margin-top: auto;
  }

  main {
    max-width: 1200px;
    margin: 0 auto;
    padding: 4rem 2rem;
  }

  .header {
    text-align: center;
    margin-bottom: 4rem;
  }

  .header h1 {
    font-size: 3rem;
    margin-bottom: 1rem;
  }

  .header p {
    font-size: 1.25rem;
    color: var(--muted);
  }

  .pricing-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 2rem;
  }

  @media (min-width: 768px) {
    .pricing-grid {
      grid-template-columns: 1fr 1fr;
    }
  }

  @media (min-width: 1024px) {
    .pricing-grid {
      grid-template-columns: 1fr 1fr 1fr 1fr;
    }
  }

  .pricing-card {
    background-color: var(--surface);
    border: var(--border);
    border-radius: 12px;
    padding: 2.5rem;
    display: flex;
    flex-direction: column;
    position: relative;
  }

  .pricing-card h2 {
    margin-top: 0;
    font-size: 1.5rem;
  }

  .price {
    font-size: 2.5rem;
    font-weight: 700;
    margin: 1rem 0;
  }

  .price span {
    font-size: 1rem;
    color: var(--muted);
    font-weight: 400;
  }

  .description {
    color: var(--muted);
    margin-bottom: 2rem;
    line-height: 1.5;
  }

  .feature-list {
    list-style: none;
    padding: 0;
    margin: 0 0 2.5rem 0;
    flex-grow: 1;
  }

  .feature-list li {
    margin-bottom: 1rem;
    color: var(--text);
  }

  .pricing-card.pro {
    border-color: var(--accent);
    box-shadow: 0 0 40px rgba(124, 58, 237, 0.15);
  }

  .popular-badge {
    position: absolute;
    top: -12px;
    right: 2rem;
    background-color: var(--accent);
    color: white;
    padding: 0.25rem 0.75rem;
    border-radius: 99px;
    font-size: 0.875rem;
    font-weight: 600;
  }
</style>
