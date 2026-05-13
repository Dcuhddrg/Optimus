# Cloud Account + Unified Asset Registry Design

**Date:** 2026-05-13
**Status:** Approved

## Overview

Phase 1 建立云账号凭证管理和统一资产注册表，作为后续 FinOps 与 DevOps 模块的共同数据基础。设计目标是**通用框架优先**：接口、数据模型、编排逻辑与云厂商无关，阿里云 ECS 作为参考实现证明接口可行，其余厂商后续按需插入。

---

## Package Layout

```
server/
├── api/v1/cloud/
│   ├── handlers.go          # NewCloudHandlers, CloudHandlerDeps
│   ├── account.go           # AccountHandler HTTP 方法
│   └── asset.go             # AssetHandler HTTP 方法
│
├── router/cloud/
│   └── enter.go             # 路由注册（全部挂到 privateGroup）
│
├── initialize/cloud/
│   └── module.go            # CloudModule 组装（SyncerRegistry、handlers、DB）
│
└── service/cloud/
    ├── usecase/
    │   ├── account.go       # CloudAccountService + AccountStore 接口
    │   ├── asset.go         # AssetRegistryService + AssetStore 接口
    │   └── syncer.go        # AssetSyncer 接口 + RawAsset + SyncerRegistry
    └── infra/
        ├── persistence/
        │   ├── account_store_gorm.go
        │   └── asset_store_gorm.go
        └── provider/
            └── aliyun/
                └── ecs_syncer.go   # 参考实现
```

`initialize/cloud/module.go` 在 `Routers()` 中与 `SystemModule` 并列初始化。

---

## Data Model

### `cloud_accounts`

| 列 | 类型 | 说明 |
|----|------|------|
| `id` | uint PK | |
| `name` | varchar(128) | 人类可读名称 |
| `provider` | varchar(32) | `"aliyun"` / `"aws"` / `"tencent"` / `"azure"` |
| `account_id` | varchar(64) | 云厂商原生账号 ID |
| `credential_type` | varchar(32) | `"ak_sk"` / `"role_arn"` / `"service_principal"` |
| `credential_json` | text | AES-256-GCM 加密后的凭证 JSON |
| `status` | varchar(16) | `"active"` / `"disabled"` / `"error"` |
| `last_synced_at` | datetime nullable | 最近成功同步时间 |
| `deleted_at` | datetime nullable | 软删除标记 |\n| `created_at` / `updated_at` | datetime | |

凭证永不以明文通过 API 返回（读时替换为 `"***"`）。加密 key 来自 config 新增字段：

```yaml
cloud:
  credential_encryption_key: ""   # 32 字节，AES-256-GCM
```

### `cloud_assets`

| 列 | 类型 | 说明 |
|----|------|------|
| `id` | uint PK | |
| `account_id` | uint FK → cloud_accounts | |
| `provider` | varchar(32) | 冗余，查询提效 |
| `region` | varchar(64) | |
| `resource_type` | varchar(64) | `"ecs_instance"` 等 |
| `resource_id` | varchar(128) | 云厂商原生资源 ID |
| `name` | varchar(256) | 显示名称 |
| `status` | varchar(32) | 云厂商原生状态 |
| `config_json` | text JSON | 提供商特有属性 |
| `tags_json` | text JSON | 标签键值对 |
| `discovered_at` | datetime | 首次发现时间 |
| `last_seen_at` | datetime | 最近同步确认存在 |
| `deleted_at` | datetime nullable | 软删除（资源消失时置值） |

索引：`(account_id, resource_type)`、`(provider, region, resource_type)`；UNIQUE `(account_id, resource_id)`。

---

## Sync Architecture

### 核心接口（`usecase/syncer.go`）

