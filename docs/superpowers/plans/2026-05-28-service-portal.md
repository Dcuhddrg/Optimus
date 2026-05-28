# 服务门户（前端 mock 阶段）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把运维中心-服务门户从静态 mockup 升级为可交互前端原型：13 张服务卡 / 5 分类 / 申请抽屉 / localStorage 工单 / setTimeout 状态机推进 / 工单列表与详情。

**Architecture:** 纯前端 + localStorage；模块级 `ticketStore` + `useSyncExternalStore` 订阅；setTimeout 状态机模块解耦；按 spec 拆 7 个组件 / 5 个 data 模块 / 2 个新页面 + 1 个改造页面 + 路由 + SideNav 改动。

**Tech Stack:** React 18.3 / react-router-dom 6.28 / TypeScript 5.6 / Vite 5.4 / CSS Modules。**不引入任何新运行时或测试依赖（无 vitest / testing-library / MSW / zod / antd）。**

**Spec:** `docs/superpowers/specs/2026-05-28-service-portal-design.md`

**关键代码事实**（实施前必读）：

- i18n 系统：`web/src/locale/messages.ts` 顶部声明 `const en = {...}` 与 `const zhCN = {...}`，文件末尾用 `type MessageKey = LeafMessageKeys<typeof en>` **从 `en` 字面量推导 key 类型**。新增 key 必须先加进 `en`，再加进 `zhCN`；TS 编译会强制要求 zhCN 与 en 同构（用 `satisfies Record<AppLocale, MessageDictionary>` 校验）
- `t()` 形参类型为 `MessageKey`（点路径字符串）。代码里把"运行时 i18nKey 字符串"传给 `t()` 时需要 `as MessageKey` 显式断言
- `useAuth()` 返回 `{ current: { user: { displayName, email, ... } } }`，作为 `applicant` snapshot 来源
- 现有路由 `web/src/router/index.tsx` 在 `AppShell` 子路由下，无懒加载；新增路由直接平铺在同级 `children` 数组
- 项目无测试框架：`web/package.json` 只有 `dev / build / preview` 三个脚本；每个任务的"验证"统一用 `cd web && npm run build` + 手动浏览器走查
- 既有 `dashboard.servicePortal.*` i18n key 经 grep 只在 `web/src/view/service/ServicePortalPage.tsx` 一处消费

**Commit 约定：** `feat(service-portal): ...` 或 `refactor(service-portal): ...`，遵循仓库现有中文 commit 风格。

---

## 文件结构

**Modify:**

- `web/src/locale/messages.ts` — 迁移 `dashboard.servicePortal.*` → `servicePortal.*`，新增 catalog / form / status / common 子树（en + zhCN）
- `web/src/components/nav/SideNav.tsx:99` — 把「工单」从 disabled `<span>` 升级为 `<NavLink to="/service-portal/tickets">`
- `web/src/router/index.tsx` — 在 `service-portal` 路由附近新增 `/service-portal/tickets` 和 `/service-portal/tickets/:id` 两条
- `web/src/view/service/ServicePortalPage.tsx` — 全量改造：13 卡 / 5 分类 + 顶部按钮接 Drawer / 「我的工单」组件化
- `web/src/view/service/ServicePortalPage.module.css` — 调整 grid（适配组件化抽出）

**Create:**

- `web/src/view/service/ticketTypes.ts` — 共享类型：`ServiceCategory` / `ServiceCatalogItem` / `FieldType` / `RequestFormField` / `ServiceFormSchemas` / `TicketStatus` / `TicketEvent` / `Ticket` / `StorageStatus` / `CreateTicketResult`
- `web/src/view/service/data/serviceCatalog.ts` — 13 条目录 + 5 个分类常量
- `web/src/view/service/data/serviceFormSchemas.ts` — 服务专属表单 + `__generic`
- `web/src/view/service/data/ticketStore.ts` — 模块级单例 + `useTicketStore()` (`useSyncExternalStore`)
- `web/src/view/service/data/ticketMockEngine.ts` — setTimeout 状态机
- `web/src/view/service/components/TicketStatusBadge.tsx` + `.module.css`
- `web/src/view/service/components/CategoryTabs.tsx` + `.module.css`
- `web/src/view/service/components/ServiceCard.tsx` + `.module.css`
- `web/src/view/service/components/MyTicketsAside.tsx` + `.module.css`
- `web/src/view/service/components/RequestDrawer.tsx` + `.module.css`
- `web/src/view/service/components/RequestForm.tsx` + `.module.css`
- `web/src/view/service/components/TicketTimeline.tsx` + `.module.css`
- `web/src/view/service/TicketListPage.tsx` + `.module.css`
- `web/src/view/service/TicketDetailPage.tsx` + `.module.css`

---

## Task 1: 迁移既有 i18n 命名空间 `dashboard.servicePortal.*` → `servicePortal.*`

**Files:**

- Modify: `web/src/locale/messages.ts`（en 第 ~216–254 行 + zhCN 第 ~523–559 行）
- Modify: `web/src/view/service/ServicePortalPage.tsx`（所有 `t('dashboard.servicePortal.*')` 调用）

- [ ] **Step 1: 在 `en.dashboard` 之外的顶层添加 `servicePortal` 节点**

打开 `web/src/locale/messages.ts`。找到 `en` 常量。在 `dashboard` 节点**内部**删除整个 `servicePortal: { ... }` 子树（包括 eyebrow / title / description / viewAll / createRequest / 5 个 category / 4 组 title+description / 8 个 tag / myTickets / myTicketsDescription / 3 组 ticket + meta / 3 个 status）。

然后在 `dashboard` 节点**之外**（同 `dashboard` / `system` / `shell` 同级），插入一个新顶级节点 `servicePortal`，把刚删的内容平移过来 + 补全新增 key。最终结构：

```ts
const en = {
  // ...existing common / auth / shell / project / cloud / dashboard / system...

  servicePortal: {
    eyebrow: 'Service portal',
    title: 'Service portal',
    description:
      'Add IT services to the current dashboard as request entry points, so business users can submit service requests and follow recent progress without leaving the workspace.',
    viewAll: 'View all services',
    createRequest: 'Create request',
    cancel: 'Cancel',
    submit: 'Submit',
    submitting: 'Submitting...',
    optional: 'Optional',
    required: 'Required',
    validationRequired: 'This field is required',
    validationTooLong: 'Maximum {max} characters',
    storageFailed: 'Local storage unavailable. The ticket only lives in memory and will be lost on refresh.',
    breadcrumbHome: 'Service portal',
    breadcrumbTickets: 'My tickets',
    backToTickets: 'Back to tickets',
    notFoundTitle: 'Ticket not found or already cleared',
    notFoundDescription:
      'The ticket may have been removed from your browser storage or never existed. Return to the tickets list to continue.',
    category: {
      label: 'Service categories',
      popular: 'Popular',
      account: 'Accounts & access',
      cloud: 'Cloud resources',
      network: 'Network & security',
      env: 'Environments',
      support: 'Support',
    },
    catalog: {
      accountRoleGrant: {
        title: 'Account role grant',
        description: 'Request a new role or elevated permission on an existing system account.',
      },
      // ...其他 12 条占位在 Task 2 中补全
    },
    form: {
      // 在 Task 5 中补全
    },
    ticket: {
      list: {
        eyebrow: 'Service portal',
        title: 'My tickets',
        description: 'All service requests submitted from this browser.',
        searchPlaceholder: 'Search by title or ticket id',
        filterLabel: 'Filter by status',
        empty: 'No tickets match the current filter.',
        columnId: 'Ticket id',
        columnTitle: 'Title',
        columnStatus: 'Status',
        columnCreated: 'Submitted at',
      },
      detail: {
        eyebrow: 'Service portal',
        applicantLabel: 'Applicant',
        serviceLabel: 'Service',
        genericServiceLabel: 'General request',
        formValuesLabel: 'Submitted fields',
        timelineLabel: 'Timeline',
      },
      myAside: {
        title: 'My recent tickets',
        description: 'Latest service requests from this browser.',
        viewAll: 'View all →',
        empty: 'No tickets yet. Pick a service from the catalog to start.',
      },
      status: {
        submitted: 'Submitted',
        approving: 'Approving',
        processing: 'Processing',
        done: 'Done',
        rejected: 'Rejected',
      },
      event: {
        submitted: 'Ticket submitted by {applicant}.',
        autoAdvanced: 'Status auto-advanced to {status}.',
      },
    },
    sla: {
      oneToTwoBusinessDays: '1-2 business days',
      sameDay: 'Same day',
      threeBusinessDays: 'Up to 3 business days',
    },
    tag: {
      approval: 'Approval',
      oneToTwoDays: '1-2 days',
      projectScope: 'Project scoped',
      costReview: 'Cost review',
      securityReview: 'Security review',
      frequent: 'Frequent',
      fastResponse: 'Fast response',
      trackable: 'Trackable',
    },
  },

  system: { ... },
};
```

zhCN 镜像对应位置，对照翻译。例如：

```ts
const zhCN = {
  // ...
  servicePortal: {
    eyebrow: '服务门户',
    title: '服务门户',
    description: '把 IT 部门可提供的服务以工单入口的形式补充到当前页面，业务用户可直接发起申请并查看最近进度。',
    viewAll: '查看全部服务',
    createRequest: '发起服务申请',
    cancel: '取消',
    submit: '提交',
    submitting: '提交中...',
    optional: '选填',
    required: '必填',
    validationRequired: '此项必填',
    validationTooLong: '不超过 {max} 个字符',
    storageFailed: '本地存储不可用。本张工单仅留在内存，刷新即丢。',
    breadcrumbHome: '服务门户',
    breadcrumbTickets: '我的工单',
    backToTickets: '返回工单列表',
    notFoundTitle: '工单不存在或已被清理',
    notFoundDescription: '该工单可能已从浏览器存储中移除或从未存在。请返回工单列表继续。',
    category: {
      label: '服务分类',
      popular: '常用服务',
      account: '账号权限',
      cloud: '云资源',
      network: '网络安全',
      env: '环境与发布',
      support: '办公支持',
    },
    catalog: {
      accountRoleGrant: {
        title: '账号角色变更',
        description: '为已有系统账号申请新角色或更高权限。',
      },
    },
    form: {},
    ticket: {
      list: {
        eyebrow: '服务门户',
        title: '我的工单',
        description: '当前浏览器提交的全部服务申请。',
        searchPlaceholder: '按标题或工单号搜索',
        filterLabel: '按状态过滤',
        empty: '当前过滤条件下没有工单。',
        columnId: '工单号',
        columnTitle: '标题',
        columnStatus: '状态',
        columnCreated: '提交时间',
      },
      detail: {
        eyebrow: '服务门户',
        applicantLabel: '申请人',
        serviceLabel: '服务',
        genericServiceLabel: '通用申请',
        formValuesLabel: '提交内容',
        timelineLabel: '处理记录',
      },
      myAside: {
        title: '我的最近工单',
        description: '当前浏览器提交的最新服务申请。',
        viewAll: '查看全部 →',
        empty: '暂无工单，从目录中选一项服务开始申请。',
      },
      status: {
        submitted: '已提交',
        approving: '审批中',
        processing: '处理中',
        done: '已完成',
        rejected: '已驳回',
      },
      event: {
        submitted: '{applicant} 提交了工单。',
        autoAdvanced: '状态自动推进至 {status}。',
      },
    },
    sla: {
      oneToTwoBusinessDays: '1-2 个工作日',
      sameDay: '当日响应',
      threeBusinessDays: '不超过 3 个工作日',
    },
    tag: {
      approval: '需审批',
      oneToTwoDays: '1-2 天',
      projectScope: '项目范围',
      costReview: '成本确认',
      securityReview: '安全审核',
      frequent: '高频',
      fastResponse: '快速响应',
      trackable: '可追踪',
    },
  },
  system: { ... },
};
```

