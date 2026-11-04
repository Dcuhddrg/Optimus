# AGENT.MD

## 目的

本文件是 Optimus 项目级 AI 规则的唯一正文。用户当前明确指令优先；`.codex/`、`.claude/` 等工具目录仅为兼容适配层。

## 读取顺序

1. 读取 `AGENT.MD`。
2. 读取 [docs/README.md](docs/README.md)。
3. 按任务读取相关 Issue、`docs/` 中的架构、模块、前后端、示例、记忆和设计文档。
4. 仅在工具需要时读取适配文件。

项目文件之间有冲突时以 `AGENT.MD` 为准。

## 仓库概览

- `server/`：Go + Gin 后端。
- `web/`：React + TypeScript + Vite 前端，使用 pnpm。
- `deploy/`：部署资产。
- `docs/`：AI 协作文档与项目设计。

## 工程规则

### 架构

- 所有模块及插件遵循 `Router → API → Service → Model`；
- enter.go 继续作为分组注册与组合入口
- API 层处理 HTTP 相关逻辑，Service 层不要依赖 gin.Context
- 对外接口的 Swagger 注释必须和真实行为保持一致
- 优先沿用 optimus 现有模式，不做无关的大改

### 前后端协作

- 明确请求与响应契约，保持现有接口、权限和项目隔离行为。
- JSON 响应照 gin-vue-admin：只用 `model/common/response` 的 `Ok*`/`Fail*`/`NoAuth`，业务结果为 HTTP 200 + `{code,data,msg}`（0 成功、7 失败），仅未认证为 HTTP 401；分页放在 data 内。详见[接口契约](docs/frontend-backend/boundary.md)。
- 前端请求通过 `web/src/api/`，view 不直接 `fetch`。
- 优先复用 `web/src/utils/` 和已有组件；跨端变更同步调用方和文档。

### 插件与模块

- 后端插件位于 `server/plugin/<name>/`。
- 前端插件位于 `web/src/plugin/<name>/`。
- 复用现有模块装配和权限机制，独立注册或自动加载先设计确认。
- 稳定的模块职责、入口和边界放入 `docs/modules/` 或相关设计。

### 示例文档

- [docs/examples/](docs/examples/README.md) 是讲解型示例层。
- 示例说明项目标准，不代表相关功能已经实现；上游模板不按项目现状反向改写。
- 新增某层文件前阅读对应示例、模块设计和同类实现。

### 记忆规则

- 业务需求、问题和反馈通过 [GitHub Issues](https://github.com/Dcuhddrg/Optimus/issues) 跟踪。
- 稳定偏好、协作方式和已确认设计放在 `docs/`，不另建进度汇总。
- 记录与维护细节见[文档约定](docs/relations/documentation-workflow.md)。

### 文档维护

- `AGENT.MD` 保留高层、稳定的规则，细节放入 `docs/`；修改项目规则时先改本文件，再按需更新适配层。
- 设计、计划和历史依据按[文档约定](docs/relations/documentation-workflow.md)维护。

### 范围与执行

- 用户负责目标、优先级和架构取舍；助手负责核查、实现、验证和记录。反馈不等于批准的需求，业务验收由用户确认。
- 已授权、清晰且可逆的任务直接执行，小步修改；非平凡任务先明确规格并制定计划。未经授权不扩需求、跨模块改动、重写结构或更换技术栈与依赖版本；仅添加任务必要的依赖。
- 不回退用户或其他任务的改动；遇到任务冲突、范围外严重问题或无法确定边界时暂停说明。不要删除注释、TODO、占位说明或修改 `.env`、`.env.example`、部署及其他环境配置；新增前端一级导航或扩大页面范围须确认。

### 代码读取约束

- 代码搜索、源码读取和影响分析统一使用 `codebase-memory-mcp`，项目名为 `home-lilin-trae-project-Optimus`。索引、允许回退条件及派发要求见[图谱查询](docs/relations/engineering-workflow.md#图谱查询)；已知路径或进入实现阶段不构成回退理由。
- 不直接读取 `node_modules/` 实现；优先查看项目调用、锁文件、配置、官方文档或公开类型。

### 验证

- 声明完成前取得实际证据，报告命令、结果、跳过项和阻塞；源码存在、测试通过和用户验收分别记录。
- 后端基线：`cd server && go test ./...`、`cd server && go build ./...`；前端基线：`cd web && pnpm build`。
- 运行效果默认通过 Compose 构建后的 nginx 验证；仅明确需要热更新或排查 Vite 时使用 `pnpm dev`。数据库、路由、文档及其他验证步骤见[工程操作参考](docs/relations/engineering-workflow.md#验证步骤)。

## AI 文档索引

[文档入口](docs/README.md) · [系统关系](docs/relations/README.md) · [模块与分层](docs/modules/README.md) · [前后端协作](docs/frontend-backend/README.md) · [开发示例](docs/examples/README.md) · [项目记忆](docs/memory/README.md) · [工程操作](docs/relations/engineering-workflow.md)
