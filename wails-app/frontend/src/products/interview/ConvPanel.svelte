<script lang="ts">
  let {
    transcriptHistory = [],
    pendingTranscripts = [],
    ragSources = [],
    isListening = false,
    isMockMode = false,
    isPTTHeld = false,
    isThinking = false,
    includeActiveDocContext = false,
    activeDocumentName = '',

    onSendChat,
    onSendChip,
    onDismissChip,
    onClearChips,
    onStarMethod,
    onCatchMeUp,
  } = $props<{
    transcriptHistory?: Array<{ role: string; text: string }>;
    pendingTranscripts?: Array<{ id: string; text: string }>;
    ragSources?: string[];
    isListening?: boolean;
    isMockMode?: boolean;
    isPTTHeld?: boolean;
    isThinking?: boolean;
    includeActiveDocContext?: boolean;
    activeDocumentName?: string;
    onSendChat?: (text: string) => void;
    onSendChip?: (chip: { id: string; text: string }) => void;
    onDismissChip?: (chipId: string) => void;
    onClearChips?: () => void;
    onStarMethod?: () => void;
    onCatchMeUp?: () => void;
  }>();

  let showHotkeys = $state(false);
  let chatText = $state("");
  let scrollEl: HTMLElement | undefined = $state();

  $effect(() => {
    const _ = transcriptHistory;
    if (scrollEl && !showHotkeys) {
      scrollEl.scrollTop = scrollEl.scrollHeight;
    }
  });

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      const text = chatText.trim();
      if (text && onSendChat) {
        onSendChat(text);
        chatText = "";
      }
    }
  }

  function handleSendClick() {
    const text = chatText.trim();
    if (text && onSendChat) {
      onSendChat(text);
      chatText = "";
    }
  }

  function truncateText(text: string, maxLen = 60) {
    if (text.length > maxLen) {
      return text.substring(0, maxLen) + '...';
    }
    return text;
  }
</script>

