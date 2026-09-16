<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime';
  import StealthTitleBar from '../../lib/components/StealthTitleBar.svelte';

  // Assuming we use standard Wails window runtime API in production or mock state in dev
  // Actually, we poll GetState from the Wails backend using `window.go.presenter.PresenterApp.GetState()`
  // but to keep it simple and compile without generated TS errors initially (since we haven't generated wails bindings for presenter app),
  // we use dynamic calling or simply any.
  let latestTranscript = '';
  let latestResponse = '';
  let pollInterval: ReturnType<typeof setInterval>;

  // Typography state controls
  let opacity = 0.85;
  let fontSize = 1.1; // rem
  let lineHeight = 1.5;

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

<div class="presenter-root">
  <div class="eyeline-indicator"></div>

  <div
    class="presenter-hud"
    style="background-color: rgba(10, 10, 10, {opacity}); --dynamic-font-size: {fontSize}rem; --dynamic-line-height: {lineHeight};"
  >
    <StealthTitleBar />

    <div class="toolbar">
      <div class="control-group">
        <label for="opacity">Opacity</label>
        <input id="opacity" type="range" min="0" max="1" step="0.05" bind:value={opacity} />
      </div>
      <div class="control-group">
        <label for="font-size">Size</label>
        <input id="font-size" type="range" min="0.5" max="3" step="0.1" bind:value={fontSize} />
      </div>
      <div class="control-group">
        <label for="line-height">Spacing</label>
        <input id="line-height" type="range" min="1" max="2.5" step="0.1" bind:value={lineHeight} />
      </div>
      <div class="header">
        <button class="clear-btn" aria-label="Clear state" on:click={handleClear}>⟳</button>
      </div>
    </div>

    <div class="content">
      <div class="transcript">{latestTranscript}</div>
      <div class="response">{latestResponse}</div>
    </div>
  </div>
</div>

<style>
  :root {
    --hud-text-gray: #aaaaaa;
    --hud-text-white: #ffffff;
    --hud-width: 640px;
    --hud-font-family: system-ui, -apple-system, sans-serif;
    --accent-color: rgba(255, 255, 255, 0.2);
  }

  /* Full screen transparent root to allow for HUD positioning and eyeline */
  .presenter-root {
    width: 100vw;
    height: 100vh;
    background: transparent;
    position: relative;
    pointer-events: none; /* Let clicks pass through the root container */
    font-family: var(--hud-font-family);
  }

  /* Eyeline indicator guide in the upper-middle of screen */
  .eyeline-indicator {
    position: absolute;
    top: 15%; /* Roughly webcam level */
    left: 0;
    width: 100%;
    height: 1px;
    background: linear-gradient(90deg, transparent 10%, rgba(255, 255, 255, 0.3) 50%, transparent 90%);
    pointer-events: none;
    z-index: 1000;
  }

  .presenter-hud {
    width: var(--hud-width);
    height: auto;
    border-radius: 8px;
    padding: 16px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    gap: 8px;
    pointer-events: auto; /* Enable clicks on HUD elements */

    /* Position top-center pinned */
    position: fixed;
    top: 10px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 9999;

    /* Smooth transition when settings change */
    transition: background-color 0.2s ease;
  }

  /* Floating toolbar */
  .toolbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 16px;
    opacity: 0;
    transition: opacity 0.2s ease-in-out;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--accent-color);
    margin-bottom: 8px;
  }

  /* Show toolbar on hover over the HUD */
  .presenter-hud:hover .toolbar {
    opacity: 1;
  }

  .control-group {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.8rem;
    color: var(--hud-text-gray);
  }

  .control-group label {
    user-select: none;
  }

  .control-group input[type="range"] {
    cursor: pointer;
    accent-color: var(--hud-text-white);
    width: 80px;
  }

  .header {
    display: flex;
    justify-content: flex-end;
    flex-grow: 1;
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
    font-size: calc(var(--dynamic-font-size) * 0.8);
    line-height: var(--dynamic-line-height);
    color: var(--hud-text-gray);
    min-height: 1.2em;
    transition: font-size 0.2s, line-height 0.2s;
  }

  .response {
    font-size: var(--dynamic-font-size);
    line-height: var(--dynamic-line-height);
    color: var(--hud-text-white);
    min-height: 1.5em;
    transition: font-size 0.2s, line-height 0.2s;
  }
</style>
