// web/src/view/service/data/ticketMockEngine.ts

import { ticketStore } from './ticketStore';

const STEP_SUBMITTED_MS = 5_000;
const STEP_APPROVAL_MS = 10_000;

const timers = new Map<string, number[]>();

export const ticketMockEngine = {
  // submitted→approving 5s 后，审批链每 10s 通过一级；最后一级通过后进入 processing。
  schedule(ticketId: string): void {
    cancelExisting(ticketId);
    const ids: number[] = [];
    ids.push(
      window.setTimeout(() => {
        ticketStore.advance(ticketId, 'submitted', 'approving');
      }, STEP_SUBMITTED_MS),
    );
    // 空链也安排一次，让 advanceApproval 把工单放行到 processing。
    const steps = Math.max(ticketStore.get(ticketId)?.approvalChain.length ?? 0, 1);
    for (let i = 1; i <= steps; i += 1) {
      ids.push(
        window.setTimeout(() => {
          ticketStore.advanceApproval(ticketId);
        }, STEP_SUBMITTED_MS + i * STEP_APPROVAL_MS),
      );
    }
    timers.set(ticketId, ids);
  },

  cancel(ticketId: string): void {
    cancelExisting(ticketId);
  },
};

function cancelExisting(ticketId: string): void {
  const ids = timers.get(ticketId);
  if (!ids) return;
  for (const id of ids) window.clearTimeout(id);
  timers.delete(ticketId);
}
