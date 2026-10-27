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
