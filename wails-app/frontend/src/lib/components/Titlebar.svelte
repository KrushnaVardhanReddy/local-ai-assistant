<script lang="ts">
  import { WindowMinimise, WindowToggleMaximise, Quit } from '../../../wailsjs/runtime/runtime';
</script>

<div class="auth-titlebar flex justify-between items-center w-full h-8 bg-[#1e1e1e] cursor-grab active:cursor-grabbing rounded-t-lg border-b border-white/5" style="--wails-draggable: drag">
  <div class="drag-handle flex-grow h-full flex items-center justify-center pointer-events-none">
    <div class="dots w-8 h-1 rounded bg-white/10 transition-colors duration-200"></div>
  </div>

  <div class="controls flex h-full" style="--wails-draggable: no-drag">
    <button onclick={WindowMinimise} aria-label="Minimize" title="Minimize" class="control-btn">
      <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M2 6H10" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
      </svg>
    </button>
    <button onclick={WindowToggleMaximise} aria-label="Maximize" title="Maximize" class="control-btn">
      <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
        <rect x="2" y="2" width="8" height="8" stroke="currentColor" stroke-width="2"/>
      </svg>
    </button>
    <button onclick={Quit} aria-label="Close" title="Close" class="control-btn close-btn">
      <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M2 2L10 10M10 2L2 10" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
      </svg>
    </button>
  </div>
</div>

<style>
  .auth-titlebar:hover .dots {
    background-color: rgba(255, 255, 255, 0.3);
  }

  .control-btn {
    background: transparent;
    border: none;
    color: #aaaaaa;
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

  .control-btn:hover {
    background-color: rgba(255, 255, 255, 0.1);
    color: #ffffff;
  }

  .close-btn:hover {
    background-color: rgba(232, 17, 35, 0.8);
    color: white;
  }
</style>
