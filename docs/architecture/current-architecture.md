# Current Architecture

## Scope

- 当前主开发面是 `server/` 下的 Go 后端骨架
- `web/` 目录仍是占位目录，不是主要交付面
- 当前阶段目标是可运行、可验证、可扩展的后端骨架，不是完整业务系统

## Repository Layout

- `server/`：当前主要实现目录与默认工作面
- `server/config`：配置结构与默认配置文件
- `server/core`：配置加载、日志初始化、HTTP 服务启动、启动输出
- `server/global`：进程级共享对象
- `server/initialize`：基础设施与路由装配
- `server/router`：路由聚合与注册
- `server/api/v1/system`：系统相关 HTTP handlers
- `server/service/system`：系统相关业务响应封装
- `server/middleware`：日志、恢复、JWT/Casbin 占位中间件
- `server/model/common`：公共响应模型
- `server/utils`：轻量响应工具与 claims
- `docs/superpowers`：设计与实施计划文档
- `.codex`：Codex 本地规则文件

## Backend Architecture

- `main`：唯一程序入口，只负责初始化系统并启动服务
- `core`：运行时核心能力，包括配置、日志、HTTP 服务启动
- `initialize`：负责把配置、依赖和路由接起来
- `router`：只负责路由分组与挂载
- `api`：处理 HTTP 请求与响应输出
- `service`：承载当前最小业务逻辑与响应数据组织
- `model/common`：放共享响应模型
- `utils`：放轻量工具，不承担大型业务逻辑
- `global`：放配置、日志、数据库、Redis、Casbin 等共享实例

当前默认分层流向为：

`main -> initialize/core -> router -> api -> service`

## Runtime Flow

当前运行时流程如下：

1. `core.Viper()` 读取配置
2. 使用 `server/config.yaml` 作为默认配置文件
3. 通过 `OPTIMUS_` 前缀环境变量覆盖配置项
4. `initialize.OtherInit()` 完成基础环境准备
5. `core.Zap()` 初始化日志
6. `initialize.Routers()` 构建 Gin 路由
7. `core.RunServer(...)` 启动 HTTP 服务

可选依赖：

- Redis 启动时可选
- Casbin 启动时可选

## Tech Stack

| 层 | 技术 |
|----|------|
| 语言 | Go 1.23 |
| HTTP 框架 | Gin 1.10 |
| ORM | GORM 1.25 |
| 数据库目标 | MySQL |
| 配置 | Viper 1.19 |
| 日志 | Zap 1.27 |
| Redis 客户端 | go-redis/v9 |
| 权限 | Casbin v2 |
| JWT | github.com/golang-jwt/jwt/v5 |
| 测试 | Go test + testify |

## Public Endpoints & Verification

当前公开接口：

- `GET /api/v1/health`
- `GET /api/v1/ping`
- `GET /api/v1/base/ping`

常用验证命令：

- `cd server && go test ./...`
- `cd server && go build ./...`
- `cd server && go run .`
- `cd server && OPTIMUS_SYSTEM_ADDR=18089 go run .`

## Non-goals

当前尚未实现，也不应被误判为已存在的内容：

- 完整登录鉴权体系
- 用户、角色、菜单等完整系统域
- Swagger
- 代码生成器
- 完整后台联动能力
- 成熟前端应用结构
