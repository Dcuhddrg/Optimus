# Project 模块结构迁移计划

目标：按 #11 将 project 的业务、数据访问和 DTO 分别收敛到 Service 与 Model，保持所有现有接口与授权行为。

- 将 usecase 中项目业务、错误、角色及仓储接口移入 `server/service/project/`；请求与响应 DTO 移入 `server/model/project/`。
- 将 project 的 GORM 数据访问迁入 `server/service/project/`，保留仓储接口供测试；同步 project API、初始化、cloud 调用和现有测试，不保留旧包。
- 运行项目单元、API、装配和权限隔离回归，后端全量测试和构建；单独记录 MySQL 缺失时跳过的持久化测试。
- 不变更 JSON 字段、HTTP 状态、鉴权、项目隔离或数据库结构；相关跨模块改动仅限调用路径与类型适配。
