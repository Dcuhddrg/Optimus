import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { apiRequest, getStoredToken, setStoredToken, setUnauthorizedHandler } from './client';
import { login } from './system';

beforeEach(() => { localStorage.clear(); setStoredToken('existing'); });
afterEach(() => { vi.unstubAllGlobals(); vi.restoreAllMocks(); });

describe('token transport', () => {
  it('sends x-token and stores a refreshed token from the new-token header', async () => {
    const fetcher = vi.fn().mockResolvedValue(new Response(JSON.stringify({ code: 0, data: {}, msg: 'ok' }), { headers: { 'new-token': 'refreshed' } }));
    vi.stubGlobal('fetch', fetcher);
    await apiRequest('/system/auth/me');
    expect(fetcher.mock.calls[0][1].headers.get('x-token')).toBe('existing');
    expect(getStoredToken()).toBe('refreshed');
  });

  it('ignores new-token when the session changed meanwhile', async () => {
    let release: (r: Response) => void = () => {};
    vi.stubGlobal('fetch', vi.fn().mockReturnValue(new Promise<Response>((resolve) => { release = resolve; })));
    const pending = apiRequest('/system/auth/me');
    setStoredToken('other-session');
    release(new Response(JSON.stringify({ code: 0, data: {}, msg: 'ok' }), { headers: { 'new-token': 'stale' } }));
    await pending;
    expect(getStoredToken()).toBe('other-session');
  });
});

describe('request authentication errors', () => {
  it('keeps existing credentials when public login fails', async () => {
    const unauthorized = vi.fn();
    setUnauthorizedHandler(unauthorized);
    const fetcher = vi.fn().mockResolvedValue(new Response('{"message":"bad login"}', { status: 401 }));
    vi.stubGlobal('fetch', fetcher);
    await expect(login('user', 'wrong')).rejects.toMatchObject({ status: 401, message: 'bad login' });
    expect(getStoredToken()).toBe('existing');
    expect(unauthorized).not.toHaveBeenCalled();
    expect(fetcher.mock.calls[0][1].headers.has('x-token')).toBe(false);
  });

  it.each([401, 403, 500])('retains HTTP %i even when the response is not JSON', async (status) => {
    const unauthorized = vi.fn();
    setUnauthorizedHandler(unauthorized);
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('<html>Error</html>', { status })));
    await expect(apiRequest('/system/auth/me')).rejects.toMatchObject({ status });
    expect(getStoredToken()).toBe(status === 401 ? null : 'existing');
    expect(unauthorized).toHaveBeenCalledTimes(status === 401 ? 1 : 0);
  });

  it.each(['replaced', 'aborted'])('ignores a late 401 from a %s request', async (mode) => {
    const unauthorized = vi.fn();
    setUnauthorizedHandler(unauthorized);
    let respond!: (response: Response) => void;
    vi.stubGlobal('fetch', vi.fn(() => new Promise<Response>((resolve) => { respond = resolve; })));
    const controller = new AbortController();
    const pending = apiRequest('/system/auth/me', { signal: controller.signal });
    if (mode === 'replaced') setStoredToken('new-session');
    else controller.abort();
    respond(new Response('{}', { status: 401 }));
    await expect(pending).rejects.toMatchObject({ status: 401 });
    expect(getStoredToken()).toBe(mode === 'replaced' ? 'new-session' : 'existing');
    expect(unauthorized).not.toHaveBeenCalled();
  });

  it('keeps credentials on a network failure without retrying', async () => {
    const fetcher = vi.fn().mockRejectedValue(new TypeError('offline'));
    vi.stubGlobal('fetch', fetcher);
    await expect(apiRequest('/system/users', { method: 'POST', body: '{}' })).rejects.toMatchObject({ kind: 'network' });
    expect(getStoredToken()).toBe('existing'); expect(fetcher).toHaveBeenCalledTimes(1);
  });

  it('reports invalid successful JSON as a protocol error', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{broken')));
    await expect(apiRequest('/system/auth/me')).rejects.toMatchObject({ status: 200, kind: 'protocol' });
    expect(getStoredToken()).toBe('existing');
  });

  it('keeps credentials on ordinary forbidden requests', async () => {
    const unauthorized = vi.fn(); setUnauthorizedHandler(unauthorized);
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{}', { status: 403 })));
    await expect(apiRequest('/system/users')).rejects.toMatchObject({ status: 403, kind: 'http' });
    expect(getStoredToken()).toBe('existing'); expect(unauthorized).not.toHaveBeenCalled();
  });
});
