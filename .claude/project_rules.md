# Project Rules

> 本文件是 Claude Code 与 Codex 在当前仓库中的共用行为契约。处理任务前先阅读，并按此约束接任务、执行、验证和汇报。
> 本仓库在相关任务中使用 Superpowers 工作流，但默认不要求更宽泛的 OMX runtime modes。

---

## BOUNDARIES

在没有明确指令时，默认禁止：

- 修改任务未涉及的文件
- 顺手扩展需求或附带实现"更好的方案"
- 更改现有技术选型、框架或依赖版本
- 删除注释、TODO、占位说明
- 修改 `.env`、`.env.example`、部署配置或其他环境文件
- 跨越模块边界执行改动
- 因为个人判断而重写已有结构

遇到以下情况必须暂停并提问：

- 需要触碰上述禁止项
- 任务描述与当前仓库现状明显冲突
- 发现严重问题但不在本次任务范围内
- 不确定应该改哪个文件或模块
- 前端任务需要新增一级导航、页面范围或技术栈，但用户未明确确认

---

## DEFAULT EXECUTION

默认执行方式：

1. 先确认任务范围与将修改的文件边界
2. 任务无歧义时直接执行，避免把简单事项扩展成长篇方案
3. 先完成用户明确要求的事项，再补充潜在问题
4. 潜在问题放到任务完成后的"注意事项"中说明

执行时默认遵循：

- 小步修改，边改边验证
- 优先贴合现有模式，不自行发明新层次
- 只在当前任务直接需要时新增依赖、测试或文档
- 不把"更完整"当成"更正确"
- 不主动扩展已确认的前端页面范围

---

## WORKFLOW

### Superpowers 工作流

- 非平凡任务优先从 brainstorming 开始，先澄清目标、边界与设计取舍
- brainstorming 形成并确认规格后进入 writing-plans，沉淀为可执行计划
- 完成规划后，实施可根据任务形态选择 subagent-driven-development 或 executing-plans
- 设计文档写入 `docs/superpowers/specs/`，实施计划写入 `docs/superpowers/plans/`
- **当前活跃**：只保留进行中阶段的 specs / plans（当前为 Phase 1 Cloud + SP 系列）
- **已完成归档**：阶段交付完成后，把对应的 spec 与 plan 移到 `docs/superpowers/archive/specs/` 与 `docs/superpowers/archive/plans/`，保留历史但不再出现在 active 目录中干扰扫描
- 不跳过设计直接写代码，除非任务明确是单文件小改动

---

## CODE DISCOVERY

本仓库已建立 codebase-memory 知识图谱，项目名 `home-lilin-trae-project-Optimus`（路径 `/home/lilin/trae-project/Optimus`）。**任何代码探索任务必须优先走 codebase-memory-mcp 工具**，仅在非代码场景才回退到 Grep/Glob/Read。

### 默认走图谱的场景

- 找函数、类、路由、handler 在哪：`search_graph(label, name_pattern)`
- 谁调用 X / X 调用了什么 / 完整调用链：`trace_path(function_name, direction=inbound|outbound|both, depth)`
- 读源码（替代 `cat` / `Read` 源文件）：`get_code_snippet(qualified_name)`
- 看项目骨架、目录结构、技术栈：`get_architecture(aspects)`
- 跨服务依赖、HTTP / async 调用关系：`query_graph(<Cypher>)`
- 评估本地改动影响面：`detect_changes()` 把 git diff 映射到受影响的符号
- 文本搜索（带图谱权重）：`search_code(pattern)`，优于纯 `grep`
- 死代码 / 高扇出 / 高扇入 / 重构候选：`search_graph` 配合 `max_degree` / `min_degree` 参数
- 不熟悉图谱时先看 `get_graph_schema()` 了解节点和边类型

### 回退到 Grep/Glob/Read 的场景

仅限以下场景走文本工具：

- `.md` / `.yml` / `.yaml` / `.json` / `.toml` / `.env*` 等非代码文件
- 配置值、注释、字符串字面量、错误信息文案的查找
- `.claude/`、`docs/`、`deploy/` 等非源码目录的内容确认
- 图谱明确未覆盖的新文件（先用 `detect_changes` 判断）

### 索引一致性

- 拉取主干、做大重构、批量新增文件后，先 `detect_changes()` 看是否漂移
- 索引明显落后时执行 `index_repository`，必要时清理后重建
- 任务开始前若不确定状态，先 `index_status(project="home-lilin-trae-project-Optimus")` 确认 `ready`

