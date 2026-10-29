# System 模块结构迁移计划

目标：按 #11 收敛 system 的业务、数据访问与适配器，保持认证、RBAC、LDAP、审计、接口和安全行为。

1. 将 JWT 签发和解析收敛至 system Service，消除根 Service 与旧 identity/usecase 的循环依赖；保留既有公开行为和测试。
2. 业务归 Service，持久化归 store，LDAP 适配归 directory；DTO 归 Model。同步 API、初始化、中间件及测试，不保留旧 usecase/infra 目录或兼容包。
3. 验证后端全量测试、构建、认证授权和项目隔离相关 race、隔离 MySQL 持久化及独立复核；未覆盖的真实 LDAP 与端到端场景如实记录。
4. 本批不修改响应契约、权限策略、数据库结构或运行配置。
