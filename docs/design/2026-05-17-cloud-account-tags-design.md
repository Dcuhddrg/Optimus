# Cloud Account User Tags Design

**Date:** 2026-05-17
**Status:** Approved (mockup confirmed; ready for plan)

## Overview

在 SP1（项目实体）+ SP2（cloud 模块项目隔离）之上，新增"用户 tag"维度作为账号分类机制。**环境（prod / dev / qa / orphan）不作为数据库字段，而是通过 tag 实现**：账号导入后由用户随时打 `environment=prod` 这类 tag；平台对 `environment` 这个 key 提供 UI 智能（颜色、过滤芯片、快捷建议），但**不强制**——其他自由 key（team / cost-center / owner 等）也支持。

设计目标：
- **最轻数据模型**：cloud_accounts 加一个 JSON 列存用户 tag，不动 cloud_assets 表
- **不预设环境字段**：避免枚举与现实业务的频繁背离（如 staging、preview、canary）
- **资产视图继承**：资产页过滤 tag 时透过所属账号查询，UI 显示"继承 tag"
- **不动 cloud_assets.tags_json**：该字段保留给云厂商原生 tag（只读）

---

## Data Model

### `cloud_accounts.tags_json`（新增列）

```sql
ALTER TABLE cloud_accounts
ADD COLUMN tags_json TEXT NOT NULL DEFAULT '{}';
```

GORM 字段：
```go
TagsJSON string `gorm:"type:text;not null;default:'{}'" json:"-"`
```

序列化为 `map[string]string`，例如：
```json
{
  "environment": "prod",
  "team": "sre",
  "cost-center": "A1"
}
```

约束：
- Tag key：`^[a-z][a-z0-9_-]{0,62}$`（小写字母开头，最多 63 字符）
- Tag value：1-128 字符，去除前后空白
- 单账号最多 32 个 tag
- 同一 key 只能出现一次（map 自然保证）

### `cloud_assets` 不变

资产继承来自账号的 tag，不在资产表冗余存储。查询时通过 `JOIN cloud_accounts` 取得。

### 与 `cloud_assets.tags_json` 区分

| 字段 | 含义 | 写入方 |
|------|------|--------|
| `cloud_assets.tags_json` | 云厂商原生 tag（同步而来） | syncer 自动写 |
| `cloud_accounts.tags_json` | 平台用户 tag（手工打标） | API 写入 |

两者**互不干扰**。资产页 UI 上可以分别显示"原生标签"和"账号继承标签"。

---

## `environment` 软约定

`environment` 是平台**推荐**但**不强制**的 key。语义层面：

- 后端层：无任何特殊处理；与其他 key 同等对待
- 前端层：
  - 输入框旁提供快捷建议 `environment=prod`、`environment=dev`、`environment=qa`、`environment=orphan`
  - tag 渲染对 `environment` 值用预设配色（prod 红 / dev 绿 / qa 棕 / orphan 灰；其他值用通用灰）
  - 列表顶部提供"environment 过滤芯片"：基于 `environment=*` 聚合的快捷视图，含"未标记"（没打 environment key 的账号）

用户可以**完全不用 environment**，或者扩展使用 `environment=staging` / `environment=preview` 等自定义值。

---

## API Surface

### 现有路由不变

`/api/v1/cloud/accounts`（GET / POST / GET-detail / PUT / DELETE / POST sync）所有现有行为保留。

### 响应字段增量

`AccountView` 增加 `tags` 字段：
```json
{
  "id": 1,
  "projectId": 1,
  "name": "aliyun-main",
  "provider": "aliyun",
  "tags": { "environment": "prod", "team": "sre" },
  ...
}
```

`POST /api/v1/cloud/accounts` 请求体可选地接受 `tags`（缺省则空）：
```json
{
  "name": "...",
  "provider": "aliyun",
  ...,
  "tags": { "environment": "prod" }
}
```

### 新增 tag 管理端点

| Method | Path | 说明 |
|--------|------|------|
| PUT | `/api/v1/cloud/accounts/:id/tags` | 全量覆盖账号 tag |

请求体：
```json
{ "tags": { "environment": "prod", "team": "sre" } }
```

行为：
- 校验 key/value 格式与数量
- 全量替换 `tags_json`（不做部分合并；前端在提交前自行合并）
- 权限：复用 `cloud.accounts.update`
- 响应：200 + 更新后的 `AccountView`

**不单独提供 PATCH/DELETE tag 接口**（YAGNI），前端拿到全量后修改再 PUT 一次。

### 列表过滤

`GET /api/v1/cloud/accounts` 新增 query 参数 `tag`，支持多次出现（AND 关系）：
```
GET /api/v1/cloud/accounts?tag=environment=prod&tag=team=sre
```

匹配语义：
- `tag=key=value`：精确匹配 key 与 value
- `tag=key=*`：账号 tags 中存在该 key（任意 value）
- `tag=__none__:environment`：账号 tags 中**不存在** environment 这个 key（保留语法，专门支持"未标记"过滤芯片）

实现：在 store 层用 `JSON_EXTRACT(tags_json, '$.<key>')` 过滤（MySQL 5.7+ 支持）。

### 资产列表过滤

