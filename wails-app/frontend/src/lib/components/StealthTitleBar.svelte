<script lang="ts">
  import { Quit, WindowMinimise } from '../../../wailsjs/runtime/runtime';

  export let title: string = '';
  export let showTitle: boolean = false;

  function handleMinimize() {
    WindowMinimise();
  }

  function handleClose() {
    Quit();
  }
</script>

<div class="stealth-titlebar" style="--wails-draggable: drag">
  {#if showTitle}
    <div class="title">{title}</div>
  {:else}
    <div class="drag-handle">
      <!-- A subtle visual indicator of grab region -->
      <div class="dots"></div>
    </div>
  {/if}
  
  <div class="controls" style="--wails-draggable: no-drag">
    <button on:click={handleMinimize} aria-label="Minimize" title="Minimize">
      <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M2 6H10" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
      </svg>
    </button>
    <button class="close-btn" on:click={handleClose} aria-label="Close" title="Close">
      <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M2 2L10 10M10 2L2 10" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
      </svg>
    </button>
  </div>
</div>

<style>
  .stealth-titlebar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    height: 32px;
    background: transparent;
    cursor: grab;
    border-top-left-radius: 8px;
    border-top-right-radius: 8px;
    /* Optional: a subtle border at the bottom of the title bar to separate it from the content */
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .stealth-titlebar:active {
    cursor: grabbing;
  }

  .title {
    color: var(--hud-text-gray, #aaaaaa);
    font-size: 0.8rem;
    font-weight: 500;
    padding-left: 12px;
    pointer-events: none; /* Let drag pass through text */
  }

  .drag-handle {
    flex-grow: 1;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .dots {
    width: 32px;
    height: 4px;
    border-radius: 2px;
    background-color: rgba(255, 255, 255, 0.1);
    transition: background-color 0.2s;
  }

  .stealth-titlebar:hover .dots {
    background-color: rgba(255, 255, 255, 0.3);
  }

  .controls {
    display: flex;
    height: 100%;
  }

  button {
    background: transparent;
    border: none;
    color: var(--hud-text-gray, #aaaaaa);
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: background-color 0.2s, color 0.2s;
    border-radius: 4px;
    margin-right: 2px;
  }

  button:hover {
    background-color: rgba(255, 255, 255, 0.1);
    color: var(--hud-text-white, #ffffff);
  }

  .close-btn:hover {
    background-color: rgba(232, 17, 35, 0.8);
    color: white;
  }
</style>