注意：`catalog.accountRoleGrant` 与 `form.*` 在本任务里只放第一条占位 / 空对象；其他 12 条 catalog 在 Task 4 一起补，form 在 Task 5 补。

- [ ] **Step 2: 改 `ServicePortalPage.tsx` 里所有 `t('dashboard.servicePortal.*')` 调用**

打开 `web/src/view/service/ServicePortalPage.tsx`，把所有 `t('dashboard.servicePortal.xxx')` 形式的字符串替换为对应的 `servicePortal.xxx` 路径。Step 4 的全量改造会把这个文件完全重写，这一步只是为了让中间编译能通过 —— 用 `sed`：

```bash
cd web/src/view/service
# 把现有所有 dashboard.servicePortal 调用替换为 servicePortal
# 注意：分类 tab key 也要映射：
#   dashboard.servicePortal.categoryPopular  -> servicePortal.category.popular
#   dashboard.servicePortal.categoryAccount  -> servicePortal.category.account
#   ... 等等
# 简单 sed 不能完成，需要手动编辑。最终所有 t() 调用必须指向新结构。
```

实际操作：手动编辑 `ServicePortalPage.tsx`，按映射表替换。映射表见 spec §7.1。这一步只是过渡，Task 11 会重写整个 page。

- [ ] **Step 3: 跑构建确认 TS / Vite 不报错**

Run: `cd web && npm run build`
Expected: 退出码 0，`tsc -b && vite build` 全部成功。

- [ ] **Step 4: Commit**

```bash
git add web/src/locale/messages.ts web/src/view/service/ServicePortalPage.tsx
git commit -m "refactor(service-portal): 迁移 i18n 命名空间至 servicePortal.* 并新增工单相关 key"
```

---

## Task 2: 补全 13 条 catalog 的 i18n key（en + zhCN）

**Files:**

- Modify: `web/src/locale/messages.ts`（en.servicePortal.catalog 与 zhCN.servicePortal.catalog）

13 条 catalog 与对应分类（与 Task 4 的数据一致）：

| serviceId | category | 中文 title | 中文 description |
| --- | --- | --- | --- |
| `account.role-grant` | account | 账号角色变更 | 为已有系统账号申请新角色或更高权限。 |
| `account.new` | account | 新建系统账号 | 为新员工申请系统账号并初始化基础权限。 |
| `account.offboarding` | account | 离职账号回收 | 在员工离职前后申请账号停用、权限回收与资产移交。 |
| `cloud.compute-resize` | cloud | 云主机规格调整 | 申请主机规格升降配，含成本评估。 |
| `cloud.storage-quota` | cloud | 存储配额扩容 | 申请存储配额或对象存储桶配额扩容。 |
| `cloud.new-instance` | cloud | 新建云主机 | 申请新购云主机并自动归属到指定项目。 |
| `cloud.release` | cloud | 资源释放回收 | 申请释放闲置资源并完成成本归账。 |
| `network.vpn-access` | network | VPN 访问开通 | 申请远程 VPN 接入或临时访问策略。 |
| `network.acl` | network | 防火墙策略调整 | 申请新增或修改安全组 / ACL 策略。 |
| `network.bastion` | network | 堡垒机白名单 | 申请堡垒机白名单或临时 SSH 访问。 |
| `env.deploy-window` | env | 申请发布窗口 | 申请生产或预发环境的发布窗口与变更评审。 |
| `support.workplace` | support | 办公技术支持 | 提交办公网络、设备或外设故障申请。 |
| `support.incident` | support | 故障/事件上报 | 提交系统异常、登录失败、资源不可用等技术支持。 |

`hot: true` 的 4 条：`account.role-grant`、`cloud.compute-resize`、`network.vpn-access`、`support.incident`。

- [ ] **Step 1: 在 `en.servicePortal.catalog` 补全 13 条**

每条形如：

```ts
catalog: {
  accountRoleGrant: {
    title: 'Account role grant',
    description: 'Request a new role or elevated permission on an existing system account.',
  },
  accountNew: {
    title: 'New system account',
    description: 'Provision a system account for a new employee with baseline permissions.',
  },
  accountOffboarding: {
    title: 'Offboarding cleanup',
    description: 'Disable the account, revoke permissions and hand over assets when an employee leaves.',
  },
  cloudComputeResize: {
    title: 'Resize compute',
    description: 'Resize compute instance up or down, with cost review.',
  },
  cloudStorageQuota: {
    title: 'Storage quota expansion',
    description: 'Request additional block or object storage quota.',
  },
  cloudNewInstance: {
    title: 'New compute instance',
    description: 'Provision a new compute instance and bind it to a project.',
  },
  cloudRelease: {
    title: 'Resource release',
    description: 'Release idle resources and finalise the cost accounting.',
  },
  networkVpnAccess: {
    title: 'VPN access',
    description: 'Request remote VPN access or a temporary access policy.',
  },
  networkAcl: {
    title: 'Firewall policy change',
    description: 'Add or modify security-group / ACL policies.',
  },
  networkBastion: {
    title: 'Bastion allowlist',
    description: 'Request bastion allowlist or temporary SSH access.',
  },
  envDeployWindow: {
    title: 'Deployment window',
    description: 'Request a production or staging deployment window and change review.',
  },
  supportWorkplace: {
    title: 'Workplace support',
    description: 'Submit office network, device or peripheral issues.',
  },
  supportIncident: {
    title: 'Incident / outage',
    description: 'Submit system errors, login failures or unavailable resources.',
  },
},
```

- [ ] **Step 2: 在 `zhCN.servicePortal.catalog` 镜像 13 条**

按上表中文 title / description 补齐，结构与 en 完全一致。

- [ ] **Step 3: 构建验证**

Run: `cd web && npm run build`
Expected: 成功（zhCN 结构与 en 不一致会被 `satisfies Record<AppLocale, MessageDictionary>` 在 TS 阶段报错；通过即说明结构对齐）。

- [ ] **Step 4: Commit**

```bash
git add web/src/locale/messages.ts
git commit -m "feat(service-portal): 补齐 13 条服务目录的中英文文案"
```

---

## Task 3: 共享类型 `ticketTypes.ts`

**Files:**

- Create: `web/src/view/service/ticketTypes.ts`

- [ ] **Step 1: 创建文件**

```ts
// web/src/view/service/ticketTypes.ts

export type ServiceCategory = 'account' | 'cloud' | 'network' | 'env' | 'support';

export interface ServiceCatalogItem {
  id: string;
  category: ServiceCategory;
  i18nKey: string; // e.g. 'servicePortal.catalog.accountRoleGrant'
  iconToken: 'shield' | 'cloud' | 'network' | 'env' | 'support';
  tags: string[]; // i18n leaf keys, e.g. 'servicePortal.tag.approval'
  hot?: boolean;
  estimatedSla: string; // i18n leaf key
}

export type FieldType = 'text' | 'textarea' | 'select' | 'date' | 'multiselect';

export interface RequestFormFieldOption {
  value: string;
  i18nLabelKey: string;
}

export interface RequestFormField {
  name: string;
  type: FieldType;
  i18nLabelKey: string;
  required: boolean;
  options?: RequestFormFieldOption[];
  placeholderI18nKey?: string;
  maxLength?: number;
}

export type ServiceFormSchemas = Record<string, RequestFormField[]>;
export const GENERIC_FORM_KEY = '__generic';

export type TicketStatus = 'submitted' | 'approving' | 'processing' | 'done' | 'rejected';

export interface TicketEvent {
  at: string; // ISO timestamp
  status: TicketStatus;
  i18nNoteKey?: string;
}

export interface Ticket {
  id: string; // 'TKT-YYYY-NNNN'
  serviceId: string | null;
  title: string;
  formValues: Record<string, unknown>;
  status: TicketStatus;
  createdAt: string; // ISO
  timeline: TicketEvent[];
  applicant: { name: string; email: string };
}

export type StorageStatus = 'persisted' | 'memory-only';

export interface CreateTicketInput {
  serviceId: string | null;
  title: string;
  formValues: Record<string, unknown>;
  applicant: { name: string; email: string };
}

export interface CreateTicketResult {
  ticket: Ticket;
  storageStatus: StorageStatus;
}
```

- [ ] **Step 2: 构建验证**

Run: `cd web && npm run build`
Expected: 成功。

- [ ] **Step 3: Commit**

```bash
git add web/src/view/service/ticketTypes.ts
git commit -m "feat(service-portal): 新增工单与服务目录共享类型"
```

---

## Task 4: 静态服务目录 `data/serviceCatalog.ts`

**Files:**

- Create: `web/src/view/service/data/serviceCatalog.ts`

- [ ] **Step 1: 创建文件**

