// web/src/view/service/data/ticketMockEngine.ts

import { ticketStore } from './ticketStore';

const STEP_5_MS = 5_000;
const STEP_30_MS = 30_000;

const timers = new Map<string, number[]>();

export const ticketMockEngine = {
  schedule(ticketId: string): void {
    cancelExisting(ticketId);
    const ids: number[] = [];
    ids.push(
      window.setTimeout(() => {
        ticketStore.advance(ticketId, 'submitted', 'approving');
      }, STEP_5_MS),
    );
    ids.push(
      window.setTimeout(() => {
        ticketStore.advance(ticketId, 'approving', 'processing');
      }, STEP_30_MS),
    );
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
