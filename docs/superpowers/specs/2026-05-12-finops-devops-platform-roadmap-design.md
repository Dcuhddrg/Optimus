# FinOps + DevOps Platform Roadmap Design

**Date:** 2026-05-12
**Status:** Approved

## Overview

Optimus is positioned as an enterprise-grade unified operations platform covering both FinOps (cloud cost governance) and DevOps (infrastructure automation). The platform is built on a **Unified Asset Registry** as the central data foundation, ensuring cost data and infrastructure state share a single source of truth.

---

## Platform Module Map

| Phase | Modules | Status |
|-------|---------|--------|
| 0 | User / Role / Permission / Audit | ✅ Done |
| 0.5 | Audit Log Upgrade | **Next delivery** |
| 1 | Cloud Account + Unified Asset Registry | Planning |
| 2 | DevOps Layer | Future |
| 3 | FinOps Layer | Future |
| 4 | Automation Layer | Future |

---

## Phase 0.5 — Audit Log Upgrade

### Motivation

The current `system_audit_logs` table is write-only with no query or export capabilities. As the platform grows to cover cloud accounts and cost data, the audit trail becomes a compliance requirement — not just a debug tool.

### Data Model Changes

**Existing columns kept as-is:** `id`, `user_id`, `action`, `resource_type`, `resource_id`, `detail`, `ip`, `user_agent`, `created_at`

**New column:**
- `actor_username VARCHAR(64)` — snapshot of the username at write time; retained even after the user is deleted or renamed

**New indexes:**
- `(result, created_at)` — filter by outcome with time range
- `(resource_type, created_at)` — filter by resource category with time range
- `FULLTEXT(detail)` — keyword search inside JSON detail blobs (MySQL InnoDB FULLTEXT)

**New column (existing schema):**
- `result VARCHAR(16)` — `"success"` | `"failure"` (already implied; make explicit and indexed)

### Query API

`GET /api/v1/system/audit-logs`

New query parameters:

| Param | Type | Description |
|-------|------|-------------|
| `page` | int | 1-based page number (default 1) |
| `pageSize` | int | Records per page (default 20, max 200) |
| `actorUserId` | uint | Filter by user ID |
| `action` | string | Exact action match (e.g. `"login"`) |
| `resourceType` | string | Exact resource type match |
| `result` | string | `"success"` or `"failure"` |
| `startTime` | RFC3339 | Lower bound on `created_at` |
| `endTime` | RFC3339 | Upper bound on `created_at` |
| `keyword` | string | FULLTEXT search against `detail` |

Response shape:
```json
{
  "items": [...],
  "total": 1234,
  "page": 1,
  "pageSize": 20
}
```

### Export API

`GET /api/v1/system/audit-logs/export`

- Same filter params as query API (no pagination)
- Returns `text/csv` with `Content-Disposition: attachment; filename="audit-logs.csv"`
- Columns: `id`, `actor_username`, `action`, `resource_type`, `resource_id`, `result`, `ip`, `created_at`, `detail`
- Capped at 10,000 rows per export to prevent memory exhaustion

### Retention Policy

- New config field: `system.audit_log_retention_days` (default `90`)
- A background goroutine (started in `SystemModule`) runs once per day at 02:00 local time and deletes records older than the retention window
- The deletion is batched (1,000 rows per batch) to avoid long-held locks

### Backend Implementation

**`service/system/usecase/audit.go`** — add:
```go
type AuditFilter struct {
    ActorUserID  *uint
    Action       *string
    ResourceType *string
    Result       *string
    StartTime    *time.Time
    EndTime      *time.Time
    Keyword      *string
    Page         int
    PageSize     int
}

type AuditPage struct {
    Items    []AuditLogEntry
    Total    int64
    Page     int
    PageSize int
}

type AuditLogEntry struct { /* mirrors table columns + ActorUsername */ }
```

`AuditService.QueryLogs(ctx, AuditFilter) (AuditPage, error)`
`AuditService.ExportLogs(ctx, AuditFilter) ([]AuditLogEntry, error)`

**`service/system/infra/persistence/audit_store_gorm.go`** — refactor:
- `QueryLogs` builds a dynamic GORM query from `AuditFilter`
- FULLTEXT search uses raw `MATCH(detail) AGAINST(? IN BOOLEAN MODE)`
- `ExportLogs` applies same filters, `LIMIT 10000`, no pagination

**`api/v1/system/audit.go`** — add:
- `AuditHandler.QueryLogs` — parse query params, call service, return `AuditPage`
- `AuditHandler.ExportLogs` — parse same params, stream CSV via `encoding/csv`

---

## Phase 1 — Cloud Account + Unified Asset Registry

