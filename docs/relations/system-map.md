# 系统关系图

本文件描述 Optimus 的主要目录与协作关系；项目规则以 [AGENT.MD](../../AGENT.MD) 为准，具体文件与调用链按需通过图谱核对。

## 根目录职责

- `server/`：Go + Gin 后端，包括路由、API、Service、Model、初始化与插件约定。
- `web/`：React + TypeScript + Vite 前端，包括路由、页面、共享状态、接口封装和组件。
- `deploy/`：部署与运行资产，当前 Compose 编排位于 `deploy/docker-compose/`。
- `docs/`：项目规则的结构化上下文、架构与产品设计；问题和需求在 GitHub Issues 跟进。

## 后端关系

```text
main → initialize → router → API → Service → Model
```

- `initialize/` 创建模块并注入依赖；`router/enter.go` 汇总 RouterGroup，各模块路由注册路径和中间件。
- `api/v1/` 处理 HTTP 输入、校验与响应；`service/` 处理业务和数据访问协作；`model/` 放置实体及请求/响应类型。
- 后端职责和注册规则详见 [后端分层规则](../modules/backend-layer-rules.md)。

## 前端关系

```text
router → view / components → api → 后端
             ↕
        auth / locale Context
```

- `web/src/router/` 组织路由和权限入口；`view/` 与 `components/` 承担页面和复用组件。
- `web/src/api/` 封装模块接口，公共请求由 `client.ts` 处理。
- `web/src/auth/` 和 `web/src/locale/` 提供现有共享上下文；React 组件状态与项目 Context 按实际范围使用。
- 中英文文案通过 `LanguageProvider` 和 `useI18n` 提供；导航与模块页面分别位于导航组件和对应 `view/<module>/`。
- 详细规则与接口边界见 [前端开发规范](../frontend-backend/frontend-rules.md) 和 [接口契约](../frontend-backend/boundary.md)。

## 插件边界

插件按项目规则分别放在 `server/plugin/<name>/` 和 `web/src/plugin/<name>/`，按现有装配方式组织；目录约定不等于当前所有插件能力均已实现。职责稳定后补充对应模块说明。

## 继续阅读

- 运行时、接口与技术栈：[当前架构](repo-profile.md)。
- 后端模块分层：[模块与分层](../modules/README.md)。
- 重要决定：[决策索引](../memory/long-term/decisions/README.md)。
