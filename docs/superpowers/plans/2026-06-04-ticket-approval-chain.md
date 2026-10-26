# 工单审批链展示 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让工单展示其多级审批链 —— 中间审批人是谁、整体走到哪一步，纯前端 mock 自动逐级推进。

**Architecture:** 扩展 `ticketTypes` 增加按服务类别快照到工单的 `approvalChain`；`ticketStore.createTicket` 创建时快照、新增 `advanceApproval` 逐级通过并在最后一级驱动 `approving→processing`；`ticketMockEngine` 改为按链长度每 ~10s 通过一级。展示层加 `ApprovalStepper`（详情页顶部横向步骤条）、增强 `TicketTimeline`（织入审批人名）、列表页加审批进度小字。派生选择器算"当前审批步/进度"，不存冗余。

**Tech Stack:** React 18 + TypeScript + Vite + CSS Modules + framer-motion；vitest + jsdom；store 沿用 `useSyncExternalStore` + localStorage 模式。

**环境备注（执行者必读）：**
- pnpm 不在 PATH，用绝对路径：`~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm`
- 测试：`cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm test`
- 类型：`cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm exec tsc -b`
- 代码文件的 Read 工具被 `cbm-code-discovery-gate` hook 拦截；读已提交代码用 `git show HEAD:<path>`，或用 codebase-memory-mcp 工具（search_graph / get_code_snippet）。
- 分支：在 `feat/ops-center-tasks` 上继续，不新建分支。

---

## Task 1: ApprovalStep 类型定义

**Files:**
- Modify: `web/src/view/service/ticketTypes.ts`

- [ ] **Step 1: 加类型**

在 `ticketTypes.ts` 中 `TicketStatus` 定义之后、`TicketEvent` 之前插入：
```ts
export type ApprovalStepStatus = 'pending' | 'approved';

export interface ApprovalStep {
  roleI18nKey: string; // 角色名 i18n leaf key
  approver: { name: string; email: string };
  status: ApprovalStepStatus;
  decidedAt?: string; // ISO，通过时写入
}
```
把现有 `TicketEvent` 接口改为（新增两个可选字段）：
```ts
export interface TicketEvent {
  at: string; // ISO timestamp
  status: TicketStatus;
  i18nNoteKey?: string;
  actor?: string; // 审批人 displayName
  roleI18nKey?: string; // 审批角色
}
```

> 注意：`Ticket.approvalChain` 字段**不在本任务加** —— 它在 Task 3 与 `createTicket` 的快照逻辑一起加，保证每次提交都能编译通过。本任务只加 `ApprovalStep` 类型与 `TicketEvent` 的两个可选字段（都是纯增量，不破坏现有构造）。

- [ ] **Step 2: 类型校验**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm exec tsc -b`
Expected: exit 0（新增类型为纯增量，未被强制使用）。

- [ ] **Step 3: Commit**

```bash
git add web/src/view/service/ticketTypes.ts
git commit -m "feat(ticket-approval): add ApprovalStep types and Ticket.approvalChain field"
```

---

## Task 2: approvalChains —— mock 审批链与审批人

**Files:**
- Create: `web/src/view/service/data/approvalChains.ts`
- Test: `web/src/view/service/data/approvalChains.test.ts`

- [ ] **Step 1: 写失败测试**

Create `web/src/view/service/data/approvalChains.test.ts`:
```ts
import { describe, expect, it } from 'vitest';

import { buildApprovalChain } from './approvalChains';

