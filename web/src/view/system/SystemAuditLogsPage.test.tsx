import { act } from 'react';
import { createRoot, type Root } from 'react-dom/client';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { LanguageProvider } from '../../locale/LanguageProvider';
import { SystemAuditLogsPage } from './SystemAuditLogsPage';

const entry = { id: 1, actorUserId: 7, actorUsername: 'alice', action: 'role.permissions.update', resourceType: 'role', resourceId: '3', result: 'success', ip: '192.0.2.1', userAgent: 'test-browser', detail: '{"note":"<script>unsafe</script>"}', createdAt: '2026-10-29T03:00:00Z' };
const page = (list = [entry], total = 45) => new Response(JSON.stringify({ code: 0, data: { list, total, page: 1, pageSize: 20 }, msg: 'success' }));
let container: HTMLDivElement; let root: Root; let fetcher: ReturnType<typeof vi.fn>;
const button = (text: string) => Array.from(container.querySelectorAll('button')).find(b => b.textContent === text)!;
async function click(text: string) { const b = button(text); expect(b).toBeTruthy(); await act(async () => b.click()); }
async function field(label: string, value: string) {
  const wrapper = Array.from(container.querySelectorAll('label')).find(l => l.textContent?.startsWith(label));
  const el = wrapper?.querySelector('input, select') as HTMLInputElement | HTMLSelectElement;
  expect(el).toBeTruthy();
  await act(async () => {
    const proto = el instanceof HTMLSelectElement ? HTMLSelectElement.prototype : HTMLInputElement.prototype;
    Object.getOwnPropertyDescriptor(proto, 'value')!.set!.call(el, value);
    el.dispatchEvent(new Event(el instanceof HTMLSelectElement ? 'change' : 'input', { bubbles: true }));
  });
}
async function mount() { await act(async () => root.render(<LanguageProvider><SystemAuditLogsPage /></LanguageProvider>)); }
const lastURL = () => new URL(fetcher.mock.calls.at(-1)![0], 'http://test');

beforeEach(() => {
  Object.assign(globalThis, { IS_REACT_ACT_ENVIRONMENT: true });
  localStorage.clear(); localStorage.setItem('optimus.locale', 'en'); localStorage.setItem('optimus.token', 'token');
  container = document.createElement('div'); document.body.appendChild(container); root = createRoot(container);
  fetcher = vi.fn().mockImplementation(() => Promise.resolve(page())); vi.stubGlobal('fetch', fetcher);
});
afterEach(async () => { await act(async () => root.unmount()); container.remove(); vi.unstubAllGlobals(); vi.restoreAllMocks(); });

