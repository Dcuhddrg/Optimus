# 统一项目规则设计

## 背景

当前仓库存在两份独立维护的 AI 助手规则文件：

- `.claude/project_rules.md` — 给 Claude Code，内容是通用模板，与 Optimus 实际项目不匹配
- `.codex/project_rules.md` — 给 Codex，已根据 Optimus 实际情况定制

两份文件的 BOUNDARIES、执行方式、验证基线等内容高度重叠，分开维护容易产生不一致。

## 目标

- 将两份规则合并为一份，以 `.claude/project_rules.md` 为唯一规则源
- `.codex/project_rules.md` 改为引用指向 `.claude/project_rules.md`
- 合并后的内容以现有 `.codex/project_rules.md` 为基础，补充 Claude Code 特有的工作流和 git 约定
- 更新架构参考段落，反映前端和部署已落地的事实

## 非目标

- 不新增规则章节
- 不改变现有规则的语义
- 不涉及代码改动
- 不修改 `.claude/settings.json` 或 `.claude/settings.local.json`

## 设计

### 文件关系

- `.claude/project_rules.md` — 唯一规则源，Claude Code 和 Codex 共用
- `.codex/project_rules.md` — 改为一行引用，指向 `.claude/project_rules.md`

### 章节结构

沿用 `.codex/project_rules.md` 的五段结构：

1. BOUNDARIES — 原样保留
2. DEFAULT EXECUTION — 原样保留，末尾追加 Superpowers 工作流约定
3. VERIFICATION — 原样保留
4. OUTPUT CONTRACT — 原样保留，末尾追加 Git Commit 约定
5. ARCHITECTURE REFERENCE — 更新前端和部署已落地事实

### DEFAULT EXECUTION 追加内容

在现有执行规则之后增加：

```
### Superpowers 工作流

- 非平凡任务优先走 brainstorming → writing-plans → executing-plans 流程
- 设计文档写入 docs/superpowers/specs/
- 实施计划写入 docs/superpowers/plans/
- 不跳过设计直接写代码，除非任务明确是单文件小改动
```

### OUTPUT CONTRACT 追加内容

在现有输出要求之后增加：

```
### Git Commit 约定

- 格式：类型(范围): 中文描述
- 常用类型：feat / fix / chore / docs / refactor
- 范围对应目录：server / web / deploy 等
- 示例：feat(web): 添加 LDAP 静态展示页
```

### ARCHITECTURE REFERENCE 更新

替换"当前前端第一版范围补充"为：

```
当前前端已落地内容：

- web/ 下 React + TypeScript + Vite 项目已搭建
- AppShell 布局含侧边栏与内容区
- 一级导航：Dashboard、System
- Dashboard 为空态页
- System > LDAP 为静态展示页
- 目录结构：view/ layout/ router/ components/ style/

当前部署已落地内容：

- deploy/docker-compose/ 下三服务编排（frontend + backend + mysql）
- 前端通过 nginx 反向代理 /api 到后端
- Caddy HTTPS 代理方案已设计（optimus.intellivis.cn）
```

### .codex/project_rules.md 替换内容

```
> 本文件已合并至 `.claude/project_rules.md`，Codex 请直接阅读该文件。
```
