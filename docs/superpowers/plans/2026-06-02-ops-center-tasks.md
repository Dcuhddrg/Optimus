# 运维中心 · 任务管理（Ops Tasks）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在运维中心补齐「任务管理」模块——把工单派生为可领取/处理/完成的运维任务，任务完成驱动工单进入 done，纯前端实现。

**Architecture:** 方案 A——新增独立 `taskStore`（沿用 `ticketStore` 的 `useSyncExternalStore` + `localStorage` + listeners 约定），任务通过 `ticketId` 关联工单。一个幂等协调桥 `startTaskBridge()` 订阅 `ticketStore`，把每个进入 `processing` 的工单补建为 `待领取` 任务；`taskStore.complete()` 调用 `ticketStore.advance(processing→done)` 实现协同。`ticketMockEngine` 去掉 `processing→done` 的定时步，该转移改由任务完成驱动。

**Tech Stack:** React 18 + TypeScript + react-router-dom 6 + framer-motion + CSS Modules；测试用 vitest + jsdom（本计划首次引入）。

> **设计细化说明（相对 spec）：** spec 写「taskStore 在模块初始化时订阅 ticketStore」。为保持 `taskStore` 纯净、便于单测，本计划把订阅抽成显式幂等函数 `startTaskBridge()`，由 `AppShell` 在 `useEffect` 中启动。行为与 spec 一致（登录后任一页面都会启动桥并回填任务），仅初始化时机更可控。

参考实现（务必对照模仿）：
- `web/src/view/service/data/ticketStore.ts`（store 结构、persist 降级、`__resetForTest`）
- `web/src/view/service/data/ticketMockEngine.ts`（定时引擎）
- `web/src/view/service/components/TicketStatusBadge.tsx` + `.module.css`（状态徽章）
- `web/src/view/service/TicketListPage.tsx` + `.module.css`（列表/筛选/表格）
- `web/src/view/service/TicketDetailPage.tsx`（详情页）
- `web/src/locale/messages.ts`（`servicePortal.ticket.*` 块，en 与 zhCN 双份）

关键约定：
- `t(key)` 只接受类型化的 `MessageKey`，**不做参数插值**；插值在调用处用 `.replace('{x}', val)` 手工完成（见 `TicketTimeline.tsx`）。
- 新增 i18n key 必须**同时**加到 `en` 与 `zhCN`（`zhCN: typeof en`，缺键会编译报错）。
- 当前登录用户：`useAuth().current?.user`（`displayName` / `username` / `email`）。

---

## File Structure

新增：
- `web/vitest.config.ts` — 测试配置（jsdom 环境）
- `web/src/view/service/data/taskTypes.ts` — 任务类型定义
- `web/src/view/service/data/taskStore.ts` — 任务 store + 协调桥 + `useTaskStore`
- `web/src/view/service/data/taskStore.test.ts` — store/桥接单测
- `web/src/view/service/data/ticketMockEngine.test.ts` — 引擎定时步单测
- `web/src/view/service/components/TaskStatusBadge.tsx` + `.module.css` — 任务状态徽章
- `web/src/view/service/TaskListPage.tsx` + `.module.css` — 任务列表页

修改：
- `web/package.json` — 加 vitest/jsdom 依赖与 test 脚本
- `web/tsconfig.json` — 构建排除测试文件
- `web/src/view/service/data/ticketMockEngine.ts` — 移除 60s `processing→done` 步
- `web/src/layout/AppShell.tsx` — 启动协调桥
- `web/src/components/nav/SideNav.tsx` — 启用「任务管理」NavLink
- `web/src/router/index.tsx` — 新增 `/service-portal/tasks` 路由
- `web/src/view/service/TicketDetailPage.tsx` — 新增「关联任务」展示块
- `web/src/locale/messages.ts` — 新增 `servicePortal.task.*`（en + zhCN）

---

## Task 1: 引入 vitest 测试基建

**Files:**
- Create: `web/vitest.config.ts`
- Modify: `web/package.json`
- Modify: `web/tsconfig.json`

- [ ] **Step 1: 安装依赖**

Run（在 `web/` 目录）:
```bash
cd web && pnpm add -D vitest@^2.1.8 jsdom@^25.0.1
```
Expected: `package.json` 的 `devDependencies` 出现 `vitest` 与 `jsdom`，pnpm 安装成功。

- [ ] **Step 2: 创建 vitest 配置**

Create `web/vitest.config.ts`:
```ts
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    include: ['src/**/*.test.{ts,tsx}'],
    clearMocks: true,
  },
});
```

- [ ] **Step 3: 加 test 脚本**

Modify `web/package.json` 的 `scripts`，加入：
```json
    "test": "vitest run",
    "test:watch": "vitest"
```
（保留已有 `dev` / `build` / `preview`。）

