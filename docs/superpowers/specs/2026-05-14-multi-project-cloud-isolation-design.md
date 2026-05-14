# 多项目隔离 + Cloud 模块前端设计

**Date:** 2026-05-14
**Status:** Approved (overall architecture; per-sub-project specs to follow)

## Overview

在已落地的 Phase 1 Cloud 模块（cloud_accounts + cloud_assets + Aliyun ECS syncer）之上，新增"项目"实体作为业务划分的一级抽象。一个用户被显式授权加入若干项目，只能看到自己项目下的云账号与资产。顶部"项目切换器"承载上下文，所有 cloud 页面在当前项目 ctx 下自动过滤。

设计目标：
- **隔离强度** = 项目级权限隔离（用户必须显式被加入项目才能访问其资源）
- **范围扩展** = 在不破坏已上线 cloud 接口契约的前提下，加一层项目维度
- **前后端协同** = 后端先就绪，再上前端

---

## Architecture Summary

```
┌─────────────────────────────────────────────────────────┐
│  顶部项目切换器（AppShell）                              │
│  ├ 当前项目: prod ▾                                      │
│  └ 切换后所有 cloud 接口自动注入 project_id              │
└─────────────────────────────────────────────────────────┘
        │
        ↓ ctx.currentProjectID + ctx.accessibleProjectIDs
┌─────────────────────────────────────────────────────────┐
│  后端 service/cloud/usecase                              │
│  - AccountStore.List(ctx) 强制 WHERE project_id IN(...)  │
│  - AssetStore.Query(ctx, f) 强制 WHERE project_id IN(...)│
│  - Create/Update 校验 target project ∈ accessible        │
└─────────────────────────────────────────────────────────┘
        │
        ↓
┌─────────────────────────────────────────────────────────┐
│  后端 service/project/usecase（新增）                    │
│  - Project CRUD                                          │
│  - UserProject 成员管理                                  │
│  - ResolveAccessibleProjects(userID) → []uint            │
└─────────────────────────────────────────────────────────┘
```

---

## Data Model

### `projects`（新增）

| 列 | 类型 | 说明 |
|----|------|------|
| `id` | uint PK | |
| `name` | varchar(128) | 显示名 |
| `code` | varchar(64) UNIQUE | 唯一代号，URL 与 API 用 |
| `description` | varchar(512) nullable | |
| `owner_id` | uint FK → system_users | 创建者，默认 owner |
| `status` | varchar(16) | `active` / `archived` |
| `created_at` / `updated_at` | datetime | |
| `deleted_at` | datetime nullable | 软删除 |

### `user_projects`（新增，多对多关联）

| 列 | 类型 | 说明 |
|----|------|------|
| `user_id` | uint FK → system_users | |
| `project_id` | uint FK → projects | |
| `project_role` | varchar(32) | `owner` / `member` / `viewer`（项目内角色，与 SystemRole 解耦） |
| `created_at` | datetime | |

UNIQUE `(user_id, project_id)`。

### `cloud_accounts`（改造，加一列）

新增：`project_id uint NOT NULL FK → projects`，索引 `(project_id, deleted_at)`。
唯一约束改：`(project_id, provider, account_id)` —— 同一云账号可以被不同项目分别接入。

`cloud_assets` 无需直接加 `project_id`，因为已有 `account_id FK → cloud_accounts`，关联即可得项目。
查询时通过 JOIN 或子查询过滤。如性能压力大可后续 denormalize。

---

## Permission Model

**两层授权叠加：**

1. **SystemRole（既有）** —— 控制 *能否使用* 某个功能模块
   - `project.menu` / `project.view` / `project.manage` / `project.members.manage`
   - 既有的 `cloud.*` 权限保留

2. **项目成员关系（新增）** —— 控制 *能访问哪些项目的数据*
   - 在 service 层强制注入 `userProjectIDs` 过滤
   - Casbin 不引入 project 维度，避免策略爆炸

**项目内三种角色（user_projects.project_role）：**
- `owner`：删项目 / 加减成员 / 全部读写
- `member`：项目内 cloud 资源读写（受 SystemRole `cloud.*` 限制）
- `viewer`：项目内只读

**ctx 注入流程：**
```
middleware/project_scope.go:
  1. 从 header 或 cookie 读 X-Project-Code
  2. 查 user_projects 校验用户对该项目有权限
  3. 将 currentProjectID + accessibleProjectIDs + projectRole 注入 ctx
  4. 后续 handler/usecase 从 ctx 取值
```

---

## API Surface

### 项目模块（新增，挂 `/api/v1/projects`）

| Method | Path | 说明 |
|--------|------|------|
| GET | `/projects` | 当前用户可见项目列表 |
| POST | `/projects` | 创建项目（创建者自动 owner） |
| GET | `/projects/:code` | 项目详情 |
| PUT | `/projects/:code` | 更新名称/描述 |
| DELETE | `/projects/:code` | 软删除（仅 owner） |
| GET | `/projects/:code/members` | 成员列表 |
| POST | `/projects/:code/members` | 添加成员 `{userId, projectRole}` |
| PATCH | `/projects/:code/members/:userId` | 改成员角色 |
| DELETE | `/projects/:code/members/:userId` | 移除成员 |

### Cloud 模块（保留路径，加 project ctx 过滤）

既有 `/api/v1/cloud/accounts`、`/api/v1/cloud/assets` 路径不变。
- 请求需带 `X-Project-Code: prod` header（或后续切换为路径参数 `/cloud/projects/:code/accounts`，第一版先用 header）
- 创建账号时 body 加 `projectCode` 字段
- 列表/查询自动按 ctx 过滤

---

## Frontend UX

