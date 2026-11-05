// web/src/view/service/data/ticketStore.ts

import { useSyncExternalStore } from 'react';

import type {
  CreateTicketInput,
  CreateTicketResult,
  StorageStatus,
  Ticket,
  TicketEvent,
  TicketStatus,
} from '../ticketTypes';
import { buildApprovalChain } from './approvalChains';

const STORAGE_KEY = 'optimus.servicePortal.tickets.v1';

type Listener = () => void;

let tickets: Ticket[] = loadInitial();
const listeners = new Set<Listener>();

function loadInitial(): Ticket[] {
  if (typeof window === 'undefined') return [];
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    // 升级前创建的工单没有 approvalChain，补为空链：步骤条退化，审批直接放行。
    return (parsed as Ticket[]).map((t) => ({ ...t, approvalChain: t.approvalChain ?? [] }));
  } catch (error) {
    console.warn('[ticketStore] failed to parse localStorage; starting empty', error);
    return [];
  }
}

function persist(next: Ticket[]): StorageStatus {
  if (typeof window === 'undefined') return 'memory-only';
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    return 'persisted';
  } catch (error) {
    console.warn('[ticketStore] failed to persist; staying in memory', error);
    return 'memory-only';
  }
}

function notify() {
  for (const listener of listeners) listener();
}

function nextTicketId(now: Date, existing: Ticket[]): string {
  const year = now.getFullYear();
  const prefix = `TKT-${year}-`;
  const sameYear = existing.filter((t) => t.id.startsWith(prefix));
  const max = sameYear.reduce((acc, t) => {
    const n = Number(t.id.slice(prefix.length));
    return Number.isFinite(n) && n > acc ? n : acc;
  }, 0);
  return `${prefix}${String(max + 1).padStart(4, '0')}`;
}

export const ticketStore = {
  list(): Ticket[] {
    return tickets;
  },

  get(id: string): Ticket | undefined {
    return tickets.find((t) => t.id === id);
  },

  subscribe(listener: Listener): () => void {
    listeners.add(listener);
    return () => listeners.delete(listener);
  },

  createTicket(input: CreateTicketInput): CreateTicketResult {
    const now = new Date();
    const id = nextTicketId(now, tickets);
    const ticket: Ticket = {
      id,
      serviceId: input.serviceId,
      title: input.title,
      formValues: input.formValues,
      status: 'submitted',
      createdAt: now.toISOString(),
      timeline: [
        {
          at: now.toISOString(),
          status: 'submitted',
          i18nNoteKey: 'servicePortal.ticket.event.submitted',
        },
      ],
      applicant: input.applicant,
      approvalChain: buildApprovalChain(input.serviceId),
    };
    const next = [ticket, ...tickets];
    const storageStatus = persist(next);
    tickets = next;
    notify();
    return { ticket, storageStatus };
  },

  advance(ticketId: string, expectedFrom: TicketStatus, nextStatus: TicketStatus): void {
    const idx = tickets.findIndex((t) => t.id === ticketId);
    if (idx === -1) return;
    const current = tickets[idx];
    if (current.status !== expectedFrom) return;
    const at = new Date().toISOString();
    const updated: Ticket = {
      ...current,
      status: nextStatus,
      timeline: [
        ...current.timeline,
        {
          at,
          status: nextStatus,
          i18nNoteKey: 'servicePortal.ticket.event.autoAdvanced',
        },
      ],
    };
    const next = [...tickets.slice(0, idx), updated, ...tickets.slice(idx + 1)];
    // 故意丢弃 persist 返回的 storageStatus：advance 是定时器触发，没有用户上下文挂 banner；
    // 失败仅记录到 console.warn，让状态机继续推进。
    persist(next);
    tickets = next;
    notify();
  },

  // 审批链通过一级；没有待审步（含空链）时直接进入 processing，避免工单卡在 approving。
  advanceApproval(ticketId: string): void {
    const idx = tickets.findIndex((t) => t.id === ticketId);
    if (idx === -1) return;
    const current = tickets[idx];
    if (current.status !== 'approving') return;
    const chain = current.approvalChain;
    const pendingIdx = chain.findIndex((s) => s.status === 'pending');
    const at = new Date().toISOString();
    const processingEvent: TicketEvent = {
      at,
      status: 'processing',
      i18nNoteKey: 'servicePortal.ticket.event.autoAdvanced',
    };
    let updated: Ticket;
    if (pendingIdx === -1) {
      updated = { ...current, status: 'processing', timeline: [...current.timeline, processingEvent] };
    } else {
      const step = chain[pendingIdx];
      const nextChain = chain.map((s, i) =>
        i === pendingIdx ? { ...s, status: 'approved' as const, decidedAt: at } : s,
      );
      const stillPending = nextChain.some((s) => s.status === 'pending');
      const approvedEvent: TicketEvent = {
        at,
        status: 'approving',
        i18nNoteKey: 'servicePortal.ticket.event.approved',
        actor: step.approver.name,
        roleI18nKey: step.roleI18nKey,
      };
      updated = {
        ...current,
        approvalChain: nextChain,
        status: stillPending ? 'approving' : 'processing',
        timeline: [...current.timeline, approvedEvent, ...(stillPending ? [] : [processingEvent])],
      };
    }
    const next = [...tickets.slice(0, idx), updated, ...tickets.slice(idx + 1)];
    // 与 advance 一样由定时器触发，持久化失败只记录 console.warn。
    persist(next);
    tickets = next;
    notify();
  },

  // 仅供测试 / 重置使用，非 UI 入口
  __resetForTest(): void {
    tickets = [];
    persist(tickets);
    notify();
  },
};

export function useTicketStore() {
  const snapshot = useSyncExternalStore(
    ticketStore.subscribe,
    ticketStore.list,
    ticketStore.list,
  );
  return {
    tickets: snapshot,
    createTicket: ticketStore.createTicket,
    getTicket: ticketStore.get,
  };
}