### 子代理继承

派发 subagent 时必须在 prompt 里写明：**代码发现先用 codebase-memory-mcp，文本/配置才回退 Grep/Read**。否则子代理会沿用默认 grep 习惯，浪费 token 且结果不结构化。

### Token 取向

图谱查询单次约 500 tokens，等价的 `grep -r` 经常 50k+。能用 `search_graph` / `trace_path` 解决的，不要先撒大网。

## VERIFICATION

默认验证原则：

- 改了什么，就验证什么
- 不声称"已完成"前，先运行对应验证
- 如验证受环境限制阻塞，需要明确说明阻塞点和已完成的验证
- 文档任务也要验证结构、引用和改动边界

后端代码改动默认验证基线：

- `cd server && go test ./...`
- `cd server && go build ./...`

前端代码改动默认验证基线：

- `cd web && pnpm build`

开发服务器验证约束：

- 当前开发服务器通过 `192.169.50.100` 供 Mac 端访问；前端运行效果验证优先使用 Docker Compose 构建后的 nginx 容器，不默认依赖 Vite dev server。
- 当用户需要从 Mac 端访问前端页面或静态资源时，执行 `docker compose -f deploy/docker-compose/docker-compose.yml build`，再执行 `docker compose -f deploy/docker-compose/docker-compose.yml up -d`。
- Caddy 已配置域名 `optimus.intellivis.cn`；Mac 端优先通过该域名访问，必要时再用 `http://192.169.50.100:8080/` 直连 Compose 暴露端口排查。
- Compose 启动后静态资源应由 nginx 返回，接口通过 nginx 的 `/api/` 代理到 `optimus-server:8888`。
- 只有在任务明确要求本地开发热更新或排查 Vite 专属问题时，才追加 `cd web && pnpm dev`。

如存在以下情况，再先做条件性环境准备：

- 依赖缺失
- lockfile 或依赖状态发生变化
- 本地环境不完整
- 此时可执行 `cd web && pnpm install`

如果任务涉及前端运行效果，再追加：

- 优先按“开发服务器验证约束”使用 Docker Compose 构建并启动
- 确认 nginx 页面、静态资源和相关 API 代理可访问

如果任务涉及启动、路由或接口行为，再追加：

- 服务启动验证
- `/api/v1/health` 可用性验证
- 相关 ping 路由返回值验证

纯文档或规则文件改动默认验证方式：

- 检查目标文件是否存在
- 检查章节结构、引用关系和内容边界是否符合任务要求
- 检查未误改无关文件

---

## OUTPUT CONTRACT

输出要求：

- 对话使用中文
- 结果汇报保持简洁、可扫描
- 不写"顺便优化了什么"
- 不在未被要求时展开大段替代方案
- 若发现额外问题，先完成当前任务，再在"注意事项"中简述
- 结论先于过程，结果先于背景
- 前端展示任务优先说明页面范围、导航边界和是否为静态壳子

默认汇报格式：

```text
## 变更说明
[1-2 句话说明做了什么]

## 影响范围
- 修改文件：[列出变更文件]
- 影响功能：[说明影响点]
- 未修改：[说明明确没动的相关区域]

## 注意事项
[需要人工关注的点；没有则写"无"]
```

### Git Commit 约定

- 提交意图行保持中文描述，直接说明这次提交为什么存在。
- 提交信息遵循仓库当前的 Lore protocol：首行写中文意图，空行后按需补充说明与 trailers。
- 常用 trailers 包括：`Constraint:`、`Rejected:`、`Directive:`、`Confidence:`、`Scope-risk:`、`Tested:`。
- 不再默认使用 `feat(scope): 中文描述` 这类 Conventional Commit 形式。
- 示例：
  ```text
  使提交规则与当前仓库风格保持一致

  Constraint: Chinese intent descriptions must remain part of the shared contract
  Confidence: high
  Scope-risk: narrow
  Tested: grep -nE 'feat\(web\)|Lore|中文描述|Output Contract|OUTPUT CONTRACT' .codex/project_rules.md
  ```

---

## ARCHITECTURE REFERENCE

当前仓库架构事实统一维护在：

- `docs/architecture/current-architecture.md`

规则文件不重复罗列架构快照，以上述文档为准。