<div class="conv-panel">
  <!-- Header -->
  <div class="header">
    <div class="header-left">
      <span class="title">Live Session</span>
      {#if isListening && !isMockMode}
        <div class="status-badge status-live pulse">● Live</div>
      {:else if isMockMode}
        <div class="status-badge status-mock pulse">● Mock</div>
      {:else}
        <div class="status-badge status-mic-off">● Mic Off</div>
      {/if}

      {#if isPTTHeld}
        <div class="status-badge ptt-badge pulse">🎤 PTT</div>
      {/if}

      {#if includeActiveDocContext && activeDocumentName}
        <div class="status-badge rag-badge">📄 {activeDocumentName}</div>
      {/if}

      {#if ragSources && ragSources.length > 0}
        <div class="status-badge rag-src-badge">📎 {ragSources.length} src</div>
      {/if}
    </div>

    <div class="header-right">
      <button class="icon-btn" title="STAR Method" onclick={() => onStarMethod?.()}>
        <span class="material-symbols-outlined">star</span>
      </button>
      <button class="icon-btn" title="Catch Me Up" onclick={() => onCatchMeUp?.()}>
        <span class="material-symbols-outlined">history</span>
      </button>
      <button class="icon-btn" class:active={showHotkeys} title="Hotkeys" onclick={() => showHotkeys = !showHotkeys}>
        <span class="material-symbols-outlined">keyboard</span>
      </button>
    </div>
  </div>

  <!-- Body -->
  <div class="body-scroll hide-scrollbar" bind:this={scrollEl}>
    {#if showHotkeys}
      <div class="hotkeys-list">
        <div class="hotkey-row">
          <span>Toggle Hotkeys Panel</span>
          <kbd>Ctrl+/</kbd>
        </div>
        <div class="hotkey-row">
          <span>Push to Talk</span>
          <kbd>Ctrl+Shift+Space</kbd>
        </div>
        <div class="hotkey-row">
          <span>Screenshot Vision</span>
          <kbd>Ctrl+Shift+S</kbd>
        </div>
        <div class="hotkey-row">
          <span>Send Transcript Chip 1-6</span>
          <kbd>Ctrl+Shift+1...6</kbd>
        </div>
        <div class="hotkey-row">
          <span>Scroll Answer Down / Up</span>
          <kbd>Ctrl+Shift+↓/↑</kbd>
        </div>
        <div class="hotkey-row">
          <span>Toggle Stealth Click-through</span>
          <kbd>Ctrl+Alt+M</kbd>
        </div>
        <div class="hotkey-row">
          <span>Session Report</span>
          <kbd>Ctrl+Shift+E</kbd>
        </div>
        <div class="hotkey-row">
          <span>Panic Clear / Hide</span>
          <kbd>Ctrl+Shift+X</kbd>
        </div>
      </div>
    {:else}
      {#if transcriptHistory.length === 0}
        <div class="empty-state">
          <span class="material-symbols-outlined ghost-icon">hearing</span>
          <p>Waiting for the interview to begin…</p>
        </div>
      {:else}
        {#each transcriptHistory as item}
          {#if item.role === 'interviewer'}
            <div class="bubble bubble-interviewer">
              <div class="bubble-label">Interviewer</div>
              <div class="text-content">{item.text}</div>
            </div>
          {:else if item.role === 'candidate'}
            <div class="bubble bubble-candidate">
              <div class="bubble-label">You</div>
              <div class="text-content">{item.text}</div>
            </div>
          {:else if item.role === 'assistant' || item.role === 'ai'}
            <div class="bubble bubble-assistant">
              <div class="bubble-label">AI</div>
              {#if isThinking && !item.text}
                <div class="thinking-pulse"></div>
              {:else if item.text}
                <div class="text-content truncated-preview">{item.text}</div>
                <span class="see-full">→ See full answer</span>
              {/if}
            </div>
          {/if}
        {/each}
      {/if}
    {/if}
  </div>

  <!-- Suggestion Chips -->
  {#if pendingTranscripts.length > 0 && !showHotkeys}
    <div class="chips-row hide-scrollbar">
      {#each pendingTranscripts as chip (chip.id)}
        <div class="chip">
          <button class="chip-text" onclick={() => onSendChip?.(chip)} title={chip.text}>
            {truncateText(chip.text)}
          </button>
          <button class="chip-close" onclick={() => onDismissChip?.(chip.id)}>
            <span class="material-symbols-outlined" style="font-size: 14px;">close</span>
          </button>
        </div>
      {/each}
      {#if pendingTranscripts.length > 1}
        <button class="clear-all" onclick={() => onClearChips?.()}>
          Clear All
        </button>
      {/if}
    </div>
  {/if}

  <!-- Footer Input -->
  {#if !showHotkeys}
    <div class="footer">
      <textarea
        bind:value={chatText}
        rows="3"
        placeholder="Ask a follow-up… (Shift+Enter for newline)"
        onkeydown={handleKeydown}
        disabled={isThinking}
      ></textarea>
      <button
        class="send-btn"
        onclick={handleSendClick}
        disabled={isThinking || !chatText.trim()}
      >
        <span class="material-symbols-outlined" style="font-size: 16px;">send</span>
      </button>
    </div>
  {/if}
</div>

<style>
  .conv-panel {
    --conv-bg: rgba(12, 12, 12, 0.95);
    --conv-border: rgba(255, 255, 255, 0.06);
    --primary: #4ade80;

    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
    background: var(--conv-bg);
    border-right: 1px solid var(--conv-border);
  }

  .header {
    height: 44px;
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    padding: 0 10px;
    border-bottom: 1px solid var(--conv-border);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 8px;
    overflow: hidden;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 4px;
    flex-shrink: 0;
  }

  .title {
    font-variant: small-caps;
    color: rgba(255, 255, 255, 0.7);
    font-size: 12px;
    font-weight: 600;
    letter-spacing: 0.5px;
    white-space: nowrap;
  }

  .status-badge {
    font-size: 10px;
    padding: 2px 6px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    gap: 4px;
    font-weight: 600;
    white-space: nowrap;
  }

  .status-live {
    background: rgba(74, 222, 128, 0.1);
    color: #4ade80;
    border: 1px solid rgba(74, 222, 128, 0.2);
  }

  .status-mock {
    background: rgba(251, 191, 36, 0.1);
    color: #fbbf24;
    border: 1px solid rgba(251, 191, 36, 0.2);
  }

  .status-mic-off {
    background: rgba(239, 68, 68, 0.1);
    color: #ef4444;
    border: 1px solid rgba(239, 68, 68, 0.2);
  }

  .pulse {
    animation: pulse 2s infinite;
  }

  @keyframes pulse {
    0% { opacity: 1; }
    50% { opacity: 0.5; }
    100% { opacity: 1; }
  }

  .ptt-badge {
    background: rgba(74, 222, 128, 0.1);
    color: #4ade80;
    border: 1px solid rgba(74, 222, 128, 0.2);
  }

  .rag-badge {
    background: rgba(74, 222, 128, 0.05);
    color: #4ade80;
    border: 1px solid rgba(74, 222, 128, 0.1);
  }

  .rag-src-badge {
    background: rgba(255, 255, 255, 0.05);
    color: rgba(255, 255, 255, 0.7);
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .icon-btn {
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    color: rgba(255, 255, 255, 0.6);
    background: transparent;
    border: none;
    cursor: pointer;
    transition: all 0.2s;
  }

  .icon-btn .material-symbols-outlined {
    font-size: 16px;
  }

  .icon-btn:hover {
    color: rgba(255, 255, 255, 0.9);
    background: rgba(255, 255, 255, 0.1);
  }

  .icon-btn.active {
    color: #4ade80;
    background: rgba(74, 222, 128, 0.1);
  }

  .body-scroll {
    flex-grow: 1;
    overflow-y: auto;
    padding: 12px 10px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .bubble {
    padding: 10px;
    border-radius: 8px;
    max-width: 90%;
    font-size: 13px;
    line-height: 1.4;
    color: rgba(255, 255, 255, 0.9);
    position: relative;
  }

  .bubble-interviewer {
    align-self: flex-start;
    background: rgba(30, 41, 59, 0.6);
    border-left: 3px solid rgba(148, 163, 184, 0.4);
  }

  .bubble-candidate {
    align-self: flex-end;
    background: rgba(20, 40, 25, 0.6);
    border-left: 3px solid rgba(74, 222, 128, 0.4);
  }

  .bubble-assistant {
    align-self: flex-start;
    background: rgba(10, 20, 14, 0.7);
    border-left: 3px solid rgba(74, 222, 128, 0.6);
    width: 100%;
    max-width: 100%;
  }

  .bubble-label {
    font-size: 9px;
    text-transform: uppercase;
    font-weight: 700;
    margin-bottom: 4px;
    letter-spacing: 0.5px;
  }

  .bubble-interviewer .bubble-label {
    color: rgba(255, 255, 255, 0.5);
  }

  .bubble-candidate .bubble-label {
    color: #4ade80;
  }

  .bubble-assistant .bubble-label {
    color: #4ade80;
  }

  .text-content {
    white-space: pre-wrap;
    word-break: break-word;
  }

  .truncated-preview {
    display: -webkit-box;
    -webkit-line-clamp: 4;
    -webkit-box-orient: vertical;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .see-full {
    font-size: 10px;
    color: rgba(74, 222, 128, 0.7);
    margin-top: 6px;
    display: inline-block;
  }

  .thinking-pulse {
    display: inline-block;
    width: 8px;
    height: 8px;
    background-color: rgba(255, 255, 255, 0.5);
    border-radius: 50%;
    animation: pulse 1s infinite;
  }

  .ghost-icon {
    font-size: 48px;
    opacity: 0.2;
    margin-bottom: 12px;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: rgba(255, 255, 255, 0.4);
    font-size: 13px;
    text-align: center;
  }

  .chips-row {
    padding: 6px 10px;
    border-top: 1px solid var(--conv-border);
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
    max-height: 150px;
    overflow-y: auto;
  }

  .chip {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    padding: 8px 10px;
    font-size: 12px;
    color: rgba(255, 255, 255, 0.85);
  }

  .chip-text {
    cursor: pointer;
    overflow: hidden;
    text-overflow: ellipsis;
    background: none;
    border: none;
    color: inherit;
    font: inherit;
    padding: 0;
    text-align: left;
  }

  .chip-text:hover {
    color: #4ade80;
  }

  .chip-close {
    display: flex;
    align-items: center;
    justify-content: center;
    margin-left: 4px;
    border-radius: 50%;
    cursor: pointer;
    color: rgba(255, 255, 255, 0.5);
    background: none;
    border: none;
    padding: 0;
  }

  .chip-close:hover {
    color: #ef4444;
    background: rgba(255, 255, 255, 0.1);
  }

  .clear-all {
    font-size: 10px;
    color: #ef4444;
    cursor: pointer;
    opacity: 0.8;
    margin-left: auto;
    background: none;
    border: none;
    padding: 0;
    white-space: nowrap;
  }

  .clear-all:hover {
    opacity: 1;
    text-decoration: underline;
  }

  .footer {
    padding: 8px 10px;
    border-top: 1px solid var(--conv-border);
    position: relative;
  }

  textarea {
    width: 100%;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    color: white;
    padding: 8px 10px;
    padding-right: 40px; /* space for send button */
    font-size: 13px;
    outline: none;
    resize: vertical;
    min-height: 84px;
    max-height: 220px;
  }

  textarea:focus {
    border-color: rgba(74, 222, 128, 0.5);
    box-shadow: 0 0 0 2px rgba(74, 222, 128, 0.1);
  }

  textarea:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .send-btn {
    position: absolute;
    bottom: 16px;
    right: 16px;
    width: 32px;
    height: 32px;
    border-radius: 50%;
    background: rgba(74, 222, 128, 0.15);
    color: var(--primary, #4ade80);
    border: 1px solid rgba(74, 222, 128, 0.3);
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.2s;
  }

  .send-btn:hover:not(:disabled) {
    background: rgba(74, 222, 128, 0.25);
  }

  .send-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Hotkeys Overlay Styles */
  .hotkeys-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 8px;
    padding: 0.75rem 1rem;
  }

  .hotkey-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.45rem 0;
    font-size: 0.82rem;
    color: rgba(255, 255, 255, 0.85);
    border-bottom: 1px solid rgba(255, 255, 255, 0.04);
  }

  .hotkey-row:last-child {
    border-bottom: none;
  }

  .hotkey-row kbd {
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 4px;
    padding: 0.2rem 0.5rem;
    font-family: 'JetBrains Mono', 'Fira Code', monospace;
    font-size: 0.75rem;
    color: var(--primary, #4ade80);
  }

  .hide-scrollbar::-webkit-scrollbar {
    display: none;
  }
  .hide-scrollbar {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }
</style>