- [ ] **Step 4: 构建排除测试文件**

Modify `web/tsconfig.json`，在顶层对象加入 `exclude`（与 `include` 同级）：
```json
  "include": ["src"],
  "exclude": ["src/**/*.test.ts", "src/**/*.test.tsx"],
```
（保留已有 `references`。这样 `pnpm build` 的 `tsc -b` 不会把测试文件纳入产物类型检查。）

- [ ] **Step 5: 冒烟验证**

Create 临时文件 `web/src/__smoke.test.ts`:
```ts
import { describe, it, expect } from 'vitest';

describe('smoke', () => {
  it('runs', () => {
    expect(1 + 1).toBe(2);
  });
});
```
Run: `cd web && pnpm test`
Expected: 1 个测试通过。随后删除该文件：`rm web/src/__smoke.test.ts`

- [ ] **Step 6: Commit**

```bash
git add web/package.json web/pnpm-lock.yaml web/vitest.config.ts web/tsconfig.json
git commit -m "chore(web): add vitest + jsdom test setup"
```

---

## Task 2: 任务类型定义

**Files:**
- Create: `web/src/view/service/data/taskTypes.ts`

- [ ] **Step 1: 写类型**

Create `web/src/view/service/data/taskTypes.ts`:
```ts
// web/src/view/service/data/taskTypes.ts

export type TaskStatus = 'pending' | 'in_progress' | 'done';

export interface TaskEvent {
  at: string; // ISO timestamp
  status: TaskStatus;
  i18nNoteKey?: string;
  actor?: string; // displayName，便于时间线展示
}

export interface OpsTask {
  id: string; // 'TASK-YYYY-NNNN'
  ticketId: string; // 关联工单
  title: string; // 创建时从工单拷贝
  serviceId: string | null;
  status: TaskStatus;
  assignee: { name: string; email: string } | null;
  createdAt: string; // ISO
  claimedAt?: string;
  completedAt?: string;
  timeline: TaskEvent[];
}

export type TaskStorageStatus = 'persisted' | 'memory-only';
```

- [ ] **Step 2: 类型编译校验**

Run: `cd web && pnpm exec tsc -b`
Expected: 无错误（纯类型文件，暂无引用）。

- [ ] **Step 3: Commit**

```bash
git add web/src/view/service/data/taskTypes.ts
git commit -m "feat(ops-tasks): add OpsTask type definitions"
```

---

## Task 3: taskStore —— reconcile 建任务

本任务建立 store 骨架与 `reconcileFromTickets`，TDD。

**Files:**
- Create: `web/src/view/service/data/taskStore.ts`
- Test: `web/src/view/service/data/taskStore.test.ts`

- [ ] **Step 1: 写失败测试**

Create `web/src/view/service/data/taskStore.test.ts`:
```ts
import { beforeEach, describe, expect, it } from 'vitest';

import { taskStore } from './taskStore';
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
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd web && pnpm test`
Expected: FAIL —— 无法解析 `./taskStore`（模块不存在）。

- [ ] **Step 3: 写最小实现**

Create `web/src/view/service/data/taskStore.ts`:
```ts
// web/src/view/service/data/taskStore.ts

import { useSyncExternalStore } from 'react';

import type { Ticket } from '../ticketTypes';
import type { OpsTask, TaskStorageStatus } from './taskTypes';

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

  reconcileFromTickets(tickets: Ticket[]): void {
    const linked = new Set(tasks.map((task) => task.ticketId));
    const created: OpsTask[] = [];
    let working = tasks;
    for (const ticket of tickets) {
      if (ticket.status !== 'processing') continue;
      if (linked.has(ticket.id)) continue;
      const now = new Date();
      const task: OpsTask = {
        id: nextTaskId(now, [...working, ...created]),
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
    const next = [...created, ...working];
    persist(next);
    tasks = next;
    working = next;
    notify();
  },

  // 仅供测试 / 重置使用，非 UI 入口
  __resetForTest(): void {
    tasks = [];
    persist(tasks);
    notify();
  },
};
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd web && pnpm test`
Expected: PASS（3 个 reconcile 测试通过）。

- [ ] **Step 5: Commit**

```bash
git add web/src/view/service/data/taskStore.ts web/src/view/service/data/taskStore.test.ts
git commit -m "feat(ops-tasks): taskStore reconcile creates pending tasks from processing tickets"
```

---

## Task 4: taskStore.claim —— 领取

**Files:**
- Modify: `web/src/view/service/data/taskStore.ts`
- Test: `web/src/view/service/data/taskStore.test.ts`

- [ ] **Step 1: 加失败测试**

在 `taskStore.test.ts` 末尾追加：
```ts
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
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd web && pnpm test`
Expected: FAIL —— `taskStore.claim is not a function`。

- [ ] **Step 3: 实现 claim**

