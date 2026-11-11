import { act } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { LanguageProvider } from '../../locale/LanguageProvider';
import { SystemLdapPage } from './SystemLdapPage';

const ok = (data: unknown) => new Response(JSON.stringify({ code: 0, data, msg: 'success' }));
const saved = {
  enabled: true, url: 'ldap://ldap', baseDn: 'dc=optimus,dc=test', bindDn: 'cn=robot', hasBindPassword: true,
  userFilter: '(uid={username})', usernameAttr: 'uid', displayNameAttr: 'cn', emailAttr: 'mail',
  idAttr: 'entryUUID', departmentAttr: 'departmentNumber', boundUsers: 0,
};
let root: Root;
let container: HTMLDivElement;
let fetcher: ReturnType<typeof vi.fn>;
let config = saved;

async function mount() { await act(async () => root.render(<LanguageProvider><SystemLdapPage /></LanguageProvider>)); }
const select = () => container.querySelector('select')!;

beforeEach(() => {
  Object.assign(globalThis, { IS_REACT_ACT_ENVIRONMENT: true });
  localStorage.clear(); localStorage.setItem('optimus.locale', 'en'); localStorage.setItem('optimus.token', 'token');
  container = document.createElement('div'); document.body.appendChild(container); root = createRoot(container);
  fetcher = vi.fn().mockImplementation(() => Promise.resolve(ok(config))); vi.stubGlobal('fetch', fetcher);
});
afterEach(async () => { await act(async () => root.unmount()); container.remove(); vi.unstubAllGlobals(); vi.restoreAllMocks(); config = saved; });

describe('LDAP page', () => {
  it('offers only whitelisted identity attributes', async () => {
    await mount();
    expect([...select().options].map((o) => o.value)).toEqual(['entryUUID', 'nsUniqueId', 'ipaUniqueID']);
    expect(select().disabled).toBe(false);
  });

  it('locks the identity attribute while accounts are bound', async () => {
    config = { ...saved, boundUsers: 3 };
    await mount();
    expect(select().disabled).toBe(true);
    expect(container.textContent).toContain('3 accounts are bound to a directory identity');
  });

  it('saves an empty department attribute as "not synced"', async () => {
    await mount();
    const department = [...container.querySelectorAll('label')].find((l) => l.textContent?.startsWith('Department attribute'))!.querySelector('input')!;
    await act(async () => {
      Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')!.set!.call(department, '');
      department.dispatchEvent(new Event('input', { bubbles: true }));
    });
    await act(async () => container.querySelector('form')!.requestSubmit());
    const [, init] = fetcher.mock.calls.find(([, i]) => i?.method === 'PUT')!;
    expect(JSON.parse(init.body).departmentAttr).toBe('');
  });
});
