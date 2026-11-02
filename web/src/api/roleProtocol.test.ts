import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { listRoles, permissionTree, saveRolePermissions } from './system';
import { setStoredToken } from './client';

beforeEach(() => { localStorage.clear(); setStoredToken('role-token'); });
afterEach(() => { vi.unstubAllGlobals(); vi.restoreAllMocks(); });
function respond(data: unknown, code = 0, status = 200) {
  const fetcher = vi.fn().mockResolvedValue(new Response(JSON.stringify({ code, data, msg: code ? 'denied' : 'success' }), { status }));
  vi.stubGlobal('fetch', fetcher);
  return fetcher;
}
describe('role and permission response contract', () => {
  it('unwraps role list and count without inventing pagination', async () => {
    const data = { list: [{ id: 1, code: 'ops', name: 'Operations', status: 'enabled' }], total: 1 };
    const fetcher = respond(data);
    expect(await listRoles()).toEqual(data);
    expect(fetcher.mock.calls[0][0]).toBe('/api/v1/system/roles');
    expect(fetcher.mock.calls[0][1].headers.get('Authorization')).toBe('Bearer role-token');
  });
  it('unwraps empty roles and the non-paginated permission collection', async () => {
    respond({ list: [], total: 0 }); expect(await listRoles()).toEqual({ list: [], total: 0 });
    const data = { list: [{ id: 8, code: 'system.roles.view', name: 'Roles' }] };
    const fetcher = respond(data); expect(await permissionTree()).toEqual(data);
    expect(fetcher.mock.calls[0][0]).toBe('/api/v1/system/permissions/tree');
  });
  it('preserves permission update request and returns empty business data', async () => {
    const fetcher = respond({}); expect(await saveRolePermissions(3, ['system.roles.view'])).toEqual({});
    const [url, options] = fetcher.mock.calls[0];
    expect(url).toBe('/api/v1/system/roles/3/permissions'); expect(options.method).toBe('PUT');
    expect(JSON.parse(options.body)).toEqual({ permissionCodes: ['system.roles.view'] });
  });
  it.each([listRoles, permissionTree, () => saveRolePermissions(1, [])])('rejects old success payload and business failures', async request => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{"items":[],"total":0}')));
    await expect(request()).rejects.toMatchObject({ status: 200, kind: 'protocol' });
    respond({}, 7); await expect(request()).rejects.toMatchObject({ status: 200, kind: 'business' });
    respond({}, 7, 403); await expect(request()).rejects.toMatchObject({ status: 403, message: 'denied' });
    expect(localStorage.getItem('optimus.token')).toBe('role-token');
  });
});
