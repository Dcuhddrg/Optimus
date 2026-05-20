# 登录认证与 RBAC 设计

## 背景

Optimus 当前已经具备 Go 后端骨架、React 前端壳子、基础 Docker Compose 部署，以及 JWT、Casbin、MySQL 等可插拔初始化能力。但完整登录鉴权体系、用户域、角色域、权限域和审计日志仍未实现。

本次目标是在现有骨架上补齐第一版可用的系统认证闭环：本地 `admin` 登录、LDAP 用户登录、JWT 鉴权、角色授权、后端 seed 权限点和基础审计日志。

## 目标

- 支持本地 `admin` 用户登录
- 支持 LDAP 用户登录与首登同步
- 支持 Bearer JWT 鉴权
- 支持用户查看、资料维护、启停状态和角色绑定管理
- 支持角色、角色授权管理
- 支持后端 seed 菜单权限点和 API 权限点
- 支持私有 API 的 RBAC 校验
- 支持 LDAP 配置读取、保存和连接测试
- 支持登录与关键系统变更的审计记录
- 前端新增登录页和 System 下的用户、角色、LDAP、审计页面

## 非目标

- 不做菜单权限点和 API 权限点在线新增/删除
- 不做组织架构、部门、岗位或复杂角色继承
- 不做任意本地密码用户创建；第一版只有 seed 出来的本地 `admin` 使用本地密码登录
- 不做 refresh token、token 黑名单或多端会话管理
- 不搭建真实 LDAP 测试环境
- 不新增前端 UI 组件库
- 不把 Dashboard 扩展成完整运维首页

## 设计原则

- 沿用现有后端分层：`router -> api -> service -> model/global/config`
- 系统域能力集中在 `system` 下，避免第一版拆出过多横向模块
- 权限点由后端 seed，运行期只允许角色授权，不允许动态改权限元数据
- RBAC 默认拒绝，公共接口必须显式声明
- LDAP 能力通过接口隔离，测试不依赖真实 LDAP 服务
- 前端保持运维后台气质，优先可用性、可读性和管理效率

## 总体架构

后端新增四个系统域：

- `auth`：登录、当前用户、JWT 签发与校验
- `user`：本地用户、LDAP 同步用户、用户状态与角色绑定
- `rbac`：角色、权限点、角色授权、权限判定
- `audit`：登录、授权、用户、角色和 LDAP 配置变更审计

认证运行时流程：

```text
Login Form
  -> Auth API
  -> Auth Service
  -> local admin password hash OR LDAP search + bind
  -> user sync
  -> JWT issue
  -> frontend stores token
  -> private API: JWT middleware -> RBAC middleware -> handler
```

`username == admin` 时走本地认证，其余用户走 LDAP。第一版不支持创建任意本地密码用户；用户列表中的非 `admin` 用户来自 LDAP 首登同步。LDAP 用户首次登录时自动创建本地用户，后续登录同步显示名、邮箱等非状态字段。用户禁用状态以本地数据库为准。

## 数据模型

### sys_users

用户主表，保存本地用户和 LDAP 同步用户。

建议字段：

- `id`
- `uuid`
- `username`
- `display_name`
- `email`
- `source`
- `status`
- `password_hash`
- `last_login_at`
- `created_at`
- `updated_at`

`source` 支持 `local` 和 `ldap`。第一版只有 seed 出来的 `admin` 是 `local` 登录用户；LDAP 用户不保存本地密码哈希。`status` 用于启用和禁用账号。

### sys_roles

角色表。

建议字段：

- `id`
- `code`
- `name`
- `description`
- `status`
- `created_at`
- `updated_at`

默认 seed `admin` 和 `user`。第一版允许新增和编辑角色，但不做角色继承。

### sys_user_roles

用户角色绑定表。

建议字段：

- `id`
- `user_id`
- `role_id`
- `created_at`

一个用户可以绑定多个角色。

### sys_permissions

权限点表，由后端 seed，不允许页面新增或删除。

建议字段：

