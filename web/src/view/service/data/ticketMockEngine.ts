// web/src/view/service/data/ticketMockEngine.ts

import { useEffect } from 'react';

import { ticketStore } from './ticketStore';

const STEP_SUBMITTED_MS = 5_000;
const STEP_APPROVAL_MS = 10_000;

// 定时器只在内存中；页面刷新后由 resumePending 为未完成的工单重新安排。
const timers = new Map<string, number[]>();

export const ticketMockEngine = {
  // submitted→approving 5s 后，审批链每 10s 通过一级；最后一级通过后进入 processing。
  schedule(ticketId: string): void {
    cancelExisting(ticketId);
    const ids = [
      window.setTimeout(() => {
        ticketStore.advance(ticketId, 'submitted', 'approving');
      }, STEP_SUBMITTED_MS),
    ];
    // 空链也安排一次，让 advanceApproval 把工单放行到 processing。
    const steps = Math.max(ticketStore.get(ticketId)?.approvalChain.length ?? 0, 1);
    ids.push(...scheduleApprovals(ticketId, 0, steps, STEP_SUBMITTED_MS));
    timers.set(ticketId, ids);
  },

  // 为没有定时器的 submitted/approving 工单接着推进：submitted 从头安排；
  // approving 按剩余待审级数安排，没有待审级（如升级前的旧工单）时立即放行到 processing。
  resumePending(): void {
    for (const ticket of ticketStore.list()) {
      if (timers.has(ticket.id)) continue;
      if (ticket.status === 'submitted') {
        ticketMockEngine.schedule(ticket.id);
      } else if (ticket.status === 'approving') {
        const firstPending = ticket.approvalChain.findIndex((s) => s.status === 'pending');
        if (firstPending === -1) {
          ticketStore.advanceApproval(ticket.id);
        } else {
          const remaining = ticket.approvalChain.length - firstPending;
          timers.set(ticket.id, scheduleApprovals(ticket.id, firstPending, remaining, 0));
        }
      }
    }
  },

  cancel(ticketId: string): void {
    cancelExisting(ticketId);
  },

  // 仅供测试使用
  __resetForTest(): void {
    for (const id of [...timers.keys()]) cancelExisting(id);
  },
};

// 服务门户各页面挂载时调用，保证刷新后未完成的工单继续推进。
export function useTicketMockEngine(): void {
  useEffect(() => {
    ticketMockEngine.resumePending();
  }, []);
}

// 第 i 个定时器只审批 fromLevel+i-1 这一级；多个标签页各自安排时不会连跳多级。
function scheduleApprovals(ticketId: string, fromLevel: number, steps: number, offsetMs: number): number[] {
  const ids: number[] = [];
  for (let i = 1; i <= steps; i += 1) {
    const level = fromLevel + i - 1;
    ids.push(
      window.setTimeout(() => {
        ticketStore.advanceApproval(ticketId, level);
      }, offsetMs + i * STEP_APPROVAL_MS),
    );
  }
  return ids;
}

function cancelExisting(ticketId: string): void {
  const ids = timers.get(ticketId);
  if (!ids) return;
  for (const id of ids) window.clearTimeout(id);
  timers.delete(ticketId);
}
