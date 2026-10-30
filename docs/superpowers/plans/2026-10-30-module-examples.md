# 模块开发示例实施计划

**Goal:** 完成 #8，以 gin-vue-admin examples 的组织与讲解方式补齐 Optimus 开发示例。

**Architecture:** 保留 backend/frontend/plugin 分类，文档引用既有规范，以真实系统/审计实现适配上游示例，不新增业务或工具依赖。

**Tech Stack:** Markdown；Go/React/TypeScript 仅为讲解片段，源码通过 codebase-memory-mcp 核对。

- [x] 固定上游提交并读取全部示例与来源许可；结合 `44d1940` 既有 MCP 核查记录说明本地链路、目录及调用，本轮实时索引异常不冒充重新全量读取。
- [x] 新建 backend/request、model、service、api、router、enter-go、plugin-go 示例和索引。
- [x] 新建 frontend/api、state、view、utils-usage 示例和索引；说明 pinia-example 对应 React 状态管理适配。
- [x] 新建 plugin/full-plugin 示例和索引，标明当前能力与缺口；同步总入口和 modules 入口。
- [x] 检查本地文件/章节链接、示例引用和代码摘录一致性，独立审阅及 git diff --check。
- [x] 在 #8 记录结果及未覆盖项；用户验收前保持打开，提交和部署按授权执行。

## 本地验证与边界（2026-10-30）

- 初版的模板读取与验证已完成；按用户后续要求，当前 examples 不附来源/改写说明、上游参考链接或独立 license 文件。初版交付记录仍可通过 #8 与 Git 历史追溯。
- 共16份examples Markdown（含各级README），state-example对应上游pinia-example；同步docs总入口和modules入口。插件页是目录与装配设计参照，不宣称已验证完整插件SDK。
- 本轮 codebase-memory-mcp 最初符号查询返回 total 有值但 results 为空，fast 重建0节点，full重建1195节点且主要为文档；精确源码查询仍未返回实现。未回退 shell 或其他服务读取本地源码，也未用历史计划中的旧代码冒充当前实现。
- 片段与引用依据此前通过 codebase-memory-mcp 核实的真实审计/认证实现；本地源码基线保持 `44d1940`，无业务源码修改。当前工具无法完成新一轮全量符号核验，限制记录到 #8，未来源码变更后应重核对。
- `python3 /tmp/optimus-check-example-docs.py`：100份Markdown、288个本地文件/章节链接有效；5个Go片段经gofmt语法解析。代码块为上下文摘录，不声称均可独立构建。
- 在web目录执行 `node /tmp/optimus-check-examples-ts.mjs`：5个TypeScript/TSX片段语法转译通过；2个完整教学组件通过实际项目导入的TypeScript类型检查。校验使用内存虚拟文件，不创建应用源码。
- 独立文档审阅指出的显式验证清单和索引限制说明已补齐；`git diff --check`通过。仅docs变更，未运行应用测试/部署，不宣称本轮重新验证业务。
- 尚未提交、推送或用户验收；#8保持打开。

## 用户验收与交付

2026-10-30，用户确认“好的，没问题”，并授权提交 PR、合并 main。上方本地验证保留当时记录；最终提交/PR及结案以 #8 为准。源码索引异常和插件样板未核实的边界保持不变。

## 后续简化（2026-10-30）

用户要求移除 examples 的 license 和上游来源说明；删除对应文件与链接，保留教学内容、技术差异和 Optimus 真实参考，不修改业务代码。此简化尚未提交。
