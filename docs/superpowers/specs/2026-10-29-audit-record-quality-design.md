# 审计记录质量与失败可观测性

- 关联：#36；用户已明确选择保留业务结果、审计失败可见的方案。
- 状态：本批用户测试确认通过并授权提交 PR、合入 main；交付及验证记录见 #36，不扩大为审计保证落库或真实外部联调验收。
- 范围：统一现有系统模块审计写入失败处理，补用户状态、用户角色、角色权限、LDAP 配置/测试与失败登录的事件内容。

## 行为与取舍

保持现有业务返回、HTTP 行为和事务边界。审计是尽力写入：失败不回滚已提交业务，通过现有 Zap 日志（未装配时标准日志回退）报告动作、资源类型/ID、结果和错误类型，不输出原始异常、SQL、凭据或整条 Detail。不宣称事件保证落库，也不引入队列、后台重试或新数据库表。

现有 Detail 字段保存 version=1 的 JSON，包含稳定 reason、白名单 changes，以及失败登录的 attemptedUsername 和可选 targetUserId。历史内容不重写。状态记录读取前值与保存返回值；角色集合与权限集合记录实际读取的前后集合并计算增加/移除，不用请求值冒充已保存结果。集合快照失败不改变业务结果，事件标记 snapshot_unavailable 且不虚构差异。快照不是与变更原子捕获，存在并发时不承诺事务级精确历史；强一致审计另行设计。

LDAP 仅记录 enabled 前后值及连接、映射、凭据更新的布尔变化，不保存 URL、DN、密码或整份配置。连接测试失败使用固定原因码，不把 err.Error 原样持久化。登录失败的尝试账号作为未认证输入限长/清理控制字符，Actor 保持未认证；已知目标用户另记 targetUserId，不冒充认证成功的操作者。登录成功与已认证操作保持原操作人快照语义。

## 边界与验证

不扩展项目/云模块或通用鉴权中间件事件，不改 #18 协议、保留期或前端展示。新增权限快照只读查询及必要仓储接口/夹具；快照查询异常不会把业务成功改为失败。测试故障注入覆盖审计失败仍保留成功/原错误、受保护操作拒绝、真实前后集合、快照缺失、LDAP 白名单、尝试账号与原始异常不泄露。运行后端测试、构建、相关 race；MySQL 条件允许时执行隔离数据库回归，受阻/跳过如实记录。

## 新事件详情格式

`version` 固定为 1；`reason` 使用稳定原因码。`changes` 仅允许 `status`、`roleIds`、`permissionCodes`、`enabled`、`connectionChanged`、`mappingChanged`、`bindPasswordChanged`，未变化的字段省略。每项包含 `before`/`after`，集合额外包含 `added`/`removed`，排序去重后记录；空集合是数组，不是 null。LDAP 的连接/映射/凭据变化只保留布尔值。

失败登录使用独立的 `attemptedUsername`（最多 128 个字符，控制字符替换为空格）和可选 `targetUserId`；Actor 不表示已认证，不能据此将尝试账号展示为已认证操作人。历史事件的 Actor 快照及旧 Detail 原文不回填。

常见原因码：`updated`、`last_admin_protected`、`admin_role_protected`、`update_failed`、`snapshot_unavailable`、`connection_test_failed`、`invalid_config`、`invalid_credentials`、`ldap_authentication_failed`。业务错误仍原样返回；详情不能序列化时使用 `detail_unavailable`，不影响业务结果。运行日志固定消息 `audit write failed`，仅包含动作、资源类型/ID、结果及 `error_type`，不输出原始错误值、账号、IP 或 Detail。
