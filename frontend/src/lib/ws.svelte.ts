import { accessToken } from '$lib/auth';

export const wsState = $state({
  transcript: "",
  response: "",
  isListening: false,
  isThinking: false,
  isConnected: false,
  error: null as string | null
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

  const targetUrl = url ?? import.meta.env.VITE_WS_URL ?? "ws://127.0.0.1:8765/ws";
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

  ws.onopen = () => {
    wsState.isConnected = true;
    wsState.isListening = true;
    retryDelay = 500; // Reset exponential backoff

    if (accessToken !== null) {
      ws?.send(JSON.stringify({ type: "auth", token: accessToken }));
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
          break;
        case "token":
          wsState.response += data.text;
          break;
        case "end":
          wsState.isThinking = false;
          break;
        case "error":
          wsState.error = data.message;
          wsState.isThinking = false;
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
