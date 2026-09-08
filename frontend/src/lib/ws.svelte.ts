import { authState, supabase } from '$lib/auth.svelte';
import { invoke } from '@tauri-apps/api/core';

import { listen } from '@tauri-apps/api/event';

export const wsState = $state({
  transcript: "",
  response: "",
  isListening: false,
  isThinking: false,
  isAnalyzingScreen: false,
  isConnected: false,
  error: null as string | null,
  sessionExpired: false,
  ragSources: [] as string[],
  isPTTHeld: false,
  pttMode: false,
  pendingTranscripts: [] as Array<{ id: number; text: string; speaker?: "interviewer" | "candidate" | null }>,
  plan: "unknown"
});

let ws: WebSocket | null = null;
let chipIdCounter = 0;
let listenersInitialized = false;

function initListeners() {
  listen("ptt-start", () => {
    wsState.isPTTHeld = true;
    fetch('http://127.0.0.1:8765/ptt/start', { method: 'POST' }).catch(console.error);
  });

  listen("ptt-stop", () => {
    wsState.isPTTHeld = false;
    fetch('http://127.0.0.1:8765/ptt/stop', { method: 'POST' }).catch(console.error);
  });

  listen("panic-clear", () => {
    wsState.transcript = "";
    wsState.response = "";
    wsState.isThinking = false;
    wsState.ragSources = [];
    wsState.pendingTranscripts = [];
    fetch('http://127.0.0.1:8765/history/clear', { method: 'POST' }).catch(console.error);
  });
}
let retryDelay = 500;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let intentionalClose = false;
let lastUrl: string | undefined;

export function connect(url?: string): void {
  // Clear any existing reconnect timer
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }

  // Idempotent: don't reconnect if we are already connected to the same URL or opening
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
    return;
  }


  let storedUrl = null;
  if (typeof localStorage !== 'undefined') {
    storedUrl = localStorage.getItem('backend_url');
    // Auto-migrate stale port 8000 → 8765
    if (storedUrl && storedUrl.includes(':8000')) {
      storedUrl = storedUrl.replace(':8000', ':8765');
      localStorage.setItem('backend_url', storedUrl);
    }
  }

  let defaultUrl = "ws://127.0.0.1:8765/ws";
  if (typeof window !== 'undefined') {
    // Check if we are in a normal browser (not Tauri)
    const isBrowser = typeof (window as any).__TAURI_INTERNALS__ === 'undefined';
    if (isBrowser) {
      const host = window.location.host;
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      defaultUrl = `${wsProtocol}//${host}/ws`;
    }
  }

  let targetUrl = url ?? storedUrl ?? import.meta.env.VITE_WS_URL ?? defaultUrl;
  if (targetUrl.startsWith('http://')) {
    targetUrl = targetUrl.replace('http://', 'ws://');
  } else if (targetUrl.startsWith('https://')) {
    targetUrl = targetUrl.replace('https://', 'wss://');
  } else if (!targetUrl.includes('://')) {
    targetUrl = 'ws://' + targetUrl;
  }

  // Always ensure the WebSocket path ends with /ws
  if (!targetUrl.endsWith('/ws')) {
    targetUrl = targetUrl.replace(/\/?$/, '/ws');
  }

  lastUrl = targetUrl;
  intentionalClose = false;
  wsState.error = null;

  if (!listenersInitialized) {
    initListeners();
    listenersInitialized = true;
  }

  try {
    ws = new WebSocket(targetUrl);
  } catch (e: any) {
    wsState.error = e?.message || "Failed to create WebSocket";
    scheduleReconnect();
    return;
  }

  ws.onopen = async () => {
    wsState.isConnected = true;
    wsState.isListening = true;
    retryDelay = 500; // Reset exponential backoff

    if (authState.accessToken !== null) {
      try {
        const machine_id: string = await invoke('get_machine_id');
        ws?.send(JSON.stringify({ type: "auth", token: authState.accessToken, machine_id }));
      } catch (err) {
        console.error("Failed to get machine id", err);
      }
    }
  };

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      switch (data.type) {
        case "transcript":
          wsState.transcript = data.text;

          // Accumulate as a clickable chip (deduplicate identical text)
          const existingIdx = wsState.pendingTranscripts.findIndex(
            (c) => c.text === data.text
          );
          if (existingIdx >= 0) {
            // Update in place (keeps position, refreshes)
            wsState.pendingTranscripts[existingIdx] = {
              id: wsState.pendingTranscripts[existingIdx].id,
              text: data.text,
              speaker: data.speaker ?? null
            };
          } else {
            wsState.pendingTranscripts.push({
              id: chipIdCounter++,
              text: data.text,
              speaker: data.speaker ?? null
            });
            // Keep max 6 chips — drop oldest
            if (wsState.pendingTranscripts.length > 6) {
              wsState.pendingTranscripts.shift();
            }
          }
          break;
        case "session_expired":
          wsState.sessionExpired = true;
          wsState.isListening = false;
          break;
        case "message_start":
          wsState.response = "";
          wsState.isThinking = true;
          wsState.ragSources = [];
          wsState.pendingTranscripts = [];
          break;
        case "rag_sources":
          wsState.ragSources = data.sources;
          break;
        case "token":
          wsState.response += data.text;
          break;
        case "end":
          wsState.isThinking = false;
          break;
        case "plan":
          wsState.plan = data.plan;
          break;
        case "error":
          if (data.message && data.message.includes("Token expired") && supabase) {
            console.log("Token expired, refreshing session...");
            supabase.auth.refreshSession().then(() => {
              disconnect();
              connect(lastUrl);
            });
          } else {
            wsState.error = data.message;
            wsState.isThinking = false;
          }
          break;
        default:
          console.warn("Unknown message type:", data.type);
      }
    } catch (err) {
      console.error("Failed to parse WebSocket message", err);
    }
  };

  ws.onclose = (event) => {
    wsState.isConnected = false;
    wsState.isListening = false;
    ws = null;

    if (!intentionalClose && event.code !== 1000) {
      scheduleReconnect();
    }
  };

  ws.onerror = () => {
    wsState.error = "WebSocket connection error";
    // onclose is called immediately after onerror in most cases,
    // so reconnect will be scheduled there.
  };
}

export function disconnect(): void {
  intentionalClose = true;
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (ws) {
    ws.close(1000);
    ws = null;
  }
  wsState.isConnected = false;
  wsState.isListening = false;
}

function scheduleReconnect(): void {
  if (intentionalClose) return;

  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
  }

  reconnectTimer = setTimeout(() => {
    connect(lastUrl);
  }, retryDelay);

  // Exponential backoff, max 5000ms
  retryDelay = Math.min(retryDelay * 2, 5000);
}

export function reconnect(url: string): void {
  disconnect();
  connect(url);
}

export function sendChat(text: string): void {
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: "chat", text }));
  } else {
    console.error("WebSocket is not connected. Cannot send chat message.");
  }
}


export function sendChip(chip: { id: number; text: string; speaker?: "interviewer" | "candidate" | null }): void {
  sendChat(chip.text);
  wsState.pendingTranscripts = wsState.pendingTranscripts.filter(
    (c) => c.id !== chip.id
  );
}

export function dismissChip(chipId: number): void {
  wsState.pendingTranscripts = wsState.pendingTranscripts.filter(
    (c) => c.id !== chipId
  );
}

export function clearAllChips(): void {
  wsState.pendingTranscripts = [];
}