```ts
// web/src/view/service/data/serviceCatalog.ts

import type { ServiceCategory, ServiceCatalogItem } from '../ticketTypes';

export const SERVICE_CATEGORIES: ServiceCategory[] = ['account', 'cloud', 'network', 'env', 'support'];

export const CATEGORY_I18N_KEYS: Record<ServiceCategory | 'popular', string> = {
  popular: 'servicePortal.category.popular',
  account: 'servicePortal.category.account',
  cloud: 'servicePortal.category.cloud',
  network: 'servicePortal.category.network',
  env: 'servicePortal.category.env',
  support: 'servicePortal.category.support',
};

export const SERVICE_CATALOG: ServiceCatalogItem[] = [
  {
    id: 'account.role-grant',
    category: 'account',
    i18nKey: 'servicePortal.catalog.accountRoleGrant',
    iconToken: 'shield',
    tags: ['servicePortal.tag.approval', 'servicePortal.tag.oneToTwoDays'],
    hot: true,
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
  {
    id: 'account.new',
    category: 'account',
    i18nKey: 'servicePortal.catalog.accountNew',
    iconToken: 'shield',
    tags: ['servicePortal.tag.approval', 'servicePortal.tag.trackable'],
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
  {
    id: 'account.offboarding',
    category: 'account',
    i18nKey: 'servicePortal.catalog.accountOffboarding',
    iconToken: 'shield',
    tags: ['servicePortal.tag.approval', 'servicePortal.tag.trackable'],
    estimatedSla: 'servicePortal.sla.threeBusinessDays',
  },
  {
    id: 'cloud.compute-resize',
    category: 'cloud',
    i18nKey: 'servicePortal.catalog.cloudComputeResize',
    iconToken: 'cloud',
    tags: ['servicePortal.tag.projectScope', 'servicePortal.tag.costReview'],
    hot: true,
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
  {
    id: 'cloud.storage-quota',
    category: 'cloud',
    i18nKey: 'servicePortal.catalog.cloudStorageQuota',
    iconToken: 'cloud',
    tags: ['servicePortal.tag.costReview'],
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
  {
    id: 'cloud.new-instance',
    category: 'cloud',
    i18nKey: 'servicePortal.catalog.cloudNewInstance',
    iconToken: 'cloud',
    tags: ['servicePortal.tag.projectScope', 'servicePortal.tag.costReview'],
    estimatedSla: 'servicePortal.sla.threeBusinessDays',
  },
  {
    id: 'cloud.release',
    category: 'cloud',
    i18nKey: 'servicePortal.catalog.cloudRelease',
    iconToken: 'cloud',
    tags: ['servicePortal.tag.approval'],
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
  {
    id: 'network.vpn-access',
    category: 'network',
    i18nKey: 'servicePortal.catalog.networkVpnAccess',
    iconToken: 'network',
    tags: ['servicePortal.tag.securityReview', 'servicePortal.tag.frequent'],
    hot: true,
    estimatedSla: 'servicePortal.sla.sameDay',
  },
  {
    id: 'network.acl',
    category: 'network',
    i18nKey: 'servicePortal.catalog.networkAcl',
    iconToken: 'network',
    tags: ['servicePortal.tag.securityReview'],
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
  {
    id: 'network.bastion',
    category: 'network',
    i18nKey: 'servicePortal.catalog.networkBastion',
    iconToken: 'network',
    tags: ['servicePortal.tag.securityReview', 'servicePortal.tag.trackable'],
    estimatedSla: 'servicePortal.sla.sameDay',
  },
  {
    id: 'env.deploy-window',
    category: 'env',
    i18nKey: 'servicePortal.catalog.envDeployWindow',
    iconToken: 'env',
    tags: ['servicePortal.tag.approval', 'servicePortal.tag.trackable'],
    estimatedSla: 'servicePortal.sla.oneToTwoBusinessDays',
  },
  {
    id: 'support.workplace',
    category: 'support',
    i18nKey: 'servicePortal.catalog.supportWorkplace',
    iconToken: 'support',
    tags: ['servicePortal.tag.fastResponse'],
    estimatedSla: 'servicePortal.sla.sameDay',
  },
  {
    id: 'support.incident',
    category: 'support',
    i18nKey: 'servicePortal.catalog.supportIncident',
    iconToken: 'support',
    tags: ['servicePortal.tag.fastResponse', 'servicePortal.tag.trackable'],
    hot: true,
    estimatedSla: 'servicePortal.sla.sameDay',
  },
];

export function getCatalogById(id: string): ServiceCatalogItem | undefined {
  return SERVICE_CATALOG.find((item) => item.id === id);
}

export function listByCategory(category: ServiceCategory | 'popular'): ServiceCatalogItem[] {
  if (category === 'popular') {
    return SERVICE_CATALOG.filter((item) => item.hot);
  }
  return SERVICE_CATALOG.filter((item) => item.category === category);
}
```

- [ ] **Step 2: 构建验证**

Run: `cd web && npm run build`
Expected: 成功。

- [ ] **Step 3: Commit**

```bash
git add web/src/view/service/data/serviceCatalog.ts
git commit -m "feat(service-portal): 新增 13 条静态服务目录数据"
```

---

## Task 5: 表单 schema 字典 `data/serviceFormSchemas.ts` 及 i18n form key

**Files:**

- Create: `web/src/view/service/data/serviceFormSchemas.ts`
- Modify: `web/src/locale/messages.ts`（en.servicePortal.form 与 zhCN.servicePortal.form）

- [ ] **Step 1: 补全 `form.*` i18n key**

为简化，所有服务共享一个字段命名集（form key 是字段名，不按 serviceId 拆）。在 `en.servicePortal.form` 内添加：

```ts
form: {
  title: { label: 'Title', placeholder: 'Brief title of your request' },
  reason: { label: 'Reason', placeholder: 'Why is this needed?' },
  projectCode: {
    label: 'Project',
    placeholder: 'Select project',
    options: { web: 'Web platform', data: 'Data platform', mobile: 'Mobile platform' },
  },
  role: {
    label: 'Role',
    placeholder: 'Select role',
    options: { developer: 'Developer', ops: 'Ops', admin: 'Administrator', viewer: 'Viewer' },
  },
  effectiveAt: { label: 'Effective date', placeholder: '' },
  instanceSpec: {
    label: 'Target spec',
    placeholder: 'e.g. 4C8G',
  },
  quotaGb: { label: 'New quota (GB)', placeholder: 'Number of GB' },
  reasonAttachment: {
    label: 'Attachment description',
    placeholder: 'File name, link or external ticket reference',
  },
  duration: {
    label: 'Duration',
    placeholder: 'Select duration',
    options: { temporary: 'Temporary (≤7 days)', longTerm: 'Long-term' },
  },
  changeWindow: { label: 'Change window', placeholder: 'YYYY-MM-DD' },
  contactPhone: { label: 'Contact phone', placeholder: 'Optional' },
},
```

zhCN 镜像翻译。

- [ ] **Step 2: 创建 schema 字典**

```ts
// web/src/view/service/data/serviceFormSchemas.ts

import type { RequestFormField, ServiceFormSchemas } from '../ticketTypes';
import { GENERIC_FORM_KEY } from '../ticketTypes';

const accountRoleGrant: RequestFormField[] = [
  {
    name: 'projectCode',
    type: 'select',
    i18nLabelKey: 'servicePortal.form.projectCode.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.projectCode.placeholder',
    options: [
      { value: 'web', i18nLabelKey: 'servicePortal.form.projectCode.options.web' },
      { value: 'data', i18nLabelKey: 'servicePortal.form.projectCode.options.data' },
      { value: 'mobile', i18nLabelKey: 'servicePortal.form.projectCode.options.mobile' },
    ],
  },
  {
    name: 'role',
    type: 'select',
    i18nLabelKey: 'servicePortal.form.role.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.role.placeholder',
    options: [
      { value: 'developer', i18nLabelKey: 'servicePortal.form.role.options.developer' },
      { value: 'ops', i18nLabelKey: 'servicePortal.form.role.options.ops' },
      { value: 'admin', i18nLabelKey: 'servicePortal.form.role.options.admin' },
      { value: 'viewer', i18nLabelKey: 'servicePortal.form.role.options.viewer' },
    ],
  },
  {
    name: 'effectiveAt',
    type: 'date',
    i18nLabelKey: 'servicePortal.form.effectiveAt.label',
    required: false,
  },
  {
    name: 'reason',
    type: 'textarea',
    i18nLabelKey: 'servicePortal.form.reason.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.reason.placeholder',
    maxLength: 500,
  },
];

const cloudComputeResize: RequestFormField[] = [
  {
    name: 'projectCode',
    type: 'select',
    i18nLabelKey: 'servicePortal.form.projectCode.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.projectCode.placeholder',
    options: [
      { value: 'web', i18nLabelKey: 'servicePortal.form.projectCode.options.web' },
      { value: 'data', i18nLabelKey: 'servicePortal.form.projectCode.options.data' },
      { value: 'mobile', i18nLabelKey: 'servicePortal.form.projectCode.options.mobile' },
    ],
  },
  {
    name: 'instanceSpec',
    type: 'text',
    i18nLabelKey: 'servicePortal.form.instanceSpec.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.instanceSpec.placeholder',
    maxLength: 60,
  },
  {
    name: 'reason',
    type: 'textarea',
    i18nLabelKey: 'servicePortal.form.reason.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.reason.placeholder',
    maxLength: 500,
  },
];

const networkVpnAccess: RequestFormField[] = [
  {
    name: 'duration',
    type: 'select',
    i18nLabelKey: 'servicePortal.form.duration.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.duration.placeholder',
    options: [
      { value: 'temporary', i18nLabelKey: 'servicePortal.form.duration.options.temporary' },
      { value: 'longTerm', i18nLabelKey: 'servicePortal.form.duration.options.longTerm' },
    ],
  },
  {
    name: 'reason',
    type: 'textarea',
    i18nLabelKey: 'servicePortal.form.reason.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.reason.placeholder',
    maxLength: 500,
  },
];

const generic: RequestFormField[] = [
  {
    name: 'title',
    type: 'text',
    i18nLabelKey: 'servicePortal.form.title.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.title.placeholder',
    maxLength: 80,
  },
  {
    name: 'reason',
    type: 'textarea',
    i18nLabelKey: 'servicePortal.form.reason.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.reason.placeholder',
    maxLength: 500,
  },
  {
    name: 'contactPhone',
    type: 'text',
    i18nLabelKey: 'servicePortal.form.contactPhone.label',
    required: false,
    placeholderI18nKey: 'servicePortal.form.contactPhone.placeholder',
    maxLength: 40,
  },
];

// 默认 schema：含 reason + reasonAttachment，用于未显式定义自己 schema 的服务
const defaultSchema: RequestFormField[] = [
  {
    name: 'reason',
    type: 'textarea',
    i18nLabelKey: 'servicePortal.form.reason.label',
    required: true,
    placeholderI18nKey: 'servicePortal.form.reason.placeholder',
    maxLength: 500,
  },
  {
    name: 'reasonAttachment',
    type: 'textarea',
    i18nLabelKey: 'servicePortal.form.reasonAttachment.label',
    required: false,
    placeholderI18nKey: 'servicePortal.form.reasonAttachment.placeholder',
    maxLength: 300,
  },
];

export const SERVICE_FORM_SCHEMAS: ServiceFormSchemas = {
  [GENERIC_FORM_KEY]: generic,
  'account.role-grant': accountRoleGrant,
  'cloud.compute-resize': cloudComputeResize,
  'network.vpn-access': networkVpnAccess,
};

export function getFormSchema(serviceId: string | null): RequestFormField[] {
  if (!serviceId) {
    return SERVICE_FORM_SCHEMAS[GENERIC_FORM_KEY];
  }
  return SERVICE_FORM_SCHEMAS[serviceId] ?? defaultSchema;
}
```