describe('audit log page', () => {
  it('shows the audit trail, translated action and safe raw details', async () => {
    await mount();
    expect(container.textContent).toContain('alice'); expect(container.textContent).toContain('192.0.2.1');
    expect(container.textContent).toContain('Update role permissions'); expect(container.querySelector('time')).not.toBeNull();
    await click('View details');
    expect(container.textContent).toContain('test-browser'); expect(container.textContent).toContain('role.permissions.update');
    expect(container.textContent).toContain('<script>unsafe</script>'); expect(container.querySelector('script')).toBeNull();
    await click('Close details'); expect(container.textContent).not.toContain('test-browser');
  });

  it('paginates, applies filters explicitly, and resets the page and filters', async () => {
    await mount(); await click('Next'); expect(lastURL().searchParams.get('page')).toBe('2');
    await field('Detail keyword', 'a & b'); expect(fetcher).toHaveBeenCalledTimes(2);
    await field('Result', 'failure'); await click('Search');
    expect(Object.fromEntries(lastURL().searchParams)).toMatchObject({ page: '1', result: 'failure', keyword: 'a & b' });
    await click('Next'); await field('Rows per page', '50');
    expect(Object.fromEntries(lastURL().searchParams)).toMatchObject({ page: '1', pageSize: '50', result: 'failure' });
    await click('Reset'); expect(lastURL().searchParams.has('keyword')).toBe(false); expect(lastURL().searchParams.has('result')).toBe(false);
  });

  it('converts local time inputs and refuses reversed ranges', async () => {
    await mount(); await field('Start time', '2026-10-29T09:00'); await field('End time', '2026-10-28T09:00'); await click('Search');
    expect(fetcher).toHaveBeenCalledTimes(1); expect(container.textContent).toContain('Start time must not be after end time');
    await field('End time', '2026-10-30T09:00'); await click('Search');
    expect(lastURL().searchParams.get('startTime')).toBe(new Date('2026-10-29T09:00').toISOString());
  });

  it('shows loading, failure, retry and empty states separately', async () => {
    let respond!: (response: Response) => void;
    fetcher.mockImplementationOnce(() => new Promise<Response>(resolve => { respond = resolve; })); await mount();
    expect(container.textContent).toContain('Loading audit logs');
    await act(async () => respond(new Response('{}', { status: 500 })));
    expect(container.textContent).toContain('Unable to load audit logs'); expect(container.textContent).not.toContain('No audit logs');
    fetcher.mockImplementationOnce(() => Promise.resolve(page([], 0))); await click('Retry');
    expect(container.textContent).toContain('No audit logs'); expect(button('Next').disabled).toBe(true);
  });

  it('ignores stale responses when a new query has completed', async () => {
    let respond!: (response: Response) => void;
    fetcher.mockImplementationOnce(() => new Promise<Response>(resolve => { respond = resolve; })); await mount();
    await field('Detail keyword', 'new'); await click('Search'); expect(container.textContent).toContain('alice');
    await act(async () => respond(page([{ ...entry, actorUsername: 'stale-user' }])));
    expect(container.textContent).not.toContain('stale-user');
  });

  it('uses unknown raw values and an honest fallback for missing historical data', async () => {
    fetcher.mockImplementationOnce(() => Promise.resolve(page([{ ...entry, action: 'future.action', actorUsername: '', actorUserId: 0, detail: '', result: 'pending' }])));
    await mount(); expect(container.textContent).toContain('future.action'); expect(container.textContent).toContain('Unknown actor');
    expect(container.textContent).toContain('pending'); await click('View details'); expect(container.textContent).toContain('Not recorded');
  });

  it('exports only applied filters and releases the object URL', async () => {
    const create = vi.fn(() => 'blob:test'); const revoke = vi.fn();
    Object.defineProperty(URL, 'createObjectURL', { configurable: true, value: create });
    Object.defineProperty(URL, 'revokeObjectURL', { configurable: true, value: revoke });
    const download = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {});
    await mount(); await field('Result', 'failure'); await click('Search'); await click('Next');
    await field('Result', 'success');
    fetcher.mockImplementationOnce(() => Promise.resolve(new Response('id\n1', { headers: { 'Content-Type': 'text/csv' } })));
    await click('Export CSV');
    expect(lastURL().pathname).toBe('/api/v1/system/audit-logs/export');
    expect(Object.fromEntries(lastURL().searchParams)).toEqual({ result: 'failure' });
    expect(download).toHaveBeenCalledTimes(1); expect(revoke).toHaveBeenCalledWith('blob:test');
    expect(container.textContent).toContain('10,000');
  });

  it('reports an export permission failure without discarding the list', async () => {
    await mount(); fetcher.mockImplementationOnce(() => Promise.resolve(new Response('{}', { status: 403 })));
    await click('Export CSV'); expect(container.textContent).toContain('You do not have permission to export'); expect(container.textContent).toContain('alice');
  });

  it('renders Chinese labels', async () => {
    localStorage.setItem('optimus.locale', 'zh-CN'); await mount();
    expect(container.textContent).toContain('操作人'); expect(container.textContent).toContain('修改角色权限');
  });

  it('returns to the last valid page when records disappear', async () => {
    await mount(); await click('Next');
    fetcher.mockImplementation(() => Promise.resolve(page([], 5)));
    await click('Next');
    expect(lastURL().searchParams.get('page')).toBe('1');
    expect(container.textContent).toContain('No audit logs');
  });

  it('cancels an export when applied filters change', async () => {
    await mount();
    let respond!: (response: Response) => void;
    const create = vi.fn(); Object.defineProperty(URL, 'createObjectURL', { configurable: true, value: create });
    fetcher.mockImplementationOnce(() => new Promise<Response>(resolve => { respond = resolve; }));
    await click('Export CSV'); const signal = fetcher.mock.calls.at(-1)![1].signal as AbortSignal;
    await field('Result', 'failure'); await click('Search'); expect(signal.aborted).toBe(true);
    await act(async () => respond(new Response('id\n1', { headers: { 'Content-Type': 'text/csv' } })));
    expect(create).not.toHaveBeenCalled(); expect(container.textContent).toContain('alice');
  });
});
