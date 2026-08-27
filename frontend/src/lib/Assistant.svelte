<script lang="ts">
  import { wsState, sendChat } from "$lib/ws.svelte";
  import { onMount } from "svelte";
  import { listen } from "@tauri-apps/api/event";
  import { invoke } from "@tauri-apps/api/core";

  let responseEl: HTMLElement | undefined = $state();

  let isBrowser = $state(false);
  let chatText = $state("");

  onMount(() => {
    // Check if we are running in a regular browser instead of Tauri
    isBrowser = typeof window !== 'undefined' && typeof (window as any).__TAURI_INTERNALS__ === 'undefined';
    const apiUrl = isBrowser ? `${window.location.protocol}//${window.location.host}` : "http://127.0.0.1:8000";

    const unlisten = listen("trigger-vision", async () => {
      if (wsState.isAnalyzingScreen) return;

      wsState.isAnalyzingScreen = true;
      try {
        const base64Image = await invoke<string>("capture_screen");

        await fetch(`${apiUrl}/vision/analyze`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json"
          },
          body: JSON.stringify({ image_base64: base64Image })
        });

      } catch (e) {
        console.error("Failed to capture screen or send to backend:", e);
      } finally {
        wsState.isAnalyzingScreen = false;
      }
    });

    return () => {
      unlisten.then(f => f());
    };
  });

  $effect(() => {
    // This effect runs whenever wsState.response changes
    if (wsState.response && responseEl) {
      responseEl.scrollTop = responseEl.scrollHeight;
    }
  });

  function clearError() {
    wsState.error = null;
  }

  function handleChatSubmit() {
    if (chatText.trim()) {
      sendChat(chatText);
      chatText = "";
    }
  }

  function handleChatKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleChatSubmit();
    }
  }
</script>

