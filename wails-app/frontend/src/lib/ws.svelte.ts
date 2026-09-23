import { apiFetch, getApiUrl, getWsUrl } from './api';
import { authState, supabase } from '$lib/auth.svelte';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { GetState } from '../../wailsjs/go/main/App';

const isCloud = typeof import.meta !== 'undefined' && import.meta.env && import.meta.env.VITE_BUILD_FLAVOR === 'cloud';

export const wsState = $state({
  transcript: "",
  response: "",
  isListening: !isCloud,
  isThinking: false,
  isAnalyzingScreen: false,
  isConnected: !isCloud,
  error: null as string | null,
  sessionExpired: false,
  ragSources: [] as string[],
  isPTTHeld: false,
  pttMode: false,
  pendingTranscripts: [] as Array<{ id: number; text: string; speaker?: "interviewer" | "candidate" | null }>,
  plan: "unknown",
  isMockMode: false,
  summaryResults: {} as Record<string, string>,
  isSummarizing: {} as Record<string, boolean>,
  rawMode: false,
  manualMode: false,
  transcriptHistory: [] as Array<{ role: string; text: string; answer?: string }>,
  pollCount: 0,
  pollError: "none",
  cacheStats: { cached_pairs: 0, estimated_tokens_saved: 0 },
  downloadTask: null as { component: string, progress: number } | null
});

export async function toggleManualMode(): Promise<void> {
  wsState.manualMode = !wsState.manualMode;
  try {
    await (window as any).go.main.App.SetManualMode(wsState.manualMode);
  } catch (e) {
    console.error('Failed to set manual mode:', e);
    wsState.manualMode = !wsState.manualMode; // Rollback on error
  }
}

export async function toggleRawMode(): Promise<void> {
  wsState.rawMode = !wsState.rawMode;
  try {
    await (window as any).go.main.App.SetRawMode(wsState.rawMode);
  } catch (e) {
    console.error('Failed to set raw mode:', e);
    wsState.rawMode = !wsState.rawMode; // Rollback on error
  }
}

let ws: WebSocket | null = null;
let chipIdCounter = 0;
let listenersInitialized = false;

// Immediately start polling if local mode
if (!isCloud) {
  setInterval(async () => {
    wsState.pollCount++;
    try {
      let state: any = null;
      if (typeof GetState === 'function') {
        state = await GetState();
      } else if ((window as any)?.go?.main?.App?.GetState) {
        state = await (window as any).go.main.App.GetState();
      } else {
        wsState.pollError = "no_binding";
        return;
      }
      if (state) {
        wsState.pollError = "ok";
        if (typeof state.transcript === 'string') {
          wsState.transcript = state.transcript;
        }
        if (typeof state.response === 'string') {
          wsState.response = state.response;
        }
        if (typeof state.thinking === 'boolean') {
          wsState.isThinking = state.thinking;
        }
        if (typeof state.cached_pairs === 'number') {
          wsState.cacheStats = {
            cached_pairs: state.cached_pairs,
            estimated_tokens_saved: state.estimated_tokens_saved ?? state.cached_pairs * 250
          };
        }
        if (typeof state.is_listening === 'boolean') {
          wsState.isListening = state.is_listening;
        }
      }
    } catch (err: any) {
      wsState.pollError = err?.message || String(err);
    }
  }, 200);
}

function handleTranscript(data: any) {
  wsState.transcript = data.text;

  // Accumulate rolling transcript history (max 10 entries)
  if (data.text && (wsState.rawMode || data.text.trim().length > 3)) {
    // Avoid duplicating the last entry
    const last = wsState.transcriptHistory[wsState.transcriptHistory.length - 1];
    if (last?.text !== data.text) {
      wsState.transcriptHistory = [
        ...wsState.transcriptHistory,
        { role: data.speaker ?? 'interviewer', text: data.text }
      ];
      if (wsState.transcriptHistory.length > 10) {
        wsState.transcriptHistory = wsState.transcriptHistory.slice(1);
      }
    }
  }

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
}

