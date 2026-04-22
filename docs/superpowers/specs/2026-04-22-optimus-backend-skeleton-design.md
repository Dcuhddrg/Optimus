# Optimus 后端项目骨架设计

## 背景

Optimus 当前仓库仍处于空骨架阶段，根目录仅包含 `server/`、`web/` 与极简说明文件，后端尚未初始化 Go 模块、目录结构与基础启动链路。

本次目标不是直接实现业务功能，而是先建立一套可运行、可扩展、贴近 `gin-vue-admin` 组织方式的后端项目骨架，为后续登录鉴权、权限管理、系统模块与业务模块开发提供稳定基础。

## 目标

- 以 `server/main.go` 作为唯一程序入口
- 参考 `gin-vue-admin` 的目录组织方式构建后端骨架
- 以 `server/core` 作为核心能力层
- 以 `server/router` 作为路由注册层
- 完成 `Gin + Viper + Zap + GORM(MySQL) + Redis + JWT + Casbin` 的可插拔初始化骨架
- 首版只提供健康检查与基础 ping 示例接口
- 支持 `config.yaml`，并允许环境变量覆盖

## 非目标

- 不实现登录、注册、用户、角色、菜单等完整系统模块
- 不引入插件系统、任务系统、资源系统、多数据库并行管理
- 不在首版中加入自动建表、初始化数据灌库、默认管理员账户
- 不在首版中实现完整 Swagger、代码生成器或管理后台联动能力

## 技术约束

- Go 1.23
- Gin 1.10.0
- GORM 1.25.12
- Casbin 2.103.0
- Viper 1.19.0
- Zap 1.27.0
- Redis 9.7.0
- JWT 5.2.2
- 默认数据库为 MySQL
- 配置方式为 `config.yaml + 环境变量覆盖`

## 设计原则

- 尽量贴近 `gin-vue-admin/server` 的主干结构，降低后续迁移和上手成本
- 首版只保留必要骨架，避免空项目一开始就引入过重层次
- 入口层、初始化层、核心层、路由层职责分离
- 业务模块先跑通最小分层链路，不提前设计完整系统域
- 默认启动流程应可读、可测、可扩展，不隐藏高副作用行为

## 目录结构

建议首版 `server/` 目录如下：

```text
server/
├── main.go
├── api/
│   └── v1/
│       └── system/
├── config/
│   ├── config.go
│   ├── system.go
│   ├── jwt.go
│   ├── redis.go
│   ├── gorm_mysql.go
│   └── casbin.go
├── core/
│   ├── viper.go
│   ├── zap.go
│   ├── server.go
│   └── banner.go
├── docs/
├── global/
│   ├── global.go
│   └── version.go
├── initialize/
│   ├── init.go
│   ├── gorm.go
│   ├── gorm_mysql.go
│   ├── redis.go
│   ├── casbin.go
│   ├── router.go
│   └── validator.go
├── middleware/
│   ├── logger.go
│   ├── jwt.go
│   ├── casbin_rbac.go
│   └── recovery.go
├── model/
│   ├── common/
│   └── system/
├── router/
│   ├── enter.go
│   └── system/
├── service/
│   ├── enter.go
│   └── system/
└── utils/
    ├── response.go
    └── claims.go
```

## 分层说明

### main.go

`main.go` 只负责两件事：

- 初始化系统依赖
- 启动 HTTP 服务

不得在入口中直接编写数据库逻辑、路由细节或业务逻辑。

### core

`core/` 负责核心能力实现，包括：

- 配置读取
- 日志初始化
- HTTP Server 启动
- 运行时启动信息输出

`core/` 不直接承载具体业务接口定义。

### initialize

`initialize/` 负责系统装配，承担“把组件接起来”的职责，包括：

- 初始化 GORM
- 初始化 Redis
- 初始化 Casbin
- 初始化校验器
- 初始化路由

这一层参考 `gin-vue-admin` 的使用方式，是后端骨架是否整齐的关键。

### router

`router/` 负责路由聚合与分组注册：

- `router/enter.go` 作为总入口
- `router/system/` 作为首版系统模块路由分组

路由层只组织路由和中间件挂载，不承载业务逻辑。

### api

`api/` 负责 HTTP handler，处理：

- 请求参数绑定
- 输入校验
- 调用 service
- 输出统一响应

首版仅提供 `system` 模块的基础接口。

### service

`service/` 负责业务流程组织，不直接依赖 Gin 上下文。

首版只保留系统基础服务，用于跑通分层结构。

### model

`model/` 负责结构体模型定义。首版建议按 `common/` 与 `system/` 拆分，为后续用户、角色、权限等系统模型扩展留位。

