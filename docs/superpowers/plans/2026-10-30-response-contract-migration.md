# 接口协议迁移首批实施计划

**Goal:** #18 公共协议能力与审计样板可评审交付，不把未迁移模块算作完成。

**Architecture:** 明确路由标记+共用HTTP响应工具；前端显式选择envelope，旧apiRequest与CSV成功保持原语义，审计data.list同步页面。

**Tech Stack:** Go/Gin、现有TypeScript/React/Vitest；不新增依赖。

- [x] 恢复codebase-memory-mcp并逐项盘点真实路由、handler和模块调用方，在#18登记批次/未知外部调用方。
- [x] 在response工具、真实router、审计handler与前端请求/页面测试中添加目标协议用例，先观察旧实现失败。
- [x] 新建server/utils/response公共助手，在initialize仅为审计方法+路由设置标记，JWT/RBAC保留未标记请求旧格式。
- [x] 新增HTTP AuditListData，切换审计查询与导出JSON失败，保留CSV及Service/Store结构。
- [x] 在client新增显式envelope解析、ApiError业务码及分类，模块API/审计类型/页面/相关测试同步list，其他调用保持旧协议。
- [x] 运行适用后端测试/构建/race与前端测试/构建；同步契约、examples和#18证据，独立审阅。#18保留其他模块批次，用户验收前不关闭或部署。
