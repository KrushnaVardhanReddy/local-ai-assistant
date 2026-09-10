let backendUrl = 'http://127.0.0.1:8765';
let transcriptLines = [];

// DOM Elements
const settingsBtn = document.getElementById('settings-btn');
const settingsPanel = document.getElementById('settings-panel');
const backendUrlInput = document.getElementById('backend-url');
const saveSettingsBtn = document.getElementById('save-settings-btn');

const tabBtns = document.querySelectorAll('.tab-btn');
const tabContents = document.querySelectorAll('.tab-content');

const transcriptLog = document.getElementById('transcript-log');
const copyTranscriptBtn = document.getElementById('copy-transcript-btn');
const clearTranscriptBtn = document.getElementById('clear-transcript-btn');

const chatLog = document.getElementById('chat-log');
const chatForm = document.getElementById('chat-form');
const chatInput = document.getElementById('chat-input');
const sendBtn = document.getElementById('send-btn');

// Initialize
chrome.storage.local.get(['backend_url'], (result) => {
  if (result.backend_url) {
    backendUrl = result.backend_url;
  }
  backendUrlInput.value = backendUrl;
});

// Settings Toggle
settingsBtn.addEventListener('click', () => {
  settingsPanel.classList.toggle('hidden');
});

saveSettingsBtn.addEventListener('click', () => {
  const newUrl = backendUrlInput.value.trim();
  if (newUrl) {
    backendUrl = newUrl;
    chrome.storage.local.set({ backend_url: backendUrl }, () => {
      settingsPanel.classList.add('hidden');
    });
  }
});

// Tabs
tabBtns.forEach(btn => {
  btn.addEventListener('click', () => {
    tabBtns.forEach(b => b.classList.remove('active'));
    tabContents.forEach(c => c.classList.remove('active'));

    btn.classList.add('active');
    document.getElementById(`${btn.dataset.tab}-tab`).classList.add('active');
  });
});

// Transcript Handling
window.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'BARNOWL_CAPTION') {
    const { text, speaker } = event.data;
    appendTranscript(speaker, text);
  }
});

function appendTranscript(speaker, text) {
  const line = `${speaker}: ${text}`;
  transcriptLines.push(line);

  const div = document.createElement('div');
  div.className = 'transcript-line';
  div.innerHTML = `<span class="speaker-name">${speaker}</span> <span class="caption-text">${text}</span>`;
  transcriptLog.appendChild(div);

  transcriptLog.scrollTop = transcriptLog.scrollHeight;
}

copyTranscriptBtn.addEventListener('click', () => {
  const text = transcriptLines.join('\n');
  navigator.clipboard.writeText(text).then(() => {
    const originalText = copyTranscriptBtn.innerText;
    copyTranscriptBtn.innerText = 'Copied!';
    setTimeout(() => { copyTranscriptBtn.innerText = originalText; }, 2000);
  });
});

clearTranscriptBtn.addEventListener('click', () => {
  transcriptLines = [];
  transcriptLog.innerHTML = '';
});

// Simple markdown converter fallback
function simpleMarkdown(text) {
  let html = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  // Bold
  html = html.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
  // Code block
  html = html.replace(/```([\s\S]*?)```/g, '<pre><code>$1</code></pre>');
  // Inline code
  html = html.replace(/`(.*?)`/g, '<code>$1</code>');
  // Newlines
  html = html.replace(/\n/g, '<br/>');
  return html;
}

// Chat Handling
function appendMessage(role, content) {
  const msgDiv = document.createElement('div');
  msgDiv.className = `message ${role}-msg`;

  const bubbleDiv = document.createElement('div');
  bubbleDiv.className = 'msg-bubble';

  if (role === 'assistant') {
    // Render basic Markdown for AI responses
    bubbleDiv.innerHTML = simpleMarkdown(content);
  } else {
    // Plain text for user
    bubbleDiv.textContent = content;
  }

  msgDiv.appendChild(bubbleDiv);
  chatLog.appendChild(msgDiv);
  chatLog.scrollTop = chatLog.scrollHeight;
  return bubbleDiv; // Return bubble so we can update it during streaming
}

chatForm.addEventListener('submit', async (e) => {
  e.preventDefault();

  const userText = chatInput.value.trim();
  if (!userText) return;

  // Get last 10 lines of transcript for context
  const context = transcriptLines.slice(-10).join('\n');

  chatInput.value = '';
  appendMessage('user', userText);

  sendBtn.disabled = true;
  chatInput.disabled = true;

  const aiBubble = appendMessage('assistant', '...');

  try {
    const response = await fetch(`${backendUrl}/chat`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ message: userText, context: context })
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    // Read stream
    const reader = response.body.getReader();
    const decoder = new TextDecoder("utf-8");
    let aiText = "";

    while (true) {
      const { value, done } = await reader.read();
      if (done) break;

      const chunk = decoder.decode(value, { stream: true });
      aiText += chunk;
      aiBubble.innerHTML = simpleMarkdown(aiText);
      chatLog.scrollTop = chatLog.scrollHeight;
    }

  } catch (err) {
    console.error('BarnOwl API Error:', err);
    aiBubble.innerHTML = `<span style="color: var(--danger-color)">Error connecting to BarnOwl backend (${backendUrl}). Make sure the local app is running.</span>`;
  } finally {
    sendBtn.disabled = false;
    chatInput.disabled = false;
    chatInput.focus();
  }
});

// Allow Enter to submit (Shift+Enter for new line)
chatInput.addEventListener('keydown', (e) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    chatForm.dispatchEvent(new Event('submit'));
  }
});