注意：spec 没要求 13 条都各有定制 schema。只为 3 条热门服务定制；其它走 `defaultSchema`（reason + 附件描述）—— YAGNI 原则。

- [ ] **Step 3: 构建验证**

Run: `cd web && npm run build`
Expected: 成功。zhCN 与 en 不同构会在 TS 阶段报错；通过即对齐。

- [ ] **Step 4: Commit**

```bash
git add web/src/view/service/data/serviceFormSchemas.ts web/src/locale/messages.ts
git commit -m "feat(service-portal): 新增表单字段字典与对应 i18n"
```

---

## Task 6: 模块级 `ticketStore` 与 `useTicketStore`

**Files:**

- Create: `web/src/view/service/data/ticketStore.ts`

- [ ] **Step 1: 创建文件**

```ts
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
```

注意：`useSyncExternalStore` 的 `getSnapshot` 必须返回**引用稳定**的值。`ticketStore.list()` 始终返回内部 `tickets` 引用，且只有在 `notify` 之前才会被替换为新引用 —— 满足 React 18 的要求。

- [ ] **Step 2: 构建验证**

Run: `cd web && npm run build`
Expected: 成功。

- [ ] **Step 3: Commit**

```bash
git add web/src/view/service/data/ticketStore.ts
git commit -m "feat(service-portal): 新增 localStorage 工单 store 与 hook"
```

---

## Task 7: 状态机 `ticketMockEngine`

**Files:**

- Create: `web/src/view/service/data/ticketMockEngine.ts`

- [ ] **Step 1: 创建文件**

```ts
// web/src/view/service/data/ticketMockEngine.ts

import { ticketStore } from './ticketStore';

const STEP_5_MS = 5_000;
const STEP_30_MS = 30_000;
const STEP_60_MS = 60_000;

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
    ids.push(
      window.setTimeout(() => {
        ticketStore.advance(ticketId, 'processing', 'done');
      }, STEP_60_MS),
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
```

注意：spec §6.3 明确"页面卸载不清 timer / 页面刷新后状态原地停摆"。因此本模块**不**在任何 React 生命周期里注册 cleanup，也**不**在模块加载时为已有工单 reschedule。仅在 `createTicket` 后由调用方显式 `schedule(id)`。

- [ ] **Step 2: 构建验证**

Run: `cd web && npm run build`
Expected: 成功。

- [ ] **Step 3: Commit**

```bash
git add web/src/view/service/data/ticketMockEngine.ts
git commit -m "feat(service-portal): 新增 setTimeout 工单状态机"
```

---

## Task 8: `TicketStatusBadge` 组件

**Files:**

- Create: `web/src/view/service/components/TicketStatusBadge.tsx`
- Create: `web/src/view/service/components/TicketStatusBadge.module.css`

- [ ] **Step 1: 创建 css**

```css
/* TicketStatusBadge.module.css */

.badge {
  display: inline-flex;
  align-items: center;
  min-height: 1.5rem;
  padding: 0 0.55rem;
  border-radius: 999px;
  font-size: 0.7rem;
  font-weight: 850;
  letter-spacing: 0.01em;
  white-space: nowrap;
}

.submitted {
  background: rgba(118, 105, 93, 0.14);
  color: #5a4a3a;
}

.approving {
  background: rgba(147, 93, 45, 0.14);
  color: #935d2d;
}

.processing {
  background: rgba(54, 93, 126, 0.14);
  color: #365d7e;
}

.done {
  background: rgba(47, 98, 92, 0.16);
  color: #2f625c;
}

.rejected {
  background: rgba(180, 60, 60, 0.14);
  color: #a23838;
}
```

- [ ] **Step 2: 创建组件**

```tsx
// TicketStatusBadge.tsx

import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { TicketStatus } from '../ticketTypes';
import styles from './TicketStatusBadge.module.css';

interface Props {
  status: TicketStatus;
}

const STATUS_I18N: Record<TicketStatus, MessageKey> = {
  submitted: 'servicePortal.ticket.status.submitted',
  approving: 'servicePortal.ticket.status.approving',
  processing: 'servicePortal.ticket.status.processing',
  done: 'servicePortal.ticket.status.done',
  rejected: 'servicePortal.ticket.status.rejected',
};

export function TicketStatusBadge({ status }: Props) {
  const { t } = useI18n();
  return <span className={`${styles.badge} ${styles[status]}`}>{t(STATUS_I18N[status])}</span>;
}
```

注意：`MessageKey` 来自 `messages.ts` 顶层 export，需要 `import type { MessageKey } from '../../../locale/messages'`。

- [ ] **Step 3: 构建验证**

Run: `cd web && npm run build`
Expected: 成功。

- [ ] **Step 4: Commit**

```bash
git add web/src/view/service/components/TicketStatusBadge.tsx web/src/view/service/components/TicketStatusBadge.module.css
git commit -m "feat(service-portal): 新增工单状态徽章组件"
```

---

## Task 9: `CategoryTabs` 与 `ServiceCard` 抽出

**Files:**

- Create: `web/src/view/service/components/CategoryTabs.tsx`
- Create: `web/src/view/service/components/CategoryTabs.module.css`
- Create: `web/src/view/service/components/ServiceCard.tsx`
- Create: `web/src/view/service/components/ServiceCard.module.css`

- [ ] **Step 1: 创建 CategoryTabs**

```css
/* CategoryTabs.module.css */
.tabs {
  display: flex;
  gap: 0.55rem;
  overflow-x: auto;
  padding-bottom: 0.1rem;
}

.tab {
  flex: 0 0 auto;
  min-height: 2.35rem;
  padding: 0 0.78rem;
  border: 1px solid rgba(86, 70, 55, 0.13);
  border-radius: 0.72rem;
  background: rgba(255, 255, 255, 0.62);
  color: #24201c;
  font: inherit;
  font-size: 0.84rem;
  font-weight: 760;
  cursor: pointer;
  white-space: nowrap;
}

.tabActive {
  border-color: #24201c;
  background: #24201c;
  color: #fffdf9;
}
```

```tsx
// CategoryTabs.tsx
import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { ServiceCategory } from '../ticketTypes';
import { CATEGORY_I18N_KEYS } from '../data/serviceCatalog';
import styles from './CategoryTabs.module.css';

export type CategoryFilter = ServiceCategory | 'popular';

interface Props {
  active: CategoryFilter;
  onChange: (next: CategoryFilter) => void;
}

const ORDER: CategoryFilter[] = ['popular', 'account', 'cloud', 'network', 'env', 'support'];

export function CategoryTabs({ active, onChange }: Props) {
  const { t } = useI18n();
  return (
    <div className={styles.tabs} role="tablist" aria-label={t('servicePortal.category.label')}>
      {ORDER.map((category) => {
        const isActive = category === active;
        return (
          <button
            key={category}
            type="button"
            role="tab"
            aria-selected={isActive}
            className={isActive ? `${styles.tab} ${styles.tabActive}` : styles.tab}
            onClick={() => onChange(category)}
          >
            {t(CATEGORY_I18N_KEYS[category] as MessageKey)}
          </button>
        );
      })}
    </div>
  );
}
```

- [ ] **Step 2: 创建 ServiceCard**

```css
/* ServiceCard.module.css */
.card {
  display: grid;
  align-content: start;
  gap: 0.65rem;
  min-height: 11rem;
  padding: 0.95rem;
  border: 1px solid rgba(86, 70, 55, 0.1);
  border-radius: 1rem;
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.66), rgba(246, 238, 228, 0.44));
  text-align: left;
  font: inherit;
  cursor: pointer;
  transition: transform 120ms ease, box-shadow 120ms ease;
}

.card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 18px rgba(47, 98, 92, 0.08);
}

.icon {
  width: 2.15rem;
  height: 2.15rem;
  border: 1px solid rgba(47, 98, 92, 0.13);
  border-radius: 0.78rem;
  background: rgba(47, 98, 92, 0.12);
}

.title {
  margin: 0;
  color: #24201c;
  font-size: 0.98rem;
  font-weight: 700;
  line-height: 1.3;
}

.description {
  margin: 0;
  color: #76695d;
  font-size: 0.78rem;
  line-height: 1.55;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.38rem;
}

.tag {
  display: inline-flex;
  align-items: center;
  min-height: 1.45rem;
  padding: 0 0.48rem;
  border-radius: 999px;
  background: rgba(47, 98, 92, 0.12);
  color: #2f625c;
  font-size: 0.7rem;
  font-weight: 800;
}
```

```tsx
// ServiceCard.tsx
import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { ServiceCatalogItem } from '../ticketTypes';
import styles from './ServiceCard.module.css';

interface Props {
  item: ServiceCatalogItem;
  onSelect: (item: ServiceCatalogItem) => void;
}

export function ServiceCard({ item, onSelect }: Props) {
  const { t } = useI18n();
  const titleKey = `${item.i18nKey}.title` as MessageKey;
  const descKey = `${item.i18nKey}.description` as MessageKey;
  return (
    <button type="button" className={styles.card} onClick={() => onSelect(item)}>
      <span className={styles.icon} aria-hidden="true" />
      <h3 className={styles.title}>{t(titleKey)}</h3>
      <p className={styles.description}>{t(descKey)}</p>
      <div className={styles.tags}>
        {item.tags.map((tag) => (
          <span key={tag} className={styles.tag}>
            {t(tag as MessageKey)}
          </span>
        ))}
      </div>
    </button>
  );
}
```