在 `taskStore.ts` 的 `taskStore` 对象中，`reconcileFromTickets` 之前加入方法（并在文件顶部按需复用已有 import）：
```ts
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
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd web && pnpm test`
Expected: PASS（claim 两个测试 + 之前全部通过）。

- [ ] **Step 5: Commit**

```bash
git add web/src/view/service/data/taskStore.ts web/src/view/service/data/taskStore.test.ts
git commit -m "feat(ops-tasks): taskStore.claim assigns task to current user"
```

---

## Task 5: taskStore.complete —— 完成并驱动工单

**Files:**
- Modify: `web/src/view/service/data/taskStore.ts`
- Test: `web/src/view/service/data/taskStore.test.ts`

- [ ] **Step 1: 加失败测试**

在 `taskStore.test.ts` 末尾追加（用真实 `ticketStore` 验证驱动）：
```ts
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
    // 未领取（仍 pending）直接 complete 应无效
    taskStore.complete(taskId);
    expect(taskStore.get(taskId)!.status).toBe('pending');
  });
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd web && pnpm test`
Expected: FAIL —— `taskStore.complete is not a function`。

- [ ] **Step 3: 实现 complete**

在 `taskStore.ts` 顶部 import 加入 `ticketStore`：
```ts
import { ticketStore } from './ticketStore';
```
在 `taskStore` 对象中 `claim` 之后加入：
```ts
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
        {
          at,
          status: 'done',
          i18nNoteKey: 'servicePortal.task.event.completed',
          actor: current.assignee?.name,
        },
      ],
    };
    const next = [...tasks.slice(0, idx), updated, ...tasks.slice(idx + 1)];
    persist(next);
    tasks = next;
    notify();
    // 协同：任务完成驱动关联工单 processing → done
    ticketStore.advance(current.ticketId, 'processing', 'done');
  },
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd web && pnpm test`
Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add web/src/view/service/data/taskStore.ts web/src/view/service/data/taskStore.test.ts
git commit -m "feat(ops-tasks): taskStore.complete drives linked ticket to done"
```

---

## Task 6: taskStore.release —— 退回/释放

**Files:**
- Modify: `web/src/view/service/data/taskStore.ts`
- Test: `web/src/view/service/data/taskStore.test.ts`

- [ ] **Step 1: 加失败测试**

在 `taskStore.test.ts` 末尾追加：
```ts
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
    taskStore.release(id); // 仍 pending
    expect(taskStore.get(id)!.status).toBe('pending');
  });
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd web && pnpm test`
Expected: FAIL —— `taskStore.release is not a function`。

- [ ] **Step 3: 实现 release**

在 `taskStore` 对象中 `complete` 之后加入：
```ts
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
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd web && pnpm test`
Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add web/src/view/service/data/taskStore.ts web/src/view/service/data/taskStore.test.ts
git commit -m "feat(ops-tasks): taskStore.release returns task to pending"
```

---

## Task 7: 持久化降级测试

验证 `localStorage.setItem` 抛错时状态仍在内存推进（与 ticketStore 一致的容错）。

**Files:**
- Test: `web/src/view/service/data/taskStore.test.ts`

- [ ] **Step 1: 加失败/回归测试**

在 `taskStore.test.ts` 顶部 import 补充 `vi`：
```ts
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
```
在文件末尾追加：
```ts
describe('taskStore 持久化降级', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('localStorage 抛错时仍在内存建任务', () => {
    const spy = vi.spyOn(window.localStorage, 'setItem').mockImplementation(() => {
      throw new Error('quota exceeded');
    });
    taskStore.reconcileFromTickets([processingTicket('TKT-2026-0030')]);
    expect(taskStore.list()).toHaveLength(1);
    expect(spy).toHaveBeenCalled();
  });
});
```

- [ ] **Step 2: 运行测试确认通过**

Run: `cd web && pnpm test`
Expected: PASS（实现已具备 try/catch 降级；本测试锁定该行为，不需改实现）。
若 FAIL，检查 `persist()` 是否吞掉异常返回 `'memory-only'`。

- [ ] **Step 3: Commit**

```bash
git add web/src/view/service/data/taskStore.test.ts
git commit -m "test(ops-tasks): taskStore survives localStorage failure"
```

---

## Task 8: 协调桥 startTaskBridge + useTaskStore hook

**Files:**
- Modify: `web/src/view/service/data/taskStore.ts`
- Test: `web/src/view/service/data/taskStore.test.ts`

- [ ] **Step 1: 加失败测试**

在 `taskStore.test.ts` 顶部 import 加上 `startTaskBridge`：
```ts
import { startTaskBridge, taskStore } from './taskStore';
```
在文件末尾追加：
```ts
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
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd web && pnpm test`
Expected: FAIL —— `startTaskBridge` 未导出。

