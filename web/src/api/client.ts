const TOKEN_KEY = 'optimus.token';
const PROJECT_CODE_KEY = 'optimus.currentProjectCode';

let onUnauthorized: (() => void) | undefined;
let currentProjectCode: string | null = window.localStorage.getItem(PROJECT_CODE_KEY);

export function getStoredToken() {
  return window.localStorage.getItem(TOKEN_KEY);
}

export function setStoredToken(token: string) {
  window.localStorage.setItem(TOKEN_KEY, token);
}

export function clearStoredToken() {
  window.localStorage.removeItem(TOKEN_KEY);
}

export function getStoredProjectCode() {
  return window.localStorage.getItem(PROJECT_CODE_KEY);
}

export function setStoredProjectCode(code: string | null) {
  currentProjectCode = code;
  if (code) {
    window.localStorage.setItem(PROJECT_CODE_KEY, code);
    return;
  }
  window.localStorage.removeItem(PROJECT_CODE_KEY);
}

export function setUnauthorizedHandler(handler: () => void) {
  onUnauthorized = handler;
}

export async function apiRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers);
  headers.set('Accept', 'application/json');

  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const token = getStoredToken();
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  if (currentProjectCode) {
    headers.set('X-Project-Code', currentProjectCode);
  }

  const response = await fetch(`/api/v1${path}`, {
    ...options,
    headers,
  });

  if (response.status === 401) {
    clearStoredToken();
    onUnauthorized?.();
  }

  const text = await response.text();
  const data = text ? JSON.parse(text) : undefined;

  if (!response.ok) {
    throw new Error(data?.message || `Request failed with ${response.status}`);
  }

  return data as T;
}
