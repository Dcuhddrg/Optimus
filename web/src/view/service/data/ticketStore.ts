// web/src/view/service/data/ticketStore.ts

import { useSyncExternalStore } from 'react';

import type {
  CreateTicketInput,
  CreateTicketResult,
  StorageStatus,
  Ticket,
  TicketStatus,
} from '../ticketTypes';

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
    return parsed as Ticket[];
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
    persist(next); // 推进失败也接受，状态机继续
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