- [ ] **Step 3: 实现桥 + hook**

在 `taskStore.ts` 末尾（`taskStore` 对象之后）加入：
```ts
let bridgeUnsub: (() => void) | null = null;

/**
 * 启动 ticket → task 协调桥：订阅 ticketStore，并立即回填一次。
 * 幂等：重复调用复用已有订阅。返回停止函数。
 */
export function startTaskBridge(): () => void {
  if (!bridgeUnsub) {
    const reconcile = () => taskStore.reconcileFromTickets(ticketStore.list());
    bridgeUnsub = ticketStore.subscribe(reconcile);
    reconcile(); // 立即回填已处于 processing 的工单
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
```

> 说明：测试里 `__resetForTest()` 不会断开桥，但因为测试用 `stop()` 主动停桥、且 `beforeEach` 会重置数据，幂等断言依然成立。`bridgeUnsub` 是模块级单例，符合「桥只订阅一次」的预期。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd web && pnpm test`
Expected: PASS（全部测试）。

- [ ] **Step 5: Commit**

```bash
git add web/src/view/service/data/taskStore.ts web/src/view/service/data/taskStore.test.ts
git commit -m "feat(ops-tasks): add startTaskBridge + useTaskStore hook"
```

---

## Task 9: ticketMockEngine 移除 60s done 步

**Files:**
- Modify: `web/src/view/service/data/ticketMockEngine.ts`
- Test: `web/src/view/service/data/ticketMockEngine.test.ts`

- [ ] **Step 1: 写失败测试（用 fake timers）**

Create `web/src/view/service/data/ticketMockEngine.test.ts`:
```ts
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { ticketMockEngine } from './ticketMockEngine';
import { ticketStore } from './ticketStore';