### 1. AppShell 顶部项目切换器
- 位于 brand 右侧，dropdown 展开显示用户可访问项目（带账号/资产计数预览）
- 选中后写入 localStorage + 全局 context，所有 API 自动带 `X-Project-Code` header
- 切换器底部 "+ 管理项目" 入口进 `/projects`

### 2. SideNav 重构
```
概览
├ 仪表盘
└ 项目                       (gated by project.menu)

当前项目 · {projectName}
├ 云账号                     (gated by cloud.accounts.menu)
└ 云资产                     (gated by cloud.assets.menu)

系统
├ 用户 / 角色 / LDAP / 审计
```
未选中项目时"当前项目"section 折叠或显示"请先选择项目"。

### 3. 项目管理页 `/projects`
- 列表：项目名 / 代号 / 成员数 / 云账号数 / 资产数 / 创建时间 / 操作（成员/编辑）
- 新建/编辑弹窗：name / code (URL slug) / description
- 删除二次确认（提示该项目下还有 N 个云账号需先解绑）

### 4. 项目成员管理 `/projects/:code/members`
- 列表卡片：头像 / 用户名 / SystemRole / 项目内角色（Owner/Member/Viewer 徽章）
- 添加成员：选 user + 选 projectRole
- 移除成员二次确认

### 5. 云账号页（已存在，加项目维度）
- 顶部面包屑 `{projectName} / 云账号`
- 不再有"项目"过滤列（顶部切换器即过滤）
- 新建弹窗自动绑定当前项目，无需用户选

### 6. 云资产页（已存在，同上）

---

## Sub-Project Decomposition

按依赖顺序拆为 4 个可独立交付的 sub-project：

### SP1. 后端：项目实体与成员管理
**新增包：** `server/service/project/{usecase,infra/persistence}` + `server/api/v1/project` + `server/router/project` + `server/initialize/project`
**交付物：** projects 表 + user_projects 表 + 9 个 REST 接口 + 权限种子（project.*）+ 单元/集成测试
**对 cloud 模块无侵入。** 完成后 cloud 模块仍按现状无项目维度运行。

### SP2. 后端：Cloud 模块绑定项目
**改动：**
- `cloud_accounts` migration 加 `project_id` 列（带数据迁移：把所有现有账号划到 "default" 项目）
- middleware/project_scope.go 注入 ctx
- AccountStore / AssetStore 改 List/Query 签名接收 `accessibleProjectIDs`
- handler 从 ctx 读 currentProjectID
- 既有 cloud 测试更新

**依赖 SP1。** 不动 API 路径，前端可平滑过渡。

### SP3. 前端：项目管理 + 顶部切换器 + AppShell 重构
**新增：**
- `web/src/view/project/ProjectsPage.tsx` + `ProjectMembersPage.tsx`
- `web/src/api/project.ts`
- `web/src/components/project/ProjectSwitcher.tsx`
- `web/src/auth/ProjectContext.tsx`（当前项目 + localStorage 持久化）
- `web/src/api/client.ts` 注入 `X-Project-Code` header
- SideNav 重构为三段式

**依赖 SP1 + SP2 接口就绪。** 上线后用户可见项目页与切换器。

### SP4. 前端：Cloud 页面项目化适配
**改动：**
- `CloudAccountsPage` 移除项目列（如有）、新建弹窗自动绑定 currentProject、显示面包屑
- `CloudAssetsPage` 移除项目过滤、显示面包屑
- 未选中项目时显示 EmptyState

**依赖 SP2 + SP3。**

---

## Architecture Decisions

1. **项目内角色（owner/member/viewer）与 SystemRole 解耦** —— SystemRole 控制功能可见性，project_role 控制项目内动作。两者 AND 关系。理由：避免 Casbin 策略爆炸，同时保留项目内职责分层。

2. **`cloud_assets` 不冗余 `project_id`** —— 通过 `account_id` JOIN cloud_accounts 即可得。理由：YAGNI；Phase 1 资产量级查询性能足够；后续 denormalize 成本低。

3. **`X-Project-Code` Header 注入 vs 路径参数** —— 选 header 方案。理由：cloud 模块已上线，改路径会破坏接口契约；header 让现有客户端无感升级。

4. **首次登录无项目时的行为** —— 若用户被任何 SystemRole 授予 `cloud.accounts.view` 但 user_projects 为空，则 cloud 页面显示 EmptyState "你还未被加入任何项目，请联系管理员"。

5. **创建项目时的默认成员** —— 创建者自动 user_projects.project_role = "owner"。

6. **删除项目** —— 软删除项目本身；硬阻止：若项目下仍有未删除的 cloud_accounts，返回 409 + 提示先解绑。

7. **既有云账号的数据迁移** —— SP2 migration 步骤：建 "default" 项目；把所有现有 cloud_accounts 划入 default；project_id 改 NOT NULL。

---

## Open Questions (deferred to sub-project brainstorming)

- **SP1**：项目 `code` 的合法字符集与长度（建议 `^[a-z][a-z0-9-]{1,62}$`）
- **SP1**：项目软删除后能否恢复，UI 是否提供恢复入口
- **SP2**：账号能否跨项目迁移（PUT /cloud/accounts/:id 是否允许改 project_id）
- **SP3**：项目切换器是否支持"全部项目"聚合视图（首版按 brainstorming 决议：不支持）
- **SP3**：未登录或 SystemRole 无 project.menu 时切换器是否仍显示

---

## Scope Notes

- 不实现 SaaS 级多租户隔离（tenant_id 全表注入）
- 不实现项目级 Quota / 用量限额
- 不实现项目模板 / 项目克隆
- 不实现项目间资源共享 / 跨项目策略
- 这些都属于未来扩展，本设计预留了空间但 Phase 1 不做
