import { beforeEach, describe, expect, it } from 'vitest';

import { taskStore } from './taskStore';
import { ticketStore } from './ticketStore';
import type { Ticket } from '../ticketTypes';

function processingTicket(id: string): Ticket {
  return {
    id,
    serviceId: 'svc-1',
    title: `工单 ${id}`,
    formValues: {},
    status: 'processing',
    createdAt: new Date().toISOString(),
    timeline: [],
    applicant: { name: '申请人', email: 'a@x.io' },
  };
}

beforeEach(() => {
  taskStore.__resetForTest();
  ticketStore.__resetForTest();
});

describe('taskStore.reconcileFromTickets', () => {
  it('为 processing 工单补建一个待领取任务', () => {
    taskStore.reconcileFromTickets([processingTicket('TKT-2026-0001')]);
    const tasks = taskStore.list();
    expect(tasks).toHaveLength(1);
    expect(tasks[0].ticketId).toBe('TKT-2026-0001');
    expect(tasks[0].status).toBe('pending');
    expect(tasks[0].assignee).toBeNull();
    expect(tasks[0].id).toMatch(/^TASK-\d{4}-\d{4}$/);
  });

  it('对已有任务的工单幂等，不重复建', () => {
    const ticket = processingTicket('TKT-2026-0002');
    taskStore.reconcileFromTickets([ticket]);
    taskStore.reconcileFromTickets([ticket]);
    expect(taskStore.list()).toHaveLength(1);
  });

  it('忽略非 processing 工单', () => {
    const ticket = { ...processingTicket('TKT-2026-0003'), status: 'submitted' as const };
    taskStore.reconcileFromTickets([ticket]);
    expect(taskStore.list()).toHaveLength(0);
  });
});
