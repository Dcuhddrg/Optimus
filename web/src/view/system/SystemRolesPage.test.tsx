import { act } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import { LanguageProvider } from '../../locale/LanguageProvider';
import { SystemRolesPage } from './SystemRolesPage';

let container: HTMLDivElement;
let root: Root;
beforeEach(() => {
  Object.assign(globalThis, { IS_REACT_ACT_ENVIRONMENT: true });
  localStorage.clear(); localStorage.setItem('optimus.locale', 'en'); localStorage.setItem('optimus.token', 'token');
  container = document.createElement('div'); document.body.appendChild(container); root = createRoot(container);
});
afterEach(async () => { await act(async () => root.unmount()); container.remove(); vi.unstubAllGlobals(); });
it('loads migrated collections and grants permissions using the existing workflow', async () => {
  const fetcher = vi.fn().mockImplementation((url: string) => {
    const data = url.endsWith('/roles')
      ? { list: [{ id: 3, code: 'ops', name: 'Operations', status: 'enabled' }], total: 1 }
      : url.endsWith('/tree') ? { list: [{ id: 8, code: 'system.roles.view', name: 'Roles' }] } : {};
    return Promise.resolve(new Response(JSON.stringify({ code: 0, data, msg: 'success' })));
  });
  vi.stubGlobal('fetch', fetcher);
  await act(async () => root.render(<LanguageProvider><SystemRolesPage /></LanguageProvider>));
  expect(container.textContent).toContain('Operations');
  await act(async () => container.querySelector('button')!.click());
  const [url, options] = fetcher.mock.calls[2];
  expect(url).toBe('/api/v1/system/roles/3/permissions'); expect(options.method).toBe('PUT');
  expect(JSON.parse(options.body)).toEqual({ permissionCodes: ['system.roles.view'] });
});