describe('buildApprovalChain', () => {
  it('account 类服务返回 1 级链（直属主管），全部 pending', () => {
    const chain = buildApprovalChain('account.role-grant');
    expect(chain).toHaveLength(1);
    expect(chain[0].roleI18nKey).toBe('servicePortal.ticket.approval.role.lineManager');
    expect(chain[0].status).toBe('pending');
    expect(chain[0].approver.name).toBeTruthy();
    expect(chain[0].decidedAt).toBeUndefined();
  });

  it('cloud 类服务返回 3 级链', () => {
    const chain = buildApprovalChain('cloud.compute-resize');
    expect(chain).toHaveLength(3);
    expect(chain.every((s) => s.status === 'pending')).toBe(true);
  });

  it('network 类服务返回 2 级链', () => {
    expect(buildApprovalChain('network.acl')).toHaveLength(2);
  });

  it('null 或未知服务返回默认 1 级链', () => {
    expect(buildApprovalChain(null)).toHaveLength(1);
    expect(buildApprovalChain('does.not.exist')).toHaveLength(1);
  });

  it('返回独立副本，改一处不影响下次', () => {
    const a = buildApprovalChain('account.role-grant');
    a[0].status = 'approved';
    const b = buildApprovalChain('account.role-grant');
    expect(b[0].status).toBe('pending');
  });
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm test`
Expected: FAIL —— `buildApprovalChain` 未定义。

- [ ] **Step 3: 实现**

Create `web/src/view/service/data/approvalChains.ts`:
```ts
// web/src/view/service/data/approvalChains.ts
import type { ApprovalStep, ServiceCategory } from '../ticketTypes';
import { getCatalogById } from './serviceCatalog';

interface ApproverDef {
  roleI18nKey: string;
  approver: { name: string; email: string };
}

const APPROVERS: Record<'lineManager' | 'security' | 'deptHead', ApproverDef> = {
  lineManager: {
    roleI18nKey: 'servicePortal.ticket.approval.role.lineManager',
    approver: { name: '王明', email: 'wang.ming@optimus.io' },
  },
  security: {
    roleI18nKey: 'servicePortal.ticket.approval.role.security',
    approver: { name: '李强', email: 'li.qiang@optimus.io' },
  },
  deptHead: {
    roleI18nKey: 'servicePortal.ticket.approval.role.deptHead',
    approver: { name: '张伟', email: 'zhang.wei@optimus.io' },
  },
};

type RoleKey = keyof typeof APPROVERS;

const CHAIN_TEMPLATES: Record<ServiceCategory, RoleKey[]> = {
  account: ['lineManager'],
  cloud: ['lineManager', 'security', 'deptHead'],
  network: ['lineManager', 'security'],
  env: ['lineManager', 'deptHead'],
  support: ['lineManager', 'security', 'deptHead'],
};

const DEFAULT_CHAIN: RoleKey[] = ['lineManager'];

export function buildApprovalChain(serviceId: string | null): ApprovalStep[] {
  const category = serviceId ? getCatalogById(serviceId)?.category : undefined;
  const roleKeys = (category && CHAIN_TEMPLATES[category]) || DEFAULT_CHAIN;
  return roleKeys.map((key) => {
    const def = APPROVERS[key];
    return {
      roleI18nKey: def.roleI18nKey,
      approver: { ...def.approver },
      status: 'pending' as const,
    };
  });
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm test`
Expected: PASS（5 个 approvalChains 测试）。

- [ ] **Step 5: Commit**

```bash
git add web/src/view/service/data/approvalChains.ts web/src/view/service/data/approvalChains.test.ts
git commit -m "feat(ticket-approval): add per-category approval chain mock data"
```

---

## Task 3: createTicket 快照审批链

**Files:**
- Modify: `web/src/view/service/data/ticketStore.ts`
- Test: `web/src/view/service/data/ticketStore.test.ts`（新建）

- [ ] **Step 1: 写失败测试**

Create `web/src/view/service/data/ticketStore.test.ts`:
```ts
import { beforeEach, describe, expect, it } from 'vitest';

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
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm test`
Expected: FAIL —— `approvalChain` 为 `undefined`。

- [ ] **Step 3: 加 Ticket.approvalChain 字段**

先在 `web/src/view/service/ticketTypes.ts` 的 `Ticket` 接口中，`applicant` 字段之后新增（此时本任务会同步更新唯一构造它的 `createTicket`，故编译保持通过）：
```ts
  approvalChain: ApprovalStep[];
```

- [ ] **Step 4: 实现 createTicket 快照**

在 `ticketStore.ts` 顶部 import 加入：
```ts
import { buildApprovalChain } from './approvalChains';
```
在 `createTicket` 内构造 `ticket` 对象时，`applicant: input.applicant,` 之后新增一行：
```ts
      approvalChain: buildApprovalChain(input.serviceId),
```

- [ ] **Step 5: 运行测试 + 类型校验确认通过**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm test && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm exec tsc -b`
Expected: 测试 PASS；`tsc -b` exit 0（`Ticket.approvalChain` 加入的同时 `createTicket` 即填充，无缺字段错误）。

- [ ] **Step 6: Commit**

```bash
git add web/src/view/service/ticketTypes.ts web/src/view/service/data/ticketStore.ts web/src/view/service/data/ticketStore.test.ts
git commit -m "feat(ticket-approval): snapshot approval chain on ticket creation"
```

---

## Task 4: ticketStore.advanceApproval —— 逐级通过并驱动工单

**Files:**
- Modify: `web/src/view/service/data/ticketStore.ts`
- Test: `web/src/view/service/data/ticketStore.test.ts`

- [ ] **Step 1: 写失败测试**

在 `ticketStore.test.ts` 末尾追加：
```ts
describe('ticketStore.advanceApproval', () => {
  function createApproving(serviceId: string | null) {
    const ticket = ticketStore.createTicket({
      serviceId,
      title: '审批测试',
      formValues: {},
      applicant: { name: '申请人', email: 'a@x.io' },
    }).ticket;
    ticketStore.advance(ticket.id, 'submitted', 'approving');
    return ticket.id;
  }

  it('逐级把 pending 标为 approved 并写 decidedAt + 事件', () => {
    const id = createApproving('network.acl'); // 2 级
    ticketStore.advanceApproval(id);
    const t1 = ticketStore.get(id)!;
    expect(t1.approvalChain[0].status).toBe('approved');
    expect(t1.approvalChain[0].decidedAt).toBeTruthy();
    expect(t1.status).toBe('approving'); // 还有第 2 级
    const ev = t1.timeline[t1.timeline.length - 1];
    expect(ev.i18nNoteKey).toBe('servicePortal.ticket.event.approved');
    expect(ev.actor).toBe(t1.approvalChain[0].approver.name);
    expect(ev.roleI18nKey).toBe(t1.approvalChain[0].roleI18nKey);
  });

  it('最后一级通过后把工单推进到 processing', () => {
    const id = createApproving('network.acl'); // 2 级
    ticketStore.advanceApproval(id);
    ticketStore.advanceApproval(id);
    const t2 = ticketStore.get(id)!;
    expect(t2.approvalChain.every((s) => s.status === 'approved')).toBe(true);
    expect(t2.status).toBe('processing');
  });

  it('非 approving 状态 no-op', () => {
    const ticket = ticketStore.createTicket({
      serviceId: 'account.role-grant',
      title: 'x',
      formValues: {},
      applicant: { name: 'a', email: 'a@x.io' },
    }).ticket; // 仍 submitted
    ticketStore.advanceApproval(ticket.id);
    expect(ticketStore.get(ticket.id)!.status).toBe('submitted');
  });
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm test`
Expected: FAIL —— `advanceApproval is not a function`。

- [ ] **Step 3: 实现**

在 `ticketStore.ts` 的 `ticketStore` 对象中，`advance(...)` 方法之后加入：
```ts
  advanceApproval(ticketId: string): void {
    const idx = tickets.findIndex((t) => t.id === ticketId);
    if (idx === -1) return;
    const current = tickets[idx];
    if (current.status !== 'approving') return;
    const chain = current.approvalChain ?? [];
    const pendingIdx = chain.findIndex((s) => s.status === 'pending');
    const at = new Date().toISOString();

    let updated: Ticket;
    if (pendingIdx === -1) {
      // 空链 / 已全通过：直接进 processing
      updated = {
        ...current,
        status: 'processing',
        timeline: [
          ...current.timeline,
          { at, status: 'processing', i18nNoteKey: 'servicePortal.ticket.event.autoAdvanced' },
        ],
      };
    } else {
      const step = chain[pendingIdx];
      const nextChain = [
        ...chain.slice(0, pendingIdx),
        { ...step, status: 'approved' as const, decidedAt: at },
        ...chain.slice(pendingIdx + 1),
      ];
      const stillPending = nextChain.some((s) => s.status === 'pending');
      const approvedEvent = {
        at,
        status: 'approving' as const,
        i18nNoteKey: 'servicePortal.ticket.event.approved',
        actor: step.approver.name,
        roleI18nKey: step.roleI18nKey,
      };
      const tail = stillPending
        ? []
        : [{ at, status: 'processing' as const, i18nNoteKey: 'servicePortal.ticket.event.autoAdvanced' }];
      updated = {
        ...current,
        approvalChain: nextChain,
        status: stillPending ? 'approving' : 'processing',
        timeline: [...current.timeline, approvedEvent, ...tail],
      };
    }

    const next = [...tickets.slice(0, idx), updated, ...tickets.slice(idx + 1)];
    persist(next);
    tickets = next;
    notify();
  },
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm test`
Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add web/src/view/service/data/ticketStore.ts web/src/view/service/data/ticketStore.test.ts
git commit -m "feat(ticket-approval): advanceApproval steps chain and drives ticket to processing"
```

---

## Task 5: ticketMockEngine 逐级推进

**Files:**
- Modify: `web/src/view/service/data/ticketMockEngine.ts`
- Test: `web/src/view/service/data/ticketMockEngine.test.ts`

- [ ] **Step 1: 写失败测试**

在 `ticketMockEngine.test.ts` 末尾追加（沿用文件已有的 `beforeEach`/fake timers 设置）：
```ts
describe('ticketMockEngine 审批链逐级推进', () => {
  it('按链长度逐级通过后进入 processing，不自动 done', () => {
    const created = ticketStore.createTicket({
      serviceId: 'cloud.compute-resize', // 3 级
      formValues: {},
      title: '逐级验证',
      applicant: { name: '申请人', email: 'a@x.io' },
    });
    const id = created.ticket.id;
    ticketMockEngine.schedule(id);

    vi.advanceTimersByTime(5_000);
    expect(ticketStore.get(id)!.status).toBe('approving');

    vi.advanceTimersByTime(10_000); // 第 1 级
    expect(ticketStore.get(id)!.approvalChain[0].status).toBe('approved');
    expect(ticketStore.get(id)!.status).toBe('approving');

    vi.advanceTimersByTime(10_000); // 第 2 级
    vi.advanceTimersByTime(10_000); // 第 3 级 → processing
    expect(ticketStore.get(id)!.status).toBe('processing');

    vi.advanceTimersByTime(120_000);
    expect(ticketStore.get(id)!.status).toBe('processing'); // 不自动 done
  });
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm test`
Expected: FAIL —— 旧引擎在 30s 一步把它推到 processing，第 1 级断言（10s 后只过 1 级、仍 approving）不成立。

- [ ] **Step 3: 实现**

把 `ticketMockEngine.ts` 整体替换为：
```ts
// web/src/view/service/data/ticketMockEngine.ts

import { ticketStore } from './ticketStore';

const STEP_5_MS = 5_000;
const STEP_APPROVAL_MS = 10_000;

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
    const chainLen = ticketStore.get(ticketId)?.approvalChain?.length ?? 0;
    const steps = Math.max(chainLen, 1);
    for (let i = 1; i <= steps; i += 1) {
      ids.push(
        window.setTimeout(() => {
          ticketStore.advanceApproval(ticketId);
        }, STEP_5_MS + i * STEP_APPROVAL_MS),
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
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm test`
Expected: PASS（含原有「推进到 processing 后不再自动 done」用例 —— 它用默认 1 级链，15s 进 processing，30s 断言仍成立）。

- [ ] **Step 5: Commit**

```bash
git add web/src/view/service/data/ticketMockEngine.ts web/src/view/service/data/ticketMockEngine.test.ts
git commit -m "feat(ticket-approval): mock engine advances approval chain level by level"
```

---

## Task 6: 审批进度派生选择器

**Files:**
- Create: `web/src/view/service/data/approvalSelectors.ts`
- Test: `web/src/view/service/data/approvalSelectors.test.ts`

- [ ] **Step 1: 写失败测试**

Create `web/src/view/service/data/approvalSelectors.test.ts`:
```ts
import { describe, expect, it } from 'vitest';

import { approvalProgress, currentApprovalStep } from './approvalSelectors';
import type { Ticket } from '../ticketTypes';

function ticketWith(chain: Ticket['approvalChain']): Ticket {
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

const step = (status: 'pending' | 'approved') => ({
  roleI18nKey: 'servicePortal.ticket.approval.role.lineManager',
  approver: { name: '王明', email: 'w@x.io' },
  status,
});

describe('approvalSelectors', () => {
  it('currentApprovalStep 返回第一个 pending', () => {
    const t = ticketWith([step('approved'), step('pending')]);
    expect(currentApprovalStep(t)?.status).toBe('pending');
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

  it('缺失链兜底为 0/0 与 null', () => {
    const t = ticketWith([]);
    expect(approvalProgress(t)).toEqual({ approved: 0, total: 0 });
    expect(currentApprovalStep(t)).toBeNull();
  });
});
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm test`
Expected: FAIL —— 选择器未定义。

- [ ] **Step 3: 实现**

Create `web/src/view/service/data/approvalSelectors.ts`:
```ts
// web/src/view/service/data/approvalSelectors.ts
import type { ApprovalStep, Ticket } from '../ticketTypes';

export function currentApprovalStep(ticket: Ticket): ApprovalStep | null {
  return (ticket.approvalChain ?? []).find((s) => s.status === 'pending') ?? null;
}

export function approvalProgress(ticket: Ticket): { approved: number; total: number } {
  const chain = ticket.approvalChain ?? [];
  return {
    approved: chain.filter((s) => s.status === 'approved').length,
    total: chain.length,
  };
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm test`
Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add web/src/view/service/data/approvalSelectors.ts web/src/view/service/data/approvalSelectors.test.ts
git commit -m "feat(ticket-approval): add approval progress selectors"
```

---

## Task 7: i18n —— servicePortal.ticket.approval.*（en + zhCN）

**Files:**
- Modify: `web/src/locale/messages.ts`

- [ ] **Step 1: 加 en 文案**

在 **en** 块 `servicePortal.ticket` 对象内、`event: { ... },` 块**之后**（与 `event` 同级，缩进 6 空格）插入：
```ts
      approval: {
        role: {
          lineManager: 'Line manager',
          security: 'Security review',
          deptHead: 'Department head',
        },
        event: {
          approved: '{actor} ({role}) approved',
        },
        stepperApplicant: 'You',
        stepperCurrent: 'Reviewing',
        stepperWaiting: 'Pending',
        stepperRejected: 'Rejected',
        listProgress: 'Approval {count}/{total} · {name}',
      },
```

- [ ] **Step 2: 加 zhCN 文案**

在 **zhCN** 块 `servicePortal.ticket` 对象内、对应 `event: { ... },` 之后插入：
```ts
      approval: {
        role: {
          lineManager: '直属主管',
          security: '安全审批',
          deptHead: '部门负责人',
        },
        event: {
          approved: '{actor}（{role}）已通过',
        },
        stepperApplicant: '本人',
        stepperCurrent: '审批中',
        stepperWaiting: '待审批',
        stepperRejected: '已驳回',
        listProgress: '审批 {count}/{total} · 当前:{name}',
      },
```

- [ ] **Step 3: 类型/结构校验**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm exec tsc -b`
Expected: exit 0（en 与 zhCN `approval` 子树键一致；`zhCN: typeof en` 不报缺键）。

- [ ] **Step 4: Commit**

```bash
git add web/src/locale/messages.ts
git commit -m "feat(ticket-approval): add servicePortal.ticket.approval i18n keys (en + zh)"
```

---

## Task 8: ApprovalStepper 组件

**Files:**
- Create: `web/src/view/service/components/ApprovalStepper.tsx`
- Create: `web/src/view/service/components/ApprovalStepper.module.css`

- [ ] **Step 1: 写组件**

Create `web/src/view/service/components/ApprovalStepper.tsx`:
```tsx
// ApprovalStepper.tsx
import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { Ticket } from '../ticketTypes';
import styles from './ApprovalStepper.module.css';

type NodeState = 'done' | 'current' | 'upcoming' | 'rejected';

interface StepNode {
  key: string;
  label: string;
  sub?: string;
  state: NodeState;
}

export function ApprovalStepper({ ticket }: { ticket: Ticket }) {
  const { t } = useI18n();
  const chain = ticket.approvalChain ?? [];
  const status = ticket.status;
  const firstPendingIdx = chain.findIndex((s) => s.status === 'pending');

  const nodes: StepNode[] = [];

  nodes.push({
    key: 'submitted',
    label: t('servicePortal.ticket.status.submitted'),
    sub: t('servicePortal.ticket.approval.stepperApplicant'),
    state: 'done',
  });

  chain.forEach((step, i) => {
    const roleLabel = t(step.roleI18nKey as MessageKey);
    let state: NodeState;
    let sub = step.approver.name;
    if (step.status === 'approved' || status === 'processing' || status === 'done') {
      state = 'done';
    } else if (status === 'rejected' && i === firstPendingIdx) {
      state = 'rejected';
      sub = `${step.approver.name} · ${t('servicePortal.ticket.approval.stepperRejected')}`;
    } else if (status === 'approving' && i === firstPendingIdx) {
      state = 'current';
      sub = `${step.approver.name} · ${t('servicePortal.ticket.approval.stepperCurrent')}`;
    } else {
      state = 'upcoming';
      sub = `${step.approver.name} · ${t('servicePortal.ticket.approval.stepperWaiting')}`;
    }
    nodes.push({ key: `appr-${i}`, label: roleLabel, sub, state });
  });

  nodes.push({
    key: 'processing',
    label: t('servicePortal.ticket.status.processing'),
    state: status === 'done' ? 'done' : status === 'processing' ? 'current' : 'upcoming',
  });

  nodes.push({
    key: 'done',
    label: t('servicePortal.ticket.status.done'),
    state: status === 'done' ? 'current' : 'upcoming',
  });

  return (
    <div className={styles.stepper} role="list">
      {nodes.map((node, i) => (
        <div key={node.key} className={styles.node} role="listitem" data-state={node.state}>
          {i > 0 ? <span className={styles.bar} data-state={node.state} aria-hidden="true" /> : null}
          <span className={styles.dot} data-state={node.state}>
            {node.state === 'done' ? '✓' : ''}
          </span>
          <span className={styles.label}>{node.label}</span>
          {node.sub ? (
            <span className={styles.sub} data-state={node.state}>
              {node.sub}
            </span>
          ) : null}
        </div>
      ))}
    </div>
  );
}
```

- [ ] **Step 2: 写样式**

Create `web/src/view/service/components/ApprovalStepper.module.css`:
```css
/* ApprovalStepper.module.css */
.stepper {
  display: flex;
  align-items: flex-start;
  gap: 0;
  padding: 0.4rem 0 0.2rem;
  overflow-x: auto;
}

.node {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1 0 auto;
  min-width: 5.5rem;
  position: relative;
  text-align: center;
}

.bar {
  position: absolute;
  top: 0.55rem;
  left: -50%;
  width: 100%;
  height: 2px;
  background: #e6dcc8;
  z-index: 0;
}

.bar[data-state='done'] {
  background: #2f625c;
}

.dot {
  position: relative;
  z-index: 1;
  width: 1.15rem;
  height: 1.15rem;
  border-radius: 50%;
  border: 2px solid #cdbfa8;
  background: #fffdf9;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.6rem;
  color: #fff;
}

.dot[data-state='done'] {
  background: #2f625c;
  border-color: #2f625c;
}

.dot[data-state='current'] {
  border-color: #365d7e;
  box-shadow: 0 0 0 3px rgba(54, 93, 126, 0.18);
}

.dot[data-state='current']::after {
  content: '';
  width: 0.45rem;
  height: 0.45rem;
  border-radius: 50%;
  background: #365d7e;
}

.dot[data-state='rejected'] {
  border-color: #a23838;
  background: #a23838;
}

.label {
  margin-top: 0.4rem;
  font-size: 0.72rem;
  color: #4a3f33;
  font-weight: 600;
}

.sub {
  font-size: 0.68rem;
  color: #8a7d6c;
  margin-top: 0.1rem;
}

.sub[data-state='current'] {
  color: #365d7e;
  font-weight: 700;
}

.sub[data-state='rejected'] {
  color: #a23838;
  font-weight: 700;
}
```

- [ ] **Step 3: 编译校验**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm exec tsc -b`
Expected: exit 0。

- [ ] **Step 4: Commit**

```bash
git add web/src/view/service/components/ApprovalStepper.tsx web/src/view/service/components/ApprovalStepper.module.css
git commit -m "feat(ticket-approval): add ApprovalStepper component"
```

---

## Task 9: TicketTimeline 织入审批人名

**Files:**
- Modify: `web/src/view/service/components/TicketTimeline.tsx`

- [ ] **Step 1: 改文案插值**

在 `TicketTimeline.tsx` 的 `.map` 回调里，把现有的 `note` 计算：
```ts
        const note = event.i18nNoteKey
          ? t(event.i18nNoteKey as MessageKey)
              .replace('{applicant}', ticket.applicant.name || ticket.applicant.email)
              .replace('{status}', statusLabel)
          : statusLabel;
```
替换为（追加 `{actor}` 与 `{role}` 两个替换）：
```ts
        const note = event.i18nNoteKey
          ? t(event.i18nNoteKey as MessageKey)
              .replace('{applicant}', ticket.applicant.name || ticket.applicant.email)
              .replace('{status}', statusLabel)
              .replace('{actor}', event.actor ?? '')
              .replace('{role}', event.roleI18nKey ? t(event.roleI18nKey as MessageKey) : '')
          : statusLabel;
```

- [ ] **Step 2: 编译校验**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm exec tsc -b`
Expected: exit 0。

- [ ] **Step 3: 运行测试确认无回归**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm test`
Expected: PASS（全部既有测试）。

- [ ] **Step 4: Commit**

```bash
git add web/src/view/service/components/TicketTimeline.tsx
git commit -m "feat(ticket-approval): weave approver name into ticket timeline events"
```

---

## Task 10: TicketDetailPage 插入步骤条

**Files:**
- Modify: `web/src/view/service/TicketDetailPage.tsx`

- [ ] **Step 1: import**

在 `TicketDetailPage.tsx` 顶部，`import { TicketTimeline } from './components/TicketTimeline';` 之后加入：
```ts
import { ApprovalStepper } from './components/ApprovalStepper';
```

- [ ] **Step 2: 渲染步骤条**

在 `return (` 之后、`<section className={styles.summary}>` 之前插入：
```tsx
      <ApprovalStepper ticket={ticket} />
```

- [ ] **Step 3: 编译校验**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm exec tsc -b`
Expected: exit 0。

- [ ] **Step 4: Commit**

```bash
git add web/src/view/service/TicketDetailPage.tsx
git commit -m "feat(ticket-approval): show ApprovalStepper on ticket detail page"
```

---

## Task 11: TicketListPage 审批进度小字

**Files:**
- Modify: `web/src/view/service/TicketListPage.tsx`

- [ ] **Step 1: import 选择器**

在 `TicketListPage.tsx` 顶部，`import { TicketStatusBadge } from './components/TicketStatusBadge';` 之后加入：
```ts
import { approvalProgress, currentApprovalStep } from './data/approvalSelectors';
```

- [ ] **Step 2: 渲染小字**

找到状态徽章单元格：
```tsx
                <td>
                  <TicketStatusBadge status={ticket.status} />
                </td>
```
替换为（徽章下方在 approving 时加一行派生小字）：
```tsx
                <td>
                  <TicketStatusBadge status={ticket.status} />
                  {ticket.status === 'approving' && currentApprovalStep(ticket) ? (
                    <div className={styles.approvalSub}>
                      {t('servicePortal.ticket.approval.listProgress')
                        .replace('{count}', String(approvalProgress(ticket).approved))
                        .replace('{total}', String(approvalProgress(ticket).total))
                        .replace('{name}', currentApprovalStep(ticket)!.approver.name)}
                    </div>
                  ) : null}
                </td>
```

- [ ] **Step 3: 加样式**

在 `web/src/view/service/TicketListPage.module.css` 末尾追加：
```css
.approvalSub {
  margin-top: 0.25rem;
  font-size: 0.68rem;
  color: #8a7d6c;
}
```

- [ ] **Step 4: 编译 + 测试**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm exec tsc -b && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm test`
Expected: exit 0；测试全绿。

- [ ] **Step 5: 手工验证（运行 app）**

Run: `cd web && ~/.npm/_npx/a1a38f5f0f780954/node_modules/.bin/pnpm dev`，登录后：
1. 在服务门户提交一个「云资源」类申请（3 级链）。
2. 约 5s 进入「审批中」；之后每 ~10s 过一级，列表行小字从 `审批 0/3 · 当前:王明` → `1/3 · 李强` → `2/3 · 张伟`。
3. 打开该工单详情：顶部步骤条逐级点亮，当前节点高亮显示当前审批人；时间线出现「王明（直属主管）已通过」等事件。
4. 3 级走完后工单进入「处理中」（任务页随之出现待领取任务，沿用既有任务模块联动）。

Expected: 行为符合。

- [ ] **Step 6: Commit**

```bash
git add web/src/view/service/TicketListPage.tsx web/src/view/service/TicketListPage.module.css
git commit -m "feat(ticket-approval): show approval progress on ticket list rows"
```

---

## Self-Review 对照

- 数据模型（ApprovalStep / Ticket.approvalChain / TicketEvent.actor+role）→ Task 1。
- mock 审批人目录 + 按类别链 + 默认链 → Task 2。
- 创建快照 → Task 3；逐级通过 + 驱动 processing + 兜底空链 → Task 4。
- mock 引擎逐级（删 30s 单步、每 ~10s 一级）→ Task 5。
- 派生「当前审批人 / 进度」→ Task 6（被 Task 8/11 复用）。
- i18n（en+zh，角色/事件/步骤条/列表小字）→ Task 7。
- 详情页步骤条（方案 A）→ Task 8 + Task 10；时间线织入审批人（方案 C）→ Task 9。
- 列表页 approving 小字 → Task 11。
- 错误处理：持久化降级沿用既有；旧数据 `approvalChain ?? []` 兜底贯穿 Task 4/6/8；空链不卡死 → Task 4。
- 测试策略：approvalChains（T2）/ createTicket 快照（T3）/ advanceApproval（T4）/ 引擎 fake timers（T5）/ 选择器（T6）。
- 范围外（rejected 交互、后端、SLA、明细面板 B）未纳入任务，符合 spec。