### Motivation

Every subsequent feature (cost analysis, pipeline targets, alerts) references a cloud resource. The Asset Registry is the single source of truth for what resources exist, which cloud account owns them, and their current state. Building this foundation first prevents each feature from inventing its own resource model.

### Cloud Account Module

**Purpose:** Credentials vault and connectivity baseline for each cloud provider account.

**Data model — `cloud_accounts`:**
| Column | Type | Notes |
|--------|------|-------|
| `id` | uint PK | |
| `name` | varchar(128) | Human-friendly label |
| `provider` | varchar(32) | `"aws"`, `"aliyun"`, `"tencent"`, `"azure"` |
| `account_id` | varchar(64) | Provider-native account/subscription ID |
| `credential_type` | varchar(32) | `"ak_sk"`, `"role_arn"`, `"service_principal"` |
| `credential_json` | text | AES-256-GCM encrypted at rest |
| `status` | varchar(16) | `"active"`, `"disabled"`, `"error"` |
| `last_synced_at` | datetime | Last successful asset sync |
| `created_at`, `updated_at` | datetime | |

**API endpoints:**
- `POST /api/v1/cloud/accounts` — register account
- `GET /api/v1/cloud/accounts` — list with status
- `GET /api/v1/cloud/accounts/:id` — detail
- `PUT /api/v1/cloud/accounts/:id` — update label or rotate credentials
- `DELETE /api/v1/cloud/accounts/:id` — soft delete
- `POST /api/v1/cloud/accounts/:id/verify` — test connectivity, return latency + IAM permission check

**Credential security:**
- Credentials never returned in plaintext via API (masked on read)
- Encryption key stored in config (`cloud.credential_encryption_key`), separate from JWT signing key
- Audit logged on every credential write

### Unified Asset Registry

**Purpose:** Normalized, provider-agnostic snapshot of all discovered cloud resources.

**Data model — `cloud_assets`:**
| Column | Type | Notes |
|--------|------|-------|
| `id` | uint PK | |
| `account_id` | uint FK → cloud_accounts | |
| `provider` | varchar(32) | Denormalized for query efficiency |
| `region` | varchar(64) | |
| `resource_type` | varchar(64) | `"ec2_instance"`, `"ecs_instance"`, `"rds_mysql"`, … |
| `resource_id` | varchar(128) | Provider-native resource ID |
| `name` | varchar(256) | Display name / tag |
| `status` | varchar(32) | Provider-native status |
| `config_json` | text (JSON) | Provider-specific attributes |
| `tags_json` | text (JSON) | Key-value tags |
| `discovered_at` | datetime | First seen |
| `last_seen_at` | datetime | Last sync confirmed existence |
| `deleted_at` | datetime | Soft delete — resource no longer exists |

**Indexes:** `(account_id, resource_type)`, `(provider, region, resource_type)`, `(resource_id)` unique per account.

**Sync architecture:**
- `AssetSyncer` interface: `Sync(ctx, account CloudAccount) ([]RawAsset, error)`
- One implementation per provider (`AWSAssetSyncer`, `AliyunAssetSyncer`, …)
- `AssetRegistryService.SyncAccount(ctx, accountID)` orchestrates: fetch → normalize → upsert
- Sync triggered manually (`POST /api/v1/cloud/accounts/:id/sync`) and on a configurable cron schedule

**Query API — `GET /api/v1/cloud/assets`:**

| Param | Description |
|-------|-------------|
| `accountId` | Filter by account |
| `provider` | Filter by provider |
| `region` | Filter by region |
| `resourceType` | Filter by resource type |
| `status` | Filter by status |
| `keyword` | Name/tag search |
| `page`, `pageSize` | Pagination |

---

## Architecture Decisions

1. **Asset Registry before FinOps / DevOps**: Cost data and pipeline targets are both views over assets. A shared registry prevents divergent resource models in each feature.

2. **Provider adapters behind interfaces**: `AssetSyncer`, `CostFetcher` (Phase 3), `PipelineRunner` (Phase 2) are interfaces. Provider implementations live in `infra/cloud/<provider>/`. New providers don't touch usecase code.

3. **Credential encryption at rest**: AES-256-GCM, key in config. Not in the database, not in JWT signing secrets. Rotation requires re-encrypting the `credential_json` column — acceptable for Phase 1.

4. **Soft deletes for assets**: Resources that disappear from a sync are marked `deleted_at`, not hard-deleted. Historical cost data in Phase 3 will reference these records.

5. **Audit log as compliance foundation**: Phase 0.5 is done before Phase 1 because cloud credential operations (rotate, delete) must be audit-logged with full filter/export capability before those credentials exist.