- `id`
- `code`
- `name`
- `type`
- `method`
- `path`
- `parent_code`
- `sort`
- `created_at`
- `updated_at`

`type` 支持 `menu` 和 `api`。菜单权限用于前端导航过滤，API 权限用于后端接口访问控制。

### sys_role_permissions

角色权限绑定表。

建议字段：

- `id`
- `role_id`
- `permission_id`
- `created_at`

页面可以修改角色拥有的权限点。

### sys_ldap_config

LDAP 配置表。

建议字段：

- `id`
- `enabled`
- `url`
- `base_dn`
- `bind_dn`
- `bind_password`
- `user_filter`
- `username_attr`
- `display_name_attr`
- `email_attr`
- `created_at`
- `updated_at`

密码字段不能在接口响应中明文返回。第一版如暂未接入 secret 管理，至少必须做响应脱敏。

### sys_audit_logs

审计日志表。

建议字段：

- `id`
- `actor_user_id`
- `action`
- `resource_type`
- `resource_id`
- `result`
- `ip`
- `user_agent`
- `detail`
- `created_at`

第一版支持写入和基础列表查询，不做复杂检索和报表。

## 迁移与 Seed

启动时通过 GORM AutoMigrate 建表，然后执行幂等 seed：

- 默认 `admin` 用户
- 默认 `admin` 和 `user` 角色
- 固定菜单权限点
- 固定 API 权限点
- `admin` 角色拥有全部权限
- 默认 `admin` 用户绑定 `admin` 角色

默认 admin 初始密码必须来自配置或环境变量，不能硬编码在代码里。首次 seed 后只保存 bcrypt 哈希。

权限点 seed 必须幂等：重复启动不能重复插入，也不能覆盖运行期的角色授权结果。

## API 设计

接口统一挂在现有 `/api/v1` 前缀下，系统能力放在 `/system/...`。

### 认证

- `POST /system/auth/login`
- `GET /system/auth/me`
- `POST /system/auth/logout`

`login` 返回 token、用户信息、角色和菜单权限。`logout` 主要由前端清理 token，后端记录审计日志。

### 用户

- `GET /system/users`
- `PUT /system/users/:id`
- `PATCH /system/users/:id/status`
- `PUT /system/users/:id/roles`

用户管理负责用户列表、基础资料、启停状态和角色绑定。第一版不提供创建任意本地密码用户的接口；除 seed 的 `admin` 外，用户由 LDAP 首登同步产生。LDAP 用户的身份字段由 LDAP 同步，不应允许随意改写来源身份。

### 角色

- `GET /system/roles`
- `POST /system/roles`
- `PUT /system/roles/:id`
- `PATCH /system/roles/:id/status`
- `PUT /system/roles/:id/permissions`

角色授权接口保存角色拥有的菜单和 API 权限点。

角色授权请求使用权限 code，而不是数据库 ID：

```json
{
  "permission_codes": ["system.users.view", "system.roles.update"]
}
```

### 权限

- `GET /system/permissions/tree`

返回后端 seed 的权限树，供角色授权页面使用。

### LDAP

- `GET /system/ldap/config`
- `PUT /system/ldap/config`
- `POST /system/ldap/test`

连接测试不保存配置。配置接口返回时必须脱敏 bind password。

`GET /system/ldap/config` 返回 `has_bind_password`，但不返回 `bind_password` 明文。

`PUT /system/ldap/config` 保存当前表单配置。`bind_password` 为空时保留已保存密码；非空时替换已保存密码。

`POST /system/ldap/test` 测试请求使用当前表单载荷，不隐式读取前端看不到的字段：

```json
{
  "url": "ldap://127.0.0.1:389",
  "base_dn": "dc=example,dc=com",
  "bind_dn": "cn=admin,dc=example,dc=com",
  "bind_password": "secret",
  "use_saved_bind_password": false,
  "user_filter": "(uid={username})",
  "username_attr": "uid",
  "display_name_attr": "cn",
  "email_attr": "mail"
}
```

