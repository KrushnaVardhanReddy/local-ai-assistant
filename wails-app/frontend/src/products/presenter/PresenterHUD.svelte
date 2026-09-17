<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { EventsOn, EventsOff } from '../../../wailsjs/runtime/runtime';
  import StealthTitleBar from '../../lib/components/StealthTitleBar.svelte';
  import WorkspaceSidebar from '../../lib/components/workspace/WorkspaceSidebar.svelte';
  import WorkspaceTabs from '../../lib/components/workspace/WorkspaceTabs.svelte';
  import IDEShell from '../../lib/components/workspace/IDEShell.svelte';
  import CodeEditor from '../../lib/components/workspace/CodeEditor.svelte';
  import PresenterSettings from './PresenterSettings.svelte';
  import type { FileNode, WorkspaceTab } from '../../lib/components/workspace/types';

  // Assuming we use standard Wails window runtime API in production or mock state in dev
  // Actually, we poll GetState from the Wails backend using `window.go.presenter.PresenterApp.GetState()`
  // but to keep it simple and compile without generated TS errors initially (since we haven't generated wails bindings for presenter app),
  // we use dynamic calling or simply any.
  let latestTranscript = $state('');
  let latestResponse = $state('');
  let script = $state('');
  let scriptWords = $state<string[]>([]);
  let currentWordIndex = $state(0);
  let activeLineIndex = $state(0);

  let isThinking = $state(false);
  let isCopilotOpen = $state(false);

  let autoScrollPaused = $state(false);
  let resumeScrollTimeout: ReturnType<typeof setTimeout>;

  let pollInterval: ReturnType<typeof setInterval>;

  // Workspace State
  let isSidebarOpen = $state(false);
  let workspaceTree = $state<FileNode[]>([]);
  let openTabs = $state<WorkspaceTab[]>([]);
  let activeDocumentPath = $state('');

  // Typography state controls
  let opacity = $state(0.85);
  let fontSize = $state(1.1); // rem
  let lineHeight = $state(1.5);

  async function fetchState() {
    try {
      // @ts-ignore
      if (window.go && window.go.presenter && window.go.presenter.PresenterApp) {
        // @ts-ignore
        const state = await window.go.presenter.PresenterApp.GetState();
        if (state) {
          latestTranscript = state.transcript || '';
          latestResponse = state.response || '';
          isThinking = state.thinking || false;
          const newScript = state.script || '';
          if (newScript !== script) {
            script = newScript;
            scriptWords = script.split(/\s+/).filter(w => w.length > 0);
            currentWordIndex = 0;
            updateScrollPosition();
          }

          // We need to map `isDir` from backend to `isDirectory` for Svelte components.
          // We also preserve user expanded folder states across polling intervals.
          const expandedMap = new Map<string, boolean>();
          const recordExpanded = (nodes: FileNode[]) => {
            for (const n of nodes) {
              if (n.isDirectory && n.isExpanded !== undefined) {
                expandedMap.set(n.path, n.isExpanded);
              }
              if (n.children) recordExpanded(n.children);
            }
          };
          recordExpanded(workspaceTree);

          const mapTree = (nodes: any[]): FileNode[] => {
            if (!nodes) return [];
            return nodes.map((n: any) => ({
              ...n,
              isDirectory: n.isDir,
              isExpanded: expandedMap.has(n.path) ? expandedMap.get(n.path) : (n.isExpanded ?? false),
              children: mapTree(n.children)
            }));
          };
          workspaceTree = mapTree(state.workspaceTree);

          const rawDocs = state.openDocuments || [];
          openTabs = rawDocs.map((doc: any) => ({
            name: doc.name,
            path: doc.path
          }));
          activeDocumentPath = state.activeDocumentPath || '';

          if (isThinking || latestResponse) {
             isCopilotOpen = true;
          }
        }
      }
    } catch (err) {
      console.error('Failed to fetch state:', err);
    }
  }

  async function handleOpenFolder() {
    try {
      // @ts-ignore
      if (window.go && window.go.presenter && window.go.presenter.PresenterApp) {
        // @ts-ignore
        await window.go.presenter.PresenterApp.PromptOpenDirectory();
      }
      await fetchState();
    } catch (e) {
      console.error(e);
    }
  }

  async function handleOpenFile() {
    try {
      // @ts-ignore
      if (window.go && window.go.presenter && window.go.presenter.PresenterApp) {
        // @ts-ignore
        await window.go.presenter.PresenterApp.PromptOpenFile();
      }
      await fetchState();
    } catch (e) {
      console.error(e);
    }
  }

  async function handleSelectNode(node: FileNode) {
    if (node.isDirectory) return; // Note: Node might have isDir or isDirectory depending on mapping
    try {
      // @ts-ignore
      if (window.go && window.go.presenter && window.go.presenter.PresenterApp) {
        // @ts-ignore
        await window.go.presenter.PresenterApp.OpenFile(node.path);
      }
      await fetchState();
    } catch (e) {
      console.error(e);
    }
  }

  async function handleTabSelect(path: string) {
    try {
      // @ts-ignore
      if (window.go && window.go.presenter && window.go.presenter.PresenterApp) {
        // @ts-ignore
        await window.go.presenter.PresenterApp.SetActiveDocument(path);
      }
      await fetchState();
    } catch (e) {
      console.error(e);
    }
  }

  async function handleTabClose(path: string) {
    try {
      // @ts-ignore
      if (window.go && window.go.presenter && window.go.presenter.PresenterApp) {
        // @ts-ignore
        await window.go.presenter.PresenterApp.CloseFile(path);
      }
      await fetchState();
    } catch (e) {
      console.error(e);
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

  async function handleLoadDoc() {
    try {
      // @ts-ignore
      if (window.go && window.go.presenter && window.go.presenter.PresenterApp) {
        // @ts-ignore
        await window.go.presenter.PresenterApp.PromptOpenFile();
        await fetchState();
      }
    } catch (err) {
      console.error('Failed to load document:', err);
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
    pauseAutoScrollTemporarily();
  }

  function handleKeyDown(e: KeyboardEvent) {
    // Escape to un-pause auto scroll immediately
    if (e.key === 'Escape') {
      autoScrollPaused = false;
      if (resumeScrollTimeout) clearTimeout(resumeScrollTimeout);
      updateScrollPosition();
      return;
    }

    // Up/Down arrows manually scroll
    if (e.code === 'ArrowUp' || e.code === 'ArrowDown' || e.code === 'PageUp' || e.code === 'PageDown') {
      pauseAutoScrollTemporarily();
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
    EventsOn('on_response_start', () => {
      isThinking = true;
      isCopilotOpen = true;
      latestResponse = '';
    });

    EventsOn('on_response_token', (token: string) => {
      latestResponse += token;
      isThinking = true;
      isCopilotOpen = true;
    });

    EventsOn('on_response_end', () => {
      isThinking = false;
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
    EventsOff('on_response_start');
    EventsOff('on_response_token');
    EventsOff('on_response_end');
    EventsOff('on_transcript');
  });
</script>

<div class="presenter-container" style="--hud-bg: rgba(10, 10, 10, {opacity});">
  <StealthTitleBar />
  <IDEShell
    {workspaceTree}
    {openTabs}
    activePath={activeDocumentPath}
    totalDocs={openTabs.length}
    onSelect={handleSelectNode}
    onTabSelect={handleTabSelect}
    onTabClose={handleTabClose}
    onFileOpen={handleOpenFile}
    onOpenFolder={handleOpenFolder}
  >
    <div class="editor-header" style="padding: 4px 8px; border-bottom: 1px solid rgba(255,255,255,0.1); display: flex; justify-content: flex-end; gap: 8px; align-items: center; background: rgba(0,0,0,0.2);">
      {#if autoScrollPaused}
        <div class="status-badge paused" style="margin-right:auto;">PAUSED</div>
      {:else}
        <div class="status-badge listening" style="margin-right:auto;">AUTO-SYNC</div>
      {/if}
      <button class="clear-btn" aria-label="Load Document" title="Load Script" onclick={handleLoadDoc}>📁</button>
      <button class="clear-btn" aria-label="Clear state" title="Clear State" onclick={handleClear}>⟳</button>
    </div>

    <div class="editor-area" onwheel={handleManualScroll}>
      {#if script}
        <CodeEditor
          content={script}
          {fontSize}
          {lineHeight}
          activeLine={activeLineIndex}
          readOnly={true}
        />
      {:else}
        <div class="empty-state">No script loaded. Open a Markdown file.</div>
      {/if}
    </div>

    {#snippet rightDrawer()}
      <div class="copilot-panel {isThinking ? 'thinking' : ''}">
        <div class="copilot-header">
          <span>Audience Copilot</span>
          <button class="close-copilot" onclick={() => { isCopilotOpen = false; latestResponse = ''; }}>✖</button>
        </div>
        <div class="copilot-body">
          {#if latestTranscript}
            <div class="transcript">Q: {latestTranscript}</div>
          {/if}
          {#if latestResponse}
            <div class="response">A: {latestResponse}</div>
          {/if}
        </div>
      </div>
    {/snippet}

    {#snippet settingsPanel()}
      <div class="presenter-settings-wrapper">
        <h3 style="margin-top:0; border-bottom: 1px solid rgba(255,255,255,0.1); padding-bottom: 8px;">Display Settings</h3>
        <div class="control-group">
          <label for="opacity">Opacity: {opacity.toFixed(2)}</label>
          <input id="opacity" type="range" min="0" max="1" step="0.05" bind:value={opacity} />
        </div>
        <div class="control-group">
          <label for="font-size">Font Size: {fontSize.toFixed(1)}rem</label>
          <input id="font-size" type="range" min="0.5" max="3" step="0.1" bind:value={fontSize} />
        </div>
        <div class="control-group">
          <label for="line-height">Spacing: {lineHeight.toFixed(1)}</label>
          <input id="line-height" type="range" min="1" max="2.5" step="0.1" bind:value={lineHeight} />
        </div>
        <PresenterSettings />
      </div>
    {/snippet}
  </IDEShell>
</div>

<style>
  .presenter-container {
    width: 100vw;
    height: 100vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    pointer-events: auto;
    background-color: var(--hud-bg, rgba(18, 18, 18, 0.95));
  }

  .editor-area {
    flex-grow: 1;
    overflow: hidden;
    position: relative;
    display: flex;
    flex-direction: column;
  }

  .empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: rgba(255,255,255,0.4);
    font-size: 1.2rem;
  }

  .copilot-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 16px;
    box-sizing: border-box;
    transition: box-shadow 0.3s ease;
  }

  .copilot-panel.thinking {
    box-shadow: inset 0 0 15px rgba(74, 222, 128, 0.4);
  }

  .copilot-body {
    flex-grow: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin-top: 16px;
  }

  .presenter-settings-wrapper {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .control-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
    color: var(--hud-text-gray, #ccc);
    font-size: 0.9rem;
  }

  .control-group input[type="range"] {
    width: 100%;
    accent-color: var(--hud-text-white, #fff);
  }

  .clear-btn {
    background: transparent;
    border: none;
    color: var(--hud-text-gray, #ccc);
    font-size: 1.2rem;
    cursor: pointer;
    transition: color 0.2s;
    padding: 4px;
    line-height: 1;
  }

  .clear-btn:hover {
    color: var(--hud-text-white, #fff);
  }

  .copilot-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 0.9rem;
    color: var(--hud-text-gray, #ccc);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    padding-bottom: 8px;
  }

  .close-copilot {
    background: transparent;
    border: none;
    color: var(--hud-text-gray, #ccc);
    cursor: pointer;
    font-size: 1rem;
    padding: 0;
  }

  .close-copilot:hover {
    color: var(--hud-text-white, #fff);
  }

  .transcript, .response {
    color: #fff;
    font-size: 0.95rem;
    line-height: 1.5;
  }

  .transcript {
    color: #aaa;
    font-style: italic;
  }

  .status-badge {
    font-size: 0.7rem;
    font-weight: bold;
    padding: 2px 6px;
    border-radius: 4px;
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
