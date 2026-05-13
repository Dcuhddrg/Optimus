# Project Rules

> 本文件是 Claude Code 与 Codex 在当前仓库中的共用行为契约。处理任务前先阅读，并按此约束接任务、执行、验证和汇报。

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
- 设计文档写入 `docs/superpowers/specs/`
- 实施计划写入 `docs/superpowers/plans/`
- 不跳过设计直接写代码，除非任务明确是单文件小改动

---

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

如存在以下情况，再先做条件性环境准备：

- 依赖缺失
- lockfile 或依赖状态发生变化
- 本地环境不完整
- 此时可执行 `cd web && pnpm install`

如果任务涉及前端运行效果，再追加：

- `cd web && pnpm dev`
- 确认 dev server 可访问

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
