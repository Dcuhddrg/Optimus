# Current Architecture

> 本文件只维护**图谱说不清的契约和约定**：边界、分层规则、命名/目录约定、运行时契约、技术选型。
> 目录树、文件清单、函数列表、API 端点表等**结构性事实**统一以 codebase-memory 图谱为准（`get_architecture` / `search_graph` / `query_graph`）。

---

## Scope

- `server/` — Go 后端
- `web/` — React + TypeScript + Vite 前端
- `deploy/docker-compose/` — 容器化部署编排
- 模块已建立：系统管理（鉴权 / 用户 / 角色 / 权限 / 审计日志）；其它模块以图谱为准

---

## Backend 分层契约

```
main → initialize → router → api/handlers → usecase/service → infra/persistence
```

**强制约束：**

- `usecase/` 只依赖标准库，定义 Store / Provider 等接口，**不感知 GORM / Gin / 任何框架**
- `infra/persistence/` 实现 Store 接口、持有 `*gorm.DB`
- `infra/<adapter>/` 实现外部协议适配器（如 LDAP、JWT、云厂商 SDK）
- `initialize/<module>_module.go` 负责依赖注入：GORM 实现 → usecase → handlers
- handler 不得直接调用 GORM；usecase 不得 import infra

**新模块必须遵循的目录约定：**

```
server/
├── api/v1/<module>/         # HTTP handlers
├── router/<module>/         # 路由注册（public / private 分组）
├── service/<module>/
│   ├── usecase/             # 接口 + 业务逻辑，可纯单元测试
│   └── infra/
│       ├── persistence/     # GORM Store 实现
│       └── provider/        # 外部 SDK 适配器
├── model/<module>/          # GORM 模型
└── initialize/<module>.go   # 模块组装
```

---

## Frontend 约定

- React 18 + TypeScript + Vite，包管理器 **pnpm**（禁用 npm install）
- 路由守卫：未登录 → `/login`；无权限 → 403 页
- i18n：`LanguageProvider` + `useI18n`，中 / 英双语
- 一级导航在 `TopNav`；模块子页面在各 `view/<module>/` 下；不在 SideNav 重复一级入口
- 页面状态 / 接口调用统一走 `src/api/` 封装，view 层不直接 `fetch`

---

## API 设计契约

- 统一前缀 `/api/v1/<module>/...`
- 健康检查：`/api/v1/health`、`/api/v1/ping`（公开）
- 鉴权：登录 `/api/v1/<module>/auth/login`（公开），其它路径走 JWT + Casbin RBAC（私有）
- 路由必须按 public / private 分组注册（参考 `router/system/`）
- 具体端点清单：`query_graph("MATCH (n)-[r:HANDLES]->(p) RETURN n.name, r.method, r.url_path")`

---

## Deployment 运行时契约

- 三服务：frontend（nginx）+ backend（Go）+ mysql，由 `deploy/docker-compose/docker-compose.yml` 编排
- nginx 反向代理 `/api/` → `optimus-server:8888`
- Mac 端访问域名 `optimus.intellivis.cn`（Caddy HTTPS），直连排查走 `http://192.169.50.100:8080/`
- 前端运行效果验证**默认走 Compose**，不依赖 Vite dev server（详见 `.claude/project_rules.md` VERIFICATION）

---

## Tech Stack

| 层 | 选型 | 备注 |
|----|------|------|
| 语言 | Go 1.23 | 后端 |
| HTTP | Gin 1.10 | |
| ORM | GORM 1.25 | |
| DB | MySQL | |
| 配置 | Viper 1.19 | |
| 日志 | Zap 1.27 | |
| 缓存 | go-redis/v9 | |
| 权限 | Casbin v2 | |
| JWT | golang-jwt/jwt v5 | |
| 测试 | Go test + testify | |
| 前端 | React 18 + TypeScript + Vite | |
| 前端包管理 | pnpm | 禁用 npm install |

技术选型变更**必须先经 brainstorming + writing-plans**，不允许任务级擅自替换。

---

## 图谱查询指引

| 想知道 | 用 | 不要 |
|--------|----|------|
| 目录骨架、模块清单 | `get_architecture(aspects=["layout","modules"])` | 看本文件目录树 |
| 函数 / 类 / 路由位置 | `search_graph(label, name_pattern)` | `grep -r` |
| 调用链 / 影响面 | `trace_path(function_name, direction="both")` | 人肉追代码 |
| 跨服务边、HTTP 调用 | `query_graph("MATCH ... HTTP_CALLS ...")` | 看本文档 |
| 本地改动影响哪些符号 | `detect_changes()` | 看 git diff 猜 |

---

## Verification Commands

```bash
# 后端基线
cd server && go test ./...
cd server && go build ./...

# 前端基线
cd web && pnpm build

# 运行效果（按 project_rules.md VERIFICATION 章节）
docker compose -f deploy/docker-compose/docker-compose.yml build
docker compose -f deploy/docker-compose/docker-compose.yml up -d
```

---

## Platform Roadmap

| 阶段 | 内容 | 状态 |
|------|------|------|
| Phase 0 | User / Role / Permission / Audit | ✅ |
| Phase 0.5 | Audit Log Upgrade（查询 / 导出 / 保留策略） | ✅ |
| Phase 1 | Cloud Account + Unified Asset Registry | 进行中 |
| Phase 2 | DevOps Layer | 规划 |
| Phase 3 | FinOps Layer | 规划 |
| Phase 4 | Automation Layer | 规划 |

详细路线图见 `docs/roadmaps/finops-devops-platform-roadmap.md`。各阶段设计文档见 `docs/superpowers/specs/`。
