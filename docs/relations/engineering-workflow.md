# 工程操作参考

最近整理：2026-10-29。由 [AGENT.MD](../../AGENT.MD) 引用，承接原规则中的工具用法和验证步骤；行为边界以根规则为准。按当前任务读取对应章节。

## 图谱查询

项目名：`home-lilin-trae-project-Optimus`；本机路径：`/home/lilin/trae-project/Optimus`。

本节工具均指 `codebase-memory-mcp` 的能力；根据任务使用 `search_code`、`search_graph`、`get_code_snippet`、`trace_path` 等具体工具，不能以其他 MCP 服务代替代码层级的工具要求。

| 目标 | 工具 |
| --- | --- |
| 找函数、类、路由和 handler | `search_graph(label, name_pattern)` |
| 调用者、被调用者和调用链 | `trace_path(function_name, direction=inbound或outbound或both, depth)` |
| 读已定位的源码 | `get_code_snippet(qualified_name)` |
| 项目骨架、模块和技术栈 | `get_architecture(aspects)` |
| 跨服务依赖、HTTP / async 关系 | `query_graph(<Cypher>)` |
| 改动影响面 | `detect_changes()` |
| 带图谱权重的文本搜索 | `search_code(pattern)` |
| 死代码、高扇入/扇出及重构候选 | `search_graph` 配合 `max_degree` / `min_degree` |
| 不熟悉图谱结构 | 先 `get_graph_schema()` |

路由端点可通过 `query_graph("MATCH (n)-[r:HANDLES]->(p) RETURN n.name, r.method, r.url_path")` 按需查询；图谱结果需结合实际实现核对。

### 索引与回退

- 状态不明时，先用 `index_status(project="home-lilin-trae-project-Optimus")` 确认 ready。
- 拉取主干、大重构或批量新增后执行 `detect_changes()`；明显落后时用 `index_repository`，必要时清理后重建。清理操作仍受根规则的授权边界约束。
- 文本工具回退仅用于：非代码文件（如 `.md`、`.yaml`、`.json`、`.toml`、`.env*`）；配置值、注释、字符串字面量或错误文案；`.claude/`、`docs/`、`deploy/` 等非源码目录；先经 `detect_changes` 确认未被图谱覆盖的新文件。
- `codebase-memory-mcp` 服务本身不可用（未启动、索引失败或重建后仍不 ready）时，停止源码检索，向用户说明原因并等待处理；不改用文本工具检索源码。
- 使用文本工具处理源码前，必须核实符合上条列出的回退条件，并说明依据；否则继续使用 `codebase-memory-mcp`。已知路径、已定位符号或进入实现阶段，均不构成回退条件。
- 派发子代理时传递 `codebase-memory-mcp` 使用要求及回退条件；定向查询并限制输出，不以固定 Token 数字推断工具成本。

## 验证步骤

基础命令见 [根规则验证章节](../../AGENT.MD#验证)。以下步骤按任务触发；部署地址和代理契约统一以 [当前架构](repo-profile.md) 为准。

后端持久化及部分业务、权限测试需要 `OPTIMUS_TEST_MYSQL_DSN`；未设置时这些测试会跳过，普通 `go test ./...` 不能证明数据库行为。使用专门的临时 MySQL 数据库设置该变量后运行 `cd server && go test -count=1 -p 1 -timeout 30m ./...`（`service/system` 包串行跑完要 10 分钟以上，超过 go test 默认的 10 分钟超时）。测试会删除并重建表，不能指向应用或共享数据库；`-p 1` 避免不同包并发重置同一测试库。记录数据库测试的实际结果和跳过项。

### 依赖准备

仅在依赖缺失、lockfile/依赖状态改变或环境不完整时执行 `cd web && pnpm install`。前端使用 pnpm，不使用 npm install。

### 前端运行效果和 Mac 访问

默认验证 Compose 构建后的 nginx；当需要从 Mac 访问页面或静态资源时，依次执行：

```bash
docker compose -f deploy/docker-compose/docker-compose.yml build
docker compose -f deploy/docker-compose/docker-compose.yml up -d
```

- 优先通过架构文档记录的 Caddy 域名访问，必要时使用直连端口排查。
- 确认 nginx 页面、静态资源及相关 API 代理可访问；资源由 nginx 返回，`/api/` 代理到后端。
- 仅明确需要本地热更新或排查 Vite 专属问题时追加 `cd web && pnpm dev`。

### 启动、路由或接口

追加服务启动验证、`/api/v1/health` 可用性验证和相关 ping 路由返回值验证。接口或注释有变化时运行 `cd server && swag init` 重新生成 `server/docs/`，并确认 `go test ./initialize/` 中的 Swagger 覆盖测试通过；`swag` CLI 版本与 `go.mod` 中的 `github.com/swaggo/swag`（v1.16.4）一致，安装：`go install github.com/swaggo/swag/cmd/swag@v1.16.4`。

### 文档和规则

检查目标文件存在、章节与引用关系正确、内容符合任务边界，且未误改无关文件；运行 `git diff --check`。链接检查或构建通过均不能代替业务验收。

## 上线检查清单

上线前逐项人工确认，结果记在对应 Issue。

- **内置 admin 密码**：确认内置 `admin` 的密码已更换为强密码。启动时若修复过内置 admin（日志提示"内置 admin 已恢复"），旧密码会重新可用，见[用户管理设计 3.4](../design/2026-11-06-user-management-design.md#34-停用规则q4)。
- **LDAP 目录 ACL（信任边界）**：能改名、移动用户条目（`modrdn`、对条目的 `write`）的账号，必须同时能重置这些用户的密码；Optimus 的绑定账号对整个目录的用户条目有读权限，包括 `id_attr`（默认 `entryUUID`）、`entryCSN`/`modifyTimestamp`、`createTimestamp`。OpenLDAP 在 `cn=config` 的 `olcAccess` 中核对，并在测试 LDAP 上核对一次。Optimus 不自动校验 ACL。见[用户管理设计 3.3](../design/2026-11-06-user-management-design.md#33-从-ldap-同步用户d2-至-d6)"写入顺序"。
- **LDAP 配置**：在 LDAP 配置页"测试连接"通过（同步范围内至少一个用户，且能读到目录标识）。
- **存量 LDAP 账号绑定**：存量 LDAP 账号没有绑定目录标识前不能登录。上线前执行一次同步并完成"待绑定"确认。