<div class="assistant-panel">
  <div class="header">
    <div class="brand">✦ Local AI</div>
    <div class="header-right">
      {#if wsState.isAnalyzingScreen}
        <div class="vision-indicator" title="Analyzing screen...">👁️</div>
      {/if}
      <div class="mic-status {wsState.isListening ? 'listening' : ''}"></div>
    </div>
  </div>

  <div class="content">
    {#if !wsState.transcript && !wsState.response && wsState.isConnected}
      <div class="idle-hint">Hold <code>Ctrl+Shift+Space</code> to speak</div>
    {/if}

    {#if wsState.transcript}
      {#key wsState.transcript}
        <div class="transcript-bubble">
          {wsState.transcript}
        </div>
      {/key}
    {/if}

    {#if wsState.response || wsState.isThinking}
      <div class="response-area" bind:this={responseEl}>
        <span class="response-text">{wsState.response}</span>
        {#if wsState.isThinking}
          <span class="cursor">▌</span>
        {/if}
        {#if wsState.ragSources && wsState.ragSources.length > 0}
          <div class="rag-sources">
            🔍 Sources: {wsState.ragSources.join(' · ')}
          </div>
        {/if}
      </div>
    {/if}
  </div>

  {#if isBrowser}
    <div class="chat-input-container">
      <textarea
        bind:value={chatText}
        onkeydown={handleChatKeydown}
        placeholder="Silent Chat Input (Helper Mode)..."
        rows="2"
      ></textarea>
      <button class="chat-submit" onclick={handleChatSubmit} disabled={!chatText.trim()}>
        Send
      </button>
    </div>
  {/if}

  <div class="status-bar">
    {#if wsState.error}
      <button class="status-error" onclick={clearError}>
        ⚠ {wsState.error}
      </button>
    {:else if wsState.isConnected}
      <span class="status-connected">● Connected</span>
    {:else}
      <span class="status-connecting">○ Reconnecting<span class="dots"></span></span>
    {/if}
  </div>
</div>

<style>
  .assistant-panel {
    background: rgba(13, 13, 13, 0.85);
    backdrop-filter: blur(12px);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 16px;
    color: #e2e8f0;
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
    overflow: hidden;
    /* Providing some initial sizing to make it visible during dev */
    min-height: 400px;
    max-width: 400px;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.5rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .brand {
    color: #7c3aed;
    font-weight: 500;
    font-size: 1.1rem;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .vision-indicator {
    font-size: 1.1rem;
    animation: pulse 1.5s infinite;
  }

  .mic-status {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    background-color: #475569; /* grey when not listening */
    transition: background-color 0.3s;
  }

  .mic-status.listening {
    background-color: #22c55e;
    animation: pulse 1.5s infinite;
  }

  @keyframes pulse {
    0% {
      box-shadow: 0 0 0 0 rgba(34, 197, 94, 0.4);
    }
    70% {
      box-shadow: 0 0 0 8px rgba(34, 197, 94, 0);
    }
    100% {
      box-shadow: 0 0 0 0 rgba(34, 197, 94, 0);
    }
  }

  .content {
    flex: 1;
    display: flex;
    flex-direction: column;
    padding: 1.5rem;
    overflow-y: auto;
    gap: 1.5rem;
  }

  .idle-hint {
    margin: auto;
    color: #94a3b8;
    font-size: 0.9rem;
    text-align: center;
  }

  .idle-hint code {
    background: rgba(255, 255, 255, 0.1);
    padding: 0.2rem 0.4rem;
    border-radius: 4px;
    font-family: inherit;
    font-size: 0.85rem;
  }

  .transcript-bubble {
    align-self: flex-end;
    background: rgba(124, 58, 237, 0.15);
    border: 1px solid rgba(124, 58, 237, 0.5);
    border-radius: 12px;
    padding: 0.75rem 1rem;
    max-width: 85%;
    font-size: 0.95rem;
    line-height: 1.4;
    animation: fade-up 0.3s ease-out forwards;
  }

  .response-area {
    font-size: 0.95rem;
    line-height: 1.5;
    color: #e2e8f0;
    max-height: 100%;
    overflow-y: auto;
    white-space: pre-wrap;
  }

  .cursor {
    color: #7c3aed;
    display: inline-block;
    animation: blink 1s step-end infinite;
    margin-left: 2px;
  }

  @keyframes blink {
    0%, 100% { opacity: 1; }
    50% { opacity: 0; }
  }

  .status-bar {
    padding: 0.75rem 1.5rem;
    border-top: 1px solid rgba(255, 255, 255, 0.05);
    font-size: 0.8rem;
    display: flex;
    align-items: center;
  }

  .status-connected {
    color: #22c55e;
  }

  .status-connecting {
    color: #94a3b8;
  }

  .dots::after {
    content: '';
    animation: dots 1.5s steps(4, end) infinite;
  }

  @keyframes dots {
    0% { content: ''; }
    25% { content: '.'; }
    50% { content: '..'; }
    75% { content: '...'; }
    100% { content: ''; }
  }

  .status-error {
    color: #f59e0b;
    background: none;
    border: none;
    padding: 0;
    font-size: inherit;
    font-family: inherit;
    cursor: pointer;
    text-align: left;
  }

  .status-error:hover {
    text-decoration: underline;
  }

  .rag-sources {
    font-size: 0.7rem;
    color: #64748b;
    margin-top: 0.5rem;
    font-style: italic;
  }

  .chat-input-container {
    display: flex;
    padding: 1rem 1.5rem;
    gap: 1rem;
    border-top: 1px solid rgba(255, 255, 255, 0.05);
  }

  .chat-input-container textarea {
    flex: 1;
    background: rgba(0, 0, 0, 0.2);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 0.75rem;
    color: #e2e8f0;
    font-family: inherit;
    font-size: 0.9rem;
    resize: none;
    outline: none;
  }

  .chat-input-container textarea:focus {
    border-color: rgba(124, 58, 237, 0.5);
  }

  .chat-submit {
    background: #7c3aed;
    color: white;
    border: none;
    border-radius: 8px;
    padding: 0 1rem;
    cursor: pointer;
    font-weight: 500;
    transition: background 0.2s;
  }

  .chat-submit:hover:not(:disabled) {
    background: #6d28d9;
  }

  .chat-submit:disabled {
    background: #475569;
    cursor: not-allowed;
    opacity: 0.7;
  }
</style>