- [ ] **Step 3: 构建验证**

Run: `cd web && npm run build`
Expected: 成功。

- [ ] **Step 4: Commit**

```bash
git add web/src/view/service/components/CategoryTabs.tsx web/src/view/service/components/CategoryTabs.module.css web/src/view/service/components/ServiceCard.tsx web/src/view/service/components/ServiceCard.module.css
git commit -m "feat(service-portal): 抽出分类 tab 与服务卡组件"
```

---

## Task 10: `MyTicketsAside` 组件

**Files:**

- Create: `web/src/view/service/components/MyTicketsAside.tsx`
- Create: `web/src/view/service/components/MyTicketsAside.module.css`

- [ ] **Step 1: 创建 css**

```css
/* MyTicketsAside.module.css */
.aside {
  display: grid;
  align-content: start;
  gap: 0.85rem;
  padding: clamp(1rem, 2.2vw, 1.25rem);
  border-left: 1px solid rgba(86, 70, 55, 0.13);
  background: rgba(246, 238, 228, 0.42);
}

.heading {
  margin: 0;
  color: #24201c;
  font-size: 1rem;
  font-weight: 700;
}

.description {
  margin: 0;
  color: #76695d;
  font-size: 0.78rem;
  line-height: 1.55;
}

.list {
  display: grid;
  gap: 0.62rem;
}

.item {
  display: grid;
  gap: 0.5rem;
  padding: 0.82rem;
  border: 1px solid rgba(86, 70, 55, 0.1);
  border-radius: 0.9rem;
  background: rgba(255, 255, 255, 0.58);
  text-decoration: none;
  color: inherit;
}

.topline {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.7rem;
}

.title {
  min-width: 0;
  color: #24201c;
  font-size: 0.84rem;
  line-height: 1.35;
  font-weight: 700;
}

.meta {
  margin: 0;
  color: #76695d;
  font-size: 0.72rem;
}

.empty {
  margin: 0;
  color: #76695d;
  font-size: 0.8rem;
  line-height: 1.5;
}

.viewAll {
  align-self: end;
  font-size: 0.82rem;
  font-weight: 700;
  color: #2f625c;
  text-decoration: none;
}
```

- [ ] **Step 2: 创建组件**

```tsx
// MyTicketsAside.tsx
import { Link } from 'react-router-dom';

import { useI18n } from '../../../locale/useI18n';
import { useTicketStore } from '../data/ticketStore';
import { TicketStatusBadge } from './TicketStatusBadge';
import styles from './MyTicketsAside.module.css';

const PREVIEW_COUNT = 5;

export function MyTicketsAside() {
  const { t } = useI18n();
  const { tickets } = useTicketStore();
  const preview = tickets.slice(0, PREVIEW_COUNT);

  return (
    <aside className={styles.aside}>
      <h3 className={styles.heading}>{t('servicePortal.ticket.myAside.title')}</h3>
      <p className={styles.description}>{t('servicePortal.ticket.myAside.description')}</p>

      {preview.length === 0 ? (
        <p className={styles.empty}>{t('servicePortal.ticket.myAside.empty')}</p>
      ) : (
        <div className={styles.list}>
          {preview.map((ticket) => (
            <Link key={ticket.id} to={`/service-portal/tickets/${ticket.id}`} className={styles.item}>
              <div className={styles.topline}>
                <strong className={styles.title}>{ticket.title}</strong>
                <TicketStatusBadge status={ticket.status} />
              </div>
              <p className={styles.meta}>
                {ticket.id} · {new Date(ticket.createdAt).toLocaleString()}
              </p>
            </Link>
          ))}
        </div>
      )}

      <Link to="/service-portal/tickets" className={styles.viewAll}>
        {t('servicePortal.ticket.myAside.viewAll')}
      </Link>
    </aside>
  );
}
```

- [ ] **Step 3: 构建验证**

Run: `cd web && npm run build`
Expected: 成功。

- [ ] **Step 4: Commit**

```bash
git add web/src/view/service/components/MyTicketsAside.tsx web/src/view/service/components/MyTicketsAside.module.css
git commit -m "feat(service-portal): 新增「我的工单」侧栏组件"
```

---

## Task 11: `RequestDrawer` 组件（portal + ESC + 遮罩）

**Files:**

- Create: `web/src/view/service/components/RequestDrawer.tsx`
- Create: `web/src/view/service/components/RequestDrawer.module.css`

- [ ] **Step 1: 创建 css**

```css
/* RequestDrawer.module.css */
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  opacity: 0;
  transition: opacity 160ms ease;
  z-index: 1000;
  pointer-events: none;
}

.overlay[data-open='true'] {
  opacity: 1;
  pointer-events: auto;
}

.drawer {
  position: fixed;
  right: 0;
  top: 0;
  bottom: 0;
  width: 520px;
  max-width: 100vw;
  background: #fffdf9;
  box-shadow: -16px 0 32px rgba(36, 32, 28, 0.16);
  transform: translateX(100%);
  transition: transform 220ms ease;
  z-index: 1001;
  display: grid;
  grid-template-rows: auto 1fr;
}

.drawer[data-open='true'] {
  transform: translateX(0);
}

.header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.1rem 1.25rem 0.9rem;
  border-bottom: 1px solid rgba(86, 70, 55, 0.13);
}

.eyebrow {
  margin: 0 0 0.25rem;
  color: #76695d;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.title {
  margin: 0;
  color: #24201c;
  font-size: 1.1rem;
  font-weight: 700;
}

.description {
  margin: 0.35rem 0 0;
  color: #76695d;
  font-size: 0.84rem;
}

.close {
  border: 0;
  background: transparent;
  font-size: 1.4rem;
  cursor: pointer;
  color: #76695d;
}

.body {
  overflow-y: auto;
  padding: 1rem 1.25rem 1.25rem;
}
```

- [ ] **Step 2: 创建组件**

```tsx
// RequestDrawer.tsx
import { useEffect, type ReactNode } from 'react';
import { createPortal } from 'react-dom';

import styles from './RequestDrawer.module.css';

interface Props {
  open: boolean;
  onClose: () => void;
  eyebrow?: string;
  title: string;
  description?: string;
  children: ReactNode;
}

export function RequestDrawer({ open, onClose, eyebrow, title, description, children }: Props) {
  useEffect(() => {
    if (!open) return;
    function onKey(event: KeyboardEvent) {
      if (event.key === 'Escape') onClose();
    }
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open, onClose]);

  if (typeof document === 'undefined') return null;

  return createPortal(
    <>
      <div className={styles.overlay} data-open={open} onClick={onClose} aria-hidden="true" />
      <aside
        className={styles.drawer}
        data-open={open}
        role="dialog"
        aria-modal="true"
        aria-label={title}
      >
        <header className={styles.header}>
          <div>
            {eyebrow ? <p className={styles.eyebrow}>{eyebrow}</p> : null}
            <h2 className={styles.title}>{title}</h2>
            {description ? <p className={styles.description}>{description}</p> : null}
          </div>
          <button type="button" className={styles.close} onClick={onClose} aria-label="Close">
            ×
          </button>
        </header>
        <div className={styles.body}>{children}</div>
      </aside>
    </>,
    document.body,
  );
}
```

- [ ] **Step 3: 构建验证**

Run: `cd web && npm run build`
Expected: 成功。

- [ ] **Step 4: Commit**

```bash
git add web/src/view/service/components/RequestDrawer.tsx web/src/view/service/components/RequestDrawer.module.css
git commit -m "feat(service-portal): 新增右侧抽屉组件（portal + ESC）"
```

---

## Task 12: `RequestForm` 组件（schema 驱动 + 校验 + 提交）

**Files:**

- Create: `web/src/view/service/components/RequestForm.tsx`
- Create: `web/src/view/service/components/RequestForm.module.css`

- [ ] **Step 1: 创建 css**

```css
/* RequestForm.module.css */
.form {
  display: grid;
  gap: 1rem;
}

.banner {
  padding: 0.7rem 0.9rem;
  border: 1px solid rgba(180, 60, 60, 0.4);
  border-radius: 0.65rem;
  background: rgba(180, 60, 60, 0.08);
  color: #a23838;
  font-size: 0.82rem;
}

.field {
  display: grid;
  gap: 0.35rem;
}

.label {
  display: flex;
  justify-content: space-between;
  gap: 0.5rem;
  color: #24201c;
  font-size: 0.82rem;
  font-weight: 700;
}

.optional {
  color: #76695d;
  font-weight: 500;
  font-size: 0.72rem;
}

.input,
.textarea,
.select {
  font: inherit;
  font-size: 0.86rem;
  padding: 0.55rem 0.7rem;
  border: 1px solid rgba(86, 70, 55, 0.2);
  border-radius: 0.55rem;
  background: #fffdf9;
  color: #24201c;
}

.textarea {
  min-height: 5.5rem;
  resize: vertical;
}

.error {
  color: #a23838;
  font-size: 0.74rem;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.55rem;
  padding-top: 0.4rem;
}

.cancel,
.submit {
  min-height: 2.35rem;
  padding: 0 0.95rem;
  border: 0;
  border-radius: 0.65rem;
  font: inherit;
  font-size: 0.84rem;
  font-weight: 760;
  cursor: pointer;
}

.cancel {
  background: rgba(255, 255, 255, 0.6);
  border: 1px solid rgba(86, 70, 55, 0.2);
  color: #24201c;
}

.submit {
  background: #2f625c;
  color: #fffdf9;
}

.submit[disabled] {
  opacity: 0.5;
  cursor: not-allowed;
}
```

- [ ] **Step 2: 创建组件**

