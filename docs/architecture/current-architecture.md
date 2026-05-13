# Current Architecture

## Scope

- `server/` 下的 Go 后端，已实现完整的系统管理域（鉴权、用户、角色、权限、审计日志）
- `web/` 下的 React + TypeScript + Vite 前端，已实现对应系统管理页面
- `deploy/docker-compose/` 下的容器化部署编排

---

## Repository Layout

```
server/
├── api/v1/system/       # system 模块 HTTP handlers
├── config/              # 配置结构（System、JWT、MySQL、Redis、Casbin、Zap）
├── core/                # 配置加载、日志初始化、HTTP 服务启动
├── global/              # 进程级共享对象（DB、Config、Logger、Casbin）
├── initialize/          # 基础设施与路由装配
│   ├── system_module.go # SystemModule 组装（handlers + capabilities）
│   ├── audit_retention.go # 审计日志定时清理 worker
│   └── startup.go       # DB 迁移、seed、worker 启动
├── middleware/          # JWT、Casbin RBAC、日志、恢复中间件
├── model/system/        # GORM 模型（SystemUser、SystemAuditLog 等）
├── router/system/       # system 路由注册（public / private 分组）
├── service/system/
│   ├── usecase/         # 业务逻辑层：接口定义 + service 实现
│   │   ├── auth.go      # AuthService（本地登录 + LDAP）
│   │   ├── user.go      # UserService
│   │   ├── role.go      # RoleService
│   │   ├── audit.go     # AuditService（QueryLogs / ExportLogs / PurgeBefore）
│   │   ├── ldap.go      # LDAPService
│   │   └── rbac.go      # RBACService
│   └── infra/
│       ├── persistence/ # GORM 实现各 Store 接口
│       ├── directory/   # LDAP 客户端适配器
│       └── identity/    # JWT 签发 / 验证适配器
└── utils/               # 轻量工具（claims、响应封装）

web/
├── src/
│   ├── api/             # 后端接口调用封装
│   ├── auth/            # AuthContext、RequireAuth 守卫
│   ├── components/      # 通用组件（导航、页头、空态、语言切换）
│   ├── layout/          # AppShell（侧边栏 + 内容区布局）
│   ├── locale/          # i18n（中/英双语）
│   ├── router/          # React Router 路由定义
│   ├── types/           # TypeScript 类型定义
│   └── view/
│       ├── auth/        # 登录页
│       ├── dashboard/   # 仪表盘（占位）
│       └── system/      # 用户管理、角色管理、LDAP 配置、审计日志
```

---

## Backend Architecture

分层流向：

```
main → initialize → router → api/handlers → usecase/service → infra/persistence
```

**关键分层原则：**

- `usecase/` 只依赖标准库，定义接口（`AuthStore`、`AuditStore` 等），不感知数据库或框架
- `infra/persistence/` 实现 Store 接口，持有 `*gorm.DB`
- `infra/directory/` 实现 LDAP 客户端接口
- `infra/identity/` 实现 JWT 签发 / 验证
- `initialize/system_module.go` 负责依赖注入：把 GORM 实现传入 usecase，再传入 handlers

**新模块目录约定（新模块参照此结构）：**

```
service/<module>/
├── usecase/         # 接口 + 业务逻辑，可纯单元测试
└── infra/
    ├── persistence/ # GORM Store 实现
    └── provider/    # 外部 SDK 适配器（如云厂商同步器）

api/v1/<module>/     # HTTP handlers
router/<module>/     # 路由注册
initialize/<module>/ # 模块组装（Module struct）
```

---

## API Endpoints

**公开（无鉴权）：**

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/v1/health` | 健康检查 |
| GET | `/api/v1/ping` | ping |
| GET | `/api/v1/base/ping` | ping（别名） |
| POST | `/api/v1/system/auth/login` | 登录（本地 / LDAP） |

**私有（JWT + Casbin RBAC）：**

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/v1/system/auth/me` | 当前用户信息 |
| POST | `/api/v1/system/auth/logout` | 登出 |
| GET | `/api/v1/system/users` | 用户列表 |
| PUT | `/api/v1/system/users/:id` | 更新用户 |
| PATCH | `/api/v1/system/users/:id/status` | 更新用户状态 |
| PUT | `/api/v1/system/users/:id/roles` | 更新用户角色 |
| GET | `/api/v1/system/roles` | 角色列表 |
| POST | `/api/v1/system/roles` | 创建角色 |
| PUT | `/api/v1/system/roles/:id` | 更新角色 |
| PATCH | `/api/v1/system/roles/:id/status` | 更新角色状态 |
| PUT | `/api/v1/system/roles/:id/permissions` | 更新角色权限 |
| GET | `/api/v1/system/permissions/tree` | 权限树 |
| GET | `/api/v1/system/ldap/config` | LDAP 配置 |
| PUT | `/api/v1/system/ldap/config` | 更新 LDAP 配置 |
| POST | `/api/v1/system/ldap/test` | 测试 LDAP 连接 |
| GET | `/api/v1/system/audit-logs` | 审计日志查询（分页 + 多维过滤） |
| GET | `/api/v1/system/audit-logs/export` | 审计日志 CSV 导出 |

---

## Frontend Architecture

- React 18 + TypeScript + Vite，包管理器为 **pnpm**
- 路由守卫：未登录重定向到 `/login`，无权限跳转 403 页
- i18n：中 / 英双语，`LanguageProvider` + `useI18n`
- 一级导航：Dashboard、System
- System 子页面：用户管理、角色管理、LDAP 配置、审计日志

---

## Deployment

- `deploy/docker-compose/` 三服务编排（frontend + backend + mysql）
- 前端 nginx 反向代理 `/api` 到后端
- Caddy HTTPS 代理方案已设计（`optimus.intellivis.cn`）

---

## Tech Stack

| 层 | 技术 |
|----|------|
| 语言 | Go 1.23 |
| HTTP 框架 | Gin 1.10 |
| ORM | GORM 1.25 |
| 数据库 | MySQL |
| 配置 | Viper 1.19 |
| 日志 | Zap 1.27 |
| Redis 客户端 | go-redis/v9 |
| 权限 | Casbin v2 |
| JWT | golang-jwt/jwt v5 |
| 测试 | Go test + testify |
| 前端 | React 18 + TypeScript + Vite |
| 前端包管理 | pnpm |

---

## Verification Commands

```bash
# 后端
cd server && go test ./...
cd server && go build ./...
cd server && go run .

# 前端
cd web && pnpm build
cd web && pnpm dev   # 开发模式
```

---

## Platform Roadmap

已完成阶段：

| 阶段 | 内容 | 状态 |
|------|------|------|
| Phase 0 | User / Role / Permission / Audit | ✅ 完成 |
| Phase 0.5 | Audit Log Upgrade（查询/导出/保留策略） | ✅ 完成 |

规划中阶段：

| 阶段 | 内容 |
|------|------|
| Phase 1 | Cloud Account + Unified Asset Registry |
| Phase 2 | DevOps Layer |
| Phase 3 | FinOps Layer |
| Phase 4 | Automation Layer |

详见 `docs/superpowers/specs/2026-05-12-finops-devops-platform-roadmap-design.md`。
