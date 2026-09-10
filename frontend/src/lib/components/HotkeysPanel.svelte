<script lang="ts">
    import { uiState } from '$lib/stores/uiState.svelte.ts';
</script>

{#if uiState.hotkeysPanelOpen}
    <!-- Backdrop -->
    <div
        class="fixed inset-0 z-[9998] bg-black/50 backdrop-blur-sm pointer-events-auto"
        onclick={() => uiState.hotkeysPanelOpen = false}
        onkeydown={(e) => e.key === 'Escape' && (uiState.hotkeysPanelOpen = false)}
        role="button"
        tabindex="0"
        aria-label="Close hotkeys panel"
    ></div>
{/if}

<!-- Drawer Panel -->
<div
    class="glass-panel fixed top-0 right-0 h-full w-96 max-w-[90vw] z-[9999] bg-surface-variant/90 backdrop-blur-md border-l border-white/10 shadow-2xl transition-transform duration-300 ease-in-out flex flex-col pointer-events-auto {uiState.hotkeysPanelOpen ? 'translate-x-0' : 'translate-x-full'}"
>
    <!-- Header -->
    <div class="flex items-center justify-between border-b border-white/10 p-6">
        <h2 class="text-xl font-headline-md text-on-background tracking-wide flex items-center gap-2">
            <span class="material-symbols-outlined text-primary">keyboard</span>
            Hotkeys
        </h2>
        <button
            class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-white/10 text-on-surface-variant hover:text-red-400 transition-colors"
            onclick={() => uiState.hotkeysPanelOpen = false}
        >
            <span class="material-symbols-outlined text-[20px]">close</span>
        </button>
    </div>

    <!-- Hotkeys List -->
    <div class="flex-1 overflow-y-auto p-6 flex flex-col gap-3 font-body-sm text-on-surface-variant">
        <div class="flex items-center justify-between py-2 border-b border-white/5">
            <span class="font-bold">Toggle Hotkeys Panel</span>
            <kbd class="px-2 py-1 bg-white/10 rounded font-mono text-sm text-on-background">Ctrl+/</kbd>
        </div>
        <div class="flex items-center justify-between py-2 border-b border-white/5">
            <span class="font-bold">Push to Talk</span>
            <kbd class="px-2 py-1 bg-white/10 rounded font-mono text-sm text-on-background">Ctrl+Shift+Space</kbd>
        </div>
        <div class="flex items-center justify-between py-2 border-b border-white/5">
            <span class="font-bold">Screenshot Vision</span>
            <kbd class="px-2 py-1 bg-white/10 rounded font-mono text-sm text-on-background">Ctrl+Shift+S</kbd>
        </div>
        <div class="flex items-center justify-between py-2 border-b border-white/5">
            <span class="font-bold">Send Transcript Chip 1-6</span>
            <kbd class="px-2 py-1 bg-white/10 rounded font-mono text-sm text-on-background">Ctrl+Shift+1...6</kbd>
        </div>
        <div class="flex items-center justify-between py-2 border-b border-white/5">
            <span class="font-bold">Scroll Answer Down / Up</span>
            <kbd class="px-2 py-1 bg-white/10 rounded font-mono text-sm text-on-background">Ctrl+Shift+↓/↑</kbd>
        </div>
        <div class="flex items-center justify-between py-2 border-b border-white/5">
            <span class="font-bold">Toggle Click-through Mode</span>
            <kbd class="px-2 py-1 bg-white/10 rounded font-mono text-sm text-on-background">Ctrl+Shift+M</kbd>
        </div>
        <div class="flex items-center justify-between py-2 border-b border-white/5">
            <span class="font-bold">Session Report</span>
            <kbd class="px-2 py-1 bg-white/10 rounded font-mono text-sm text-on-background">Ctrl+Shift+E</kbd>
        </div>
        <div class="flex items-center justify-between py-2">
            <span class="font-bold">Panic Clear / Hide</span>
            <kbd class="px-2 py-1 bg-white/10 rounded font-mono text-sm text-on-background">Ctrl+Shift+X</kbd>
        </div>
    </div>
</div>

<style>
  .glass-panel {
      background: rgba(10, 10, 14, 0.88);
      backdrop-filter: blur(24px) saturate(1.4);
      -webkit-backdrop-filter: blur(24px) saturate(1.4);
      border-left: 1px solid rgba(255, 255, 255, 0.10);
      box-shadow: 0 24px 48px rgba(0, 0, 0, 0.7), inset 0 1px 0 rgba(255,255,255,0.06);
      /* Since we use transition-transform in tailwind class, we only transition the visual props here if needed, but not transform */
      transition: background 0.3s ease, backdrop-filter 0.3s ease, border-color 0.3s ease, box-shadow 0.3s ease, transform 0.3s ease-in-out;
  }
</style>
