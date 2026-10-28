# AGENT.MD

## 目的

本文件是 Optimus 仓库内 AI 协作规则的唯一正文。`.codex/`、`.claude/` 下的规则仅为兼容适配层，不保存独立 project rule 副本。用户当前明确指令优先；仓库内规则冲突以本文件为准。
本文件规定新旧模块共同遵循的架构目标；存量差异必须列入迁移清单，分批整改。文档更新不代表代码已完成迁移，迁移不得混入无关功能改动。

## 读取顺序

1. `AGENT.MD`。
2. [docs/README.md](docs/README.md)。
3. 按任务读取相关 GitHub Issue、能力状态、架构、模块设计及实施计划，不加载全部上下文。
4. 仅在当前工具确实依赖时，读取工具目录下的适配文件。

## 仓库概览

- `server/`：Go + Gin 后端
- `web/`：React + TypeScript + Vite 前端，使用 pnpm
- `deploy/`：部署资产
- `docs/`：架构、设计、工程上下文与索引

## 工程规则

### 架构

- 新模块与存量模块统一采用 `Router -> API -> Service -> Model`，同时统一目录、包依赖与注册入口。
- 旧 `usecase/infra` 必须纳入 [#11](https://github.com/Dcuhddrg/Optimus/issues/11) 的迁移清单，不再作为长期可选结构保留；目标与验收边界见 [ADR-002](docs/relations/decisions/ADR-002-backend-structure-migration.md)。
- 结构迁移保持现有接口及权限行为，接口协议变更按独立 Issue 实施，不与结构迁移混批。
- `enter.go` 继续作为分组注册与组合入口。
- API 层处理 HTTP 相关逻辑，Service 层不要依赖 `gin.Context`。
- 对外接口的 Swagger 注释必须和真实行为保持一致。
- 优先沿用 optimus 现有模式，复用已有模块和基础设施，不做无关大改。

### 前后端协作

- 明确请求与响应契约，保持统一响应结构：`{ code, data, msg }`。
- 保持统一分页结构：`{ page, pageSize, total, list }`。
- 前后端字段名、数据类型和错误处理保持一致；接口变更同步调用方和文档。
- 请求通过 `web/src/api/` 封装，view 不直接 `fetch`；优先复用 `web/src/utils/` 的工具及现有组件。
- 跨栈边界变更同步相关模块设计、接口约定和能力证据；未要求行为变化时保留业务逻辑。

### 插件与模块

- 后端插件放在 `server/plugin/<name>/`，前端插件放在 `web/src/plugin/<name>/`。
- 插件沿用现有注册和装配方式；模块职责、入口和边界写入 `docs/modules/` 或相关 spec。
- 新增模块或跨模块改动先明确边界，避免绕过既有 API、Service 和注册入口。

### 示例与文档维护

- 新增某层文件前先阅读对应模块设计和同类实现；示例用于说明项目标准，不要求逐字复制。
- 本文件只保存高层、稳定、跨任务规则，工具用法和验证细节按需读取 [工程操作参考](docs/relations/engineering-workflow.md)。
- spec / plan 放在 `docs/superpowers/specs/`、`docs/superpowers/plans/`；已交付阶段归档到 `docs/superpowers/archive/`，待核实项保持活跃。
- 项目规则先改本文件，再按需更新适配层；相关设计保留决定理由、历史及替代关系，状态维护遵循 [文档维护约定](docs/relations/documentation-workflow.md)。

### 需求与协作记忆

- 每项业务需求、问题和反馈在 [GitHub Issues](https://github.com/Dcuhddrg/Optimus/issues) 记录、讨论和跟进；业务需求包括模块、接口、页面、流程、业务规则及插件能力的新增或修改。
- 长期稳定偏好提炼到本文件或对应设计，不只留在聊天中；Issue 记录单项需求，docs 保存确定后的设计和能力证据，不双写问题状态。
- README、映射文件只承担入口和索引职责；更新具体 Issue 或相关文档，不用重写汇总正文覆盖历史记录。远端不可用时说明阻塞。

### 范围与执行

- 已授权、清晰且可逆的任务直接执行，先明确文件边界，小步修改；非平凡任务按 brainstorming 确认规格 → writing-plans → 执行推进，单文件小改动可直接实施。
- 未经授权不扩需求、修改任务外文件、跨模块改动、重写结构、更换技术栈或依赖版本；仅添加任务必要的依赖、测试和文档。
- 不删除注释、TODO、占位说明，不改 `.env`、`.env.example`、部署或其他环境配置；新增前端一级导航或扩大页面范围须确认。
- 不回退用户或其他任务的未提交改动；任务与现状冲突、发现范围外严重问题、无法确定修改模块或需要突破边界时，暂停提问并说明原因。额外问题登记反馈，不顺手实现。

### 代码读取约束

- 代码探索优先 codebase-memory-mcp，项目为 `home-lilin-trae-project-Optimus`；状态不明先查 `index_status`，拉取主干、大重构或批量新增后用 `detect_changes` 核对漂移并更新落后索引。
- 探索前读取 [图谱操作参考](docs/relations/engineering-workflow.md#图谱查询)；文本工具仅用于非代码文件、文字查找或确认图谱未覆盖的新文件，不先宽泛扫描源码。派发子代理须传递相同要求。
- 不直接读取 `node_modules/` 中的实现代码；第三方行为先看项目调用、锁文件、配置、官方文档和公开类型/说明，不逐文件翻依赖实现。

### 验证

- 成功声明前取得实际结果并说明命令、覆盖范围、跳过项和阻塞；源码存在、测试通过、产品验收分别记录。
- 后端基线：`cd server && go test ./...`、`cd server && go build ./...`；前端基线：`cd web && pnpm build`。
- 运行效果默认通过 Compose 构建后的 nginx 验证，仅明确需要热更新或排查 Vite 时用 `pnpm dev`；运行效果、Mac 访问、启动/路由/API 和依赖准备按工程操作参考执行。
- 文档和规则检查文件、结构、引用、变更边界及 `git diff --check`；需要业务验收的 Issue 验收后再关闭。

## 文档索引

- [文档入口](docs/README.md) · [能力状态](docs/project-status.md) · [平台路线图](docs/memory/platform-roadmap.md)
- [系统关系](docs/relations/README.md) · [模块与分层](docs/modules/README.md) · [前后端协作](docs/frontend-backend/README.md) · [决策索引](docs/relations/decisions/README.md)
- [文档维护](docs/relations/documentation-workflow.md) · [工程操作](docs/relations/engineering-workflow.md)
- [设计](docs/superpowers/specs/) · [计划](docs/superpowers/plans/) · [归档](docs/superpowers/archive/)