```tsx
// RequestForm.tsx
import { useState, type FormEvent } from 'react';

import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import { useTicketStore } from '../data/ticketStore';
import { ticketMockEngine } from '../data/ticketMockEngine';
import type { RequestFormField, StorageStatus } from '../ticketTypes';
import styles from './RequestForm.module.css';

interface Props {
  serviceId: string | null;
  serviceTitle: string;
  schema: RequestFormField[];
  applicant: { name: string; email: string };
  onSubmitted: () => void;
  onCancel: () => void;
}

type Values = Record<string, string | string[]>;

export function RequestForm({ serviceId, serviceTitle, schema, applicant, onSubmitted, onCancel }: Props) {
  const { t } = useI18n();
  const { createTicket } = useTicketStore();
  const [values, setValues] = useState<Values>(() => initValues(schema));
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [submitting, setSubmitting] = useState(false);
  const [storageBanner, setStorageBanner] = useState<StorageStatus | null>(null);

  function setField(name: string, value: string | string[]) {
    setValues((prev) => ({ ...prev, [name]: value }));
  }

  function validate(): boolean {
    const next: Record<string, string> = {};
    for (const field of schema) {
      const raw = values[field.name];
      const empty =
        raw == null ||
        (typeof raw === 'string' && raw.trim() === '') ||
        (Array.isArray(raw) && raw.length === 0);
      if (field.required && empty) {
        next[field.name] = t('servicePortal.validationRequired');
        continue;
      }
      if (typeof raw === 'string' && field.maxLength && raw.length > field.maxLength) {
        next[field.name] = t('servicePortal.validationTooLong').replace('{max}', String(field.maxLength));
      }
    }
    setErrors(next);
    return Object.keys(next).length === 0;
  }

  function handleSubmit(event: FormEvent) {
    event.preventDefault();
    if (submitting) return;
    if (!validate()) return;
    setSubmitting(true);
    const title = deriveTitle(serviceId, serviceTitle, values);
    const result = createTicket({ serviceId, title, formValues: values, applicant });
    ticketMockEngine.schedule(result.ticket.id);
    if (result.storageStatus === 'memory-only') {
      setStorageBanner('memory-only');
      // 表单保留已提交态，按钮 disabled 不变；不关 Drawer
      return;
    }
    onSubmitted();
  }

  return (
    <form className={styles.form} onSubmit={handleSubmit}>
      {storageBanner === 'memory-only' ? (
        <div className={styles.banner} role="alert">
          {t('servicePortal.storageFailed')}
        </div>
      ) : null}

      {schema.map((field) => (
        <div className={styles.field} key={field.name}>
          <label className={styles.label} htmlFor={`field-${field.name}`}>
            <span>{t(field.i18nLabelKey as MessageKey)}</span>
            {field.required ? null : <span className={styles.optional}>{t('servicePortal.optional')}</span>}
          </label>
          {renderField(field, values[field.name] ?? '', setField, t, `field-${field.name}`)}
          {errors[field.name] ? <span className={styles.error}>{errors[field.name]}</span> : null}
        </div>
      ))}

      <div className={styles.actions}>
        <button type="button" className={styles.cancel} onClick={onCancel} disabled={submitting}>
          {t('servicePortal.cancel')}
        </button>
        <button type="submit" className={styles.submit} disabled={submitting}>
          {submitting ? t('servicePortal.submitting') : t('servicePortal.submit')}
        </button>
      </div>
    </form>
  );
}

function initValues(schema: RequestFormField[]): Values {
  const result: Values = {};
  for (const field of schema) {
    result[field.name] = field.type === 'multiselect' ? [] : '';
  }
  return result;
}

function deriveTitle(serviceId: string | null, serviceTitle: string, values: Values): string {
  if (serviceId) return serviceTitle;
  const raw = values.title;
  if (typeof raw === 'string' && raw.trim()) return raw.trim();
  return serviceTitle; // fallback
}

function renderField(
  field: RequestFormField,
  value: string | string[],
  setField: (name: string, value: string | string[]) => void,
  t: (key: MessageKey) => string,
  inputId: string,
) {
  const placeholder = field.placeholderI18nKey ? t(field.placeholderI18nKey as MessageKey) : undefined;
  switch (field.type) {
    case 'textarea':
      return (
        <textarea
          id={inputId}
          className={styles.textarea}
          value={typeof value === 'string' ? value : ''}
          maxLength={field.maxLength}
          placeholder={placeholder}
          onChange={(e) => setField(field.name, e.target.value)}
        />
      );
    case 'select':
      return (
        <select
          id={inputId}
          className={styles.select}
          value={typeof value === 'string' ? value : ''}
          onChange={(e) => setField(field.name, e.target.value)}
        >
          <option value="">{placeholder ?? ''}</option>
          {field.options?.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {t(opt.i18nLabelKey as MessageKey)}
            </option>
          ))}
        </select>
      );
    case 'multiselect':
      return (
        <select
          id={inputId}
          multiple
          className={styles.select}
          value={Array.isArray(value) ? value : []}
          onChange={(e) => {
            const next = Array.from(e.target.selectedOptions).map((o) => o.value);
            setField(field.name, next);
          }}
        >
          {field.options?.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {t(opt.i18nLabelKey as MessageKey)}
            </option>
          ))}
        </select>
      );
    case 'date':
      return (
        <input
          id={inputId}
          type="date"
          className={styles.input}
          value={typeof value === 'string' ? value : ''}
          onChange={(e) => setField(field.name, e.target.value)}
        />
      );
    case 'text':
    default:
      return (
        <input
          id={inputId}
          type="text"
          className={styles.input}
          value={typeof value === 'string' ? value : ''}
          maxLength={field.maxLength}
          placeholder={placeholder}
          onChange={(e) => setField(field.name, e.target.value)}
        />
      );
  }
}
```

注意：spec §8.1 要求"提交按钮 onClick 后立即设 disabled"。本实现里通过 `setSubmitting(true)` 实现；当 `storageStatus === 'memory-only'` 时不重置 submitting，按钮保持 disabled —— 满足 spec "持久化失败时按钮保持 disabled"。

- [ ] **Step 3: 构建验证**

Run: `cd web && npm run build`
Expected: 成功。

- [ ] **Step 4: Commit**

```bash
git add web/src/view/service/components/RequestForm.tsx web/src/view/service/components/RequestForm.module.css
git commit -m "feat(service-portal): 新增 schema 驱动的申请表单组件"
```

---

## Task 13: 改造 `ServicePortalPage`（首页 = 目录 + 我的工单侧栏 + Drawer 入口）

**Files:**

- Modify (rewrite): `web/src/view/service/ServicePortalPage.tsx`
- Modify: `web/src/view/service/ServicePortalPage.module.css`（移除已被组件接管的 token；保留 panel / panelHeader / panelBody / actions）

- [ ] **Step 1: 调整 ServicePortalPage.module.css**

**保留**：`.page` / `.portalPanel` / `.panelHeader` / `.panelHeader h2` / `.panelHeader p` / `.actions` / `.primaryAction` / `.secondaryAction` / `.panelBody` / `.catalogArea`

**删除**（已被独立组件接管）：`.categoryTabs` / `.categoryTab` / `.categoryTabActive` / `.serviceGrid` / `.serviceCard` 系列 / `.serviceIcon` / `.serviceTags` / `.serviceTag` / `.ticketArea` 系列 / `.ticketList` / `.ticketItem` / `.ticketTopline` / `.ticketStatus` / `.ticketStatusWarning`

**调整**：`.panelBody` 的 grid-template-columns 改为 `minmax(0, 1fr) minmax(18rem, 0.36fr)`；保留 `@media (max-width: 980px)` 处的栈式回退。**新增** `.serviceGrid` 简化为 `display: grid; grid-template-columns: repeat(auto-fill, minmax(13.5rem, 1fr)); gap: 0.8rem; padding: clamp(1rem, 2.2vw, 1.25rem);` 以适配 13 张卡。

具体完整内容（取代现文件）：

```css
.page {
  --service-ink: #24201c;
  --service-muted: #76695d;
  --service-line: rgba(86, 70, 55, 0.13);
  --service-accent: #2f625c;
  --service-accent-soft: rgba(47, 98, 92, 0.12);
  display: grid;
  gap: 1.35rem;
}

.portalPanel {
  overflow: hidden;
  border: 1px solid var(--service-line);
  border-radius: 1.55rem;
  background: rgba(255, 253, 249, 0.74);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72);
}

.panelHeader {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  padding: clamp(1.1rem, 2.4vw, 1.45rem);
  border-bottom: 1px solid var(--service-line);
}

.panelHeader h2,
.panelHeader p {
  margin: 0;
}

.panelHeader h2 {
  color: var(--service-ink);
  font-size: clamp(1.25rem, 2vw, 1.65rem);
  font-weight: 680;
  line-height: 1.12;
}

.panelHeader p {
  max-width: 62ch;
  margin-top: 0.55rem;
  color: var(--service-muted);
  line-height: 1.65;
}

.actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 0.55rem;
}

.primaryAction,
.secondaryAction {
  min-height: 2.35rem;
  padding: 0 0.9rem;
  border: 0;
  border-radius: 0.72rem;
  font: inherit;
  font-size: 0.84rem;
  font-weight: 760;
  cursor: pointer;
  white-space: nowrap;
}

.primaryAction {
  background: var(--service-accent);
  color: #fffdf9;
  box-shadow: 0 10px 22px rgba(47, 98, 92, 0.14);
}

.secondaryAction {
  border: 1px solid rgba(86, 70, 55, 0.13);
  background: rgba(255, 255, 255, 0.62);
  color: var(--service-ink);
}

.panelBody {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(18rem, 0.36fr);
}

.catalogArea {
  display: grid;
  gap: 1rem;
  min-width: 0;
  padding: clamp(1rem, 2.2vw, 1.25rem);
}

.serviceGrid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(13.5rem, 1fr));
  gap: 0.8rem;
}

@media (max-width: 980px) {
  .panelBody {
    grid-template-columns: 1fr;
  }
}
```

- [ ] **Step 2: 重写 ServicePortalPage.tsx**