```go
type RawAsset struct {
    ResourceType string
    ResourceID   string
    Name         string
    Region       string
    Status       string
    ConfigJSON   string
    TagsJSON     string
}

type AssetSyncer interface {
    Sync(ctx context.Context, account CloudAccount) ([]RawAsset, error)
}

type SyncerRegistry struct{ ... }
func (r *SyncerRegistry) Register(provider string, s AssetSyncer)
func (r *SyncerRegistry) Get(provider string) (AssetSyncer, bool)
```

`CloudAccount` 是 usecase 层值类型（非 GORM model），凭证字段传入前已解密。

### 同步编排（`usecase/asset.go`）

`AssetRegistryService.SyncAccount(ctx, accountID)` 步骤：
1. 从 `AccountStore` 读取账号，解密 `credential_json`
2. 从 `SyncerRegistry` 按 `account.Provider` 查找同步器
3. 调用 `syncer.Sync(ctx, account)` → `[]RawAsset`
4. 逐条 upsert 到 `AssetStore`：已存在 → 更新 `last_seen_at`；新增 → 写入 `discovered_at`
5. 成功 → 更新 `account.last_synced_at`；失败 → 更新 `account.status = "error"`

### Provider 注册（`initialize/cloud/module.go`）

```go
registry := usecase.NewSyncerRegistry()
registry.Register("aliyun", aliyun.NewECSSyncer())
// 将来：registry.Register("aws", aws.NewEC2Syncer())
```

### 阿里云参考实现（`infra/provider/aliyun/ecs_syncer.go`）

调用 ECS `DescribeInstances` API，将实例列表映射为 `[]RawAsset`（`resource_type = "ecs_instance"`）。凭证从解密后的 `credential_json` 读取：`{"access_key_id": "...", "access_key_secret": "..."}`.

---

## API Endpoints

所有路由挂到 `privateGroup`（JWT + Casbin RBAC）。

### Cloud Account

| Method | Path | 说明 |
|--------|------|------|
| POST | `/api/v1/cloud/accounts` | 注册账号 |
| GET | `/api/v1/cloud/accounts` | 账号列表（含状态） |
| GET | `/api/v1/cloud/accounts/:id` | 账号详情 |
| PUT | `/api/v1/cloud/accounts/:id` | 更新名称或轮换凭证 |
| DELETE | `/api/v1/cloud/accounts/:id` | 软删除 |
| POST | `/api/v1/cloud/accounts/:id/sync` | 手动触发资产同步 |

读取账号时 `credential_json` 不出现在响应中，替换为 `"credential": "***"`。

### Cloud Assets

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/v1/cloud/assets` | 资产列表（分页 + 过滤） |

查询参数：`accountId`、`provider`、`region`、`resourceType`、`status`、`keyword`、`page`（默认 1）、`pageSize`（默认 20，最大 200）。

Response：`{"items": [...], "total": 1234, "page": 1, "pageSize": 20}`

### 权限种子

| Code | 说明 |
|------|------|
| `cloud.accounts.view` | 查看账号列表/详情 |
| `cloud.accounts.manage` | 创建/更新/删除账号 |
| `cloud.accounts.sync` | 触发同步 |
| `cloud.assets.view` | 查看资产列表 |

admin 角色默认拥有全部 cloud 权限。

---

## Architecture Decisions

1. **账号级全量拉取**：`AssetSyncer.Sync` 一次返回账号下所有资产，provider 内部迭代资源类型。接口极简，框架对资源类型无感知。按资源类型拆分是未来可选演进，不影响当前数据模型。

2. **凭证 AES-256-GCM 加密**：key 在 config，不在数据库，不与 JWT signing key 共用。轮换需重新加密 `credential_json` 列，Phase 1 可接受。

3. **软删除 assets**：消失的资源打 `deleted_at` 不硬删，Phase 3 成本数据需要引用历史资源。

4. **SyncerRegistry 运行时注册**：新厂商不改 usecase 代码，只在 `initialize/cloud/module.go` 加一行 `Register`。

5. **第一期不做定时同步**：手动触发足够验证框架。定时调度结构已预留（`AssetRegistryService.SyncAccount` 可被任意触发方调用）。
