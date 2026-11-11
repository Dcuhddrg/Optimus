import { act } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { LanguageProvider } from '../../locale/LanguageProvider';
import { SystemRolesPage } from './SystemRolesPage';

const ok = (data: unknown) => new Response(JSON.stringify({ code: 0, data, msg: 'success' }));
const roles = [
  { id: 1, code: 'admin', name: 'Administrator', description: '', approvable: false, permissionCodes: ['system.users.menu', 'system.users.view', 'cloud.accounts.view'] },
  { id: 2, code: 'user', name: 'User', description: '', approvable: false, permissionCodes: [] },
  { id: 3, code: 'security-approver', name: 'Security approver', description: '', approvable: true, permissionCodes: [] },
  { id: 4, code: 'ops', name: 'Ops', description: '', approvable: false, permissionCodes: ['cloud.accounts.view'] },
];
const permissions = [
  { code: 'system.users.menu', name: 'Users', type: 'menu' },
  { code: 'system.users.view', name: 'View users', type: 'api' },
  { code: 'cloud.accounts.view', name: 'View cloud accounts', type: 'api' },
];
let root: Root;
let container: HTMLDivElement;
let fetcher: ReturnType<typeof vi.fn>;

function respond(url: string, init?: RequestInit) {
  if (url.endsWith('/system/roles') && (!init?.method || init.method === 'GET')) return ok({ list: roles, total: roles.length });
  if (url.endsWith('/system/permissions/tree')) return ok({ list: permissions });
  return ok({});
}
const card = (code: string) => container.querySelector(`[data-role="${code}"]`)!;
const buttonIn = (el: Element, text: string) => [...el.querySelectorAll('button')].find((b) => b.textContent === text);
async function mount() { await act(async () => root.render(<LanguageProvider><SystemRolesPage /></LanguageProvider>)); }
async function click(el: HTMLElement) { await act(async () => el.click()); }
const calls = (method: string) => fetcher.mock.calls.filter(([, init]) => init?.method === method);

beforeEach(() => {
  Object.assign(globalThis, { IS_REACT_ACT_ENVIRONMENT: true });
  localStorage.clear(); localStorage.setItem('optimus.locale', 'en'); localStorage.setItem('optimus.token', 'token');
  container = document.createElement('div'); document.body.appendChild(container); root = createRoot(container);
  fetcher = vi.fn().mockImplementation((url: string, init?: RequestInit) => Promise.resolve(respond(url, init))); vi.stubGlobal('fetch', fetcher);
});
afterEach(async () => { await act(async () => root.unmount()); container.remove(); vi.unstubAllGlobals(); vi.restoreAllMocks(); });

describe('roles page', () => {
  it('shows approval roles and hides delete and the approval toggle where not allowed', async () => {
    await mount();
    expect(card('security-approver').textContent).toContain('Approval role');
    expect(buttonIn(card('admin'), 'Delete')).toBeUndefined();
    expect(buttonIn(card('user'), 'Delete')).toBeUndefined();
    expect(buttonIn(card('ops'), 'Delete')).toBeDefined();
    expect(card('admin').querySelector('input[type="checkbox"]')).toBeNull();
  });

  it('locks the admin role system permissions and always saves them', async () => {
    await mount();
    await click(buttonIn(card('admin'), 'Edit permissions')!);
    const boxes = [...card('admin').querySelectorAll<HTMLInputElement>('fieldset input[type="checkbox"]')];
    expect(boxes.filter((b) => b.disabled)).toHaveLength(2);
    const cloud = boxes.find((b) => !b.disabled)!;
    await click(cloud); // revoke the non-system permission
    await click(buttonIn(card('admin'), 'Save')!);
    const [url, init] = calls('PUT').at(-1)!;
    expect(url).toContain('/system/roles/1/permissions');
    expect(JSON.parse(init.body).permissionCodes.sort()).toEqual(['system.users.menu', 'system.users.view']);
  });

  it('deletes a role after confirmation and shows server errors', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    await mount();
    fetcher.mockImplementationOnce(() => Promise.resolve(new Response(JSON.stringify({ code: 7, data: {}, msg: 'role is still assigned to users; remove it from them first' }))));
    await click(buttonIn(card('ops'), 'Delete')!);
    expect(calls('DELETE')[0][0]).toContain('/system/roles/4');
    expect(container.querySelector('[role="alert"]')?.textContent).toContain('still assigned');
  });

  it('creates an approval role', async () => {
    await mount();
    const inputs = container.querySelectorAll<HTMLInputElement>('form input');
    const set = (el: HTMLInputElement, value: string) => { Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value')!.set!.call(el, value); el.dispatchEvent(new Event('input', { bubbles: true })); };
    await act(async () => { set(inputs[0], 'audit-approver'); set(inputs[1], 'Audit approver'); });
    await click(inputs[3]);
    await act(async () => { container.querySelector('form')!.requestSubmit(); });
    const [, init] = calls('POST').at(-1)!;
    expect(JSON.parse(init.body)).toEqual({ code: 'audit-approver', name: 'Audit approver', description: '', approvable: true });
  });
});
