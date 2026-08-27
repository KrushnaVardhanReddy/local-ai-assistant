<script lang="ts">
  import { onMount } from 'svelte';

  let featuresSection: HTMLElement;
  let hasScrolled = false;

  onMount(() => {
    const observer = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add('visible');
        }
      });
    }, { threshold: 0.1 });

    const featureCards = document.querySelectorAll('.feature-card');
    featureCards.forEach((card) => observer.observe(card));

    return () => {
      featureCards.forEach((card) => observer.unobserve(card));
    };
  });
</script>

<svelte:head>
  <title>Local AI Assistant</title>
</svelte:head>

<nav class="navbar">
  <div class="logo">✦ Local AI</div>
  <div class="nav-links">
    <a href="/download" class="btn btn-secondary">Download</a>
    <a href="/login" class="btn btn-primary">Sign In</a>
  </div>
</nav>

<main>
  <section class="hero">
    <h1>Your Private AI Assistant</h1>
    <p class="subtitle">
      Real-time voice AI that runs on your hardware. No cloud. No data leaks. No subscriptions.
    </p>
    <div class="cta-group">
      <a href="/download" class="btn btn-primary btn-large">Download Free</a>
      <a href="/pricing" class="link-secondary">Or use the cloud dashboard →</a>
    </div>
  </section>

  <section class="features" bind:this={featuresSection}>
    <div class="feature-card">
      <h3>100% Private</h3>
      <p>Data never leaves your machine. Your conversations stay strictly confidential.</p>
    </div>
    <div class="feature-card">
      <h3>Any LLM</h3>
      <p>Seamlessly integrate with Ollama, LM Studio, OpenAI, Groq, and more.</p>
    </div>
    <div class="feature-card">
      <h3>Screen Share Safe</h3>
      <p>Invisible to Zoom, Teams, and OBS. Keep your AI assistance truly stealthy.</p>
    </div>
  </section>

  <section class="stealth-features">
    <h2>Stealth Mode</h2>
    <ol class="stealth-list">
      <li><span>01</span> Screen Share Safe</li>
      <li><span>02</span> Dock Hidden</li>
      <li><span>03</span> Task Manager Invisible</li>
      <li><span>04</span> Tab Switch Silent</li>
      <li><span>05</span> Cursor Undetectable</li>
    </ol>
  </section>

  <section class="providers">
    <p>Supported Providers</p>
    <div class="logo-row">
      <span>Ollama</span>
      <span class="dot">·</span>
      <span>LM Studio</span>
      <span class="dot">·</span>
      <span>llama.cpp</span>
      <span class="dot">·</span>
      <span>OpenAI</span>
      <span class="dot">·</span>
      <span>Groq</span>
      <span class="dot">·</span>
      <span>Gemini</span>
      <span class="dot">·</span>
      <span>Anthropic</span>
    </div>
  </section>
</main>

<footer>
  <a href="https://github.com" target="_blank" rel="noreferrer">GitHub</a>
  <a href="/privacy">Privacy Policy</a>
  <span>© 2025 Local AI Assistant</span>
</footer>

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

  .btn-large {
    padding: 0.75rem 1.5rem;
    font-size: 1.1rem;
  }

  .link-secondary {
    color: var(--muted);
    font-size: 0.9rem;
    text-decoration: none;
  }

  .link-secondary:hover {
    color: var(--text);
  }

  main {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
  }

  .hero {
    text-align: center;
    padding: 6rem 0;
  }

  .hero h1 {
    font-size: 4rem;
    margin-bottom: 1.5rem;
    line-height: 1.2;
    background: linear-gradient(to right, #fff, #a5b4fc);
    background-clip: text;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .hero .subtitle {
    font-size: 1.25rem;
    color: var(--muted);
    max-width: 600px;
    margin: 0 auto 3rem auto;
    line-height: 1.6;
  }

  .cta-group {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1.5rem;
  }

  .features {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 2rem;
    padding: 4rem 0;
  }

  .feature-card {
    background-color: var(--surface);
    border: var(--border);
    padding: 2rem;
    border-radius: 8px;
    opacity: 0;
    transform: translateY(20px);
    transition: opacity 0.6s ease-out, transform 0.6s ease-out;
  }

  :global(.feature-card.visible) {
    opacity: 1;
    transform: translateY(0);
  }

  .feature-card h3 {
    margin-top: 0;
    color: var(--accent);
    font-size: 1.25rem;
  }

  .feature-card p {
    color: var(--muted);
    line-height: 1.5;
    margin-bottom: 0;
  }

  .stealth-features {
    padding: 4rem 0;
    border-top: var(--border);
  }

  .stealth-features h2 {
    text-align: center;
    font-size: 2rem;
    margin-bottom: 3rem;
  }

  .stealth-list {
    list-style: none;
    padding: 0;
    max-width: 600px;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .stealth-list li {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem;
    background-color: var(--surface);
    border: var(--border);
    border-radius: 4px;
    font-size: 1.1rem;
  }

  .stealth-list li span {
    color: var(--accent);
    font-family: monospace;
    font-weight: 700;
  }

  .providers {
    text-align: center;
    padding: 4rem 0;
    color: var(--muted);
  }

  .providers p {
    margin-bottom: 2rem;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    font-size: 0.875rem;
  }

  .logo-row {
    display: flex;
    justify-content: center;
    flex-wrap: wrap;
    gap: 1.5rem;
    align-items: center;
    font-weight: 500;
  }

  .dot {
    color: var(--surface);
  }

  footer {
    padding: 2rem;
    display: flex;
    justify-content: center;
    gap: 2rem;
    border-top: var(--border);
    color: var(--muted);
    font-size: 0.875rem;
  }

  footer a:hover {
    color: var(--text);
  }
</style>
