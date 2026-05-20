# Docker Compose 部署设计

## 背景

当前仓库已经具备可运行的 Go 后端骨架与一个最小可运行的前端壳子，但还没有任何正式的容器化部署产物。`deploy/docker-compose/` 目录已经存在，但目前为空，尚未定义前后端和数据库的统一编排方式。

本次目标不是建设完整生产级部署体系，而是先在 `deploy/docker-compose/` 下落一套最小可用的容器部署方案，打通：

- 前端静态构建与服务
- 后端 Go 服务构建与运行
- MySQL 容器提供数据库
- 前端通过 `nginx` 反向代理访问后端 API

## 目标

- 在 `deploy/docker-compose/` 下定义前后端与 MySQL 的最小可用部署结构
- 采用 `docker-compose` 编排 `frontend + backend + mysql`
- 前端使用“静态构建 + nginx”方式运行
- 后端使用 Go 构建后二进制运行
- 前端通过 `nginx` 同域代理 `/api` 到 backend
- 对外同时暴露 `nginx` 与 `backend` 端口

## 非目标

- 不构建完整生产级运维体系
- 不加入 `.env.example`
- 不加入部署 README
- 不加入初始化 SQL
- 不加入额外脚本文件
- 不加入反向代理之外的网关层
- 不在本轮中处理 HTTPS、域名、证书、监控、日志采集

## 已确认的设计决策

### 服务范围

本次 compose 只编排三个服务：

- `frontend`
- `backend`
- `mysql`

### 前端运行方式

前端采用：

- Node 阶段构建
- nginx 阶段运行静态产物

不采用：

- `vite preview`
- 额外 Node 静态服务

### 数据库策略

MySQL 本轮只负责提供数据库能力：

- 后端能够连接数据库即可
- 不执行初始化 SQL
- 不额外引入初始化脚本

### API 访问方式

前端访问后端推荐走：

- `nginx` 同域反向代理

即：

- 页面资源由 nginx 提供
- `/api` 请求由 nginx 转发到 backend

这样既避免跨域，也保持前端访问入口稳定。

### 对外暴露端口

对外同时暴露：

- `nginx` 端口：正常访问入口
- `backend` 端口：调试与排查入口

## 文件设计

本轮只新增四个文件：

- `deploy/docker-compose/docker-compose.yml`
- `deploy/docker-compose/Dockerfile.frontend`
- `deploy/docker-compose/Dockerfile.backend`
- `deploy/docker-compose/nginx.conf`

不新增其他部署文件。

## 容器职责设计

### 1. frontend

`frontend` 由 `Dockerfile.frontend` 负责，采用多阶段构建：

#### 构建阶段

- 使用 Node 镜像
- 安装 `web/` 依赖
- 执行前端构建命令
- 生成静态产物

#### 运行阶段

- 使用 nginx 镜像
- 拷贝前端构建产物
- 通过 `nginx.conf` 提供静态文件
- 同时代理 `/api` 到 backend

该服务在 compose 中对外暴露标准访问端口。

### 2. backend

`backend` 由 `Dockerfile.backend` 负责：

#### 构建阶段

- 使用 Go 镜像
- 在 `server/` 下拉取依赖并构建二进制

#### 运行阶段

- 使用轻量运行镜像
- 仅携带后端二进制与必要配置
- 容器启动后直接运行服务

backend 在 compose 中应：

- 通过环境变量接收数据库连接信息
- 通过 compose 网络访问 `mysql`
- 对外额外暴露端口，方便调试

### 3. mysql

`mysql` 直接使用官方镜像：

- 提供数据库能力
- 使用数据卷持久化
- 不挂初始化 SQL
- 不承担前端访问职责

## 连接关系设计

建议连接关系如下：

```text
Browser
  ├── http://nginx
  │     ├── /        -> frontend static assets
  │     └── /api/*   -> backend
  └── http://backend -> direct debug access

backend -> mysql
```

具体原则：

- 浏览器默认通过 `nginx` 访问前端
- 前端 API 统一通过同域 `/api` 访问
- `nginx` 负责把 `/api` 转发给 backend
- backend 通过服务名访问 mysql
- mysql 不直接暴露给前端

## Compose 设计原则

### 1. 保持最小可用

第一版 compose 只解决“能跑”：

- 前端能打开
- API 请求有明确代理路径
- backend 能连接 mysql

### 2. 保持可演进

虽然这次不加更多文件，但 compose 设计不能阻断后续扩展，例如：

- 后续加入 healthcheck
- 后续加入 env 文件
- 后续加入初始化 SQL
- 后续加入更正式的部署文档

### 3. 避免范围膨胀

本次不因为“顺手”把部署体系做成完整生产方案。

## 运行时配置建议

backend 需要的数据库配置建议通过 compose `environment` 注入，至少包括：

- 数据库主机
- 数据库端口
- 数据库名
- 用户名
- 密码

前端不直接持有 backend 容器地址，而是统一通过 nginx 代理访问 `/api`。

## 风险与控制

### 风险 1：前端直连 backend 导致入口分裂

如果前端直接请求 backend 暴露端口，会导致浏览器侧地址管理复杂，也容易产生跨域问题。

控制方式：

- 强制通过 nginx 代理 `/api`

### 风险 2：第一版部署文件过多

如果一开始加入 `.env.example`、README、SQL、脚本等，部署目录会立刻变重。

控制方式：

- 第一版只保留四个核心文件

### 风险 3：把调试配置和生产配置混在一起

同时暴露 nginx 与 backend 端口容易让默认部署边界变模糊。

控制方式：

- 对外主入口仍定义为 nginx
- backend 暴露端口明确标记为调试用途

### 风险 4：数据库初始化职责不清

如果在没有明确初始化策略前加入 SQL，会让部署边界变复杂。

控制方式：

- 当前版本 mysql 只提供数据库能力，不做初始化

## 验收标准

完成后应满足：

- `deploy/docker-compose/` 下只新增四个文件
- compose 中只包含 `frontend`、`backend`、`mysql`
- 前端通过 nginx 提供静态资源
- `/api` 由 nginx 代理到 backend
- backend 能通过 compose 网络访问 mysql
- nginx 与 backend 端口都对外暴露
- 不包含初始化 SQL、部署说明、额外脚本或 env 文件
