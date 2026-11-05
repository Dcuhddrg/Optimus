# 后端架构一致性整改：首批路由与授权边界

- 日期：2026-10-27；父事项：[#11](https://github.com/Dcuhddrg/Optimus/issues/11)。
- 范围：用户已授权的审计与分批整改，本批处理 [#12](https://github.com/Dcuhddrg/Optimus/issues/12)、[#13](https://github.com/Dcuhddrg/Optimus/issues/13)。
- 用户已验收 #6、#12、#13，并授权提交、推送及创建 PR；架构文档与代码整改分别提交，交付记录在对应 Issue 跟进。

## 证据与目标

1. ProjectScope 挂在全部 privateGroup 上，而前端所有请求携带当前项目选择。过期、无权限或解析失败的项目选择会阻断系统及项目恢复入口。
2. RouterPrefix 支持自定义，权限键仍以 `/api/v1` 为规范路径；RBAC 对运行时 FullPath 精确匹配造成错误拒绝。
3. 图谱定向检索未见 Service 依赖 Gin、API 引入 GORM/SQL，实际路由为单一注册链；未据此宣称全面合规，也不搬迁 usecase/infra。

## 决定

- JWT/RBAC 继续覆盖所有私有接口，ProjectScope 仅装配到云资源路由组。项目管理保留按路径项目、用户成员资格的独立授权，项目服务本身的故障仍正常返回错误。
- 保留原 `CasbinHandler(authorizer)`，新增显式 `CasbinHandlerWithPrefix(authorizer, runtimePrefix, permissionPrefix)`。装配层传入 Gin 规范化后的 BasePath 和已有 defaultRouterPrefix；只映射权限查询路径，不修改请求路径、权限数据或配置。
- 映射保留 HTTP 方法、参数路由和角色；不匹配的前缀、未匹配路由及未知权限继续拒绝。
- 响应 envelope、分页、错误码迁移待 #7 确认；本批不改字段、HTTP 状态语义、依赖版本、数据库结构或环境文件。

## 验收与测试

- 回归测试先在现有 Routers 上复现失效项目选择和自定义前缀错误拒绝，再实施最小装配修复。
- 系统入口在项目解析故障时仍可执行，项目列表不被无关 header 拦截；云接口保留 403/404/500，项目路径自身的越权检查继续生效。
- 测试默认、自定义、根、尾斜杠和参数路由；前缀边界不得误匹配，未知权限不得放行。
- 不使用应用数据库跑测试：为本轮创建唯一 `optimus_issue11_*` 隔离库，以 `go test -count=1 -p 1 -json ./...` 串行执行，避免既有清表夹具跨包冲突。凭据仅在进程环境中传递，不写入仓库或日志。
- 基线：未配置测试库时 109 测试通过、40 跳过；配置隔离 MySQL 后 149 测试通过、0 跳过。基线 `go build ./...` 通过。

## 后续批次

项目集合类型与服务端真实字段不一致、审计分页入口缺失、会话恢复错误分类和全局响应迁移，先登记审计结果并按契约/行为设计单独实施，不混入本批。
