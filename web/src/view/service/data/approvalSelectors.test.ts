import { describe, expect, it } from 'vitest';

import { approvalProgress, currentApprovalStep } from './approvalSelectors';
import type { ApprovalStep, Ticket } from '../ticketTypes';

function ticketWith(chain: ApprovalStep[]): Ticket {
  return {
    id: 'TKT-2026-0001',
    serviceId: 'network.acl',
    title: 't',
    formValues: {},
    status: 'approving',
    createdAt: '2026-06-04T00:00:00.000Z',
    timeline: [],
    applicant: { name: 'a', email: 'a@x.io' },
    approvalChain: chain,
  };
}

const step = (status: ApprovalStep['status'], name = '王明'): ApprovalStep => ({
  roleI18nKey: 'servicePortal.ticket.approval.role.lineManager',
  approver: { name, email: 'w@x.io' },
  status,
});

describe('approvalSelectors', () => {
  it('currentApprovalStep 返回第一个 pending', () => {
    const t = ticketWith([step('approved', '王明'), step('pending', '李强'), step('pending', '张伟')]);
    expect(currentApprovalStep(t)?.approver.name).toBe('李强');
  });

  it('全部 approved 时 currentApprovalStep 为 null', () => {
    expect(currentApprovalStep(ticketWith([step('approved')]))).toBeNull();
  });

  it('approvalProgress 返回已通过/总数', () => {
    expect(approvalProgress(ticketWith([step('approved'), step('pending')]))).toEqual({
      approved: 1,
      total: 2,
    });
  });

  it('空链为 0/0 与 null', () => {
    const t = ticketWith([]);
    expect(approvalProgress(t)).toEqual({ approved: 0, total: 0 });
    expect(currentApprovalStep(t)).toBeNull();
  });
});
