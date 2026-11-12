import { act } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { LanguageProvider } from '../../locale/LanguageProvider';
import { SystemUsersPage } from './SystemUsersPage';

const ok = (data: unknown) => new Response(JSON.stringify({ code: 0, data, msg: 'success' }));
const base = { uuid: 'u', email: '', source: 'ldap', bound: true, status: 'enabled', roles: [] };
const users = [
  { ...base, id: 1, username: 'alice', displayName: 'Alice', personType: 'employee' },
  { ...base, id: 2, username: 'eve', displayName: 'Eve', personType: 'external', expiresOn: null, sponsorId: null },
  { ...base, id: 3, username: 'old', displayName: 'Old', personType: 'external', expiresOn: '2026-01-31', sponsorId: 4, sponsor: 'gone', sponsorInvalid: true, expired: true, restrictedRoles: ['admin'], ownedProjects: ['prod'] },
  { ...base, id: 4, username: 'gone', displayName: 'Gone', personType: 'employee', status: 'disabled' },
];
let root: Root;
let container: HTMLDivElement;
let fetcher: ReturnType<typeof vi.fn>;
const row = (username: string) => container.querySelector(`[data-username="${username}"]`)!;
const buttonIn = (el: Element, text: string) => [...el.querySelectorAll('button')].find((b) => b.textContent === text)!;

beforeEach(() => {
  Object.assign(globalThis, { IS_REACT_ACT_ENVIRONMENT: true });
  localStorage.clear(); localStorage.setItem('optimus.locale', 'en'); localStorage.setItem('optimus.token', 'token');
  container = document.createElement('div'); document.body.appendChild(container); root = createRoot(container);
  fetcher = vi.fn().mockImplementation((url: string) => Promise.resolve(ok(String(url).endsWith('/system/users') ? { list: users, total: users.length } : users[1])));
  vi.stubGlobal('fetch', fetcher);
});
afterEach(async () => { await act(async () => root.unmount()); container.remove(); vi.unstubAllGlobals(); vi.restoreAllMocks(); });

describe('System users page', () => {
  it('shows external controls and saves expiry and sponsor', async () => {
    await act(async () => root.render(<LanguageProvider><SystemUsersPage /></LanguageProvider>));

    expect(row('alice').textContent).not.toContain('External');
    expect(row('eve').textContent).toContain('No permissions until expiry and sponsor are set');
    expect(row('old').textContent).toContain('Expired');
    expect(row('old').textContent).toContain('Sponsor no longer valid');
    expect(row('old').textContent).toContain('Roles to remove (not applied): admin');
    expect(row('old').textContent).toContain('Projects to hand over: prod');

    await act(async () => buttonIn(row('eve'), 'Expiry and sponsor').click());
    const sponsorSelect = row('eve').querySelector('select')!;
    const options = [...sponsorSelect.options].map((o) => o.value);
    expect(options).toEqual(['', '1'], 'only enabled employees can sponsor');

    const date = row('eve').querySelector('input[type="date"]') as HTMLInputElement;
    await act(async () => {
      const setValue = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')!.set!;
      setValue.call(date, '2026-12-31');
      date.dispatchEvent(new Event('input', { bubbles: true }));
      const setSelect = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value')!.set!;
      setSelect.call(sponsorSelect, '1');
      sponsorSelect.dispatchEvent(new Event('change', { bubbles: true }));
    });
    await act(async () => buttonIn(row('eve'), 'Save').click());

    const [, init] = fetcher.mock.calls.find(([url]) => String(url).endsWith('/system/users/2/external'))!;
    expect(init.method).toBe('PUT');
    expect(JSON.parse(init.body)).toEqual({ expiresOn: '2026-12-31', sponsorId: 1 });
  });
});
