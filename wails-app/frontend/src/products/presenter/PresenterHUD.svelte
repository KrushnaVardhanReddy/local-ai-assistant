<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { EventsOn, EventsOff } from '../../../../wailsjs/runtime/runtime';

  // Assuming we use standard Wails window runtime API in production or mock state in dev
  // Actually, we poll GetState from the Wails backend using `window.go.presenter.PresenterApp.GetState()`
  // but to keep it simple and compile without generated TS errors initially (since we haven't generated wails bindings for presenter app),
  // we use dynamic calling or simply any.
  let latestTranscript = '';
  let latestResponse = '';
  let pollInterval: ReturnType<typeof setInterval>;

  async function fetchState() {
    try {
      // @ts-ignore
      if (window.go && window.go.presenter && window.go.presenter.PresenterApp) {
        // @ts-ignore
        const state = await window.go.presenter.PresenterApp.GetState();
        if (state) {
          latestTranscript = state.transcript || '';
          latestResponse = state.response || '';
        }
      }
    } catch (err) {
      console.error('Failed to fetch state:', err);
    }
  }

  async function handleClear() {
    try {
      // @ts-ignore
      if (window.go && window.go.presenter && window.go.presenter.PresenterApp) {
        // @ts-ignore
        await window.go.presenter.PresenterApp.ClearState();
      }
      latestTranscript = '';
      latestResponse = '';
    } catch (err) {
      console.error('Failed to clear state:', err);
    }
  }

  onMount(() => {
    // Re-use existing on_response_token / on_response_end Wails events
    EventsOn('on_response_token', (token: string) => {
      latestResponse += token;
    });

    EventsOn('on_response_end', () => {
      fetchState(); // sync state fully when response ends
    });

    pollInterval = setInterval(fetchState, 200);
  });

  onDestroy(() => {
    if (pollInterval) clearInterval(pollInterval);
    EventsOff('on_response_token');
    EventsOff('on_response_end');
  });
</script>

<div class="presenter-hud">
  <div class="header">
    <button class="clear-btn" aria-label="Clear state" on:click={handleClear}>⟳</button>
  </div>
  <div class="content">
    <div class="transcript">{latestTranscript}</div>
    <div class="response">{latestResponse}</div>
  </div>
</div>

<style>
  :root {
    --hud-bg: rgba(10, 10, 10, 0.85);
    --hud-text-gray: #aaaaaa;
    --hud-text-white: #ffffff;
    --hud-width: 640px;
    --hud-font-family: system-ui, -apple-system, sans-serif;
  }

  .presenter-hud {
    width: var(--hud-width);
    height: auto;
    background-color: var(--hud-bg);
    border-radius: 8px;
    padding: 16px;
    box-sizing: border-box;
    font-family: var(--hud-font-family);
    display: flex;
    flex-direction: column;
    gap: 8px;

    /* Position top-center pinned */
    position: fixed;
    top: 10px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 9999;
  }

  .header {
    display: flex;
    justify-content: flex-end;
  }

  .clear-btn {
    background: transparent;
    border: none;
    color: var(--hud-text-gray);
    font-size: 1.2rem;
    cursor: pointer;
    transition: color 0.2s;
    padding: 4px;
    line-height: 1;
  }

  .clear-btn:hover {
    color: var(--hud-text-white);
  }

  .content {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .transcript {
    font-size: 0.9rem;
    color: var(--hud-text-gray);
    min-height: 1.2em;
  }

  .response {
    font-size: 1.1rem;
    color: var(--hud-text-white);
    min-height: 1.5em;
  }
</style>
