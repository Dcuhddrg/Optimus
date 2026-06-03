import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { startTaskBridge, taskStore } from './taskStore';
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

describe('taskStore.claim', () => {
  const user = { name: '小李', email: 'li@x.io' };

  function seedPending(ticketId: string): string {
    taskStore.reconcileFromTickets([processingTicket(ticketId)]);
    return taskStore.list()[0].id;
  }

  it('pending → in_progress 并记录 assignee/claimedAt', () => {
    const id = seedPending('TKT-2026-0010');
    taskStore.claim(id, user);
    const task = taskStore.get(id)!;
    expect(task.status).toBe('in_progress');
    expect(task.assignee).toEqual(user);
    expect(task.claimedAt).toBeTruthy();
  });

  it('对非 pending 任务幂等忽略', () => {
    const id = seedPending('TKT-2026-0011');
    taskStore.claim(id, user);
    taskStore.claim(id, { name: '别人', email: 'other@x.io' });
    expect(taskStore.get(id)!.assignee).toEqual(user);
  });
});

describe('taskStore.complete', () => {
  const user = { name: '小李', email: 'li@x.io' };

  function seedProcessingTicketWithTask(): { ticketId: string; taskId: string } {
    const created = ticketStore.createTicket({
      serviceId: 'svc-1',
      title: '扩容申请',
      formValues: {},
      applicant: { name: '申请人', email: 'a@x.io' },
    });
    const ticketId = created.ticket.id;
    ticketStore.advance(ticketId, 'submitted', 'approving');
    ticketStore.advance(ticketId, 'approving', 'processing');
    taskStore.reconcileFromTickets(ticketStore.list());
    const taskId = taskStore.list().find((tk) => tk.ticketId === ticketId)!.id;
    return { ticketId, taskId };
  }

  it('in_progress → done 并把工单推进到 done', () => {
    const { ticketId, taskId } = seedProcessingTicketWithTask();
    taskStore.claim(taskId, user);
    taskStore.complete(taskId);
    const task = taskStore.get(taskId)!;
    expect(task.status).toBe('done');
    expect(task.completedAt).toBeTruthy();
    expect(ticketStore.get(ticketId)!.status).toBe('done');
  });

  it('对非 in_progress 任务幂等忽略', () => {
    const { taskId } = seedProcessingTicketWithTask();
    taskStore.complete(taskId);
    expect(taskStore.get(taskId)!.status).toBe('pending');
  });
});

describe('taskStore.release', () => {
  const user = { name: '小李', email: 'li@x.io' };

  it('in_progress → pending 并清空 assignee', () => {
    taskStore.reconcileFromTickets([processingTicket('TKT-2026-0020')]);
    const id = taskStore.list()[0].id;
    taskStore.claim(id, user);
    taskStore.release(id);
    const task = taskStore.get(id)!;
    expect(task.status).toBe('pending');
    expect(task.assignee).toBeNull();
  });

  it('对非 in_progress 任务幂等忽略', () => {
    taskStore.reconcileFromTickets([processingTicket('TKT-2026-0021')]);
    const id = taskStore.list()[0].id;
    taskStore.release(id);
    expect(taskStore.get(id)!.status).toBe('pending');
  });
});

describe('taskStore 持久化降级', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('localStorage 抛错时仍在内存建任务', () => {
    const spy = vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('quota exceeded');
    });
    taskStore.reconcileFromTickets([processingTicket('TKT-2026-0030')]);
    expect(taskStore.list()).toHaveLength(1);
    expect(spy).toHaveBeenCalled();
  });
});

describe('startTaskBridge', () => {
  it('启动后工单进入 processing 会自动建任务', () => {
    const stop = startTaskBridge();
    const created = ticketStore.createTicket({
      serviceId: 'svc-1',
      title: '桥接验证',
      formValues: {},
      applicant: { name: '申请人', email: 'a@x.io' },
    });
    ticketStore.advance(created.ticket.id, 'submitted', 'approving');
    ticketStore.advance(created.ticket.id, 'approving', 'processing');
    expect(taskStore.list().some((tk) => tk.ticketId === created.ticket.id)).toBe(true);
    stop();
  });

  it('重复启动幂等，只订阅一次', () => {
    const stop1 = startTaskBridge();
    const stop2 = startTaskBridge();
    const created = ticketStore.createTicket({
      serviceId: 'svc-1',
      title: '幂等验证',
      formValues: {},
      applicant: { name: '申请人', email: 'a@x.io' },
    });
    ticketStore.advance(created.ticket.id, 'submitted', 'approving');
    ticketStore.advance(created.ticket.id, 'approving', 'processing');
    expect(taskStore.list().filter((tk) => tk.ticketId === created.ticket.id)).toHaveLength(1);
    stop1();
    stop2();
  });
});
