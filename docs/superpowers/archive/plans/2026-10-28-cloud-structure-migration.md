# Cloud 模块结构迁移计划

目标：依 #11，将云账号、资产和同步业务迁入 cloud Service，HTTP DTO 迁入 cloud Model，数据访问与 Provider 适配明确归属；保持接口、凭据加密、同步和项目隔离行为。

1. 迁移 `service/cloud/usecase` 业务至 `service/cloud`，持久化至 `service/cloud/store`，阿里云适配至 `service/cloud/provider/aliyun`；同步 API、初始化、适配器和测试引用。
2. 迁移请求和响应 DTO 至 `model/cloud`，保留 Service 内仓储接口、同步协议和业务错误；保持 JSON 字段不变。
3. 运行 cloud 定向测试、后端全量测试、构建、隔离 MySQL 持久化回归与路由/项目隔离 race 检查；独立审阅并如实记录未覆盖场景。
4. 该分支依赖已验收但尚未合入 main 的 project 分支；两批单独提交与验收，cloud 不改变现有 HTTP 协议、权限或数据库结构。
