import { cloudAuthState } from './auth.svelte';

export function getApiUrl(): string {
  if (typeof import.meta !== 'undefined' && import.meta.env && import.meta.env.VITE_API_BASE) {
    return import.meta.env.VITE_API_BASE as string;
  }
  return 'http://127.0.0.1:8765';
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