当 `use_saved_bind_password` 为 `true` 时，后端可以使用已保存密码；否则必须使用请求中的 `bind_password`。未保存密码且请求也未提供密码时返回校验错误。

测试响应：

```json
{
  "success": true,
  "message": "LDAP connection test passed"
}
```

### 审计

- `GET /system/audit-logs`

第一版提供基础分页列表。

## 前端设计

新增页面：

- `/login`：独立登录页，不进入 `AppShell`
- `/`：需要登录，默认 Dashboard
- `System > Users`：用户列表、启停状态、角色绑定
- `System > Roles`：角色列表、角色表单、权限树授权
- `System > LDAP`：从静态页升级为 LDAP 配置和连接测试页
- `System > Audit Logs`：审计日志列表

前端状态策略：

- token 存在 `localStorage`
- 请求层统一加 `Authorization: Bearer <token>`
- 401 时清理 token 并跳转 `/login`
- `/system/auth/me` 返回当前用户、角色和菜单权限
- 侧边栏按菜单权限过滤
- 无权限路由显示无权限页

前端不引入新 UI 组件库，继续使用 React、React Router、CSS Modules 和现有组件风格。

## 错误处理

- 登录失败统一返回“用户名或密码错误”
- 用户禁用返回明确“账号已禁用”
- JWT 缺失、无效或过期返回 401
- 权限不足返回 403
- LDAP 配置错误和连接失败在审计日志或服务日志中记录内部原因，接口返回可操作的通用错误
- LDAP 测试连接失败不保存配置
- 审计写入失败不能阻断主流程

## RBAC 匹配规则

API 权限点的唯一运行时匹配键为：

```text
METHOD + " " + ROUTE_PATTERN
```

其中：

- `METHOD` 使用大写 HTTP method，例如 `GET`、`POST`、`PUT`
- `ROUTE_PATTERN` 使用 Gin 注册路由模板，而不是原始请求路径，例如 `/api/v1/system/users/:id`
- `sys_permissions.method` 保存 method
- `sys_permissions.path` 保存 route pattern
- `sys_permissions.code` 是授权和前端展示使用的稳定标识

RBAC 中间件处理流程：

1. JWT 中间件解析 token，得到用户 ID 和角色集合
2. RBAC 中间件读取 `c.FullPath()` 和 `c.Request.Method`
3. 如果路由是公共接口或认证例外，跳过 RBAC
4. 用 `method + path` 查找 seed 的 API 权限点
5. 查不到权限点时默认返回 403
6. 用户任一启用角色拥有该权限点时放行
7. 否则返回 403

Casbin 可作为判定引擎承载 `role_code, path, method` 策略，但数据库中的 `sys_role_permissions` 是授权事实来源。角色授权保存后需要刷新内存策略或在下一次请求前保证判定使用最新授权。

## 安全边界

- admin 初始密码来自配置或环境变量
- 密码只保存 bcrypt 哈希
- LDAP bind password 不明文返回给前端
- 生产环境必须覆盖默认 JWT signing key
- 私有 API 没有匹配权限点时默认拒绝
- 公共接口和认证接口必须显式列入例外
- 至少保留一个启用的 admin-capable 用户，避免系统被锁死

admin-capable 用户定义为：用户状态启用，并绑定状态启用的 `admin` 角色。`admin` 角色是系统保留角色，默认拥有全部权限。第一版不允许删除 `admin` 角色，也不允许把 `admin` 角色禁用。

以下变更必须校验不能破坏 admin-capable 不变量：

- 禁用用户
- 修改用户角色绑定
- 禁用角色
- 修改 `admin` 角色权限

第一版可以直接禁止移除 `admin` 角色的核心系统权限，核心权限至少包括用户、角色、权限、LDAP 和审计相关管理接口。

## 审计事件

第一版必须记录这些审计事件：

