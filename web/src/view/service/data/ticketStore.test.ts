import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ticketStore } from './ticketStore';

beforeEach(() => {
  ticketStore.__resetForTest();
});

function create(serviceId: string | null) {
  return ticketStore.createTicket({
    serviceId,
    title: '测试工单',
    formValues: {},
    applicant: { name: '申请人', email: 'a@x.io' },
  }).ticket;
}

function createApproving(serviceId: string | null) {
  const ticket = create(serviceId);
  ticketStore.advance(ticket.id, 'submitted', 'approving');
  return ticket.id;
}

describe('ticketStore.createTicket approvalChain 快照', () => {
  it('按服务类别快照审批链，全部 pending', () => {
    const ticket = create('cloud.compute-resize');
    expect(ticket.approvalChain).toHaveLength(3);
    expect(ticket.approvalChain.every((s) => s.status === 'pending')).toBe(true);
  });

  it('null 服务给默认 1 级链', () => {
    expect(create(null).approvalChain).toHaveLength(1);
  });
});

describe('ticketStore.advanceApproval', () => {
  it('逐级把 pending 标为 approved 并写 decidedAt + 事件', () => {
    const id = createApproving('network.acl'); // 2 级
    ticketStore.advanceApproval(id);
    const t1 = ticketStore.get(id)!;
    expect(t1.approvalChain[0].status).toBe('approved');
    expect(t1.approvalChain[0].decidedAt).toBeTruthy();
    expect(t1.approvalChain[1].status).toBe('pending');
    expect(t1.status).toBe('approving');
    const ev = t1.timeline[t1.timeline.length - 1];
    expect(ev.i18nNoteKey).toBe('servicePortal.ticket.event.approved');
    expect(ev.actor).toBe(t1.approvalChain[0].approver.name);
    expect(ev.roleI18nKey).toBe(t1.approvalChain[0].roleI18nKey);
  });

  it('最后一级通过后把工单推进到 processing', () => {
    const id = createApproving('network.acl');
    ticketStore.advanceApproval(id);
    ticketStore.advanceApproval(id);
    const t2 = ticketStore.get(id)!;
    expect(t2.approvalChain.every((s) => s.status === 'approved')).toBe(true);
    expect(t2.status).toBe('processing');
    expect(t2.timeline.map((e) => e.status)).toEqual([
      'submitted',
      'approving',
      'approving',
      'approving',
      'processing',
    ]);
  });

  it('非 approving 状态 no-op', () => {
    const ticket = create('account.role-grant'); // 仍 submitted
    ticketStore.advanceApproval(ticket.id);
    const after = ticketStore.get(ticket.id)!;
    expect(after.status).toBe('submitted');
    expect(after.approvalChain[0].status).toBe('pending');
  });
});

describe('ticketStore 旧数据兼容', () => {
  it('缺少 approvalChain 的旧工单补为空链，approving 时直接放行到 processing', async () => {
    window.localStorage.setItem(
      'optimus.servicePortal.tickets.v1',
      JSON.stringify([
        {
          id: 'TKT-2026-0001',
          serviceId: null,
          title: '旧工单',
          formValues: {},
          status: 'approving',
          createdAt: '2026-06-01T00:00:00.000Z',
          timeline: [],
          applicant: { name: 'a', email: 'a@x.io' },
        },
      ]),
    );
    vi.resetModules();
    const { ticketStore: fresh } = await import('./ticketStore');
    expect(fresh.get('TKT-2026-0001')!.approvalChain).toEqual([]);
    fresh.advanceApproval('TKT-2026-0001');
    expect(fresh.get('TKT-2026-0001')!.status).toBe('processing');
    fresh.__resetForTest();
  });
});
