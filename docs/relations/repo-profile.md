# 项目画像与运行边界

> 规范依据：[AGENT.MD](../../AGENT.MD) 是项目规则唯一正文；本文展开其架构职责和边界，发生冲突时以根规则为准。
> 现行架构依据见 [ADR-002](../memory/long-term/decisions/ADR-002-backend-structure-migration.md)，迁移结果见 [#11](https://github.com/Dcuhddrg/Optimus/issues/11)，历史验证基线见 [#2](https://github.com/Dcuhddrg/Optimus/issues/2)。

> 本文件维护项目定位、技术栈与运行时边界；系统结构和工程操作分别见 [系统关系图](system-map.md) 与 [工程操作参考](engineering-workflow.md)。

---

## 项目定位

Optimus 以治理、云资源、运维和成本能力为长期方向；当前实现与验证边界以相关 Issue/PR 的记录为依据，历史基线见 [#2](https://github.com/Dcuhddrg/Optimus/issues/2)，不能依据路线图推断功能已交付。

## 系统关系

目录职责、后端分层和前端协作流程见 [系统关系图](system-map.md)。

---

## API 设计契约

- 前后端响应、分页、错误和兼容边界见 [接口契约](../frontend-backend/boundary.md)。
- 统一前缀 `/api/v1/<module>/...`
- 健康检查：`/api/v1/health`、`/api/v1/ping`（公开）
- 鉴权：登录 `/api/v1/<module>/auth/login`（公开），其它路径走 JWT + Casbin RBAC（私有）
- 路由必须按 public / private 分组注册（参考 `router/system/`）
- 端点发现方式见 [工程操作参考](engineering-workflow.md)。

---

## Deployment 运行时契约

- 三服务：frontend（nginx）+ backend（Go）+ mysql，由 `deploy/docker-compose/docker-compose.yml` 编排
- nginx 反向代理 `/api/` → `optimus-server:8888`
- Mac 端访问域名 `optimus.intellivis.cn`（Caddy HTTPS），直连排查走 `http://192.169.50.100:8080/`
- 运行验证步骤见 [工程操作参考](engineering-workflow.md)。

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

## 工程操作

图谱查询与按条件触发的验证步骤见 [工程操作参考](engineering-workflow.md)；基础验证要求见 [AGENT.MD](../../AGENT.MD#验证)。

---

## Platform Roadmap

长期阶段见 [路线图](../memory/long-term/platform-roadmap.md)，各功能实现、验证及验收结果在对应 [GitHub Issue](https://github.com/Dcuhddrg/Optimus/issues) 中跟进，决策理由见 [决策索引](../memory/long-term/decisions/README.md)。本文件维护架构契约，不重复维护项目进度。
