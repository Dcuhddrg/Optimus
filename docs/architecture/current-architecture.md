# Current Architecture

> 规范依据：[AGENT.MD](../../AGENT.MD) 是项目规则唯一正文；本文展开其架构职责和边界，发生冲突时以根规则为准。
> 最近规范整理：2026-10-27，关联 [Issue #6](https://github.com/Dcuhddrg/Optimus/issues/6)。当前规范、存量实现与后续迁移分别说明，不以文档更新宣称代码已迁移。

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

遵循根规则确认的统一逻辑分层：

```text
Router -> API -> Service -> Model
```

| 层 | 职责与位置 | 依赖边界 |
| --- | --- | --- |
| Router | `server/router/<module>/`，注册路径、中间件与 public/private 分组 | 连接 API，不承担业务处理 |
| API | `server/api/v1/<module>/`，绑定和校验 HTTP 输入、调用服务、组织响应 | 通过 Service 处理业务，不绕过服务直接访问持久化 |
| Service | `server/service/<module>/`，业务规则、流程编排和数据访问协作 | 使用 Model 及所需基础设施，不依赖 `gin.Context`，不反向调用 API/Router |
| Model | `server/model/<module>/`，实体、请求/响应数据结构和持久化映射 | 不反向依赖 Router/API/Service |

- 保留现有 `enter.go` 分组注册和组合入口，新增分组沿用对应模块的注册方式；不得绕过已有入口另建平行注册链。
- `main` 与 `initialize` 负责启动和模块装配；它们不改变上述业务职责划分。已有初始化、依赖注入和数据访问封装按任务复用，不要求重写。
- Service 不接收 `gin.Context`；HTTP 处理留在 API，所需业务参数显式传入。Swagger 注释与真实接口行为保持一致。
- 新模块按上述四层职责组织；具体文件与装配示例由后续 [模块开发指引事项 #8](https://github.com/Dcuhddrg/Optimus/issues/8) 整理，不将存量内部目录强加为新的必选层。

### 存量实现与历史方案

2026-10-27 通过 codebase-memory 确认，`server/service/system/`、`server/service/project/`、`server/service/cloud/` 下仍存在 `usecase/` 和 `infra/`。这是目录级事实核对，不代表已全面检查各模块行为或完成验收。

这些目录属于存量服务内部组织。旧文档中的“usecase 仅标准库、infra 实现 Store、所有新模块必须拆成这套目录”等要求，已由根规则的统一分层取代；原文及取舍依据保留在 [ADR-001](decisions/ADR-001-backend-layering.md)。

维护存量模块时先核对其已有接口、初始化和依赖关系；本次只统一文档，不删除目录、不改变接口或注入流程。代码迁移需要另行明确范围与验证条件，不能因为规则更新自动重构模块。

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
- 前端运行效果验证**默认走 Compose**，不依赖 Vite dev server（详见 [工程操作参考](../governance/engineering.md) 的验证步骤）

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

# 运行效果（按工程操作参考的验证步骤）
docker compose -f deploy/docker-compose/docker-compose.yml build
docker compose -f deploy/docker-compose/docker-compose.yml up -d
```

---

## Platform Roadmap

最近状态核查：2026-10-26。历史阶段完成标记不替代当前验收证据。

长期阶段见 [路线图](../roadmaps/finops-devops-platform-roadmap.md)，当前实现及验证状态统一见 [能力状态](../project-status.md)，决策理由见 [决策索引](decisions/README.md)。本文件维护架构契约，不重复维护项目进度。
