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
  return () => {
    if (onUnauthorized === handler) onUnauthorized = undefined;
  };
}

export class ApiError extends Error {
  constructor(message: string, public readonly status: number | undefined, public readonly kind: 'http' | 'network' | 'protocol') {
    super(message);
    this.name = 'ApiError';
  }
}

type RequestOptions = RequestInit & { auth?: 'required' | 'none' };

async function requestResponse(path: string, options: RequestOptions, accept: string): Promise<Response> {
  const { auth = 'required', ...requestOptions } = options;
  const headers = new Headers(options.headers);
  headers.set('Accept', accept);

  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const token = auth === 'required' ? getStoredToken() : null;
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  if (currentProjectCode) {
    headers.set('X-Project-Code', currentProjectCode);
  }

  let response: Response;
  try {
    response = await fetch(`/api/v1${path}`, { ...requestOptions, headers });
  } catch {
    throw new ApiError('Unable to reach the server', undefined, 'network');
  }

  // A cancelled request or a response for an older token cannot expire the current session.
  if (response.status === 401 && token && token === getStoredToken() && !options.signal?.aborted) {
    clearStoredToken();
    onUnauthorized?.();
  }

  if (!response.ok) {
    let message = `Request failed with ${response.status}`;
    try {
      const data = JSON.parse(await response.text());
      if (typeof data?.message === 'string') message = data.message;
    } catch {
      // Non-JSON error bodies still retain their HTTP status.
    }
    throw new ApiError(message, response.status, 'http');
  }
  return response;
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const response = await requestResponse(path, options, 'application/json');
  let data;
  try {
    const text = await response.text();
    data = text ? JSON.parse(text) : undefined;
  } catch {
    throw new ApiError(`Request failed with ${response.status}`, response.status, 'protocol');
  }

  return data as T;
}

export async function apiDownloadCSV(path: string, options: RequestOptions = {}): Promise<Blob> {
  const response = await requestResponse(path, options, 'text/csv');
  if (!/^text\/csv(?:\s*;|$)/i.test(response.headers.get('Content-Type') ?? '')) {
    throw new ApiError('Expected a CSV response', response.status, 'protocol');
  }
  try {
    return await response.blob();
  } catch {
    throw new ApiError('Unable to read the CSV response', response.status, 'network');
  }
}