```tsx
// ServicePortalPage.tsx
import { useMemo, useState } from 'react';
import { Link } from 'react-router-dom';

import { useAuth } from '../../auth/AuthContext';
import { useI18n } from '../../locale/useI18n';
import type { MessageKey } from '../../locale/messages';
import { PageHeader } from '../../components/page/PageHeader';
import { CategoryTabs, type CategoryFilter } from './components/CategoryTabs';
import { MyTicketsAside } from './components/MyTicketsAside';
import { RequestDrawer } from './components/RequestDrawer';
import { RequestForm } from './components/RequestForm';
import { ServiceCard } from './components/ServiceCard';
import { SERVICE_CATALOG, listByCategory } from './data/serviceCatalog';
import { getFormSchema } from './data/serviceFormSchemas';
import type { ServiceCatalogItem } from './ticketTypes';
import styles from './ServicePortalPage.module.css';

type DrawerSelection =
  | { kind: 'service'; item: ServiceCatalogItem }
  | { kind: 'generic' }
  | null;

export function ServicePortalPage() {
  const { t } = useI18n();
  const { current } = useAuth();
  const [activeCategory, setActiveCategory] = useState<CategoryFilter>('popular');
  const [drawer, setDrawer] = useState<DrawerSelection>(null);

  const visibleItems = useMemo(() => listByCategory(activeCategory), [activeCategory]);

  const applicant = {
    name: current?.user.displayName ?? current?.user.username ?? '',
    email: current?.user.email ?? '',
  };

  function openService(item: ServiceCatalogItem) {
    setDrawer({ kind: 'service', item });
  }

  function openGeneric() {
    setDrawer({ kind: 'generic' });
  }

  function closeDrawer() {
    setDrawer(null);
  }

  const drawerProps = useMemo(() => {
    if (!drawer) return null;
    if (drawer.kind === 'service') {
      const titleKey = `${drawer.item.i18nKey}.title` as MessageKey;
      const descKey = `${drawer.item.i18nKey}.description` as MessageKey;
      return {
        serviceId: drawer.item.id,
        serviceTitle: t(titleKey),
        description: t(descKey),
        schema: getFormSchema(drawer.item.id),
      };
    }
    return {
      serviceId: null,
      serviceTitle: t('servicePortal.createRequest'),
      description: undefined as string | undefined,
      schema: getFormSchema(null),
    };
  }, [drawer, t]);

  return (
    <div className={styles.page}>
      <PageHeader
        eyebrow={t('servicePortal.eyebrow')}
        title={t('servicePortal.title')}
        description={t('servicePortal.description')}
      />

      <section className={styles.portalPanel}>
        <div className={styles.panelHeader}>
          <div>
            <h2>{t('servicePortal.title')}</h2>
            <p>{t('servicePortal.description')}</p>
          </div>
          <div className={styles.actions}>
            <Link className={styles.secondaryAction} to="/service-portal/tickets">
              {t('servicePortal.viewAll')}
            </Link>
            <button className={styles.primaryAction} type="button" onClick={openGeneric}>
              {t('servicePortal.createRequest')}
            </button>
          </div>
        </div>

        <div className={styles.panelBody}>
          <div className={styles.catalogArea}>
            <CategoryTabs active={activeCategory} onChange={setActiveCategory} />
            <div className={styles.serviceGrid}>
              {visibleItems.map((item) => (
                <ServiceCard key={item.id} item={item} onSelect={openService} />
              ))}
            </div>
          </div>

          <MyTicketsAside />
        </div>
      </section>

      <RequestDrawer
        open={drawer !== null}
        onClose={closeDrawer}
        eyebrow={t('servicePortal.eyebrow')}
        title={drawerProps?.serviceTitle ?? ''}
        description={drawerProps?.description}
      >
        {drawerProps ? (
          <RequestForm
            serviceId={drawerProps.serviceId}
            serviceTitle={drawerProps.serviceTitle}
            schema={drawerProps.schema}
            applicant={applicant}
            onSubmitted={closeDrawer}
            onCancel={closeDrawer}
          />
        ) : null}
      </RequestDrawer>
    </div>
  );
}
```

- [ ] **Step 3: 构建验证**

Run: `cd web && npm run build`
Expected: 成功。

- [ ] **Step 4: 浏览器走查**

```bash
cd web && npm run dev
```

登录后打开 `/service-portal`：
- 5 个分类 tab 切换不抖；popular 显示 4 条 hot
- 任一服务卡点开 → Drawer 滑出 → 填表 → 提交：Drawer 关，「我的工单」侧栏顶部出现新工单（状态「已提交」）
- 5s 后状态变为「审批中」，30s「处理中」，60s「已完成」
- 顶部「创建申请」按钮 → 打开通用表单 Drawer，提交后行为同上

- [ ] **Step 5: Commit**

```bash
git add web/src/view/service/ServicePortalPage.tsx web/src/view/service/ServicePortalPage.module.css
git commit -m "feat(service-portal): 改造首页为目录 + 我的工单 + Drawer 入口"
```

---

## Task 14: `TicketListPage`（路由 + 搜索 + 多选状态过滤）

**Files:**

- Create: `web/src/view/service/TicketListPage.tsx`
- Create: `web/src/view/service/TicketListPage.module.css`

- [ ] **Step 1: 创建 css**

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

.table {
  width: 100%;
  border-collapse: collapse;
  background: #fffdf9;
  border: 1px solid rgba(86, 70, 55, 0.13);
  border-radius: 0.9rem;
  overflow: hidden;
}

.table th,
.table td {
  text-align: left;
  padding: 0.7rem 0.9rem;
  font-size: 0.84rem;
  border-bottom: 1px solid rgba(86, 70, 55, 0.08);
}

.table th {
  background: rgba(246, 238, 228, 0.42);
  font-weight: 760;
  color: #24201c;
}

.row {
  cursor: pointer;
}

.row:hover {
  background: rgba(47, 98, 92, 0.04);
}

