# 角色与权限响应契约迁移

关联 [#18](https://github.com/Dcuhddrg/Optimus/issues/18)，延续已确认的[接口契约](../../frontend-backend/boundary.md)及分批迁移方案，不新增业务功能。

## 范围与契约

本批迁移 GET/POST `/system/roles`、PUT `/system/roles/:id`、PATCH `/system/roles/:id/status`、PUT `/system/roles/:id/permissions`、GET `/system/permissions/tree` 六个端点。

- 成功使用 `{code:0,data,msg}`；角色集合为 `data:{list,total}`，空列表为 `[]`；权限集合为 `data:{list}`。两者均不提供分页。
- 创建、修改和状态更新保留原角色 DTO 和 HTTP 200；保存权限为 `data:{}`。
- 保留 HTTP 400/401/403/404/500/503、既有错误摘要与管理员保护规则；失败 code=7，普通错误 data={}，401 data=null。
- 复用公共响应助手和 JWT 前的方法/规范路由精确标记；不变更 Service、Store、RBAC、请求结构或项目隔离。

## 调用方、发布与回退

codebase-memory-mcp 核查：仓库导出的调用为 `listRoles`、`permissionTree`、`saveRolePermissions`，消费者为 `SystemRolesPage`。模块 API 显式使用 `apiRequestEnvelope`，页面改读 `list`。角色新增/编辑/状态接口未发现对应导出函数，不据此排除外部调用方。LDAP、project、cloud 继续旧协议。

沿用既有方案，以前后端同批发布；本次只实现和本地验证，不部署。仓库外客户端、旧标签页和发布顺序仍需发布前核实；存在独立消费者时另定兼容方案。回退同时回退角色 handler、六条协议标记、前端模块 API 和角色页面；公共助手保留。

## 验证

先失败再实现：真实 Router 下默认/自定义/根前缀的六路 JWT/RBAC 拒绝；真实 Service 配合可控 Store 的成功、空集合、参数/JSON 错误、管理员保护、未找到、存储错误及缺服务。旧 LDAP 和探活协议保持回归。

前端用模拟 HTTP 测试真实 API 解析、请求方法/正文、空集合、业务/HTTP/协议失败；角色页面测试加载和授权操作。运行后端测试/构建、相关 race 和前端测试/构建，环境问题单独记录，不等同真实部署验收。