`GET /api/v1/cloud/assets` 同样接受 `tag` 参数；语义为"账号继承 tag 匹配"，通过现有 JOIN cloud_accounts 加 JSON_EXTRACT 条件实现。

---

## Frontend UX

### 云账号列表页

新增 **Tags 列**（在 `名称` 之后、`厂商` 之前）：
- 每条账号显示其所有 tag 为 chip；`environment=X` chip 使用预设配色
- 未打任何 tag 的账号显示"未标记任何 tag"灰色虚线 chip
- 行级操作菜单加 `改 tag` 按钮

**顶部 environment 过滤芯片**：
```
environment：  全部 12  ▾env=prod 4  ▾env=dev 5  ▾env=qa 2  ▾未标记 1
```
点击芯片在 URL query 中追加 `tag=environment=...`，等价于过滤栏的下拉。

**顶部通用 tag 过滤输入框**：
- 支持自由输入 `environment=prod team=ops`，按空格分隔多条
- 输入时显示下拉建议（基于当前已有 tag 的 key/value 集合）

### Tag 编辑弹窗

字段：
- 显示当前 tag 列表（可移除 chip）
- 新增行：key 输入 + value 输入 + "添加"按钮
- 提示文字：推荐使用 `environment` 作为环境分类
- 快捷建议 chip：`environment=prod` / `environment=dev` / `environment=qa` / `environment=orphan`，点击直填

校验：
- key 不符合正则 → 输入框红框 + 下方红色提示
- value 为空 → 同上
- 超过 32 个 tag → 添加按钮禁用，提示上限
- 重复 key → 提示"该 key 已存在，将覆盖原值"

提交 → 调用 `PUT /api/v1/cloud/accounts/:id/tags` → 关闭弹窗 → 刷新列表。

### 云资产列表页

资产展示**继承 tag 列**：从所属账号读取 tag，显示为 chip。

过滤栏 tag 输入框与账号页同语法；查询语义透传到后端。

---

## Migration

### Schema 迁移

`AutoMigrateCloud` 自动创建 `tags_json` 列（GORM AutoMigrate 会 add column with default `'{}'`）。无需手写 DDL。

### 历史账号 backfill

历史账号的 `tags_json` 默认为 `'{}'`（DB default）。**不自动填 `environment=orphan` 或任何值**——保持"未标记"语义，由运维显式打标。

`BackfillCloudProjectScope`（SP2 已落地）和本设计的迁移逻辑互不影响。

---

## Architecture Decisions

1. **环境不作为枚举字段**：避免日后业务多变时 schema migration（如增加 staging、canary）。tag-as-environment 牺牲了类型安全，换来了灵活性。前端用 `environment` 软约定提供 80% 场景的体验。

2. **Tag 打在账号级，资产继承显示**：
   - 与 SP2 项目隔离粒度对齐（账号是项目下的最小单元，资产是账号下的派生物）
   - 数据模型简单，不引入资产级 user tag 表
   - 同账号下 prod/dev 混合的场景视为反模式（建议拆账号）

3. **不动 `cloud_assets.tags_json`**：该字段是云厂商原生 tag，syncer 负责维护。用户 tag 与原生 tag 分层显示，避免混淆。

4. **PUT 全量覆盖而非 PATCH**：前端在弹窗内可视化所有 tag 后整体提交，简化语义。原子性好。32 个 tag 上限保证全量 PUT 不会膨胀。

5. **`environment` 是软约定而非保留字**：后端不限制 environment 的合法 value（不限定 prod/dev/qa/orphan）。前端给这 4 个值预设配色和建议；其他值用通用灰显示。

6. **不实现 tag 历史 / 审计**：YAGNI；如果将来需要，可在 `cloud_audit_logs` 加 `tag_changed` 事件。

7. **JSON_EXTRACT 而非独立 tag 表**：tag 数量小（32/账号上限）且查询模式以 key=value 精确匹配为主。引入 `cloud_account_tags(account_id, key, value)` 表会带来 JOIN 开销且不显著提升查询能力。如果未来需要全文搜索或聚合，可再加。

---

## Out of Scope

- 资产级用户 tag（资产 inherit-only）
- 多 value 的 tag（key 唯一，单值）
- Tag 模板 / 全局 tag 治理（如"所有账号必须打 environment 才能用"这类策略）
- Tag 自动推断（从云厂商原生 tag 自动同步到用户 tag）
- Tag 审计日志
- Tag-based RBAC（按 tag 控制访问）

---

## Open Questions

- **Tag 排序**：列表渲染时按字母序、插入顺序、还是 `environment` 优先 + 其余字母序？默认 `environment` 优先 + 字母序。
- **Backfill 现有云账号的 environment**：design 当前是不填，由运维手动打。若运维量大，是否提供批量 import CSV？延后到运维反馈再决定。
- **Tag 删除时是否保留历史**：当前 PUT 全量覆盖，删除即丢失。若有审计需求，未来加事件流。

---

## Scope Notes

本 spec 与 SP4（前端 cloud 页面项目化适配）合并实施：cloud 页面是 SP4 的范围，tag 是 SP4 在 mockup 上的增量。后端的 tags_json 列 + PUT 接口 + 过滤参数也作为同一个 sub-project 落地。原 spec 中的 SP4 概念扩展为"前端 cloud 页面 + 用户 tag（含 environment 软约定）"。
