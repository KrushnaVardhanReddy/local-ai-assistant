import { cloudAuthState } from './auth.svelte';

export function getApiUrl(): string {
  const isCloud = typeof import.meta !== 'undefined' && import.meta.env && import.meta.env.VITE_BUILD_FLAVOR === 'cloud';

  if (isCloud) {
    if (typeof import.meta !== 'undefined' && import.meta.env && import.meta.env.VITE_API_BASE) {
      return import.meta.env.VITE_API_BASE as string;
    }
    return 'https://ai.krushnavardhan.workers.dev';
  }

  // Local Edition now uses standard HTTP/WS API
  return 'http://localhost:8080/api';
}

export function getWsUrl(): string {
  const isCloud = typeof import.meta !== 'undefined' && import.meta.env && import.meta.env.VITE_BUILD_FLAVOR === 'cloud';

  if (!isCloud) {
    return 'ws://localhost:8080/ws';
  }

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
