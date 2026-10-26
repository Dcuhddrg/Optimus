# 工单管理 · 审批链展示（中间审批人 + 当前进度）设计

- 日期：2026-06-04
- 范围：服务门户「工单管理」展示增强 —— 中间审批人 + 当前进度（前端优先，纯 mock）
- 状态：原文记录已通过设计评审；[实施计划已存在](../plans/2026-06-04-ticket-approval-chain.md)，部分实现，尚未完成交付验收。
- 最近核查：2026-10-26；跟进 [#4](https://github.com/Dcuhddrg/Optimus/issues/4)，证据见 [能力状态](../../project-status.md)。
- 分支：在 `feat/ops-center-tasks` 基础上继续（不另开分支）

> 演进关系：本设计拟将 [任务阶段](2026-06-02-ops-center-tasks-design.md) 的固定审批推进改为逐级通过，保留任务完成驱动工单 done。当前仅有类型等增量，不能把本文中的逐级审批步骤当成已实现行为。

## 背景

工单现状：状态机 `submitted → approving → processing → done`（另有 `rejected`），由 `ticketMockEngine` 定时自动推进（`submitted→approving` 5s、`approving→processing` 30s）；`processing→done` 已在任务模块改为由任务完成驱动。`Ticket` 只有 `applicant`，**没有任何审批人概念**，`approving` 是一个不可见的黑盒；详情页用 `TicketTimeline` 顺序展示状态事件，进度仅靠状态徽章体现。

本设计补齐：谁在审批（中间审批人）、整体走到哪（当前进度）。后端后续再补，本期为纯前端展示，mock 引擎自动逐级推进，不做交互式「通过/驳回」。

## 决策摘要

- **审批模型**：按服务类别配置的多级审批链（1~3 级）。"中间审批人" = 当前待审那一级的审批人。
- **范围**：纯展示。mock 引擎按链逐级自动通过；无用户交互审批。
- **数据来源**：静态 mock —— 审批人目录 + 按类别的链定义；创建工单时把链**快照**进工单，之后稳定不变。
- **详情页展示**：顶部横向步骤条（方案 A）+ 增强现有时间线织入审批人名（方案 C）。不做独立明细面板（B，信息重叠，YAGNI）。
- **列表页展示**：`approving` 行在状态徽章下加一行小字 `审批 m/n · 当前:<name>`，派生计算，不加列。
- **推进节奏**：`submitted→approving` 5s 后，审批链每 ~10s 逐级通过，走完最后一级进入 `processing`。

## 数据模型（`web/src/view/service/ticketTypes.ts` 扩展）

```ts
export type ApprovalStepStatus = 'pending' | 'approved';

export interface ApprovalStep {
  roleI18nKey: string;                       // 角色名 i18n leaf key
  approver: { name: string; email: string };
  status: ApprovalStepStatus;
  decidedAt?: string;                        // ISO，通过时写入
}

// Ticket 新增字段：
//   approvalChain: ApprovalStep[];          // 创建时快照，全部初始 pending
//
// TicketEvent 新增可选字段（供时间线织入审批人名，镜像任务模块的 actor 模式）：
//   actor?: string;                         // 审批人 displayName
//   roleI18nKey?: string;                   // 审批角色
```

**全部"当前/进度"派生，不存冗余状态**：
- 当前审批步 = `approvalChain` 中第一个 `status === 'pending'` 的步；无则审批已完成。
- 进度 = `approved` 数 / 链总长。

派生逻辑集中在选择器（建议放 `data/approvalSelectors.ts` 或就近的纯函数）：`currentApprovalStep(ticket)`、`approvalProgress(ticket) → { approved, total }`。

## Mock 数据（新增 `web/src/view/service/data/approvalChains.ts`）

- **审批人目录**：少量固定 mock 人（如 王明·直属主管、李强·安全审批、张伟·部门负责人 等），每人 `{ name, email }`。
- **按类别的链**：`ServiceCategory`（account / cloud / network / env / support）各配一条有序链；级数 1~3 按类别区分（如 account=1 级主管；network/env=2 级；生产/支持变更=3 级）。
- **默认链**：`serviceId` 为 `null` 或未在目录中的服务 → 默认 1 级（主管）。
- `buildApprovalChain(serviceId: string | null): ApprovalStep[]`：经 `serviceCatalog` 解析 `category`，返回对应链的**新副本**（每步 `status: 'pending'`，无 `decidedAt`）。

> 链是按类别定义的常量模板；`buildApprovalChain` 负责取模板并深拷贝出可写快照，避免共享引用被工单改写。

## Store + 引擎行为

### ticketStore
- `createTicket`：调用 `buildApprovalChain(input.serviceId)`，把链快照写入新工单 `approvalChain`（全 pending）。其余不变。
- 新增 `advanceApproval(ticketId: string): void`：
  1. 若 `status !== 'approving'` → no-op（防御，与 `advance` 一致）。
  2. 找到第一个 `pending` 步：
     - **有**：标记 `approved` + `decidedAt = now`，追加一条 timeline 事件 `{ at, status: 'approving', i18nNoteKey: 'servicePortal.ticket.event.approved', actor, roleI18nKey }`。
     - **无**（空链/缺失链的旧工单）：跳过审批标记。
  3. 标记后若链中**已无任何 pending 步**（含上面"无 pending"的情形）：追加 `approving→processing` 推进（status 改 `processing` + processing 事件，复用既有 `advance` 风格）。
  4. `persist()` 后 `notify()`；持久化降级沿用既有 `memory-only` 容错。

### ticketMockEngine
- 保留 `submitted→approving`（5s）。
- **删除**固定的 `approving→processing` 30s 单步。
- `approving` 后，按 `ticket.approvalChain.length` 安排 N 次 `advanceApproval`，每次间隔 ~10s（如第 i 级在 `5s + i*10s` 触发）；最后一次内部自动进 `processing`。
- 链长在 `schedule` 时从工单读取（创建后快照已存在）。

## 组件与展示面

- **新增 `components/ApprovalStepper.tsx` + `.module.css`**：详情页顶部横向步骤条。节点序列 = `提交` → 各审批级（用 `roleI18nKey` 作标签、审批人名作副标题）→ `处理中` → `完成`。节点态由 `ticket.status + approvalChain` 计算：已过=绿勾、当前=蓝圈高亮 + "审批中"、未到=灰。被 rejected 的工单按当前级标红（展示层处理，本期引擎不产生 rejected）。
- **增强 `components/TicketTimeline.tsx`**：事件文案织入审批人名 —— 当 `event.actor` 存在时，对 `i18nNoteKey` 文案做 `.replace('{actor}', actor).replace('{role}', t(roleI18nKey))`，得到 `王明（直属主管）已通过`。无 actor 的旧事件行为不变。
- **`TicketDetailPage.tsx`**：在 `<section className={styles.summary}>` 之前插入 `<ApprovalStepper ticket={ticket} />`。
- **`TicketListPage.tsx`**：行内状态徽章下方，当 `ticket.status === 'approving'` 时渲染一行小字：`审批 {approved}/{total} · {currentRole}:{currentName}`（派生，不新增列）。

## i18n（`web/src/locale/messages.ts`，en + zhCN）

新增 `servicePortal.ticket.approval.*`：
- `role.*`：lineManager / security / deptHead 等角色名。
- `event.approved`：带 `{actor}` `{role}` 占位（手动 `.replace`，沿用现有约定，无插值类型）。
- `stepper.*`：步骤条固定节点标签（提交 / 处理中 / 完成 / 审批中 / 待审批）。
- `list.approvalProgress`：列表小字模板（含 `{count}` `{total}` `{role}` `{name}` 占位）。

en 与 zhCN 两块同步新增，保持 `zhCN: typeof en` 结构一致（`tsc` 校验）。

## 错误处理

- 持久化降级：沿用工单既有 `memory-only` 容错，审批推进失败仍在内存生效。
- 旧数据兼容：升级前已存在的工单 `approvalChain` 缺失 → 读取时一律按 `approvalChain ?? []` 处理，步骤条退化为「提交 → 处理 → 完成」（不渲染审批级），不报错。`advanceApproval` 的规则统一为：`status === 'approving'` 且**无 pending 步**（含空链/缺失链）时，直接 `approving→processing`；这样处于 `approving` 的旧工单不会卡死，新工单走完最后一级也复用同一路径。
- 空链/默认链：保证每条链 ≥1 级，避免 `approving` 卡死。

## 测试策略（vitest，聚焦纯逻辑）

1. `approvalChains`：各类别返回正确链；`null`/未知服务返回默认 1 级；返回的是独立副本（改一个工单的链不影响模板）。
2. `createTicket`：新工单 `approvalChain` 按服务类别快照、全 `pending`。
3. `advanceApproval`：逐级把 pending → approved 并写 `decidedAt` 与带 actor/role 的事件；最后一级触发 `approving→processing`；非 `approving` 状态 no-op；`approving` 但空链/无 pending 时直接进 `processing`（兜底不卡死）。
4. `ticketMockEngine`（fake timers）：从 approving 起每 ~10s 通过一级，走完 N 级进入 `processing`，且不自动 `done`。
5. 派生选择器：`currentApprovalStep` / `approvalProgress` 在链不同进度下返回正确值。

详情页步骤条、时间线文案、列表小字的视觉通过运行 app 手工验证。

## 范围外（后续）

- 后端持久化与真实审批人/权限。
- 交互式「通过/驳回」、加签/转审、驳回回退到指定级。
- 审批 SLA 计时、超时升级、提醒。
- 独立的审批链明细面板（方案 B）。

## 关键文件清单

新增：
- `web/src/view/service/data/approvalChains.ts`
- `web/src/view/service/components/ApprovalStepper.tsx` / `.module.css`
- 测试：`approvalChains.test.ts`、`ticketStore` 审批相关测试、`ticketMockEngine` 审批推进测试（在既有/新建 test 文件中）

改动：
- `web/src/view/service/ticketTypes.ts`（ApprovalStep + Ticket/TicketEvent 字段）
- `web/src/view/service/data/ticketStore.ts`（createTicket 快照、advanceApproval）
- `web/src/view/service/data/ticketMockEngine.ts`（逐级推进，删 30s 单步）
- `web/src/view/service/components/TicketTimeline.tsx`（织入审批人名）
- `web/src/view/service/TicketDetailPage.tsx`（插入步骤条）
- `web/src/view/service/TicketListPage.tsx`（审批小字）
- `web/src/locale/messages.ts`（approval.* en + zh）
