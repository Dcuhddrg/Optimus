---
tool: codex
role: compatibility-adapter
canonical_source: /AGENT.MD
structured_context: /docs
---

# Codex 规则适配层

本文件只用于兼容 Codex 现有规则入口。以下 `/` 开头的路径均相对仓库根目录。

## 真实规则入口

请按下面顺序读取：

1. `/AGENT.MD`
2. `/docs/README.md`
3. `/docs/relations/`、`/docs/memory/`、`/docs/modules/`、`/docs/frontend-backend/`、`/docs/superpowers/specs/`、`/docs/superpowers/plans/` 中与当前任务相关的文件

## 适配层约束

- 不要在这里扩写项目级规则
- 项目级规则或项目上下文变更时，先更新 `/AGENT.MD` 与 `/docs/` 中的对应内容
- 工具目录只保留薄适配层职责，不再保存独立 project rule 副本
