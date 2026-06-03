// web/src/view/service/data/taskStore.ts

import { useSyncExternalStore } from 'react';
import type { Ticket } from '../ticketTypes';
import type { OpsTask, TaskStorageStatus } from './taskTypes';
import { ticketStore } from './ticketStore';

const STORAGE_KEY = 'optimus.opsTasks.v1';

type Listener = () => void;

let tasks: OpsTask[] = loadInitial();
const listeners = new Set<Listener>();

function loadInitial(): OpsTask[] {
  if (typeof window === 'undefined') return [];
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed as OpsTask[];
  } catch (error) {
    console.warn('[taskStore] failed to parse localStorage; starting empty', error);
    return [];
  }
}

function persist(next: OpsTask[]): TaskStorageStatus {
  if (typeof window === 'undefined') return 'memory-only';
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    return 'persisted';
  } catch (error) {
    console.warn('[taskStore] failed to persist; staying in memory', error);
    return 'memory-only';
  }
}

function notify() {
  for (const listener of listeners) listener();
}

function nextTaskId(now: Date, existing: OpsTask[]): string {
  const year = now.getFullYear();
  const prefix = `TASK-${year}-`;
  const sameYear = existing.filter((task) => task.id.startsWith(prefix));
  const max = sameYear.reduce((acc, task) => {
    const n = Number(task.id.slice(prefix.length));
    return Number.isFinite(n) && n > acc ? n : acc;
  }, 0);
  return `${prefix}${String(max + 1).padStart(4, '0')}`;
}

export const taskStore = {
  list(): OpsTask[] {
    return tasks;
  },

  get(id: string): OpsTask | undefined {
    return tasks.find((task) => task.id === id);
  },

  subscribe(listener: Listener): () => void {
    listeners.add(listener);
    return () => listeners.delete(listener);
  },

  claim(id: string, user: { name: string; email: string }): void {
    const idx = tasks.findIndex((task) => task.id === id);
    if (idx === -1) return;
    const current = tasks[idx];
    if (current.status !== 'pending') return;
    const at = new Date().toISOString();
    const updated: OpsTask = {
      ...current,
      status: 'in_progress',
      assignee: { name: user.name, email: user.email },
      claimedAt: at,
      timeline: [
        ...current.timeline,
        { at, status: 'in_progress', i18nNoteKey: 'servicePortal.task.event.claimed', actor: user.name },
      ],
    };
    const next = [...tasks.slice(0, idx), updated, ...tasks.slice(idx + 1)];
    persist(next);
    tasks = next;
    notify();
  },

  complete(id: string): void {
    const idx = tasks.findIndex((task) => task.id === id);
    if (idx === -1) return;
    const current = tasks[idx];
    if (current.status !== 'in_progress') return;
    const at = new Date().toISOString();
    const updated: OpsTask = {
      ...current,
      status: 'done',
      completedAt: at,
      timeline: [
        ...current.timeline,
        { at, status: 'done', i18nNoteKey: 'servicePortal.task.event.completed', actor: current.assignee?.name },
      ],
    };
    const next = [...tasks.slice(0, idx), updated, ...tasks.slice(idx + 1)];
    persist(next);
    tasks = next;
    notify();
    ticketStore.advance(current.ticketId, 'processing', 'done');
  },

  release(id: string): void {
    const idx = tasks.findIndex((task) => task.id === id);
    if (idx === -1) return;
    const current = tasks[idx];
    if (current.status !== 'in_progress') return;
    const at = new Date().toISOString();
    const updated: OpsTask = {
      ...current,
      status: 'pending',
      assignee: null,
      claimedAt: undefined,
      timeline: [
        ...current.timeline,
        { at, status: 'pending', i18nNoteKey: 'servicePortal.task.event.released' },
      ],
    };
    const next = [...tasks.slice(0, idx), updated, ...tasks.slice(idx + 1)];
    persist(next);
    tasks = next;
    notify();
  },

  reconcileFromTickets(tickets: Ticket[]): void {
    const linked = new Set(tasks.map((task) => task.ticketId));
    const created: OpsTask[] = [];
    for (const ticket of tickets) {
      if (ticket.status !== 'processing') continue;
      if (linked.has(ticket.id)) continue;
      const now = new Date();
      const task: OpsTask = {
        id: nextTaskId(now, [...tasks, ...created]),
        ticketId: ticket.id,
        title: ticket.title,
        serviceId: ticket.serviceId,
        status: 'pending',
        assignee: null,
        createdAt: now.toISOString(),
        timeline: [
          { at: now.toISOString(), status: 'pending', i18nNoteKey: 'servicePortal.task.event.created' },
        ],
      };
      created.push(task);
      linked.add(ticket.id);
    }
    if (created.length === 0) return;
    const next = [...created, ...tasks];
    persist(next);
    tasks = next;
    notify();
  },

  // 仅供测试 / 重置使用，非 UI 入口
  __resetForTest(): void {
    tasks = [];
    persist(tasks);
    notify();
  },
};

let bridgeUnsub: (() => void) | null = null;

/**
 * 启动 ticket → task 协调桥：订阅 ticketStore，并立即回填一次。
 * 幂等：重复调用复用已有订阅。返回停止函数。
 */
export function startTaskBridge(): () => void {
  if (!bridgeUnsub) {
    const reconcile = () => taskStore.reconcileFromTickets(ticketStore.list());
    bridgeUnsub = ticketStore.subscribe(reconcile);
    reconcile();
  }
  return () => {
    bridgeUnsub?.();
    bridgeUnsub = null;
  };
}

export function useTaskStore() {
  const snapshot = useSyncExternalStore(taskStore.subscribe, taskStore.list, taskStore.list);
  return {
    tasks: snapshot,
    claim: taskStore.claim,
    complete: taskStore.complete,
    release: taskStore.release,
    getTask: taskStore.get,
  };
}
