import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { ticketMockEngine } from './ticketMockEngine';
import { ticketStore } from './ticketStore';

beforeEach(() => {
  ticketStore.__resetForTest();
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('ticketMockEngine', () => {
  it('推进到 processing 后不再自动 done', () => {
    const created = ticketStore.createTicket({
      serviceId: 'svc-1',
      title: '定时验证',
      formValues: {},
      applicant: { name: '申请人', email: 'a@x.io' },
    });
    const id = created.ticket.id;
    ticketMockEngine.schedule(id);

    vi.advanceTimersByTime(5_000);
    expect(ticketStore.get(id)!.status).toBe('approving');

    vi.advanceTimersByTime(25_000); // 累计 30s
    expect(ticketStore.get(id)!.status).toBe('processing');

    vi.advanceTimersByTime(120_000); // 再过很久
    expect(ticketStore.get(id)!.status).toBe('processing'); // 不再自动 done
  });
});

describe('ticketMockEngine 审批链逐级推进', () => {
  it('按链长度每 10s 通过一级，走完进入 processing，不自动 done', () => {
    const id = ticketStore.createTicket({
      serviceId: 'cloud.compute-resize', // 3 级
      title: '逐级验证',
      formValues: {},
      applicant: { name: '申请人', email: 'a@x.io' },
    }).ticket.id;
    ticketMockEngine.schedule(id);

    vi.advanceTimersByTime(5_000);
    expect(ticketStore.get(id)!.status).toBe('approving');

    vi.advanceTimersByTime(10_000); // 第 1 级
    expect(ticketStore.get(id)!.approvalChain.map((s) => s.status)).toEqual(['approved', 'pending', 'pending']);
    expect(ticketStore.get(id)!.status).toBe('approving');

    vi.advanceTimersByTime(10_000); // 第 2 级
    expect(ticketStore.get(id)!.status).toBe('approving');

    vi.advanceTimersByTime(10_000); // 第 3 级 → processing
    expect(ticketStore.get(id)!.status).toBe('processing');

    vi.advanceTimersByTime(120_000);
    expect(ticketStore.get(id)!.status).toBe('processing');
  });
});
