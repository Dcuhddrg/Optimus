import { act } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { LanguageProvider } from '../../locale/LanguageProvider';
import { LdapSyncPanel } from './LdapSyncPanel';

const ok = (data: unknown) => new Response(JSON.stringify({ code: 0, data, msg: 'success' }));
const account = (id: number, username: string) => ({ id, username, displayName: username, email: '', department: '', externalId: '', status: 'enabled', createdAt: '2026-01-01T00:00:00Z', roles: ['user'] });
const entry = (username: string, externalId: string) => ({ dn: `uid=${username}`, username, displayName: username, email: '', department: '', externalId, createdAt: '2025-01-01T00:00:00Z' });
const preview = {
  previewId: 'p-1',
  expiresAt: '2026-01-01T00:30:00Z',
  disablesBlocked: false,
  configChanged: false,
  items: [
    { category: 'create', entry: entry('hank', 'id-hank') },
    { category: 'pendingBinding', account: account(6, 'erin'), entry: entry('erin', 'id-erin') },
    { category: 'pendingBinding', account: account(7, 'later'), entry: entry('later', 'id-later'), suspicious: true },
    { category: 'conflict', reason: 'username_reused', account: account(8, 'gina'), entry: entry('gina', 'id-new-gina') },
    { category: 'outOfScope', account: account(4, 'dan') },
    { category: 'conflict', reason: 'local_account', account: { ...account(1, 'admin'), roles: ['admin'] }, entry: entry('admin', 'id-admin') },
  ],
};
let root: Root;
let container: HTMLDivElement;
let fetcher: ReturnType<typeof vi.fn>;

function respond(url: string) {
  if (url.endsWith('/system/ldap/sync/preview')) return ok(preview);
  if (url.endsWith('/system/ldap/sync/execute')) return ok({ status: 'completed', outcomes: [{ action: 'create', username: 'hank', outcome: 'done' }] });
  return ok({});
}
const button = (text: string) => [...container.querySelectorAll('button')].find((b) => b.textContent === text)!;
const row = (username: string) => container.querySelector(`li[data-username="${username}"]`)!;
async function click(el: HTMLElement) { await act(async () => el.click()); }

beforeEach(() => {
  Object.assign(globalThis, { IS_REACT_ACT_ENVIRONMENT: true });
  localStorage.clear(); localStorage.setItem('optimus.locale', 'en'); localStorage.setItem('optimus.token', 'token');
  container = document.createElement('div'); document.body.appendChild(container); root = createRoot(container);
  fetcher = vi.fn().mockImplementation((url: string) => Promise.resolve(respond(url))); vi.stubGlobal('fetch', fetcher);
});
afterEach(async () => { await act(async () => root.unmount()); container.remove(); vi.unstubAllGlobals(); vi.restoreAllMocks(); });

describe('LDAP sync panel', () => {
  it('previews, bulk-binds only unflagged rows, and executes the choices', async () => {
    const onDone = vi.fn();
    await act(async () => root.render(<LanguageProvider><LdapSyncPanel onDone={onDone} /></LanguageProvider>));
    await click(button('Preview sync'));

    expect(container.querySelector('[data-category="create"]')!.textContent).toContain('hank');
    expect(row('later').textContent).toContain('uid may have been reused');
    await click(button('Confirm all unflagged as the same person'));
    expect(row('erin').querySelector('select')!.value).toBe('bind');
    expect(row('later').querySelector('select')!.value).toBe('leave');
    expect(row('admin').querySelector('input')).toBeNull();
    expect(row('admin').textContent).toContain('rename the directory entry');

    await click(row('gina').querySelector('input')!);
    await click(row('dan').querySelector('input')!);
    await click(button('Execute'));

    const [, init] = fetcher.mock.calls.find(([url]) => String(url).endsWith('/execute'))!;
    expect(JSON.parse(init.body)).toEqual({ previewId: 'p-1', disable: [4], bind: [6], separate: [8] });
    expect(container.textContent).toContain('Sync completed');
    expect(onDone).toHaveBeenCalled();
  });

  it('disables the disable choices when disables are blocked', async () => {
    preview.disablesBlocked = true;
    await act(async () => root.render(<LanguageProvider><LdapSyncPanel onDone={() => {}} /></LanguageProvider>));
    await click(button('Preview sync'));
    preview.disablesBlocked = false;
    expect(container.textContent).toContain('this sync disables nobody');
    expect(row('dan').querySelector<HTMLInputElement>('input')!.disabled).toBe(true);
  });
});