beforeEach(() => {
  ticketStore.__resetForTest();
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('ticketMockEngine', () => {
  it('推进到 processing 后不再自动 done', () => {
    const created = ticketStore.createTicket({
      serviceId: 'svc-1',
      title: '定时验证',
      formValues: {},
      applicant: { name: '申请人', email: 'a@x.io' },
    });
    const id = created.ticket.id;
    ticketMockEngine.schedule(id);

    vi.advanceTimersByTime(5_000);
    expect(ticketStore.get(id)!.status).toBe('approving');

    vi.advanceTimersByTime(25_000); // 累计 30s
    expect(ticketStore.get(id)!.status).toBe('processing');

    vi.advanceTimersByTime(120_000); // 再过很久
    expect(ticketStore.get(id)!.status).toBe('processing'); // 不再自动 done
  });
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd web && pnpm test`
Expected: FAIL —— 最后断言为 `done`（旧实现 60s 自动推进）。

- [ ] **Step 3: 移除 60s 步**

Modify `web/src/view/service/data/ticketMockEngine.ts`：
1. 删除常量 `const STEP_60_MS = 60_000;`
2. 删除以下整段 `ids.push(...)`：
```ts
    ids.push(
      window.setTimeout(() => {
        ticketStore.advance(ticketId, 'processing', 'done');
      }, STEP_60_MS),
    );
```
保留 `STEP_5_MS`、`STEP_30_MS` 两步与其余结构不变。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd web && pnpm test`
Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add web/src/view/service/data/ticketMockEngine.ts web/src/view/service/data/ticketMockEngine.test.ts
git commit -m "feat(ops-tasks): ticket done is driven by task completion, not timer"
```

---

## Task 10: i18n —— servicePortal.task.*（en + zhCN）

**Files:**
- Modify: `web/src/locale/messages.ts`

- [ ] **Step 1: 加 en 文案**

在 `web/src/locale/messages.ts` 的 **en** 块里，`servicePortal.ticket` 对象**之后**、`sla:` 之前插入 `task` 对象（与 `ticket` 同级缩进，注意 `ticket` 块结尾的 `},` 后插入）：
```ts
      task: {
        list: {
          eyebrow: 'Ops center',
          title: 'Ops tasks',
          description: 'Execution tasks derived from approved tickets.',
          searchPlaceholder: 'Search by task id, title or ticket id',
          filterLabel: 'Filter by status',
          mineOnly: 'My tasks only',
          empty: 'No tasks match the current filter.',
          columnId: 'Task id',
          columnTitle: 'Title / Ticket',
          columnService: 'Service',
          columnStatus: 'Status',
          columnAssignee: 'Assignee',
          columnActions: 'Actions',
          unassigned: 'Unassigned',
          claim: 'Claim',
          complete: 'Complete',
          release: 'Release',
          view: 'View ticket',
        },
        status: {
          pending: 'Pending',
          in_progress: 'In progress',
          done: 'Done',
        },
        event: {
          created: 'Task created from ticket.',
          claimed: 'Claimed by {actor}.',
          completed: 'Completed by {actor}.',
          released: 'Task released back to pending.',
        },
        ticketLink: {
          label: 'Linked task',
          none: 'No task yet (ticket not in processing).',
        },
      },
```

- [ ] **Step 2: 加 zhCN 文案**

在 **zhCN** 块里，对应 `servicePortal.ticket` 之后、`sla:` 之前插入：
```ts
      task: {
        list: {
          eyebrow: '运维中心',
          title: '任务管理',
          description: '由已审批工单派生的执行任务。',
          searchPlaceholder: '按任务号、标题或工单号搜索',
          filterLabel: '按状态过滤',
          mineOnly: '只看我的',
          empty: '当前过滤条件下没有任务。',
          columnId: '任务号',
          columnTitle: '标题 / 工单',
          columnService: '服务',
          columnStatus: '状态',
          columnAssignee: '领取人',
          columnActions: '操作',
          unassigned: '未领取',
          claim: '领取',
          complete: '完成',
          release: '退回',
          view: '查看工单',
        },
        status: {
          pending: '待领取',
          in_progress: '进行中',
          done: '已完成',
        },
        event: {
          created: '已由工单派生任务。',
          claimed: '{actor} 领取了任务。',
          completed: '{actor} 完成了任务。',
          released: '任务已退回待领取。',
        },
        ticketLink: {
          label: '关联任务',
          none: '暂无任务（工单未进入处理中）。',
        },
      },
```

- [ ] **Step 3: 类型/编译校验**

Run: `cd web && pnpm exec tsc -b`
Expected: 无错误（en 与 zhCN 结构一致；`zhCN: typeof en` 不报缺键）。

- [ ] **Step 4: Commit**

```bash
git add web/src/locale/messages.ts
git commit -m "feat(ops-tasks): add servicePortal.task i18n keys (en + zh)"
```

---

## Task 11: TaskStatusBadge 组件

**Files:**
- Create: `web/src/view/service/components/TaskStatusBadge.tsx`
- Create: `web/src/view/service/components/TaskStatusBadge.module.css`

- [ ] **Step 1: 写组件（镜像 TicketStatusBadge）**

Create `web/src/view/service/components/TaskStatusBadge.tsx`:
```tsx
// TaskStatusBadge.tsx
import { AnimatePresence, motion } from 'framer-motion';

import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { TaskStatus } from '../data/taskTypes';
import styles from './TaskStatusBadge.module.css';

interface Props {
  status: TaskStatus;
}

const STATUS_I18N: Record<TaskStatus, MessageKey> = {
  pending: 'servicePortal.task.status.pending',
  in_progress: 'servicePortal.task.status.in_progress',
  done: 'servicePortal.task.status.done',
};

export function TaskStatusBadge({ status }: Props) {
  const { t } = useI18n();
  return (
    <AnimatePresence mode="wait" initial={false}>
      <motion.span
        key={status}
        className={`${styles.badge} ${styles[status]}`}
        initial={{ scale: 0.85, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.85, opacity: 0 }}
        transition={{ type: 'spring', stiffness: 380, damping: 24 }}
      >
        {t(STATUS_I18N[status])}
      </motion.span>
    </AnimatePresence>
  );
}
```

- [ ] **Step 2: 写样式（复用工单徽章配色语义）**

Create `web/src/view/service/components/TaskStatusBadge.module.css`:
```css
/* TaskStatusBadge.module.css */

.badge {
  display: inline-flex;
  align-items: center;
  min-height: 1.5rem;
  padding: 0 0.55rem;
  border-radius: 999px;
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.01em;
  white-space: nowrap;
}

.pending {
  background: rgba(147, 93, 45, 0.14);
  color: #935d2d;
}

.in_progress {
  background: rgba(54, 93, 126, 0.14);
  color: #365d7e;
}

.done {
  background: rgba(47, 98, 92, 0.16);
  color: #2f625c;
}
```

- [ ] **Step 3: 编译校验**

Run: `cd web && pnpm exec tsc -b`
Expected: 无错误。

- [ ] **Step 4: Commit**

```bash
git add web/src/view/service/components/TaskStatusBadge.tsx web/src/view/service/components/TaskStatusBadge.module.css
git commit -m "feat(ops-tasks): add TaskStatusBadge component"
```

---

## Task 12: TaskListPage 列表页

**Files:**
- Create: `web/src/view/service/TaskListPage.tsx`
- Create: `web/src/view/service/TaskListPage.module.css`

- [ ] **Step 1: 写页面（镜像 TicketListPage，加领取人列与行内操作）**

Create `web/src/view/service/TaskListPage.tsx`:
```tsx
// TaskListPage.tsx
import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { useAuth } from '../../auth/AuthContext';
import { useI18n } from '../../locale/useI18n';
import type { MessageKey } from '../../locale/messages';
import { PageHeader } from '../../components/page/PageHeader';
import { TaskStatusBadge } from './components/TaskStatusBadge';
import { useTaskStore } from './data/taskStore';
import type { TaskStatus } from './data/taskTypes';
import styles from './TaskListPage.module.css';

const ALL_STATUSES: TaskStatus[] = ['pending', 'in_progress', 'done'];

export function TaskListPage() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const { current } = useAuth();
  const { tasks, claim, complete, release } = useTaskStore();
  const [search, setSearch] = useState('');
  const [activeStatuses, setActiveStatuses] = useState<Set<TaskStatus>>(new Set());
  const [mineOnly, setMineOnly] = useState(false);

  const me = {
    name: current?.user.displayName ?? current?.user.username ?? '',
    email: current?.user.email ?? '',
  };

  const filtered = useMemo(() => {
    const query = search.trim().toLowerCase();
    return tasks.filter((task) => {
      if (activeStatuses.size > 0 && !activeStatuses.has(task.status)) return false;
      if (mineOnly && task.assignee?.email !== me.email) return false;
      if (query) {
        const hay = `${task.id} ${task.title} ${task.ticketId}`.toLowerCase();
        if (!hay.includes(query)) return false;
      }
      return true;
    });
  }, [tasks, search, activeStatuses, mineOnly, me.email]);

  function toggleStatus(status: TaskStatus) {
    setActiveStatuses((prev) => {
      const next = new Set(prev);
      if (next.has(status)) next.delete(status);
      else next.add(status);
      return next;
    });
  }

  return (
    <div className={styles.page}>
      <PageHeader
        eyebrow={t('servicePortal.task.list.eyebrow')}
        title={t('servicePortal.task.list.title')}
        description={t('servicePortal.task.list.description')}
      />

      <div className={styles.controls}>
        <input
          className={styles.search}
          placeholder={t('servicePortal.task.list.searchPlaceholder')}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <div className={styles.chips} role="group" aria-label={t('servicePortal.task.list.filterLabel')}>
          {ALL_STATUSES.map((status) => {
            const active = activeStatuses.has(status);
            return (
              <button
                key={status}
                type="button"
                className={active ? `${styles.chip} ${styles.chipActive}` : styles.chip}
                onClick={() => toggleStatus(status)}
              >
                {t(`servicePortal.task.status.${status}` as MessageKey)}
              </button>
            );
          })}
          <button
            type="button"
            className={mineOnly ? `${styles.chip} ${styles.chipActive}` : styles.chip}
            onClick={() => setMineOnly((v) => !v)}
          >
            {t('servicePortal.task.list.mineOnly')}
          </button>
        </div>
      </div>

      <table className={styles.table}>
        <thead>
          <tr>
            <th>{t('servicePortal.task.list.columnId')}</th>
            <th>{t('servicePortal.task.list.columnTitle')}</th>
            <th>{t('servicePortal.task.list.columnService')}</th>
            <th>{t('servicePortal.task.list.columnStatus')}</th>
            <th>{t('servicePortal.task.list.columnAssignee')}</th>
            <th>{t('servicePortal.task.list.columnActions')}</th>
          </tr>
        </thead>
        <tbody>
          {filtered.length === 0 ? (
            <tr>
              <td colSpan={6} className={styles.empty}>
                {t('servicePortal.task.list.empty')}
              </td>
            </tr>
          ) : (
            filtered.map((task) => {
              const isMine = task.assignee?.email === me.email;
              return (
                <tr key={task.id}>
                  <td className={styles.mono}>{task.id}</td>
                  <td>
                    <div className={styles.titleCell}>
                      <span>{task.title}</span>
                      <span className={styles.mono}>{task.ticketId}</span>
                    </div>
                  </td>
                  <td>{task.serviceId ?? '—'}</td>
                  <td>
                    <TaskStatusBadge status={task.status} />
                  </td>
                  <td>{task.assignee?.name ?? t('servicePortal.task.list.unassigned')}</td>
                  <td>
                    <div className={styles.actions}>
                      {task.status === 'pending' ? (
                        <button type="button" className={styles.action} onClick={() => claim(task.id, me)}>
                          {t('servicePortal.task.list.claim')}
                        </button>
                      ) : null}
                      {task.status === 'in_progress' && isMine ? (
                        <>
                          <button type="button" className={styles.action} onClick={() => complete(task.id)}>
                            {t('servicePortal.task.list.complete')}
                          </button>
                          <button
                            type="button"
                            className={styles.actionGhost}
                            onClick={() => release(task.id)}
                          >
                            {t('servicePortal.task.list.release')}
                          </button>
                        </>
                      ) : null}
                      <button
                        type="button"
                        className={styles.actionGhost}
                        onClick={() => navigate(`/service-portal/tickets/${task.ticketId}`)}
                      >
                        {t('servicePortal.task.list.view')}
                      </button>
                    </div>
                  </td>
                </tr>
              );
            })
          )}
        </tbody>
      </table>
    </div>
  );
}
```

- [ ] **Step 2: 写样式（复用 TicketListPage 配色，加操作按钮）**

Create `web/src/view/service/TaskListPage.module.css`:
```css
.page {
  display: grid;
  gap: 1.2rem;
}

.controls {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  align-items: center;
}

.search {
  flex: 1 1 18rem;
  min-height: 2.35rem;
  padding: 0 0.85rem;
  border: 1px solid rgba(86, 70, 55, 0.2);
  border-radius: 0.65rem;
  background: #fffdf9;
  font: inherit;
  color: #24201c;
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
}

.chip {
  min-height: 1.85rem;
  padding: 0 0.7rem;
  border: 1px solid rgba(86, 70, 55, 0.2);
  border-radius: 999px;
  background: #fffdf9;
  color: #24201c;
  font: inherit;
  font-size: 0.78rem;
  cursor: pointer;
}

.chipActive {
  background: #24201c;
  color: #fffdf9;
  border-color: #24201c;
}

.chip:active {
  transform: translateY(1px);
  transition: transform 80ms ease;
}

.table {
  width: 100%;
  border-collapse: collapse;
}

.table th,
.table td {
  text-align: left;
  padding: 0.85rem 0.4rem;
  font-size: 0.84rem;
  border-bottom: 1px solid rgba(86, 70, 55, 0.08);
  vertical-align: middle;
}

.table th {
  background: transparent;
  font-weight: 700;
  color: #76695d;
  font-size: 0.7rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  border-bottom: 1px solid rgba(86, 70, 55, 0.18);
}

.table tbody tr:last-child td {
  border-bottom: 0;
}

.titleCell {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
}

.empty {
  padding: 1.5rem;
  text-align: center;
  color: #76695d;
}

.mono {
  font-family: var(--font-mono);
  font-feature-settings: 'tnum' on;
  letter-spacing: -0.01em;
  color: #4a3f33;
  font-size: 0.78rem;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.action {
  min-height: 1.75rem;
  padding: 0 0.65rem;
  border: 1px solid #24201c;
  border-radius: 0.5rem;
  background: #24201c;
  color: #fffdf9;
  font: inherit;
  font-size: 0.76rem;
  cursor: pointer;
}

.actionGhost {
  min-height: 1.75rem;
  padding: 0 0.65rem;
  border: 1px solid rgba(86, 70, 55, 0.25);
  border-radius: 0.5rem;
  background: #fffdf9;
  color: #24201c;
  font: inherit;
  font-size: 0.76rem;
  cursor: pointer;
}

.action:active,
.actionGhost:active {
  transform: translateY(1px);
  transition: transform 80ms ease;
}
```

- [ ] **Step 3: 编译校验**

Run: `cd web && pnpm exec tsc -b`
Expected: 无错误。

- [ ] **Step 4: Commit**

```bash
git add web/src/view/service/TaskListPage.tsx web/src/view/service/TaskListPage.module.css
git commit -m "feat(ops-tasks): add TaskListPage with claim/complete/release actions"
```

---

## Task 13: 路由 + 启动桥 + 启用导航

**Files:**
- Modify: `web/src/router/index.tsx`
- Modify: `web/src/layout/AppShell.tsx`
- Modify: `web/src/components/nav/SideNav.tsx`

- [ ] **Step 1: 注册路由**

Modify `web/src/router/index.tsx`：
1. 在顶部 import 区（与其他 view import 同组）加入：
```ts
import { TaskListPage } from '../view/service/TaskListPage';
```
2. 在 `service-portal/tickets/:id` 路由对象**之后**，新增：
```tsx
              {
                path: 'service-portal/tasks',
                element: <TaskListPage />,
              },
```

- [ ] **Step 2: AppShell 启动协调桥**

Modify `web/src/layout/AppShell.tsx`：
1. 顶部 import 加入：
```ts
import { useEffect } from 'react';
import { startTaskBridge } from '../view/service/data/taskStore';
```
2. 在 `AppShell` 组件体内（`const { t } = useI18n();` 之后）加入：
```ts
  useEffect(() => {
    const stop = startTaskBridge();
    return stop;
  }, []);
```

- [ ] **Step 3: 启用 SideNav「任务管理」NavLink**

Modify `web/src/components/nav/SideNav.tsx`：把 `tasks` 的 disabled span
```tsx
          <span className={styles.navEntryDisabled} title={t('shell.comingSoon')}>
            {t('shell.tasks')}
          </span>
```
替换为 NavLink（与同组其它 NavLink 同样式）：
```tsx
          <NavLink
            to="/service-portal/tasks"
            className={({ isActive }) =>
              [styles.navEntry, isActive ? styles.navEntryActive : '']
                .filter(Boolean)
                .join(' ')
            }
          >
            {t('shell.tasks')}
          </NavLink>
```
（保留 `myDuty`、`bigScreen` 两个 disabled span 不动。）

- [ ] **Step 4: 编译 + 全量测试**

Run: `cd web && pnpm exec tsc -b && pnpm test`
Expected: 编译无错误，测试全绿。

- [ ] **Step 5: 手工验证（运行 app）**

Run: `cd web && pnpm dev`，浏览器登录后：
1. 左侧「任务管理」可点击进入 `/service-portal/tasks`。
2. 在服务门户提交一个申请，约 30s 后工单进入「处理中」，任务页出现一条「待领取」任务。
3. 点「领取」→ 变「进行中」并显示当前用户为领取人。
4. 点「完成」→ 任务变「已完成」，进入对应工单详情可见状态已为「已完成」。
5. 「只看我的」「状态筛选」「搜索」均生效。

Expected: 上述行为全部符合。

- [ ] **Step 6: Commit**

```bash
git add web/src/router/index.tsx web/src/layout/AppShell.tsx web/src/components/nav/SideNav.tsx
git commit -m "feat(ops-tasks): wire tasks route, bridge startup and nav entry"
```

---

## Task 14: TicketDetailPage 关联任务展示（双向协同）

**Files:**
- Modify: `web/src/view/service/TicketDetailPage.tsx`

- [ ] **Step 1: 引入 task store 与徽章**

Modify `web/src/view/service/TicketDetailPage.tsx` 顶部 import：
```ts
import { TaskStatusBadge } from './components/TaskStatusBadge';
import { useTaskStore } from './data/taskStore';
```

- [ ] **Step 2: 取关联任务**

在组件体内 `const ticket = getTicket(id);` 之后加入：
```ts
  const { tasks } = useTaskStore();
  const linkedTask = tasks.find((task) => task.ticketId === id) ?? null;
```

- [ ] **Step 3: 在 summary 区展示关联任务**

在 `<section className={styles.summary}>` 内、申请人 `field` 之后，追加一个 field：
```tsx
        <div className={styles.field}>
          <span className={styles.fieldLabel}>{t('servicePortal.task.ticketLink.label')}</span>
          <span className={styles.fieldValue}>
            {linkedTask ? (
              <>
                <TaskStatusBadge status={linkedTask.status} />
                {linkedTask.assignee ? ` · ${linkedTask.assignee.name}` : ''}
              </>
            ) : (
              t('servicePortal.task.ticketLink.none')
            )}
          </span>
        </div>
```

- [ ] **Step 4: 编译 + 测试**

Run: `cd web && pnpm exec tsc -b && pnpm test`
Expected: 编译无错误，测试全绿。

- [ ] **Step 5: 手工验证**

`pnpm dev`：打开一个处于「处理中」的工单详情，「关联任务」显示任务状态徽章与领取人；任务完成后该工单详情状态徽章变「已完成」。

- [ ] **Step 6: Commit**

```bash
git add web/src/view/service/TicketDetailPage.tsx
git commit -m "feat(ops-tasks): show linked task status on ticket detail"
```

---

## Self-Review

**Spec coverage:**
- 概念/生命周期（pending→in_progress→done + release）→ Task 3/4/5/6。
- 任务驱动工单（complete→ticket done）→ Task 5；mock 引擎去 60s → Task 9。
- 架构 A（独立 taskStore + 桥）→ Task 3–8；桥 Task 8/13。
- 数据模型 OpsTask → Task 2（含 timeline，由 claim/complete/release 写入并被测试断言，非死字段）。
- 列表页（筛选/搜索/只看我的/行内操作）→ Task 12。
- TaskStatusBadge → Task 11；路由/导航 → Task 13；TicketDetail 协同 → Task 14。
- i18n（en+zh）→ Task 10。
- 测试策略（vitest + store/桥/引擎覆盖）→ Task 1 + 各 store 任务。
- 错误处理（持久化降级、领取竞态幂等、reconcile 幂等）→ Task 3/4/7/8。
- 范围外（值班/大屏/优先级/SLA/后端）→ 未触碰。✔ 全覆盖。

**Placeholder scan:** 无 TBD/TODO；每个代码步骤含完整代码与确切命令、预期输出。

**Type consistency:** `TaskStatus = 'pending'|'in_progress'|'done'`、`OpsTask`、`taskStore.{list,get,subscribe,claim,complete,release,reconcileFromTickets,__resetForTest}`、`startTaskBridge`、`useTaskStore` 在各任务间签名一致；i18n key 命名 `servicePortal.task.*` 与页面/组件引用一致；`claim(id, user)` 的 `user` 形状 `{name,email}` 与 `TaskListPage` 传入 `me` 一致。