.empty {
  padding: 1.5rem;
  text-align: center;
  color: #76695d;
}
```

- [ ] **Step 2: 创建组件**

```tsx
// TicketListPage.tsx
import { useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { useI18n } from '../../locale/useI18n';
import type { MessageKey } from '../../locale/messages';
import { PageHeader } from '../../components/page/PageHeader';
import { TicketStatusBadge } from './components/TicketStatusBadge';
import { useTicketStore } from './data/ticketStore';
import type { TicketStatus } from './ticketTypes';
import styles from './TicketListPage.module.css';

const ALL_STATUSES: TicketStatus[] = ['submitted', 'approving', 'processing', 'done', 'rejected'];

export function TicketListPage() {
  const { t } = useI18n();
  const navigate = useNavigate();
  const { tickets } = useTicketStore();
  const [search, setSearch] = useState('');
  const [activeStatuses, setActiveStatuses] = useState<Set<TicketStatus>>(new Set());

  const filtered = useMemo(() => {
    const query = search.trim().toLowerCase();
    return tickets.filter((ticket) => {
      if (activeStatuses.size > 0 && !activeStatuses.has(ticket.status)) return false;
      if (query) {
        const hay = `${ticket.id} ${ticket.title}`.toLowerCase();
        if (!hay.includes(query)) return false;
      }
      return true;
    });
  }, [tickets, search, activeStatuses]);

  function toggleStatus(status: TicketStatus) {
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
        eyebrow={t('servicePortal.ticket.list.eyebrow')}
        title={t('servicePortal.ticket.list.title')}
        description={t('servicePortal.ticket.list.description')}
      />

      <div className={styles.controls}>
        <input
          className={styles.search}
          placeholder={t('servicePortal.ticket.list.searchPlaceholder')}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <div className={styles.chips} role="group" aria-label={t('servicePortal.ticket.list.filterLabel')}>
          {ALL_STATUSES.map((status) => {
            const active = activeStatuses.has(status);
            return (
              <button
                key={status}
                type="button"
                className={active ? `${styles.chip} ${styles.chipActive}` : styles.chip}
                onClick={() => toggleStatus(status)}
              >
                {t(`servicePortal.ticket.status.${status}` as MessageKey)}
              </button>
            );
          })}
        </div>
      </div>

      <table className={styles.table}>
        <thead>
          <tr>
            <th>{t('servicePortal.ticket.list.columnId')}</th>
            <th>{t('servicePortal.ticket.list.columnTitle')}</th>
            <th>{t('servicePortal.ticket.list.columnStatus')}</th>
            <th>{t('servicePortal.ticket.list.columnCreated')}</th>
          </tr>
        </thead>
        <tbody>
          {filtered.length === 0 ? (
            <tr>
              <td colSpan={4} className={styles.empty}>
                {t('servicePortal.ticket.list.empty')}
              </td>
            </tr>
          ) : (
            filtered.map((ticket) => (
              <tr
                key={ticket.id}
                className={styles.row}
                onClick={() => navigate(`/service-portal/tickets/${ticket.id}`)}
              >
                <td>{ticket.id}</td>
                <td>{ticket.title}</td>
                <td>
                  <TicketStatusBadge status={ticket.status} />
                </td>
                <td>{new Date(ticket.createdAt).toLocaleString()}</td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  );
}
```

- [ ] **Step 3: 构建验证**

Run: `cd web && npm run build`
Expected: 成功。

- [ ] **Step 4: Commit**

```bash
git add web/src/view/service/TicketListPage.tsx web/src/view/service/TicketListPage.module.css
git commit -m "feat(service-portal): 新增工单列表页（搜索 + 状态多选过滤）"
```

---

## Task 15: `TicketTimeline` 与 `TicketDetailPage`

**Files:**

- Create: `web/src/view/service/components/TicketTimeline.tsx`
- Create: `web/src/view/service/components/TicketTimeline.module.css`
- Create: `web/src/view/service/TicketDetailPage.tsx`
- Create: `web/src/view/service/TicketDetailPage.module.css`

- [ ] **Step 1: 创建 TicketTimeline**

```css
/* TicketTimeline.module.css */
.timeline {
  display: grid;
  gap: 0.65rem;
  padding-left: 0.5rem;
  border-left: 2px solid rgba(86, 70, 55, 0.18);
}

.item {
  display: grid;
  gap: 0.2rem;
  padding-left: 0.85rem;
  position: relative;
}

.item::before {
  content: '';
  position: absolute;
  left: -0.43rem;
  top: 0.3rem;
  width: 0.6rem;
  height: 0.6rem;
  border-radius: 999px;
  background: #2f625c;
}

.timestamp {
  color: #76695d;
  font-size: 0.74rem;
}

.text {
  margin: 0;
  font-size: 0.86rem;
  color: #24201c;
}
```

```tsx
// TicketTimeline.tsx
import { useI18n } from '../../../locale/useI18n';
import type { MessageKey } from '../../../locale/messages';
import type { Ticket } from '../ticketTypes';
import styles from './TicketTimeline.module.css';

interface Props {
  ticket: Ticket;
}

export function TicketTimeline({ ticket }: Props) {
  const { t } = useI18n();
  return (
    <ol className={styles.timeline}>
      {ticket.timeline.map((event, index) => {
        const statusLabel = t(`servicePortal.ticket.status.${event.status}` as MessageKey);
        const note = event.i18nNoteKey
          ? t(event.i18nNoteKey as MessageKey)
              .replace('{applicant}', ticket.applicant.name || ticket.applicant.email)
              .replace('{status}', statusLabel)
          : statusLabel;
        return (
          <li key={`${event.at}-${index}`} className={styles.item}>
            <span className={styles.timestamp}>{new Date(event.at).toLocaleString()}</span>
            <p className={styles.text}>{note}</p>
          </li>
        );
      })}
    </ol>
  );
}
```

- [ ] **Step 2: 创建 TicketDetailPage**

```css
/* TicketDetailPage.module.css */
.page {
  display: grid;
  gap: 1.2rem;
}

.breadcrumb {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  color: #76695d;
  font-size: 0.82rem;
}

.breadcrumb a {
  color: #2f625c;
  text-decoration: none;
}

.breadcrumb a:hover {
  text-decoration: underline;
}

.summary {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 1fr);
  gap: 1.25rem;
  padding: 1.1rem 1.25rem;
  border: 1px solid rgba(86, 70, 55, 0.13);
  border-radius: 1rem;
  background: rgba(255, 253, 249, 0.74);
}

.field {
  display: grid;
  gap: 0.25rem;
}

.fieldLabel {
  color: #76695d;
  font-size: 0.74rem;
  font-weight: 700;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.fieldValue {
  color: #24201c;
  font-size: 0.9rem;
}

.section {
  display: grid;
  gap: 0.6rem;
  padding: 1.1rem 1.25rem;
  border: 1px solid rgba(86, 70, 55, 0.13);
  border-radius: 1rem;
  background: rgba(255, 253, 249, 0.74);
}

.section h3 {
  margin: 0;
  color: #24201c;
  font-size: 0.95rem;
  font-weight: 700;
}

.formValues {
  display: grid;
  gap: 0.45rem;
  margin: 0;
}

.formValues dt {
  font-weight: 700;
  font-size: 0.82rem;
  color: #24201c;
}

.formValues dd {
  margin: 0 0 0.4rem;
  color: #4a4338;
  font-size: 0.82rem;
}

.emptyState {
  display: grid;
  gap: 0.65rem;
  justify-items: start;
  padding: 1.5rem;
  border: 1px dashed rgba(86, 70, 55, 0.25);
  border-radius: 1rem;
}

@media (max-width: 760px) {
  .summary {
    grid-template-columns: 1fr;
  }
}
```

```tsx
// TicketDetailPage.tsx
import { Link, useParams } from 'react-router-dom';

import { useI18n } from '../../locale/useI18n';
import type { MessageKey } from '../../locale/messages';
import { PageHeader } from '../../components/page/PageHeader';
import { TicketStatusBadge } from './components/TicketStatusBadge';
import { TicketTimeline } from './components/TicketTimeline';
import { getCatalogById } from './data/serviceCatalog';
import { useTicketStore } from './data/ticketStore';
import styles from './TicketDetailPage.module.css';

export function TicketDetailPage() {
  const { id = '' } = useParams<{ id: string }>();
  const { t } = useI18n();
  const { getTicket } = useTicketStore();
  const ticket = getTicket(id);

  if (!ticket) {
    return (
      <div className={styles.page}>
        <div className={styles.breadcrumb}>
          <Link to="/service-portal">{t('servicePortal.breadcrumbHome')}</Link>
          <span>/</span>
          <Link to="/service-portal/tickets">{t('servicePortal.breadcrumbTickets')}</Link>
        </div>
        <div className={styles.emptyState}>
          <h3>{t('servicePortal.notFoundTitle')}</h3>
          <p>{t('servicePortal.notFoundDescription')}</p>
          <Link to="/service-portal/tickets">{t('servicePortal.backToTickets')}</Link>
        </div>
      </div>
    );
  }

  const catalog = ticket.serviceId ? getCatalogById(ticket.serviceId) : null;
  const serviceLabel = catalog
    ? t(`${catalog.i18nKey}.title` as MessageKey)
    : t('servicePortal.ticket.detail.genericServiceLabel');

  return (
    <div className={styles.page}>
      <div className={styles.breadcrumb}>
        <Link to="/service-portal">{t('servicePortal.breadcrumbHome')}</Link>
        <span>/</span>
        <Link to="/service-portal/tickets">{t('servicePortal.breadcrumbTickets')}</Link>
        <span>/</span>
        <span>{ticket.title}</span>
      </div>

      <PageHeader
        eyebrow={t('servicePortal.ticket.detail.eyebrow')}
        title={ticket.title}
        description={`${ticket.id} · ${new Date(ticket.createdAt).toLocaleString()}`}
        actions={<TicketStatusBadge status={ticket.status} />}
      />

      <section className={styles.summary}>
        <div className={styles.field}>
          <span className={styles.fieldLabel}>{t('servicePortal.ticket.detail.serviceLabel')}</span>
          <span className={styles.fieldValue}>{serviceLabel}</span>
        </div>
        <div className={styles.field}>
          <span className={styles.fieldLabel}>{t('servicePortal.ticket.detail.applicantLabel')}</span>
          <span className={styles.fieldValue}>
            {ticket.applicant.name || ticket.applicant.email}
          </span>
        </div>
      </section>

      <section className={styles.section}>
        <h3>{t('servicePortal.ticket.detail.formValuesLabel')}</h3>
        <dl className={styles.formValues}>
          {Object.entries(ticket.formValues).map(([key, value]) => (
            <div key={key}>
              <dt>{key}</dt>
              <dd>{Array.isArray(value) ? value.join(', ') : String(value ?? '')}</dd>
            </div>
          ))}
        </dl>
      </section>

      <section className={styles.section}>
        <h3>{t('servicePortal.ticket.detail.timelineLabel')}</h3>
        <TicketTimeline ticket={ticket} />
      </section>
    </div>
  );
}
```

- [ ] **Step 3: 构建验证**

Run: `cd web && npm run build`
Expected: 成功。

- [ ] **Step 4: Commit**

```bash
git add web/src/view/service/components/TicketTimeline.tsx web/src/view/service/components/TicketTimeline.module.css web/src/view/service/TicketDetailPage.tsx web/src/view/service/TicketDetailPage.module.css
git commit -m "feat(service-portal): 新增工单详情页与时间轴组件"
```

---

## Task 16: 路由与 SideNav 接入

**Files:**

- Modify: `web/src/router/index.tsx`
- Modify: `web/src/components/nav/SideNav.tsx`

- [ ] **Step 1: 注册新路由**

打开 `web/src/router/index.tsx`。在文件顶部 import 区加：

```ts
import { TicketDetailPage } from '../view/service/TicketDetailPage';
import { TicketListPage } from '../view/service/TicketListPage';
```

在 `AppShell` 的 `children` 数组里、`service-portal` 路由后紧跟两条：

```tsx
{
  path: 'service-portal',
  element: <ServicePortalPage />,
},
{
  path: 'service-portal/tickets',
  element: <TicketListPage />,
},
{
  path: 'service-portal/tickets/:id',
  element: <TicketDetailPage />,
},
```

- [ ] **Step 2: SideNav 升级「工单」**

打开 `web/src/components/nav/SideNav.tsx`。在第 99 行附近的当前实现：

```tsx
<span className={styles.navEntryDisabled} title={t('shell.comingSoon')}>
  {t('shell.tickets')}
</span>
```

替换为：

```tsx
<NavLink
  to="/service-portal/tickets"
  className={({ isActive }) =>
    [styles.navEntry, isActive ? styles.navEntryActive : '']
      .filter(Boolean)
      .join(' ')
  }
>
  {t('shell.tickets')}
</NavLink>
```

确保文件顶部已 import `NavLink`（已存在）。其他三个占位（任务 / 我的值班 / 大屏）保持 disabled，不动。

- [ ] **Step 3: 构建验证**

Run: `cd web && npm run build`
Expected: 成功。

- [ ] **Step 4: 浏览器走查**

```bash
cd web && npm run dev
```

- 进 `/service-portal` 顶部「查看全部」点击 → 路由变 `/service-portal/tickets`，列表页渲染
- SideNav 「工单」点击 → 同样到列表页，且 NavLink 高亮
- 列表里点一行 → 进 `/service-portal/tickets/TKT-...`，详情 + 时间轴显示
- 手输 `/service-portal/tickets/TKT-X-9999` → 空状态页面，不崩

- [ ] **Step 5: Commit**

```bash
git add web/src/router/index.tsx web/src/components/nav/SideNav.tsx
git commit -m "feat(service-portal): 接入工单列表/详情路由与 SideNav 入口"
```

---

## Task 17: 全量手动验证（spec §9.2 清单）

**Files:** 无代码改动；用 dev server 跑 spec §9.2 全部 11 项手动验证。

- [ ] **Step 1: 启动 dev server**

```bash
cd web && npm run dev
```

- [ ] **Step 2: 跑 spec §9.2 第 1–11 项**

照抄 spec §9.2 全部 11 项：

1. ✅ `/service-portal`：13 张卡分 5 类（含 popular 显示 4 条 hot），tab 切换不抖
2. ✅ 点任一卡 → Drawer 滑出 → 填表 → 提交：Drawer 关、侧栏顶部多出新工单（「已提交」）
3. ✅ 等 5s → 「审批中」；30s → 「处理中」；60s → 「已完成」
4. ✅ 顶部「创建申请」→ Drawer 打开通用表单（不绑服务）→ 提交后同上
5. ✅ 顶部「查看全部」→ `/service-portal/tickets`：状态多选 chips 过滤、`title + id` 搜索均可用
6. ✅ 列表点一行 → 详情 → 时间轴显示已发生事件
7. ✅ SideNav「工单」直链同样到列表页
8. ✅ 刷新首页：localStorage 工单还在，状态固定在刷新前那一档（不再续推）
9. ✅ 手输不存在的 `/service-portal/tickets/TKT-X-9999`：空状态而非崩溃
10. ✅ 浏览器禁用 localStorage / 只读 → 提交时 Drawer 顶部出现红色 banner「本地存储不可用…」，页面不崩
    - 测试方法：DevTools → Application → Storage → Disable storage；或在 Console 跑 `Storage.prototype.setItem = () => { throw new Error('blocked'); }`
11. ✅ 切换登录账号后回到服务门户：仍能看到之前账号提交的工单（mock 不按用户隔离）

- [ ] **Step 3: 跑 grep 确认 i18n 残留为零**

```bash
cd /home/lilin/trae-project/Optimus
grep -rn "dashboard\.servicePortal" web/src/ || echo "no residue"
```
Expected: `no residue`

- [ ] **Step 4: 跑构建确认**

```bash
cd web && npm run build
```
Expected: 退出码 0。

- [ ] **Step 5: 如有偏差，回到对应 Task 修复并补 commit；若全部通过，不需要 commit**

---

## 完成后

所有任务做完，本期范围（spec §1）的服务门户前端 mock 升级即告完成。spec §10 的"未来工作"不在本期。
