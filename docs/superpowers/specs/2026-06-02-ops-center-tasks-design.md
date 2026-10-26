# 运维中心 · 任务管理（Ops Tasks）设计

- 日期：2026-06-02
- 范围：运维中心前端「任务管理」模块（前端优先，无后端）
- 状态：原文记录已通过设计评审；[实施计划已存在](../plans/2026-06-02-ops-center-tasks.md)，有实现提交，交付验收待核实。
- 最近核查：2026-10-26；证据与后续验证见 [能力状态](../../project-status.md) 和 [#2](https://github.com/Dcuhddrg/Optimus/issues/2)。

> 演进关系：本文取代早期 [服务门户设计](2026-05-28-service-portal-design.md) 的 processing 定时自动 done 行为。以下“背景”保留设计时事实；当前引擎已删除自动 done 定时器，任务完成调用工单推进（本轮仅源码抽查）。后续 [审批链设计](2026-06-04-ticket-approval-chain-design.md) 规划改变审批节奏，但尚未完整落地。

## 背景

运维中心（`SideNav` 的 `opsCenter` 分组）规划了五个入口，已建成 `服务门户`、`工单管理`，其余三个为 disabled 占位：`任务管理`、`我的值班`、`大屏展示`。三者是相互独立的子系统，按子项目逐个推进，本 spec 只覆盖**任务管理**。

现状关键约束：

- 整个服务门户/工单是**纯前端 mock**，无后端路由。数据靠客户端 store + `localStorage` 持久化：`web/src/view/service/data/ticketStore.ts`（`useSyncExternalStore` + listeners + `persisted/memory-only` 降级），生命周期由 `ticketMockEngine.ts` 定时推进。
- 工单状态机：`submitted → approving → processing → done`（另有 `rejected`）。当前 `ticketMockEngine` 在创建后定时自动推进：`submitted→approving`(5s)、`approving→processing`(30s)、`processing→done`(60s)，由 `RequestForm.tsx` 在提交时调用 `ticketMockEngine.schedule(ticketId)`。
- 代码库有「不凭空捏造尚不存在的数据/指标」的纪律（见 dashboard 占位文案）。本模块的数据全部由真实的工单流程派生，不引入虚构指标。

## 概念与生命周期

任务（OpsTask）是**工单派生的执行项**，通过 `ticketId` 显式关联工单，实现工单流程与运维任务的协同：

1. 工单进入 `processing` → 自动生成一个 `待领取` 任务。
2. 当前登录工程师**领取**任务 → 任务变 `进行中`，`assignee` 记为当前用户。
3. 工程师点**完成** → 任务变 `已完成`，并驱动关联工单 `processing → done`。
4. （可选）**退回/释放**：`进行中 → 待领取`，解除领取，供他人接手。

任务状态机：

```
待领取(pending) ──领取──▶ 进行中(in_progress) ──完成──▶ 已完成(done)
       ▲                        │
       └────────退回────────────┘
```

「任务驱动工单」是核心：`done` 不再由定时器推进，而由任务完成触发。

## 架构（方案 A：独立 store + 协调桥）

### 数据模型

新增 `web/src/view/service/data/taskTypes.ts`：

```ts
export type TaskStatus = 'pending' | 'in_progress' | 'done';

export interface TaskEvent {
  at: string;            // ISO
  status: TaskStatus;
  i18nNoteKey?: string;
  actor?: string;        // displayName，便于时间线展示
}

export interface OpsTask {
  id: string;            // 'TASK-YYYY-NNNN'
  ticketId: string;      // 关联工单
  title: string;         // 创建时从工单拷贝
  serviceId: string | null;
  status: TaskStatus;
  assignee: { name: string; email: string } | null;
  createdAt: string;     // ISO
  claimedAt?: string;
  completedAt?: string;
  timeline: TaskEvent[];
}

export type TaskStorageStatus = 'persisted' | 'memory-only';
```

### taskStore

新增 `web/src/view/service/data/taskStore.ts`，沿用 `ticketStore` 的全部约定：

- `localStorage` key：`optimus.opsTasks.v1`；解析失败/无值 → 空数组并 `console.warn`。
- 模块级 `tasks: OpsTask[]` + `listeners: Set<Listener>` + `persist()` 返回 `persisted | memory-only`。
- ID 生成器 `nextTaskId(now, existing)`：`TASK-YYYY-NNNN`，镜像 `nextTicketId`。

API：

```ts
taskStore.list(): OpsTask[]
taskStore.get(id): OpsTask | undefined
taskStore.subscribe(listener): () => void
taskStore.claim(id, user): void          // pending → in_progress，记 assignee/claimedAt
taskStore.complete(id): void             // in_progress → done，记 completedAt，并驱动工单
taskStore.release(id): void              // in_progress → pending，清 assignee
taskStore.reconcileFromTickets(tickets): void  // 幂等协调（见下）
taskStore.__resetForTest(): void

export function useTaskStore(): {
  tasks: OpsTask[];
  claim: typeof taskStore.claim;
  complete: typeof taskStore.complete;
  release: typeof taskStore.release;
  getTask: typeof taskStore.get;
}
```

所有状态变更都 `persist()` 后 `notify()`；非法转移（状态不匹配、任务不存在）静默忽略，保持与 `ticketStore.advance` 一致的防御风格。

### 协调桥（ticket → task）

`taskStore` 在模块初始化时 `ticketStore.subscribe(reconcile)`，并立即执行一次 `reconcileFromTickets(ticketStore.list())`：

- 对每个 `status === 'processing'` 且**尚无对应 `ticketId` 任务**的工单，补建一个 `待领取` 任务（标题/serviceId 从工单拷贝）。这同时覆盖：定时器推进到 processing、以及旧 `localStorage` 数据回填。
- reconcile **幂等**：已存在任务的工单跳过，不重复建。
- `complete(id)` 内部调用 `ticketStore.advance(ticketId, 'processing', 'done')`，使工单 timeline 追加完成事件，实现双向协同。complete 触发的工单变化会再次唤起 reconcile，但此时工单已离开 processing，不会新建任务（幂等成立）。

> 循环安全：taskStore 变更只通知 task listeners；ticketStore 变更通知 ticket listeners → reconcile。complete 引发的是 ticket 变化，reconcile 不再建任务，无无限循环。

### Mock 引擎改动

`ticketMockEngine.ts` 删除 `processing → done` 的 60s 定时步，仅保留 `submitted→approving`(5s)、`approving→processing`(30s)。`done` 由任务完成驱动。`STEP_60_MS` 常量及对应 `setTimeout` 移除。

## 组件与路由

- **新页面** `web/src/view/service/TaskListPage.tsx` + `TaskListPage.module.css`，路由 `/service-portal/tasks`（在 `router/index.tsx` 的 AppShell children 中新增，紧邻 tickets 路由）。
- **视图**：复用 `PageHeader`；顶部状态筛选 chips（`全部 / 待领取 / 进行中 / 已完成`）+ 「只看我的」开关 + 搜索框（按任务号/标题/工单号过滤）；表格行展示 `任务号 · 标题 / 关联工单 / 服务 / 状态 / 领取人 / 操作`。行内操作按状态条件渲染：`pending → 领取`、`in_progress(本人) → 完成 / 退回`、`查看`。
- **TaskStatusBadge** 组件（镜像 `TicketStatusBadge`），三态配色。
- **协同**：`查看` 跳转关联工单详情 `/service-portal/tickets/:id`；`TicketDetailPage` 补充一小块「关联任务」展示任务状态与领取人，体现双向链接。
- **导航**：`SideNav.tsx` 把 `任务管理` 的 `navEntryDisabled` span 换成指向 `/service-portal/tasks` 的 `NavLink`（与 tickets 同样式）。
- **i18n**：`messages.ts` 的 zh 与 en 两个块补 `task.*` 文案（页面标题/列头/状态名/操作/时间线事件/空态），命名沿用 `servicePortal.*` 风格。

## 错误处理

- 持久化降级：沿用工单的 `memory-only` banner 模式，领取/完成/退回失败时提示但不阻断内存态。
- 领取竞态：`claim` 对非 `pending` 任务幂等忽略。
- reconcile 幂等：重复触发不重复建任务。
- 旧数据：升级前已 `done` 的工单不回溯建任务；处于 `processing` 的旧工单由首次 reconcile 回填。

## 测试策略

`web/` 当前无测试运行器，且无既有前端测试。本模块引入 **vitest**（devDependency + `test` script + 最小 config），聚焦风险最高的 store/桥接纯逻辑：

1. `reconcileFromTickets` 对 processing 工单建任务、对已存在任务幂等。
2. `claim` 置 in_progress 并记 assignee；对非 pending 幂等。
3. `complete` 置 done 并驱动 `ticketStore.advance(processing→done)`。
4. `release` 回到 pending 并清 assignee。
5. 持久化降级（`localStorage` 抛错 → `memory-only`，状态仍推进）。

页面 UI（筛选/搜索/行内操作/导航/协同展示）通过运行 app 手工与视觉验证。

## 范围外（后续子项目）

- `我的值班`、`大屏展示` 模块。
- 任务的多人协作/指派他人、优先级、SLA 计时、批量操作。
- 任何后端持久化或 API。

## 关键文件清单

新增：
- `web/src/view/service/data/taskTypes.ts`
- `web/src/view/service/data/taskStore.ts`
- `web/src/view/service/TaskListPage.tsx` / `.module.css`
- `web/src/view/service/components/TaskStatusBadge.tsx`
- `web/src/view/service/data/taskStore.test.ts`（vitest）

改动：
- `web/src/view/service/data/ticketMockEngine.ts`（移除 60s 步）
- `web/src/view/service/TicketDetailPage.tsx`（关联任务展示）
- `web/src/components/nav/SideNav.tsx`（启用 任务管理 NavLink）
- `web/src/router/index.tsx`（新增 tasks 路由）
- `web/src/locale/messages.ts`（`task.*` zh + en）
- `web/package.json`（vitest + test script）