- `auth.login.success`
- `auth.login.failure`
- `auth.logout`
- `user.update`
- `user.status.update`
- `user.roles.update`
- `role.create`
- `role.update`
- `role.status.update`
- `role.permissions.update`
- `ldap.config.update`
- `ldap.test.success`
- `ldap.test.failure`

审计 `result` 使用 `success` 或 `failure`。登录失败审计不记录明文密码。`detail` 可以记录失败类型、变更前后摘要和请求参数摘要，但必须避免保存密码、token、LDAP bind password 等敏感明文。

## 测试策略

后端测试：

- 本地 admin 登录成功和失败
- LDAP 登录成功和失败
- LDAP 首登同步
- 禁用用户无法登录
- JWT 缺失、无效和过期
- RBAC 允许和拒绝
- 权限 seed 幂等
- 审计日志写入
- LDAP 配置脱敏

LDAP 测试通过 mock 接口完成，不依赖真实 LDAP 服务。

前端验证：

- 登录态路由保护
- 401 跳转登录页
- 菜单权限过滤
- 用户、角色、LDAP、审计页面构建通过
- `pnpm build` 通过

整体验证：

- `cd server && go test ./...`
- `cd server && go build ./...`
- `cd web && pnpm build`
- Docker Compose 配置在后续实施收尾阶段单独核对

## 实施分期

### 阶段一：后端认证基础

新增 system 模型、迁移和 seed；实现密码哈希、JWT 服务、本地 admin 登录、当前用户接口和基础审计写入。

### 阶段二：权限闭环

实现 JWT 中间件、RBAC 判定、权限点 seed、角色授权接口和私有 API 默认拒绝策略。

### 阶段三：LDAP

实现 LDAP 配置模型、配置 API、连接测试、LDAP 登录和首登同步。LDAP 逻辑通过接口封装，测试使用 mock。

### 阶段四：前端闭环

新增登录页、请求层、登录态保护、System 用户、角色、LDAP 和审计页面。

### 阶段五：验证与部署收尾

运行后端测试、后端构建和前端构建。核对当前 Docker Compose 端口示例和 pnpm 重试修改，决定单独提交或并入部署收尾提交。

## 验收标准

- 可以用默认 admin 登录系统
- 可以读取当前用户、角色和菜单权限
- 未登录访问私有接口返回 401
- 已登录但无权限访问私有接口返回 403
- admin 角色默认拥有全部权限
- LDAP 用户首次登录后自动同步本地用户
- 可以查看、编辑、禁用用户并修改用户角色绑定
- 第一版不能创建任意本地密码登录用户
- 可以创建、编辑、禁用非保留角色
- 可以给角色保存权限
- LDAP 配置可以保存和测试连接
- LDAP 配置读取不会返回 bind password 明文
- 权限判定按 `METHOD + ROUTE_PATTERN` 执行，未 seed 的私有 API 默认拒绝
- 登录、登出、用户变更、角色变更、授权变更、LDAP 配置和 LDAP 测试写入审计日志
- 前端可以完成登录、进入后台、按权限显示菜单
- 后端测试和构建通过
- 前端构建通过

## 风险与控制

### 风险 1：权限元数据管理范围膨胀

如果第一版允许在线新增和删除菜单/API 权限点，会显著增加前后端复杂度和误配置风险。

控制方式：

- 权限点只由后端 seed
- 页面只做角色授权

### 风险 2：LDAP 环境不可控

真实 LDAP 服务配置差异大，容易让测试依赖外部环境。

控制方式：

- LDAP 逻辑通过接口隔离
- 单元测试使用 mock LDAP
- 手工验证只覆盖配置保存和连接测试能力

### 风险 3：admin 账号锁死

如果禁用最后一个 admin 或移除最后一个 admin 权限，可能导致系统无法管理。

控制方式：

- 保存用户状态和角色授权时校验至少保留一个启用 admin 用户
- admin 角色默认拥有全部权限

### 风险 4：审计影响主流程

审计写入失败不应导致登录或授权变更失败。

控制方式：

- 主流程成功后写审计
- 审计失败只记录服务日志