### middleware

`middleware/` 负责 Gin 中间件，包括：

- 请求日志
- panic 恢复
- JWT 鉴权
- Casbin 权限校验

首版要求先完成接线和可插拔能力，不要求完整授权流程。

### global

`global/` 负责全局共享对象，例如：

- 配置对象
- 日志实例
- GORM DB
- Redis 客户端
- Casbin Enforcer

这样可以延续 `gin-vue-admin` 风格，降低初始化与调用成本。

### utils

`utils/` 放与业务无关的基础工具，首版仅建议包含：

- 统一响应工具
- JWT Claims 工具

避免过早形成一个无边界的大工具目录。

## 启动链路

首版启动顺序建议如下：

1. `core.Viper()` 读取配置，并写入 `global`
2. `initialize.OtherInit()` 做基础环境准备
3. `core.Zap()` 初始化日志
4. `initialize.Gorm()` 初始化 MySQL 连接
5. `initialize.Redis()` 初始化 Redis 客户端
6. `initialize.Casbin()` 初始化 Casbin Enforcer
7. `initialize.Routers()` 构建 Gin Engine
8. `core.RunServer()` 启动 HTTP Server

该链路遵循“先基础设施，后 HTTP 暴露”的顺序，避免启动流程变得混乱。

## 配置设计

`server/config/` 下按模块拆分配置结构体，再通过聚合结构统一管理。首版建议包含以下配置块：

- `System`
- `Zap`
- `Mysql`
- `Redis`
- `JWT`
- `Casbin`

### System

建议字段：

- `env`
- `addr`
- `db-type`
- `use-redis`
- `use-casbin`
- `router-prefix`

### Mysql

建议字段：

- `path`
- `port`
- `db-name`
- `username`
- `password`
- `config`
- `max-idle-conns`
- `max-open-conns`
- `log-mode`

### Redis

建议字段：

- `addr`
- `password`
- `db`

### JWT

建议字段：

- `signing-key`
- `expires-time`
- `buffer-time`
- `issuer`

### Casbin

建议字段：

- `enable`
- `model-path`
- `policy-source`

### 配置加载策略

- 默认读取 `server/config.yaml`
- 支持通过环境变量覆盖配置
- 环境变量采用统一前缀，例如 `OPTIMUS_SYSTEM_ADDR`

此方式兼顾本地开发可读性和部署环境可配置性。

## 路由设计

首版保留最小路由集合：

- `GET /health`
- `GET /base/ping`

推荐统一支持可配置前缀，例如：

- `/api/v1/health`
- `/api/v1/base/ping`

其中：

- 公开路由组用于健康检查和基础示例接口
- 私有路由组预留 JWT 与 Casbin 中间件接入位

首版不要求必须具备登录态即可验证服务启动成功。

## 首版交付范围

首版后端骨架应包含：

- 可执行的 `main.go`
- 完整 `go.mod`
- 基础配置结构与 `config.yaml`
- `Viper / Zap / Gorm(MySQL) / Redis / Casbin` 初始化代码
- `global` 全局对象
- `Gin` 路由注册代码
- 健康检查接口与基础 ping 接口
- 统一响应工具
- 至少一条覆盖路由注册或基础处理器的测试

## 不采纳的方案

### 方案一：轻量单体骨架

使用 `cmd / handler / service / repo / model` 的普通分层方式，初期实现更快，但与 `gin-vue-admin` 风格偏差较大，不利于维持你指定的目录组织习惯。

### 方案二：重型 DDD 骨架

一开始就拆分 `domain / application / infrastructure / interfaces` 等复杂层次，虽然扩展性强，但对于当前空仓库属于过度设计。

## 风险与注意事项

- 使用 `global/` 会降低依赖传递成本，但后续测试隔离需要注意全局状态重置
- Casbin 首版仅做初始化与中间件接线，不应过早承诺完整 RBAC 行为
- Redis、Casbin 等组件应允许按配置关闭，避免本地启动门槛过高
- 首版不做自动建表，后续引入迁移方案时需要明确数据库生命周期策略

## 验收标准

- `server/main.go` 可作为后端唯一启动入口
- 目录结构与 `gin-vue-admin` 主干组织方式保持一致
- 服务可正常启动
- 配置文件可读取，且支持环境变量覆盖
- MySQL 初始化链路存在且结构完整
- Redis、JWT、Casbin 初始化能力可插拔
- 路由可注册成功，至少暴露健康检查和 ping 接口
- 代码结构支持后续直接接入登录、权限和系统模块
