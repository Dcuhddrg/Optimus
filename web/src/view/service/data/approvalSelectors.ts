// web/src/view/service/data/approvalSelectors.ts
// 当前审批步与进度均由 approvalChain 派生，不存冗余状态。

import type { ApprovalStep, Ticket } from '../ticketTypes';

export function currentApprovalStep(ticket: Ticket): ApprovalStep | null {
  return ticket.approvalChain.find((s) => s.status === 'pending') ?? null;
}

export function approvalProgress(ticket: Ticket): { approved: number; total: number } {
  return {
    approved: ticket.approvalChain.filter((s) => s.status === 'approved').length,
    total: ticket.approvalChain.length,
  };
}