function initListeners() {
  console.log('[WS] initListeners() called — registering Wails EventsOn handlers');

  // Fallback to imported EventsOn if window.runtime is missing (e.g. dev mode without Wails)
  const onEvent = (window as any).runtime?.EventsOn || EventsOn;
  const onEventAll = (window as any).runtime?.EventsOnAll || null;

  if (onEventAll) {
    onEventAll((eventName: string, ...data: any[]) => {
      // Dump everything to the transcript string just so we can see it on screen!
      wsState.transcript = `[EVENT] ${eventName}: ${JSON.stringify(data)}`;
    });
  }

  onEvent("ptt-start", () => {
    wsState.isPTTHeld = true;
    apiFetch(`${getApiUrl()}/ptt/start`, { method: 'POST' }).catch(console.error);
  });

  onEvent("ptt-stop", () => {
    wsState.isPTTHeld = false;
    apiFetch(`${getApiUrl()}/ptt/stop`, { method: 'POST' }).catch(console.error);
  });

  onEvent("panic-clear", () => {
    wsState.transcript = "";
    wsState.response = "";
    wsState.isThinking = false;
    wsState.ragSources = [];
    wsState.pendingTranscripts = [];
    wsState.transcriptHistory = [];
    apiFetch(`${getApiUrl()}/history/clear`, { method: 'POST' }).catch(console.error);
  });

  onEvent("on_transcript", (data: any) => {
    console.log('[WS] on_transcript fired:', data);
    handleTranscript(data);
    console.log('[WS] wsState.transcript is now:', wsState.transcript);
  });

  onEvent("on_response_start", () => {
    console.log('[WS] on_response_start fired');
    wsState.response = "";
    wsState.isThinking = true;
    wsState.ragSources = [];
  });

  onEvent("on_response_token", (data: any) => {
    console.log('[WS] on_response_token:', data?.text?.slice(0, 20));
    wsState.response += data.text;
    wsState.isThinking = true;
  });

  onEvent("on_summary_start", (data: any) => {
    wsState.isSummarizing[data.id] = true;
    wsState.summaryResults[data.id] = "";
  });

  onEvent("on_summary_token", (data: any) => {
    if (!wsState.summaryResults[data.id]) {
      wsState.summaryResults[data.id] = "";
    }
    wsState.summaryResults[data.id] += data.text;
  });

  onEvent("on_summary_end", (data: any) => {
    wsState.isSummarizing[data.id] = false;
  });

  onEvent("on_download_progress", (data: any) => {
    if (data.progress >= 100) {
      wsState.downloadTask = null;
    } else {
      wsState.downloadTask = { component: data.component, progress: data.progress };
    }
  });

  onEvent("on_response_end", () => {
    console.log('[WS] on_response_end fired. Final response length:', wsState.response.length);
    wsState.isThinking = false;

    // Store the completed answer against the most recent unanswered transcript
    if (wsState.response) {
      for (let i = wsState.transcriptHistory.length - 1; i >= 0; i--) {
        const entry = wsState.transcriptHistory[i];
        if (!entry.answer) {
          wsState.transcriptHistory[i] = { ...entry, answer: wsState.response };
          break;
        }
      }
    }
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

  // Bypass WebSocket connection entirely if running in Wails (Local Edition)
  if (!isCloud) {
    wsState.isConnected = true;
    wsState.isListening = true;
    
    if (!listenersInitialized) {
      initListeners();
      listenersInitialized = true;
    }
    return;
  }


  let defaultUrl = getWsUrl();
  let targetUrl = url ?? import.meta.env.VITE_WS_URL ?? defaultUrl;
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

  const customProvider = localStorage.getItem("custom_provider") || "auto";
  const customApiKey = localStorage.getItem("custom_api_key") || "";
  const openRouterModel = localStorage.getItem("openrouter_model") || "anthropic/claude-3.5-sonnet:beta";

  let params = new URLSearchParams();
  if (customProvider !== "auto") {
    params.append("custom_provider", customProvider);
    if (customApiKey) params.append("custom_key", customApiKey);
    if (customProvider === "openrouter" && openRouterModel) {
      params.append("openrouter_model", openRouterModel);
    }
    const sep = targetUrl.includes('?') ? '&' : '?';
    targetUrl = `${targetUrl}${sep}${params.toString()}`;
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
        const machine_id: string = await (window as any).go.main.App.GetMachineId();
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
          handleTranscript(data);
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
        case "mock_audio":
          if (data.data) {
            try {
              const audio = new Audio(`data:audio/mp3;base64,${data.data}`);
              audio.play().catch(e => console.error("Failed to play mock audio:", e));
            } catch (e) {
              console.error("Failed to parse mock audio:", e);
            }
          }
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
        case "device_limit_reached":
          wsState.error = `Device limit reached (max ${data.max}). Visit parakeet.app/dashboard/devices to manage your devices.`;
          break;
        case "already_active":
          wsState.error = "Another session is active. Close your other device or wait 5 minutes.";
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

export function toggleMockMode(enabled: boolean): void {
  wsState.isMockMode = enabled;
  if (ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: "mock_mode_toggle", enabled }));
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
