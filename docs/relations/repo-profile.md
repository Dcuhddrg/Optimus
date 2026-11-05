# 项目画像与运行边界

项目规则以 [AGENT.MD](../../AGENT.MD) 为准；系统结构和工程操作分别见 [系统关系图](system-map.md) 与 [工程操作参考](engineering-workflow.md)，模块分组见[模块说明索引](../modules/module-index.md)。

## 项目定位

Optimus 以治理、云资源、运维和成本能力为[长期方向](../memory/long-term/platform-roadmap.md)。当前实现和验证以相关 Issue/PR 为依据，历史基线见 [#2](https://github.com/Dcuhddrg/Optimus/issues/2)；架构决定及迁移记录见 [ADR-002](../adr/0002-backend-structure-migration.md) 和 [#11](https://github.com/Dcuhddrg/Optimus/issues/11)。

## API 设计契约

- 前后端响应、分页、错误和兼容边界见 [接口契约](../frontend-backend/boundary.md)。
- 默认前缀 `/api/v1/<module>/...`，可通过配置调整
- 健康检查：`/api/v1/health`、`/api/v1/ping`（公开）；数据库不可用或启动迁移/种子失败时服务仍启动，但 `/health` 返回 503 `{"status":"unhealthy"}`（错误详情只写日志），`/ping` 始终 200 作存活探针
- 鉴权：登录 `/api/v1/system/auth/login`（公开）；私有路由使用 JWT + Casbin RBAC
- 路由必须按 public / private 分组注册（参考 `router/system/`）
- 端点发现方式见 [工程操作参考](engineering-workflow.md)。

## Deployment 运行时契约

- 三服务：frontend（nginx）+ backend（Go）+ mysql，由 `deploy/docker-compose/docker-compose.yml` 编排
- nginx 反向代理 `/api/` → `optimus-server:8888`
- Mac 端访问域名 `optimus.intellivis.cn`（Caddy HTTPS），直连排查走 `http://192.169.50.100:8080/`
- 运行验证步骤见 [工程操作参考](engineering-workflow.md)。

## Tech Stack

| 层 | 选型 | 备注 |
|----|------|------|
| 语言 | Go 1.24 | 后端 |
| HTTP | Gin 1.10 | |
| ORM | GORM 1.31 | |
| DB | MySQL | |
| 配置 | Viper 1.19 | |
| 日志 | Zap 1.27 | |
| 权限 | Casbin v3 + gorm-adapter（casbin_rule） | |
| JWT | golang-jwt/jwt v5 | |
| 测试 | Go test + testify | |
| 前端 | React 18 + TypeScript + Vite | |
| 前端包管理 | pnpm | 禁用 npm install |

技术选型变更**必须先写设计并经用户确认**，不允许任务级擅自替换。
