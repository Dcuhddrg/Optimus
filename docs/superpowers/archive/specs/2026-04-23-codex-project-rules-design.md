# Codex Project Rules 优化设计

## 背景

当前 `.codex/project_rules.md` 已经从通用模板收敛到面向 Codex 的规则文件，但仍然混入了较多仓库架构事实、技术栈说明与目录介绍，导致“行为约束”和“项目现状”耦合在同一份文档中。

这种混合写法有两个直接问题：

- 对 Codex 来说，规则文件信号不够集中，执行时容易把项目介绍误当成行为指令
- 对仓库维护者来说，规则和架构事实更新频率不同，长期容易漂移

本次优化目标不是增加更多规则，而是重新划分职责边界：把 `.codex/project_rules.md` 变成纯行为契约，把仓库当前架构迁移到 `docs/architecture/`。

## 目标

- 将 `.codex/project_rules.md` 收缩为纯 Codex 行为规则
- 将当前仓库架构、技术栈、运行链路与验证命令迁移到 `docs/architecture/current-architecture.md`
- 强化两类能力：
  - 更强约束：减少 Codex 越界、顺手多做、擅自扩需求
  - 更强执行：让 Codex 默认更主动完成任务、补齐验证、减少不必要确认
- 建立规则文档与架构文档之间的单向引用关系，避免内容重复

## 非目标

- 不调整 `server/` 当前代码结构
- 不新增新的业务模块、接口或测试策略
- 不修改 `README.md`、`docs/superpowers/` 既有设计/计划文档
- 不把 `docs/architecture/` 拆成多份子文档

## 设计原则

- 规则文件只回答“Codex 应该怎么做”
- 架构文档只回答“当前仓库是什么”
- 同一事实只保留一个权威来源，避免镜像复制
- 优先增强执行稳定性，而不是扩充文档篇幅
- 对当前仓库阶段保持克制，不提前为未来复杂系统设计过重结构

## 产物设计

### 1. `.codex/project_rules.md`

该文件只保留面向 Codex 的操作约束，不再承载项目架构正文。

建议保留以下五个区块：

#### Boundaries

定义硬边界，明确在无明确指令时 Codex 不能做什么，包括：

- 不修改任务未涉及文件
- 不擅自更改技术选型
- 不删除注释或 TODO
- 不自动升级依赖
- 不修改环境配置文件
- 不跨越模块边界
- 不顺手扩展需求

同时列出必须暂停并提问的场景：

- 需要违反硬边界
- 任务描述与仓库现状冲突
- 发现严重问题但不在本次范围
- 不确定修改落点

#### Default Execution

把 Codex 的默认执行顺序写清楚：

1. 先复述任务范围
2. 明确将修改的文件边界
3. 无歧义时直接执行，而不是先给长篇建议
4. 优先完成用户要求的任务本身
5. 潜在问题放到任务完成后的说明部分

这一节的重点是强化“默认执行”，避免规则文件只强调保守、不强调完成。

#### Verification

定义默认验证基线，当前仓库以后端为主，因此至少包括：

- `cd server && go test ./...`
- `cd server && go build ./...`

如果任务涉及启动、路由或接口行为，再追加：

- 服务启动验证
- `/api/v1/health` 可用性验证
- 相关 ping 路由 JSON 验证

#### Output Contract

约束 Codex 的输出习惯：

- 对话使用中文
- 汇报尽量简洁、可扫描
- 不写“顺便优化了什么”
- 不在未被要求时展开大段替代方案
- 默认以固定结构汇报：

```text
## 变更说明
## 影响范围
## 注意事项
```

#### Architecture Reference

只保留一句短引用，例如：

> 当前仓库架构、技术栈、启动链路与验证命令见 `docs/architecture/current-architecture.md`

这一节不重复架构正文。

### 2. `docs/architecture/current-architecture.md`

该文件作为当前仓库架构的权威说明，描述“当前事实”，不描述历史方案推演。

建议包含以下章节：

#### Scope

- 当前主开发面是 `server/`
- `web/` 仍是占位目录
- 当前阶段目标是可运行的 Go 后端骨架，而不是完整业务系统

#### Repository Layout

只列关键目录及职责，例如：

- `server/config`
- `server/core`
- `server/global`
- `server/initialize`
- `server/router`
- `server/api/v1/system`
- `server/service/system`
- `server/middleware`
- `server/model/common`
- `server/utils`
- `docs/superpowers`
- `.codex`

#### Backend Architecture

说明当前默认分层与边界：

- `main`
- `core`
- `initialize`
- `router`
- `api`
- `service`
- `model/common`
- `utils`
- `global`

重点描述职责，不展开具体实现细节。

#### Runtime Flow

写清当前启动链路与配置机制：

- `server/config.yaml`
- `OPTIMUS_` 环境变量覆盖
- Redis/Casbin 可选初始化
- 路由构建与 HTTP 服务启动流程

#### Tech Stack

明确当前实际技术栈：

- Go 1.23
- Gin 1.10
- GORM 1.25
- MySQL
- Viper 1.19
- Zap 1.27
- go-redis/v9
- Casbin v2
- `github.com/golang-jwt/jwt/v5`
- Go test + testify

#### Public Endpoints & Verification

列出当前公开接口与常用验证命令：

- `GET /api/v1/health`
- `GET /api/v1/ping`
- `GET /api/v1/base/ping`
- `cd server && go test ./...`
- `cd server && go build ./...`
- `cd server && go run .`
- `cd server && OPTIMUS_SYSTEM_ADDR=<port> go run .`

#### Non-goals

明确当前尚未实现的内容，避免后续 agent 高估系统成熟度，例如：

- 完整登录鉴权体系
- 用户/角色/菜单管理
- Swagger
- 代码生成器
- 完整后台联动

## 文档边界与同步策略

为避免未来漂移，采用以下规则：

- `.codex/project_rules.md` 不再复制架构内容
- `docs/architecture/current-architecture.md` 不再复制行为规则
- 当仓库结构变化时，优先更新 `docs/architecture/current-architecture.md`
- 当 Codex 行为要求变化时，只更新 `.codex/project_rules.md`

## 风险与控制

### 风险 1：规则文件再次膨胀

如果后续继续把技术细节写回 `.codex/project_rules.md`，会重新回到职责混乱状态。

控制方式：

- 规则文件限定为行为契约
- 新增仓库事实时，只补到 `docs/architecture/current-architecture.md`

### 风险 2：架构文档变成历史流水账

如果把 brainstorming、plan、实现历史不断追加进架构文档，文档会失去“当前状态总览”的价值。

控制方式：

- 架构文档只描述当前事实
- 历史方案与计划继续留在 `docs/superpowers/`

### 风险 3：执行约束过强导致 Codex 变得被动

如果只强调限制，不强调默认执行，Codex 可能变得保守。

控制方式：

- 在规则文件中单独设置 `Default Execution`
- 明确“无歧义时直接执行，完成后再补充潜在问题”

## 验收标准

完成后应满足：

- `.codex/project_rules.md` 中不再出现大段项目架构正文
- `.codex/project_rules.md` 能直接指导 Codex 接任务、执行、验证和汇报
- `docs/architecture/current-architecture.md` 能独立说明当前仓库结构与运行事实
- 两个文件之间不存在大段重复内容
- 规则文件中的架构引用能够指向新的架构文档
