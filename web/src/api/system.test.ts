import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { exportAuditLogs, listAuditLogs } from './system';
import { setStoredToken, setUnauthorizedHandler } from './client';

beforeEach(() => { localStorage.clear(); setStoredToken('test-token'); });
afterEach(() => { vi.unstubAllGlobals(); vi.restoreAllMocks(); });

describe('audit queries and CSV downloads', () => {
  it.each([
    { list: [], total: 0 },
    { list: [], total: -1, page: 1, pageSize: 20 },
    { list: [], total: 0, page: '1', pageSize: 20 },
    { list: [], total: 0, page: 1, pageSize: 0 },
    { list: [null], total: 1, page: 1, pageSize: 20 },
  ])('rejects malformed migrated page data: %j', async data => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ code: 0, data, msg: 'success' }))));
    await expect(listAuditLogs()).rejects.toMatchObject({ status: 200, kind: 'protocol' });
  });
  it('encodes all filters and preserves real pagination fields', async () => {
    const page = { list: [], total: 42, page: 2, pageSize: 20 };
    const fetcher = vi.fn().mockResolvedValue(new Response(JSON.stringify({ code: 0, data: page, msg: 'success' })));
    vi.stubGlobal('fetch', fetcher);
    expect(await listAuditLogs({ page: 2, pageSize: 20, actorUserId: 7, action: 'user.update', resourceType: 'user', result: 'failure', keyword: 'a & b', startTime: '2026-10-28T00:00:00.000Z', endTime: '2026-10-29T00:00:00.000Z' })).toEqual(page);
    const url = new URL(fetcher.mock.calls[0][0], 'http://test');
    expect(Object.fromEntries(url.searchParams)).toMatchObject({ page: '2', pageSize: '20', actorUserId: '7', action: 'user.update', resourceType: 'user', result: 'failure', keyword: 'a & b', startTime: '2026-10-28T00:00:00.000Z', endTime: '2026-10-29T00:00:00.000Z' });
  });

  it('omits blank filters and downloads CSV without pagination', async () => {
    const fetcher = vi.fn().mockResolvedValue(new Response('id,action\n1,user.update', { headers: { 'Content-Type': 'text/csv; charset=utf-8' } }));
    vi.stubGlobal('fetch', fetcher);
    const blob = await exportAuditLogs({ result: 'success', keyword: '', page: 3, pageSize: 50 });
    expect(await blob.text()).toContain('1,user.update');
    const [path, options] = fetcher.mock.calls[0]; const url = new URL(path, 'http://test');
    expect(url.pathname).toBe('/api/v1/system/audit-logs/export');
    expect(Object.fromEntries(url.searchParams)).toEqual({ result: 'success' });
    expect(options.headers.get('Authorization')).toBe('Bearer test-token');
    expect(options.headers.get('Accept')).toBe('text/csv');
  });

  it.each([403, 500])('rejects HTTP %i JSON errors instead of saving a CSV', async (status) => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{"message":"export denied"}', { status, headers: { 'Content-Type': 'application/json' } })));
    await expect(exportAuditLogs({})).rejects.toMatchObject({ status, message: 'export denied' });
    expect(localStorage.getItem('optimus.token')).toBe('test-token');
  });

  it('expires credentials on an authenticated CSV 401', async () => {
    const unauthorized = vi.fn(); setUnauthorizedHandler(unauthorized);
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{}', { status: 401 })));
    await expect(exportAuditLogs({})).rejects.toMatchObject({ status: 401 });
    expect(localStorage.getItem('optimus.token')).toBeNull(); expect(unauthorized).toHaveBeenCalledTimes(1);
  });

  it.each(['application/json', 'text/html'])('rejects successful %s downloads', async (contentType) => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{}', { headers: { 'Content-Type': contentType } })));
    await expect(exportAuditLogs({})).rejects.toMatchObject({ kind: 'protocol' });
  });

  it.each(['replaced', 'cancelled'])('does not expire a session for a %s CSV request', async (mode) => {
    const unauthorized = vi.fn(); setUnauthorizedHandler(unauthorized);
    let respond!: (value: Response) => void;
    vi.stubGlobal('fetch', vi.fn(() => new Promise<Response>(resolve => { respond = resolve; })));
    const controller = new AbortController(); const pending = exportAuditLogs({}, controller.signal);
    if (mode === 'replaced') setStoredToken('new-token'); else controller.abort();
    respond(new Response('{}', { status: 401 }));
    await expect(pending).rejects.toMatchObject({ status: 401 });
    expect(unauthorized).not.toHaveBeenCalled();
    expect(localStorage.getItem('optimus.token')).toBe(mode === 'replaced' ? 'new-token' : 'test-token');
  });
});
