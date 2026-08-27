import { authState, supabase } from '$lib/auth.svelte';
import { invoke } from '@tauri-apps/api/core';

export const wsState = $state({
  transcript: "",
  response: "",
  isListening: false,
  isThinking: false,
  isAnalyzingScreen: false,
  isConnected: false,
  error: null as string | null,
  ragSources: [] as string[]
});

let ws: WebSocket | null = null;
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

  lastUrl = targetUrl;
  intentionalClose = false;
  wsState.error = null;

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
          wsState.response = "";
          wsState.isThinking = true;
          wsState.ragSources = [];
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
