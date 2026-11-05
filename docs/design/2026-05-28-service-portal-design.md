# 服务门户 完善设计（前端 mock 阶段）

- 状态：草稿（待 spec 复审）
- 日期：2026-05-28
- 适用范围：本期仅前端，纯 mock data；后端工单系统在后续迭代再补
- 关联模块：`web/src/view/service/`、`web/src/components/nav/SideNav.tsx`、`web/src/router/index.tsx`

> 2026-10-26 维护说明：上方草稿状态为历史记录，未补造批准结论。本文的“现状”和非目标描述对应 2026-05-28 阶段。后续 [任务管理设计](2026-06-02-ops-center-tasks-design.md) 将自动 done 改为任务完成驱动，并扩展任务入口；[审批链设计](2026-06-04-ticket-approval-chain-design.md) 进一步规划多级审批（尚未完整实现），任务阶段已引入测试基建。因此“无任务”“无多级审批”“不新增测试依赖”不能脱离原阶段范围作为永久规则。原文保留；当前实现和验证见 [运维流程验证 #2](https://github.com/Dcuhddrg/Optimus/issues/2)。

## 1. 背景与目标

运维中心目前只有一个 `ServicePortalPage` 静态 mockup：5 个分类 tab、4 张服务卡、3 条假工单，按钮无行为，后端无任何工单 / 请求 model。SideNav 里「运维中心」组的「工单 / 任务 / 我的值班 / 大屏」全部 disabled 占位。

本次要把服务门户从静态 mockup 升级为**可交互的前端原型**：服务目录可点击、申请表单 Drawer 可填可提交、我的工单列表可看状态自动推进、工单列表与详情页打通。**纯前端、纯 mock**，所有数据走 `localStorage`；后端工单系统在后续迭代再接，届时 mock 层会被替换。

**非目标**：

- 不动后端；不引入新依赖（运行时与测试依赖都不新增；不引 MSW / zod / antd / vitest / testing-library 等）
- 不做审批流配置 / 多级审批 / SLA / 通知
- 不做「任务 / 我的值班 / 大屏」三个 SideNav 占位
- 不做端到端测试
- 不做移动端响应式适配（桌面优先；Drawer 固定 520px，窄屏体验本期不保证）
- 不做用户级数据隔离（mock 数据在浏览器内全局共享；详见 §6.5）

## 2. 范围对齐

| 维度 | 决定 |
| --- | --- |
| 后端 | 不动；纯前端 mock |
| 服务目录 | 10–15 张卡 / 4–5 个分类（实际 13 条 / 5 类） |
| 申请入口 | 点卡 → 右侧 Drawer + 该服务专属表单 |
| 通用入口 | 顶部「创建申请」→ 通用表单 Drawer（无服务绑定） |
| 工单持久化 | localStorage |
| 状态机 | 提交 5s → `审批中`，30s → `处理中`，60s → `已完成`；纯前端 setTimeout，刷新即停摆 |
| 工单详情 | 独立路由 `/service-portal/tickets/:id` |
| 工单列表 | 独立路由 `/service-portal/tickets`；搜索匹配 `title + id`（模糊）；过滤为状态多选 chips |
| 前端架构 | 方案 1：模块级 TicketStore + Hooks wrapper，无新依赖 |

## 3. 路由与导航

### 3.1 路由

在 `web/src/router/index.tsx` 现有 service-portal 路由处扩展：

```
/service-portal              ServicePortalPage     目录 + 我的工单侧栏（首页）
/service-portal/tickets      TicketListPage        全部工单列表（按状态过滤、搜索）
/service-portal/tickets/:id  TicketDetailPage      单条工单详情 + 状态时间轴
```

「创建」动作走 Drawer，不占 URL —— 跟 `/cloud/assets` 现有交互保持一致。

### 3.2 SideNav

`web/src/components/nav/SideNav.tsx` 现有「工单」从 disabled 升级为 NavLink，指向 `/service-portal/tickets`。其余三个占位（任务 / 我的值班 / 大屏）继续保持 disabled，不在本期范围。

「工单」NavLink 本期**不**接 `hasMenu()` 权限码检查，保持跟其他占位一致；待后端 `ops.tickets.menu` 等权限码就绪时整组接入。

### 3.3 面包屑

详情页头部：`服务门户 / 我的工单 / [工单标题]`，前两段可点击。

## 4. 数据模型

### 4.1 ServiceCatalogItem（静态目录，hardcode）

```ts
type ServiceCategory = 'account' | 'cloud' | 'network' | 'env' | 'support';

interface ServiceCatalogItem {
  id: string;                 // 例 'account.role-grant'
  category: ServiceCategory;
  i18nKey: string;            // 例 'servicePortal.catalog.accountRoleGrant'
  iconToken: string;          // 前端 SVG token：'shield' | 'cloud' | 'network' …
  tags: string[];             // i18n key 数组，例 ['tag.approval', 'tag.1to2days']
  hot?: boolean;              // 「热门」分类下展示
  estimatedSla: string;       // i18n key，例 'sla.1to2BusinessDays'
}
```

「热门」是虚拟分类，展示**全部** `hot: true` 的卡片（不分页、不限数量），由聚合而成、不重复存。

### 4.2 ServiceFormSchemas（表单字段字典，独立模块）

```ts
type FieldType = 'text' | 'textarea' | 'select' | 'date' | 'multiselect';
// 不含 'file'：File 对象不可序列化进 localStorage，本期不做文件上传 mock；
// 表单里需要"附件"语义时降级为 textarea 让用户填文件名/链接描述。

interface RequestFormField {
  name: string;
  type: FieldType;
  i18nLabelKey: string;
  required: boolean;
  options?: { value: string; i18nLabelKey: string }[];
  placeholderI18nKey?: string;
  maxLength?: number;
}

type ServiceFormSchemas = Record<string, RequestFormField[]>;
// 特殊 key '__generic' 对应「通用申请」表单
```

Drawer 查表：`ServiceFormSchemas[serviceId] ?? ServiceFormSchemas.__generic`。

目录与 schema 分文件管理（`data/serviceCatalog.ts` / `data/serviceFormSchemas.ts`），各自可单测，孤儿 schema 易识别。

### 4.3 Ticket（mock，存 localStorage）

```ts
type TicketStatus = 'submitted' | 'approving' | 'processing' | 'done' | 'rejected';

interface Ticket {
  id: string;                              // 'TKT-2026-0001'，简单递增
  serviceId: string | null;                // null = 通用申请
  title: string;
  formValues: Record<string, unknown>;
  status: TicketStatus;
  createdAt: string;                       // ISO
  timeline: TicketEvent[];
  applicant: { name: string; email: string };  // 从 AuthContext 取
}

interface TicketEvent {
  at: string;                              // ISO
  status: TicketStatus;
  i18nNoteKey?: string;
}
```

`rejected` 状态本期不会被状态机自动到达，但保留在类型中为详情页时间轴渲染做准备。

### 4.4 localStorage 形态

- 单个 key：`optimus.servicePortal.tickets.v1`
- 值：`Ticket[]` 的 JSON 序列化
- 版本后缀 `.v1`：未来 schema 变化时方便清理旧版

## 5. 组件分解

```
web/src/view/service/
├── ServicePortalPage.tsx           ← 改造现有；首页 = 目录 + 我的工单侧栏
├── ServicePortalPage.module.css    ← 现有，按需补样式
├── TicketListPage.tsx              ← 新增；全部工单列表
├── TicketListPage.module.css
├── TicketDetailPage.tsx            ← 新增；详情 + 时间轴
├── TicketDetailPage.module.css
├── components/
│   ├── ServiceCard.tsx             ← 目录卡（从 page 抽出）
│   ├── CategoryTabs.tsx            ← 分类 tab（从 page 抽出）
│   ├── MyTicketsAside.tsx          ← 首页右侧「我的工单」侧栏
│   ├── TicketStatusBadge.tsx       ← 跨页复用
│   ├── TicketTimeline.tsx          ← 详情页时间轴
│   ├── RequestDrawer.tsx           ← 右侧抽屉容器（layout + 关闭交互）
│   └── RequestForm.tsx             ← Drawer 内部表单（按 schema 动态渲染）
└── data/
    ├── serviceCatalog.ts           ← 13 条目录 + 5 个分类常量
    ├── serviceFormSchemas.ts       ← 表单 schema 字典 + __generic
    ├── ticketStore.ts              ← 模块级 ticketStore + useTicketStore() wrapper + localStorage 持久化
    ├── ticketMockEngine.ts         ← setTimeout 状态机
    └── ticketTypes.ts              ← Ticket / TicketStatus 等共享类型
```

### 职责边界

- **ServicePortalPage**：组合 `<CategoryTabs>` + `<ServiceCard>` grid + `<MyTicketsAside>`；不持有数据逻辑
- **RequestDrawer**：portal 渲染到 `document.body`；管动画、遮罩、关闭交互（按钮 / ESC / 遮罩点击）；内部委托 `<RequestForm>` 渲染
- **RequestForm**：按 `RequestFormField[]` 动态渲染 → useState 收值 → 校验 → 调 `useTicketStore().createTicket()`；持久化成功则关 Drawer（不弹 toast，详见 §8.1）；持久化失败时 Drawer 顶部 banner（详见 §8.2）
- **ticketStore / useTicketStore**：`ticketStore` 是唯一的工单读 / 写入口；`useTicketStore()` 只负责把模块级 store 暴露给 React 组件订阅

### 复用现有

`PageHeader`、`useI18n`、`useAuth`（取申请人姓名邮箱）。

### 新依赖

零。Drawer 用 CSS transform + `createPortal` 原生实现。

## 6. 数据流与状态机

### 6.1 写入路径

```
RequestForm
  ├─ 本地 useState 收集字段
  ├─ 校验：required / maxLength
  └─ onSubmit
        └─→ useTicketStore().createTicket({serviceId, title, formValues})
                ├─ 生成 id（'TKT-' + 年 + 4 位流水）
                ├─ status = 'submitted'，timeline 写一条 submitted 事件
                ├─ 生成 nextTickets（新工单不可变插入列表头部）
                ├─ 同步尝试 localStorage.setItem(nextTickets)
                ├─ 更新模块级内存 state + notify subscribers
                ├─ 返回 { ticket, storageStatus: 'persisted' | 'memory-only' }
                └─ ticketMockEngine.schedule(ticket.id)
                                              setTimeout 链：
                                              5s  → advance('approving')
                                              30s → advance('processing')
                                              60s → advance('done')
```

### 6.2 读取路径

`ticketStore` 是模块级单例：

- 模块首次加载时从 localStorage 反序列化一次，得到内存 `tickets`
- store 内部维护 `listeners: Set<() => void>`，每次变更后通知订阅者
- `useTicketStore()` 使用 React 18 内置 `useSyncExternalStore` 订阅模块级 store，不持有独立 ticket state
- `ServicePortalPage`、`MyTicketsAside`、`TicketListPage`、`TicketDetailPage` 同时挂载时共享同一份内存状态
- `ticketMockEngine` 不依赖 React hook，只调用模块级 `ticketStore.advance(ticketId, expectedFrom, nextStatus)`

### 6.3 状态机（ticketMockEngine.ts）

- 模块级 `Map<ticketId, number[]>`：记录每条工单未触发的 timeout id
- `schedule(ticketId)`：注册 5s / 30s / 60s 三个 setTimeout，每次触发调 `ticketStore.advance(ticketId, expectedFrom, nextStatus)`
- `cancel(ticketId)`：清掉对应 timer（本期不暴露 UI，留 API）
- **页面卸载不清 timer**，路由切换时定时器继续跑
- **页面刷新后状态原地停摆**（不做"按时间差补推"）；下次进入页面看到的就是 localStorage 里写入的最后状态
- `ticketStore.advance` 检查 `currentStatus === expectedFrom`，否则 no-op

### 6.4 通用申请

`serviceId === null`，标题用户填入；其余流程与服务申请一致。

### 6.5 多用户 mock 共享

localStorage **不按用户分**，浏览器内一份全局 mock 数据。换号登录后看到的仍是同一份工单列表，且 `applicant` 字段是提交时刻从 AuthContext 取的 snapshot —— 换号后会看到"自己的名字下挂着别人提交的工单"的小偏差。

可接受：演示原型阶段的妥协，接真实后端时由服务端自然按用户隔离。

### 6.6 「我的工单」侧栏

首页 `<MyTicketsAside>` 展示最新 5 条工单（按 `createdAt` 倒序），底部带「查看全部 →」链接到 `/service-portal/tickets`。

## 7. i18n 与样式

### 7.1 i18n

- 新命名空间：`servicePortal.*`（独立于 `dashboard.*`）
- 现有 `dashboard.servicePortal.*` 文案平移到新前缀；旧 key 在**同一 PR 中删除**（grep 已确认仅 `ServicePortalPage.tsx` 引用，迁移后无残留消费方）
- 三层结构：
  - `servicePortal.catalog.{serviceId}.{title|description}`
  - `servicePortal.form.{fieldName}.{label|placeholder|option.X}`
  - `servicePortal.ticket.status.{submitted|approving|processing|done|rejected}`
- 中英双语同步加（按现有 i18n 文件结构）

### 7.2 样式

- 沿用 CSS Modules
- Drawer 用 `RequestDrawer.module.css`：遮罩 `position:fixed;inset:0;background:rgba(0,0,0,.4)`；抽屉 `position:fixed;right:0;top:0;bottom:0;width:520px;transform:translateX(100%)`；`[data-open=true]` 切换 `translateX(0)` + transition
- 不引入新色板，复用现有 `ServicePortalPage.module.css` 的 token
- `TicketStatusBadge` 5 个状态 → 5 个 CSS class：`submitted`/`approving`/`processing`/`done`/`rejected`，颜色复用 warning / info / success / neutral / danger（先 grep 找现有 token，能复用就复用）

## 8. 错误处理与边界

### 8.1 表单提交

- 客户端校验：必填项 / maxLength → 字段下方红字
- 提交按钮 onClick 后立即设 disabled（即使内部 `createTicket` 同步完成，也防止 ~16ms 内重复点击）
- `createTicket()` 返回 `storageStatus === 'persisted'`：直接关 Drawer；视觉反馈靠"工单列表头部多出新行"完成 —— **不引入 toast 组件**（项目当前没有 toast 实现，且不在本期范围内新建）
- `createTicket()` 返回 `storageStatus === 'memory-only'`：不关 Drawer，顶部展示 inline banner；表单保持已提交态，提交按钮保持 disabled，避免用户重复创建同一张工单

### 8.2 localStorage 异常

| 场景 | 处理 |
| --- | --- |
| `JSON.parse` 抛错 | 当空列表，console.warn，**不主动清旧数据** |
| `setItem` 抛错（含 QuotaExceededError、浏览器禁用 / 只读 localStorage） | catch，`createTicket()` 返回 `storageStatus: 'memory-only'`；**在 Drawer 顶部展示 inline banner**「本地存储不可用，工单仅留在内存，刷新即丢」；新工单仍 push 到内存 state、Drawer 不关，提交按钮保持 disabled |
| key 不存在 | 正常初始化空列表 |

### 8.3 路由

- `/service-portal/tickets/:id` 找不到对应工单 → 原地空状态「工单不存在或已被清理」+ 回列表按钮
- 不做 404 跳转

### 8.4 状态机

- `advance()` 发现工单已被删除 → no-op
- 不做幂等保护（与"刷新即停摆"配套）

## 9. 测试策略

### 9.1 自动验证（不新增测试依赖）

本仓库当前 `web/package.json` 没有前端测试框架。本期遵守"不引入新依赖"，因此自动化验证限定为现有命令：

- `cd web && npm run build`：覆盖 TypeScript 编译与 Vite 构建，验证路由、组件、类型、i18n key 类型约束
- 静态自检：实现时用项目现有搜索工具确认 `dashboard.servicePortal.*` 迁移后无残留引用

### 9.2 手动验证清单

1. dev server，登录后进 `/service-portal`：13 张卡分 5 类，tab 切换不抖
2. 点任一卡 → Drawer 滑出 → 填表 → 提交：Drawer 自动关、侧栏顶部多出新工单（「已提交」）
3. 等 5s → 「审批中」；30s → 「处理中」；60s → 「已完成」
4. 顶部「创建申请」→ Drawer 打开通用表单（不绑服务）→ 提交后同上
5. 顶部「查看全部」→ `/service-portal/tickets`：状态多选 chips 过滤、`title + id` 搜索均可用
6. 列表点一行 → 详情 → 时间轴显示已发生事件
7. SideNav「工单」直链同样到列表页
8. 刷新首页：localStorage 工单还在，状态固定在刷新前那一档（不再续推）
9. 手输不存在的 `/service-portal/tickets/TKT-X-9999`：空状态而非崩溃
10. 浏览器禁用 localStorage / 只读 → 提交时 Drawer 顶部出现红色 banner「本地存储不可用…」，页面不崩
11. 切换登录账号后回到服务门户：仍能看到之前账号提交的工单（mock 不按用户隔离，§6.5）

### 9.3 不测的（明确范围外）

- E2E / 视觉回归 / 接口契约（无后端）
- Vitest / Testing Library 单测与组件测试（当前项目无测试栈，本期不新增测试依赖）

## 10. 未来工作（不在本期）

- 接真实后端：把 `useTicketStore` 内部改成调 `apiRequest`，删除 `ticketMockEngine.ts`；接口形状与本期函数签名保持一致
- 审批流配置 / 多级审批 / SLA / 邮件通知
- `ops.tickets.menu` 等权限码接入 SideNav
- 「任务 / 我的值班 / 大屏」三个 SideNav 占位
- 如项目后续统一引入前端测试栈，再补 `ticketStore`、`ticketMockEngine`、`RequestForm`、`RequestDrawer`、`TicketStatusBadge` 的 Vitest / Testing Library 覆盖
