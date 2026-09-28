import { cloudAuthState } from './auth.svelte';

export function getApiUrl(): string {
  const isCloud = typeof import.meta !== 'undefined' && import.meta.env && import.meta.env.VITE_BUILD_FLAVOR === 'cloud';

  if (isCloud) {
    if (typeof import.meta !== 'undefined' && import.meta.env && import.meta.env.VITE_API_BASE) {
      return import.meta.env.VITE_API_BASE as string;
    }
    return 'https://ai.krushnavardhan.workers.dev';
  }

  throw new Error("Local Edition uses Wails IPC, not HTTP/WS API.");
}

export function getWsUrl(): string {
  const baseUrl = getApiUrl();
  if (baseUrl.startsWith('https://')) {
    return baseUrl.replace('https://', 'wss://');
  }
  if (baseUrl.startsWith('http://')) {
    return baseUrl.replace('http://', 'ws://');
  }
  return baseUrl;
}

export async function apiFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  const url = typeof input === 'string' ? input : input instanceof URL ? input.toString() : input.url;

  const options = { ...init };
  const headers = new Headers(options.headers || {});

  if (url.startsWith(getApiUrl()) && cloudAuthState.apiKey) {
    headers.set('Authorization', `Bearer ${cloudAuthState.apiKey}`);
  }

  options.headers = headers;

  return fetch(input, options);
}

// ── NEW TYPED REST CLIENT FOR LOCAL APP ──────────────────────────────

let baseURL = (typeof import.meta !== 'undefined' && import.meta.env && import.meta.env.VITE_API_BASE_URL) || "http://127.0.0.1:8080";

// Listen for the port emitted by Wails on startup
if (typeof window !== 'undefined' && (window as any).runtime?.EventsOn) {
  (window as any).runtime.EventsOn("api_port", (port: number) => {
    baseURL = `http://127.0.0.1:${port}`;
    console.log("[API] Port set to", baseURL);
  });
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${baseURL}${path}`);
  if (!res.ok) throw new Error(`GET ${path} → ${res.status}`);
  const body = await res.json();
  return body.data as T;
}

async function post<T>(path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${baseURL}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) throw new Error(`POST ${path} → ${res.status}`);
  const json = await res.json();
  if (json.error) throw new Error(json.error);
  return json.data as T;
}

export function connectWebSocket(onMessage: (type: string, payload: unknown) => void): WebSocket {
  const wsURL = baseURL.replace("http://", "ws://") + "/ws";
  const ws = new WebSocket(wsURL);
  ws.onmessage = (ev) => {
    try {
      const { type, payload } = JSON.parse(ev.data);
      onMessage(type, payload);
    } catch {}
  };
  ws.onclose = () => {
    // Auto-reconnect after 2 seconds
    setTimeout(() => connectWebSocket(onMessage), 2000);
  };
  return ws;
}

export const api = {
  getState:       () => get<Record<string, unknown>>("/api/v1/state"),
  getSystemStatus:() => get<Record<string, unknown>>("/api/v1/status"),
  getCacheStats:  () => get<Record<string, unknown>>("/api/v1/cache/stats"),
  getCacheItems:  () => get<unknown[]>("/api/v1/cache/items"),
  getIndexedPaths:() => get<string[]>("/api/v1/indexed-paths"),
  getBuddyURL:    () => get<string>("/api/v1/buddy/url"),
  getIDEState:    () => get<Record<string, unknown>>("/api/v1/ide/state"),

  toggleMic:          () => post<boolean>("/api/v1/mic/toggle"),
  flushBuffer:        () => post<null>("/api/v1/llm/flush"),
  appendToBuffer:     (text: string) => post<null>("/api/v1/llm/append", { text }),
  sendChat:           (text: string) => post<null>("/api/v1/llm/chat", { text }),
  clearState:         () => post<string>("/api/v1/session/clear"),
  exportSession:      () => post<string>("/api/v1/session/export"),
  endSession:         () => post<Record<string, unknown>>("/api/v1/session/end"),
  summarizeTranscript:() => post<null>("/api/v1/session/summarize"),
  setAppMode:         (mode: string) => post<null>("/api/v1/mode/app", { mode }),
  setAudioMode:       (mode: string) => post<null>("/api/v1/mode/audio", { mode }),
  setManualMode:      (enabled: boolean) => post<string>("/api/v1/mode/manual", { enabled }),
  setRawMode:         (enabled: boolean) => post<string>("/api/v1/mode/raw", { enabled }),
  setMockMode:        (enabled: boolean, tts: boolean) => post<string>("/api/v1/mode/mock", { enabled, tts }),
  clearCache:         () => post<null>("/api/v1/cache/clear"),
  deleteCacheItems:   (ids: string[]) => post<null>("/api/v1/cache/delete", { ids }),
  startBuddyMode:     () => post<null>("/api/v1/buddy/start"),
  stopBuddyMode:      () => post<null>("/api/v1/buddy/stop"),
  captureScreen:      () => post<string>("/api/v1/screen/capture"),
  analyzeVision:      (image: string, prompt: string) => post<null>("/api/v1/screen/analyze", { image, prompt }),
};
