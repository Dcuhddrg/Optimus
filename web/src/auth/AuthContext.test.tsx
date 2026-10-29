import { act } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { MemoryRouter, Route, Routes, useLocation } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { AuthProvider, useAuth } from './AuthContext';
import { RequireAuth } from './RequireAuth';
import { LanguageProvider } from '../locale/LanguageProvider';

const user = { user: { id: 1, username: 'tester' }, roles: [], menuPermissions: [] };
let root: Root;
let container: HTMLDivElement;
let auth: ReturnType<typeof useAuth>;
let pathname: string;

function Probe() { auth = useAuth(); pathname = useLocation().pathname; return null; }
async function mount() {
  await act(async () => {
    root.render(<LanguageProvider><MemoryRouter initialEntries={['/private']} future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
      <AuthProvider><Probe /><Routes>
        <Route path="/login" element={<div>LOGIN</div>} />
        <Route element={<RequireAuth />}><Route path="/private" element={<div>PROTECTED</div>} /></Route>
      </Routes></AuthProvider>
    </MemoryRouter></LanguageProvider>);
  });
}
beforeEach(() => {
  Object.assign(globalThis, { IS_REACT_ACT_ENVIRONMENT: true });
  localStorage.clear(); localStorage.setItem('optimus.token', 'existing');
  container = document.createElement('div'); document.body.appendChild(container); root = createRoot(container);
});
afterEach(async () => { await act(async () => root.unmount()); container.remove(); vi.unstubAllGlobals(); vi.restoreAllMocks(); });

describe('session recovery through real requests and route guard', () => {
  it.each(['network', '500', '403', 'invalid-json', 'invalid-user'])('preserves credentials and blocks protected content after %s', async (failure) => {
    const fetcher = vi.fn().mockImplementation(() => {
      if (failure === 'network') return Promise.reject(new TypeError('offline'));
      return Promise.resolve(new Response(failure === 'invalid-json' ? '{broken' : '{}', { status: Number(failure) || 200 }));
    });
    vi.stubGlobal('fetch', fetcher);
    await mount();
    expect(localStorage.getItem('optimus.token')).toBe('existing');
    expect(pathname).toBe('/private');
    expect(auth.current).toBeNull();
    expect(container.textContent).not.toContain('PROTECTED');
    const retry = container.querySelector('button'); expect(retry).not.toBeNull();
    let resolve!: (value: Response) => void;
    fetcher.mockImplementationOnce(() => new Promise<Response>((r) => { resolve = r; }));
    await act(async () => retry!.click());
    expect(container.textContent).not.toContain('PROTECTED');
    await act(async () => resolve(new Response(JSON.stringify(user))));
    expect(container.textContent).toContain('PROTECTED');
    expect(fetcher).toHaveBeenCalledTimes(2);
  });

  it('expires a confirmed 401 once and sends the user to login', async () => {
    const remove = vi.spyOn(Storage.prototype, 'removeItem');
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{}', { status: 401 })));
    await mount();
    expect(auth.token).toBeNull(); expect(pathname).toBe('/login');
    expect(remove.mock.calls.filter(([key]) => key === 'optimus.token')).toHaveLength(1);
    remove.mockRestore();
  });

  it('keeps a validated session on failed login and still clears it on explicit logout', async () => {
    const fetcher = vi.fn().mockImplementation(() => Promise.resolve(new Response(JSON.stringify(user))));
    vi.stubGlobal('fetch', fetcher); await mount();
    fetcher.mockResolvedValueOnce(new Response('{"message":"bad login"}', { status: 401 }));
    await act(async () => { await expect(auth.login('user', 'wrong')).rejects.toMatchObject({ status: 401 }); });
    expect(auth.token).toBe('existing'); expect(container.textContent).toContain('PROTECTED');
    fetcher.mockRejectedValueOnce(new TypeError('offline'));
    await act(async () => { await expect(auth.logout()).rejects.toThrow(); });
    expect(localStorage.getItem('optimus.token')).toBeNull(); expect(pathname).toBe('/login');
  });

  it('does not restore or expire an old request after logout', async () => {
    let respond!: (response: Response) => void;
    const fetcher = vi.fn().mockImplementationOnce(() => new Promise<Response>((r) => { respond = r; }))
      .mockImplementation(() => Promise.resolve(new Response('{}')));
    vi.stubGlobal('fetch', fetcher); await mount();
    await act(async () => auth.logout());
    localStorage.setItem('optimus.token', 'another-session');
    await act(async () => respond(new Response('{}', { status: 401 })));
    expect(localStorage.getItem('optimus.token')).toBe('another-session');
    expect(auth.current).toBeNull(); expect(pathname).toBe('/login');
  });

  it('redirects an absent credential without requesting a session', async () => {
    localStorage.clear(); const fetcher = vi.fn(); vi.stubGlobal('fetch', fetcher);
    await mount(); expect(pathname).toBe('/login'); expect(fetcher).not.toHaveBeenCalled();
  });

  it('renders recovery errors and retry in Chinese', async () => {
    localStorage.setItem('optimus.locale', 'zh-CN');
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{}', { status: 403 })));
    await mount();
    expect(container.textContent).toContain('无法访问当前用户信息');
    expect(container.querySelector('button')?.textContent).toBe('重试');
  });

  it.each(['existing', 'new-session'])('completes login while recovery is pending with token %s', async (newToken) => {
    let respond!: (response: Response) => void;
    const fetcher = vi.fn().mockImplementationOnce(() => new Promise<Response>((r) => { respond = r; }))
      .mockImplementation(() => Promise.resolve(new Response(JSON.stringify(user))));
    vi.stubGlobal('fetch', fetcher); await mount();
    fetcher.mockResolvedValueOnce(new Response(JSON.stringify({ ...user, token: newToken })));
    await act(async () => auth.login('tester', 'password'));
    expect(auth.loading).toBe(false); expect(auth.token).toBe(newToken);
    expect(container.textContent).toContain('PROTECTED');
    await act(async () => respond(new Response('{}', { status: 401 })));
    expect(auth.token).toBe(newToken); expect(container.textContent).toContain('PROTECTED');
  });
});
