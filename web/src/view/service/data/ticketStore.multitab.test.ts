// 用两份独立加载的模块模拟两个标签页，它们共用 jsdom 的 localStorage。
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

const STORAGE_KEY = 'optimus.servicePortal.tickets.v1';

async function openTab() {
  vi.resetModules();
  const { ticketStore } = await import('./ticketStore');
  const { ticketMockEngine } = await import('./ticketMockEngine');
  return { store: ticketStore, engine: ticketMockEngine };
}

function input(title: string, serviceId: string | null = 'account.role-grant') {
  return { serviceId, title, formValues: {}, applicant: { name: '申请人', email: 'a@x.io' } };
}

function storedIds(): string[] {
  return (JSON.parse(window.localStorage.getItem(STORAGE_KEY) ?? '[]') as { id: string }[]).map((t) => t.id);
}

beforeEach(() => {
  window.localStorage.clear();
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
  window.localStorage.clear();
});

describe('ticketStore 多标签页', () => {
  it('B 恢复的定时器推进工单时，不覆盖 A 之后新建的工单', async () => {
    const a = await openTab();
    const t1 = a.store.createTicket(input('T1')).ticket.id;

    const b = await openTab();
    b.engine.resumePending(); // B 打开门户，为 T1 安排定时器

    const t2 = a.store.createTicket(input('T2')).ticket.id;
    expect(t2).not.toBe(t1);

    vi.advanceTimersByTime(5_000); // B 的定时器把 T1 推进到 approving 并写回
    expect(storedIds()).toEqual(expect.arrayContaining([t1, t2]));
    expect(b.store.get(t1)!.status).toBe('approving');
    expect(b.store.get(t2)).toBeDefined();
  });

  it('两个标签页新建工单不会撞号', async () => {
    const a = await openTab();
    const b = await openTab();
    const t1 = a.store.createTicket(input('T1')).ticket.id;
    const t2 = b.store.createTicket(input('T2')).ticket.id;
    expect(t2).not.toBe(t1);
    expect(storedIds()).toHaveLength(2);
  });

  it('两个标签页都安排定时器时，审批仍逐级推进、不连跳', async () => {
    const a = await openTab();
    const id = a.store.createTicket(input('T', 'cloud.compute-resize')).ticket.id; // 3 级
    a.engine.schedule(id);
    const b = await openTab();
    b.engine.resumePending();

    vi.advanceTimersByTime(5_000);
    vi.advanceTimersByTime(10_000); // 两个标签页都触发第 1 级
    expect(b.store.get(id)!.approvalChain.map((s) => s.status)).toEqual(['approved', 'pending', 'pending']);

    vi.advanceTimersByTime(20_000);
    const done = b.store.get(id)!;
    expect(done.status).toBe('processing');
    expect(done.timeline.filter((e) => e.i18nNoteKey === 'servicePortal.ticket.event.approved')).toHaveLength(3);
  });

  it('其他标签页写入后通过 storage 事件同步', async () => {
    const a = await openTab();
    const b = await openTab();
    const listener = vi.fn();
    b.store.subscribe(listener);

    a.store.createTicket(input('T1'));
    window.dispatchEvent(new StorageEvent('storage', { key: STORAGE_KEY }));

    expect(b.store.list()).toHaveLength(1);
    expect(listener).toHaveBeenCalled();
  });
});
