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

1. 先用一句话复述任务范围
2. 明确说明将修改的文件边界
3. 任务无歧义时直接执行，不先输出长篇方案
4. 先完成用户明确要求的事项，再补充潜在问题
5. 潜在问题放到任务完成后的"注意事项"中说明

执行时默认遵循：

- 小步修改，边改边验证
- 优先贴合现有模式，不自行发明新层次
- 只在当前任务直接需要时新增依赖、测试或文档
- 不把"更完整"当成"更正确"
- 当前前端默认边界是最小壳子，不主动把 `Dashboard + System > LDAP` 扩成完整平台

### Superpowers 工作流

- 非平凡任务优先走 brainstorming → writing-plans → executing-plans 流程
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

- `cd web && npm install`
- `cd web && npm run build`

如果任务涉及前端运行效果，再追加：

- `cd web && npm run dev`
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

- 格式：`类型(范围): 中文描述`
- 常用类型：feat / fix / chore / docs / refactor
- 范围对应目录：server / web / deploy 等
- 示例：`feat(web): 添加 LDAP 静态展示页`

---

## ARCHITECTURE REFERENCE

当前仓库事实请看：

- `docs/architecture/current-architecture.md`

当前前端已落地内容：

- `web/` 下 React + TypeScript + Vite 项目已搭建
- AppShell 布局含侧边栏与内容区
- 一级导航：Dashboard、System
- Dashboard 为空态页
- System > LDAP 为静态展示页
- 目录结构：view/ layout/ router/ components/ style/

当前部署已落地内容：

- `deploy/docker-compose/` 下三服务编排（frontend + backend + mysql）
- 前端通过 nginx 反向代理 `/api` 到后端
- Caddy HTTPS 代理方案已设计（optimus.intellivis.cn）
