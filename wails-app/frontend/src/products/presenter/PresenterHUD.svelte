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
  let script = '';
  let scriptWords: string[] = [];
  let currentWordIndex = 0;

  let autoScrollPaused = false;
  let resumeScrollTimeout: ReturnType<typeof setTimeout>;

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
          const newScript = state.script || '';
          if (newScript !== script) {
            script = newScript;
            scriptWords = script.split(/\s+/).filter(w => w.length > 0);
            currentWordIndex = 0;
            updateScrollPosition();
          }
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
      script = '';
      scriptWords = [];
      currentWordIndex = 0;
    } catch (err) {
      console.error('Failed to clear state:', err);
    }
  }

  let scriptContainer: HTMLElement;

  function updateScrollPosition() {
    if (!scriptContainer) return;

    // Use querySelectorAll to find all words
    const words = scriptContainer.querySelectorAll('.script-word');
    if (words && words.length > currentWordIndex) {
      const activeWord = words[currentWordIndex] as HTMLElement;
      if (activeWord) {
        // smooth scroll the container so the active word is roughly centered
        if (!autoScrollPaused) {
          activeWord.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
      }
    }
  }

  function advanceReadingPosition(transcript: string) {
    if (!transcript || scriptWords.length === 0 || autoScrollPaused) return;

    // Simple heuristic: count words in the incoming transcript and advance the index
    // In a real STT, you'd want something more robust matching words textually,
    // but for now we'll match by counting how many words have been spoken.

    // A slightly better heuristic: find the transcript words in the next N words of the script
    const spokenWords = transcript.toLowerCase().split(/\s+/).filter(w => w.length > 0);
    if (spokenWords.length === 0) return;

    // Take the last 3-4 words of the transcript to find our place
    const recentSpoken = spokenWords.slice(-4);

    // Look ahead in the script to find a match
    const lookaheadLimit = Math.min(currentWordIndex + 100, scriptWords.length);
    let bestMatchIndex = currentWordIndex;

    for (let i = currentWordIndex; i < lookaheadLimit; i++) {
      const scriptWordLower = scriptWords[i].toLowerCase().replace(/[.,!?;:]/g, '');
      const spokenWordLower = recentSpoken[recentSpoken.length - 1]?.replace(/[.,!?;:]/g, '');

      if (scriptWordLower === spokenWordLower) {
         bestMatchIndex = i;
         break; // found the latest word
      }
    }

    if (bestMatchIndex > currentWordIndex) {
       currentWordIndex = bestMatchIndex;
       updateScrollPosition();
    }
  }

  function handleManualScroll(e: Event) {
    // Determine if it was manual scroll (wheel) or keys
    pauseAutoScrollTemporarily();
  }

  function handleKeyDown(e: KeyboardEvent) {
    // Pause toggle on spacebar
    if (e.code === 'Space') {
      e.preventDefault(); // prevent page scroll
      autoScrollPaused = !autoScrollPaused;
      if (!autoScrollPaused) {
        updateScrollPosition(); // snap back immediately
      }
      return;
    }

    // Up/Down arrows manually scroll
    if (e.code === 'ArrowUp' || e.code === 'ArrowDown') {
      pauseAutoScrollTemporarily();
      if (scriptContainer) {
        const scrollAmount = 40; // px
        scriptContainer.scrollTop += (e.code === 'ArrowDown' ? scrollAmount : -scrollAmount);
      }
    }
  }

  function pauseAutoScrollTemporarily() {
    autoScrollPaused = true;
    if (resumeScrollTimeout) clearTimeout(resumeScrollTimeout);

    // Resume auto-scroll after 5 seconds of inactivity
    resumeScrollTimeout = setTimeout(() => {
      autoScrollPaused = false;
      updateScrollPosition();
    }, 5000);
  }

  onMount(() => {
    // Add manual scroll overrides
    window.addEventListener('keydown', handleKeyDown);

    // Re-use existing on_response_token / on_response_end Wails events
    EventsOn('on_response_token', (token: string) => {
      latestResponse += token;
    });

    EventsOn('on_response_end', () => {
      fetchState(); // sync state fully when response ends
    });

    EventsOn('on_transcript', (data: any) => {
      latestTranscript = data.text || '';
      if (latestTranscript) {
        advanceReadingPosition(latestTranscript);
      }
    });

    pollInterval = setInterval(fetchState, 200);
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleKeyDown);
    if (resumeScrollTimeout) clearTimeout(resumeScrollTimeout);
    if (pollInterval) clearInterval(pollInterval);
    EventsOff('on_response_token');
    EventsOff('on_response_end');
    EventsOff('on_transcript');
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
        {#if autoScrollPaused}
          <div class="status-badge paused">PAUSED</div>
        {:else}
          <div class="status-badge listening">AUTO-SYNC</div>
        {/if}
        <button class="clear-btn" aria-label="Clear state" on:click={handleClear}>⟳</button>
      </div>
    </div>

    <div class="content">
      {#if scriptWords.length > 0}
        <div class="script-display" bind:this={scriptContainer} on:wheel={handleManualScroll}>
          {#each scriptWords as word, i}
             <span class="script-word {i === currentWordIndex ? 'active' : ''} {i < currentWordIndex ? 'read' : ''}">
                {word}{' '}
             </span>
          {/each}
        </div>
      {/if}
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

  .script-display {
    max-height: 300px;
    overflow-y: auto;
    font-size: var(--dynamic-font-size);
    line-height: var(--dynamic-line-height);
    padding: 10px;
    background: rgba(0, 0, 0, 0.2);
    border-radius: 4px;

    /* Hide scrollbar for stealth */
    scrollbar-width: none; /* Firefox */
  }

  .script-display::-webkit-scrollbar {
    display: none; /* Chrome, Safari and Opera */
  }

  .script-word {
    color: var(--hud-text-gray);
    transition: color 0.2s;
  }

  .script-word.read {
    color: rgba(255, 255, 255, 0.3);
  }

  .script-word.active {
    color: #4ade80; /* bright green for current position */
    font-weight: bold;
    text-shadow: 0 0 4px rgba(74, 222, 128, 0.4);
  }

  .status-badge {
    font-size: 0.7rem;
    font-weight: bold;
    padding: 2px 6px;
    border-radius: 4px;
    margin-right: 8px;
    display: flex;
    align-items: center;
    letter-spacing: 0.5px;
  }

  .status-badge.paused {
    background: rgba(239, 68, 68, 0.2);
    color: #ef4444; /* red */
  }

  .status-badge.listening {
    background: rgba(74, 222, 128, 0.2);
    color: #4ade80; /* green */
  }
</style>
