# Audit Log Upgrade (Phase 0.5) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current unfiltered 100-record audit log list with a filter-capable paginated query, CSV export, username snapshot at write time, and a daily retention worker.

**Architecture:** `ActorUsername` is snapshotted into `system_audit_logs` at write time (via `AuditMeta.ActorUsername` populated by the HTTP layer from JWT claims). `AuditStore` gains `QueryLogs`/`ExportLogs`/`DeleteBefore`. A background goroutine purges old rows daily.

**Tech Stack:** Go 1.23, Gin, GORM (MySQL), `encoding/csv`, stdlib `time`.

---

## File Map

| File | Change |
|------|--------|
| `server/model/system/user.go` | Add `ActorUsername` to `SystemAuditLog` |
| `server/service/system/usecase/audit.go` | Add `ActorUsername` to `AuditMeta`/`AuditEvent`; new types + interface |
| `server/service/system/usecase/audit_test.go` | New — unit tests for `AuditService` |
| `server/service/system/usecase/auth.go` | Pass `ActorUsername` from user in login events |
| `server/service/system/usecase/user.go` | Pass `ActorUsername: meta.ActorUsername` in events |
| `server/service/system/usecase/role.go` | Pass `ActorUsername: meta.ActorUsername` in events |
| `server/service/system/usecase/ldap.go` | Pass `ActorUsername: meta.ActorUsername` in events |
| `server/service/system/infra/persistence/auth_store_gorm.go` | Copy `ActorUsername` in `WriteAudit` |
| `server/service/system/infra/persistence/audit_store_gorm.go` | Full rewrite: dynamic query, export, delete |
| `server/service/system/infra/persistence/audit_store_gorm_test.go` | Full rewrite: filter + export + purge tests |
| `server/api/v1/system/user.go` | `auditMeta()` extracts username from JWT claims |
| `server/api/v1/system/audit.go` | Replace `AuditLogs` with `QueryAuditLogs` + `ExportAuditLogs` |
| `server/api/v1/system/handlers_test.go` | Add nil-service + filter parse tests |
| `server/router/system/enter.go` | Update route + add export route |
| `server/config/system.go` | Add `AuditLogRetentionDays int` |
| `server/initialize/audit_retention.go` | New — daily purge worker |
| `server/initialize/startup.go` | Start retention worker |

---

## Task 1: Snapshot ActorUsername on the write path

**Files:**
- Modify: `server/model/system/user.go`
- Modify: `server/service/system/usecase/audit.go`
- Modify: `server/service/system/infra/persistence/auth_store_gorm.go`
- Modify: `server/api/v1/system/user.go`
- Modify: `server/service/system/usecase/auth.go`
- Modify: `server/service/system/usecase/user.go`
- Modify: `server/service/system/usecase/role.go`
- Modify: `server/service/system/usecase/ldap.go`

---

- [ ] **Step 1.1: Write the failing integration test**

Add to `server/service/system/infra/persistence/audit_store_gorm_test.go` (below the existing test):

```go
func TestGormAuditStoreQueryLogsPreservesActorUsername(t *testing.T) {
	db := openPersistenceTestDB(t)
	authStore := NewGormAuthStore(db)

	require.NoError(t, authStore.WriteAudit(context.Background(), usecase.AuditEvent{
		ActorUserID:   42,
		ActorUsername: "alice",
		Action:        "login",
		Result:        "success",
	}))

	store := NewGormAuditStore(db)
	items, _, err := store.ListAuditLogs(context.Background())
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "alice", items[0].ActorUsername)
}
```

- [ ] **Step 1.2: Run test to confirm it fails**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  OPTIMUS_TEST_MYSQL_DSN="root:root@tcp(127.0.0.1:3306)/optimus_test?parseTime=true" \
  go test ./service/system/infra/persistence/... -run TestGormAuditStoreQueryLogsPreservesActorUsername -v
```

Expected: FAIL — `items[0].ActorUsername` is empty (field does not exist yet).

- [ ] **Step 1.3: Add `ActorUsername` to `SystemAuditLog` in `server/model/system/user.go`**

Locate `type SystemAuditLog struct` (starts around line 95). Add the new field after `ActorUserID`:

```go
type SystemAuditLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ActorUserID   uint      `gorm:"index" json:"actorUserId"`
	ActorUsername string    `gorm:"size:64" json:"actorUsername"`
	Action        string    `gorm:"size:128;not null;index" json:"action"`
	ResourceType  string    `gorm:"size:128;index" json:"resourceType"`
	ResourceID    string    `gorm:"size:128;index" json:"resourceId"`
	Result        string    `gorm:"size:32;not null;index" json:"result"`
	IP            string    `gorm:"size:64" json:"ip"`
	UserAgent     string    `gorm:"size:255" json:"userAgent"`
	Detail        string    `gorm:"type:text" json:"detail"`
	CreatedAt     time.Time `json:"createdAt"`
}
```

- [ ] **Step 1.4: Add `ActorUsername` to `AuditMeta` and `AuditEvent` in `server/service/system/usecase/audit.go`**

Change:
```go
type AuditMeta struct {
	IP        string
	UserAgent string
}

type AuditEvent struct {
	ActorUserID  uint
	Action       string
	ResourceType string
	ResourceID   string
	Result       string
	IP           string
	UserAgent    string
	Detail       string
}
```

To:
```go
type AuditMeta struct {
	IP            string
	UserAgent     string
	ActorUsername string
}

type AuditEvent struct {
	ActorUserID   uint
	ActorUsername string
	Action        string
	ResourceType  string
	ResourceID    string
	Result        string
	IP            string
	UserAgent     string
	Detail        string
}
```

- [ ] **Step 1.5: Copy `ActorUsername` in `GormAuthStore.WriteAudit`**

In `server/service/system/infra/persistence/auth_store_gorm.go`, update `WriteAudit`:

```go
func (s *GormAuthStore) WriteAudit(ctx context.Context, event usecase.AuditEvent) error {
	log := systemModel.SystemAuditLog{
		ActorUserID:   event.ActorUserID,
		ActorUsername: event.ActorUsername,
		Action:        event.Action,
		ResourceType:  event.ResourceType,
		ResourceID:    event.ResourceID,
		Result:        event.Result,
		IP:            event.IP,
		UserAgent:     event.UserAgent,
		Detail:        event.Detail,
	}
	return s.db.WithContext(ctx).Create(&log).Error
}
```

- [ ] **Step 1.6: Update `auditMeta` in `server/api/v1/system/user.go` to extract username from JWT claims**

Change:
```go
func auditMeta(c *gin.Context) systemUsecase.AuditMeta {
	return systemUsecase.AuditMeta{IP: c.ClientIP(), UserAgent: c.Request.UserAgent()}
}
```

To:
```go
func auditMeta(c *gin.Context) systemUsecase.AuditMeta {
	meta := systemUsecase.AuditMeta{IP: c.ClientIP(), UserAgent: c.Request.UserAgent()}
	if raw, ok := c.Get(utils.ClaimsContextKey); ok {
		if cc, ok := raw.(*utils.CustomClaims); ok {
			meta.ActorUsername = cc.BaseClaims.Username
		}
	}
	return meta
}
```

Ensure `"optimus/server/utils"` is still in the import block (it already is).

- [ ] **Step 1.7: Pass `ActorUsername` from user object in auth login events**

In `server/service/system/usecase/auth.go`, update `issueLoginResult` to populate the username in its success audit event:

```go
s.writeAudit(ctx, AuditEvent{
	ActorUserID:   user.ID,
	ActorUsername: user.Username,
	Action:        "auth.login.success",
	Result:        "success",
	IP:            meta.IP,
	UserAgent:     meta.UserAgent,
})
```

Also update the two "user disabled" events that have `user.ID` available:

```go
// After "if user.Status != defaultUserStatus" in Login():
s.writeAudit(ctx, AuditEvent{
	ActorUserID:   user.ID,
	ActorUsername: user.Username,
	Action:        "auth.login.failure",
	Result:        "failure",
	IP:            meta.IP,
	UserAgent:     meta.UserAgent,
	Detail:        "user disabled",
})

// After "if user.Source != "local" || ..." check in Login():
s.writeAudit(ctx, AuditEvent{
	ActorUserID:   user.ID,
	ActorUsername: user.Username,
	Action:        "auth.login.failure",
	Result:        "failure",
	IP:            meta.IP,
	UserAgent:     meta.UserAgent,
	Detail:        "invalid credentials",
})
```

For `loginLDAP` success, update `issueLoginResult`'s call (already covered above since it goes through `issueLoginResult`). The LDAP disabled-user event also has `user.ID` and `user.Username`:

```go
s.writeAudit(ctx, AuditEvent{
	ActorUserID:   user.ID,
	ActorUsername: user.Username,
	Action:        "auth.login.failure",
	Result:        "failure",
	IP:            meta.IP,
	UserAgent:     meta.UserAgent,
	Detail:        "user disabled",
})
```

- [ ] **Step 1.8: Pass `ActorUsername: meta.ActorUsername` in user service audit events**

In `server/service/system/usecase/user.go`, update all three `WriteAudit` calls. Example for `UpdateUser`:

```go
_ = s.store.WriteAudit(ctx, AuditEvent{
	ActorUserID:   actorID,
	ActorUsername: meta.ActorUsername,
	Action:        "user.update",
	ResourceType:  "user",
	ResourceID:    resourceID(userID),
	Result:        "success",
	IP:            meta.IP,
	UserAgent:     meta.UserAgent,
})
```

Apply the same `ActorUsername: meta.ActorUsername` addition to `UpdateUserStatus` and `UpdateUserRoles`.

- [ ] **Step 1.9: Pass `ActorUsername: meta.ActorUsername` in role service audit events**

In `server/service/system/usecase/role.go`, add `ActorUsername: meta.ActorUsername` to all four `WriteAudit` calls (`CreateRole`, `UpdateRole`, `UpdateRoleStatus`, `UpdateRolePermissions`).

- [ ] **Step 1.10: Pass `ActorUsername: meta.ActorUsername` in LDAP service audit events**

In `server/service/system/usecase/ldap.go`, add `ActorUsername: meta.ActorUsername` to all three `WriteAudit` calls (`SaveConfig`, `TestConfig` failure, `TestConfig` success).

- [ ] **Step 1.11: Run the previously failing test to confirm it now passes**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  OPTIMUS_TEST_MYSQL_DSN="root:root@tcp(127.0.0.1:3306)/optimus_test?parseTime=true" \
  go test ./service/system/infra/persistence/... -run TestGormAuditStoreQueryLogsPreservesActorUsername -v
```

Expected: PASS.

- [ ] **Step 1.12: Full build + existing test suite**

```bash
cd /home/lilin/trae-project/Optimus/server && go build ./... && go test ./...
```

Expected: all pass (integration tests skip if DSN absent).

- [ ] **Step 1.13: Commit**

```bash
cd /home/lilin/trae-project/Optimus && \
git add server/model/system/user.go \
        server/service/system/usecase/audit.go \
        server/service/system/usecase/auth.go \
        server/service/system/usecase/user.go \
        server/service/system/usecase/role.go \
        server/service/system/usecase/ldap.go \
        server/service/system/infra/persistence/auth_store_gorm.go \
        server/service/system/infra/persistence/audit_store_gorm_test.go \
        server/api/v1/system/user.go && \
git commit -m "feat(audit): snapshot actor_username at write time"
```

---

## Task 2: Extend AuditStore interface with filter/export/purge + AuditService unit tests

**Files:**
- Modify: `server/service/system/usecase/audit.go`
- Create: `server/service/system/usecase/audit_test.go`
- Modify: `server/service/system/infra/persistence/audit_store_gorm.go` (compile stubs only)

---

- [ ] **Step 2.1: Write the failing unit tests in new file `server/service/system/usecase/audit_test.go`**

```go
package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeAuditStore struct {
	page    AuditPage
	entries []AuditLogEntry
	deleted time.Time
}

func (f *fakeAuditStore) QueryLogs(_ context.Context, _ AuditFilter) (AuditPage, error) {
	return f.page, nil
}

func (f *fakeAuditStore) ExportLogs(_ context.Context, _ AuditFilter) ([]AuditLogEntry, error) {
	return f.entries, nil
}

func (f *fakeAuditStore) DeleteBefore(_ context.Context, before time.Time) error {
	f.deleted = before
	return nil
}

func TestAuditServiceQueryLogsReturnsPageFromStore(t *testing.T) {
	expected := AuditPage{
		Items:    []AuditLogEntry{{ID: 1, Action: "login", ActorUsername: "admin"}},
		Total:    1,
		Page:     1,
		PageSize: 20,
	}
	svc := NewAuditService(&fakeAuditStore{page: expected})
	got, err := svc.QueryLogs(context.Background(), AuditFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, expected, got)
}

func TestAuditServiceExportLogsReturnsEntriesFromStore(t *testing.T) {
	expected := []AuditLogEntry{{ID: 1, Action: "login", ActorUsername: "admin"}}
	svc := NewAuditService(&fakeAuditStore{entries: expected})
	got, err := svc.ExportLogs(context.Background(), AuditFilter{})
	require.NoError(t, err)
	require.Equal(t, expected, got)
}

func TestAuditServicePurgeBeforeCallsDeleteBefore(t *testing.T) {
	store := &fakeAuditStore{}
	svc := NewAuditService(store)
	cutoff := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, svc.PurgeBefore(context.Background(), cutoff))
	require.Equal(t, cutoff, store.deleted)
}
```

- [ ] **Step 2.2: Run tests to confirm they fail**

```bash
cd /home/lilin/trae-project/Optimus/server && go test ./service/system/usecase/... -run "TestAuditService" -v
```

Expected: FAIL — `AuditFilter`, `AuditPage`, `AuditLogEntry`, `QueryLogs`, `ExportLogs`, `PurgeBefore` not defined.

- [ ] **Step 2.3: Rewrite `server/service/system/usecase/audit.go` with new types and service**

```go
package usecase

import (
	"context"
	"time"
)

type AuditMeta struct {
	IP            string
	UserAgent     string
	ActorUsername string
}

type AuditEvent struct {
	ActorUserID   uint
	ActorUsername string
	Action        string
	ResourceType  string
	ResourceID    string
	Result        string
	IP            string
	UserAgent     string
	Detail        string
}

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
	Items    []AuditLogEntry `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

type AuditLogEntry struct {
	ID            uint      `json:"id"`
	ActorUserID   uint      `json:"actorUserId"`
	ActorUsername string    `json:"actorUsername"`
	Action        string    `json:"action"`
	ResourceType  string    `json:"resourceType"`
	ResourceID    string    `json:"resourceId"`
	Result        string    `json:"result"`
	IP            string    `json:"ip"`
	UserAgent     string    `json:"userAgent"`
	Detail        string    `json:"detail"`
	CreatedAt     time.Time `json:"createdAt"`
}

type AuditStore interface {
	QueryLogs(ctx context.Context, f AuditFilter) (AuditPage, error)
	ExportLogs(ctx context.Context, f AuditFilter) ([]AuditLogEntry, error)
	DeleteBefore(ctx context.Context, before time.Time) error
}

type AuditService struct {
	store AuditStore
}

func NewAuditService(store AuditStore) *AuditService {
	return &AuditService{store: store}
}

func (s *AuditService) QueryLogs(ctx context.Context, f AuditFilter) (AuditPage, error) {
	return s.store.QueryLogs(ctx, f)
}

func (s *AuditService) ExportLogs(ctx context.Context, f AuditFilter) ([]AuditLogEntry, error) {
	return s.store.ExportLogs(ctx, f)
}

func (s *AuditService) PurgeBefore(ctx context.Context, before time.Time) error {
	return s.store.DeleteBefore(ctx, before)
}
```

Note: `AuditLogListResult` and `ListAuditLogs` are intentionally removed. The `systemModel` import is gone too.

- [ ] **Step 2.4: Add compile stubs to `GormAuditStore` so the project builds**

`GormAuditStore` must satisfy the new `AuditStore` interface or `NewAuditService(persistence.NewGormAuditStore(db))` in `system_module.go` will fail to compile. Add stubs to `server/service/system/infra/persistence/audit_store_gorm.go`:

Replace the entire file with:

```go
package persistence

import (
	"context"
	"time"

	systemModel "optimus/server/model/system"
	"optimus/server/service/system/usecase"

	"gorm.io/gorm"
)

type GormAuditStore struct {
	db *gorm.DB
}

func NewGormAuditStore(db *gorm.DB) *GormAuditStore {
	return &GormAuditStore{db: db}
}

// ListAuditLogs is kept temporarily to avoid breaking the existing test; removed in Task 3.
func (s *GormAuditStore) ListAuditLogs(ctx context.Context) ([]systemModel.SystemAuditLog, int64, error) {
	var total int64
	if err := s.db.WithContext(ctx).Model(&systemModel.SystemAuditLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []systemModel.SystemAuditLog
	err := s.db.WithContext(ctx).Order("created_at DESC").Limit(100).Find(&items).Error
	return items, total, err
}

func (s *GormAuditStore) QueryLogs(_ context.Context, _ usecase.AuditFilter) (usecase.AuditPage, error) {
	return usecase.AuditPage{}, nil // implemented in Task 3
}

func (s *GormAuditStore) ExportLogs(_ context.Context, _ usecase.AuditFilter) ([]usecase.AuditLogEntry, error) {
	return nil, nil // implemented in Task 3
}

func (s *GormAuditStore) DeleteBefore(_ context.Context, _ time.Time) error {
	return nil // implemented in Task 3
}
```

- [ ] **Step 2.5: Run the unit tests to confirm they now pass**

```bash
cd /home/lilin/trae-project/Optimus/server && go test ./service/system/usecase/... -run "TestAuditService" -v
```

Expected: all 3 PASS.

- [ ] **Step 2.6: Full build + full test suite**

```bash
cd /home/lilin/trae-project/Optimus/server && go build ./... && go test ./...
```

Expected: all pass (DSN-gated integration tests skip).

- [ ] **Step 2.7: Commit**

```bash
cd /home/lilin/trae-project/Optimus && \
git add server/service/system/usecase/audit.go \
        server/service/system/usecase/audit_test.go \
        server/service/system/infra/persistence/audit_store_gorm.go && \
git commit -m "feat(audit): add AuditFilter/AuditPage types and filter-capable AuditService"
```

---

## Task 3: Implement dynamic query in GormAuditStore + integration tests

**Files:**
- Modify: `server/service/system/infra/persistence/audit_store_gorm.go`
- Modify: `server/service/system/infra/persistence/audit_store_gorm_test.go`

---

- [ ] **Step 3.1: Write failing integration tests**

Replace `server/service/system/infra/persistence/audit_store_gorm_test.go` entirely:

```go
package persistence

import (
	"context"
	"testing"
	"time"

	"optimus/server/service/system/usecase"

	"github.com/stretchr/testify/require"
)

func TestGormAuditStoreQueryLogsOrdersNewestFirst(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "first", Result: "success"}))
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "second", Result: "success"}))

	store := NewGormAuditStore(db)
	page, err := store.QueryLogs(context.Background(), usecase.AuditFilter{Page: 1, PageSize: 20})

	require.NoError(t, err)
	require.Equal(t, int64(2), page.Total)
	require.Len(t, page.Items, 2)
	require.True(t,
		page.Items[0].CreatedAt.After(page.Items[1].CreatedAt) ||
			page.Items[0].CreatedAt.Equal(page.Items[1].CreatedAt),
		"items must be newest-first")
}

func TestGormAuditStoreQueryLogsPaginates(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	for i := 0; i < 5; i++ {
		require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "x", Result: "success"}))
	}

	store := NewGormAuditStore(db)
	page, err := store.QueryLogs(context.Background(), usecase.AuditFilter{Page: 1, PageSize: 2})

	require.NoError(t, err)
	require.Equal(t, int64(5), page.Total)
	require.Len(t, page.Items, 2)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 2, page.PageSize)
}

func TestGormAuditStoreQueryLogsFiltersByResult(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "a", Result: "success"}))
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "b", Result: "failure"}))

	store := NewGormAuditStore(db)
	result := "success"
	page, err := store.QueryLogs(context.Background(), usecase.AuditFilter{
		Page: 1, PageSize: 20, Result: &result,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "success", page.Items[0].Result)
}

func TestGormAuditStoreQueryLogsFiltersByAction(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "login", Result: "success"}))
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "logout", Result: "success"}))

	store := NewGormAuditStore(db)
	action := "login"
	page, err := store.QueryLogs(context.Background(), usecase.AuditFilter{
		Page: 1, PageSize: 20, Action: &action,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "login", page.Items[0].Action)
}

func TestGormAuditStoreQueryLogsFiltersByTimeRange(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "login", Result: "success"}))

	store := NewGormAuditStore(db)
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	// no result: window entirely in the past
	pagePast, err := store.QueryLogs(context.Background(), usecase.AuditFilter{
		Page: 1, PageSize: 20, EndTime: &past,
	})
	require.NoError(t, err)
	require.Equal(t, int64(0), pagePast.Total)

	// one result: window contains now
	pageNow, err := store.QueryLogs(context.Background(), usecase.AuditFilter{
		Page: 1, PageSize: 20, StartTime: &past, EndTime: &future,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), pageNow.Total)
}

func TestGormAuditStoreExportLogsReturnsSortedEntries(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{
		ActorUsername: "alice", Action: "login", Result: "success",
	}))

	store := NewGormAuditStore(db)
	entries, err := store.ExportLogs(context.Background(), usecase.AuditFilter{})

	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "alice", entries[0].ActorUsername)
	require.Equal(t, "login", entries[0].Action)
}

func TestGormAuditStoreDeleteBefore(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "old", Result: "success"}))

	store := NewGormAuditStore(db)
	future := time.Now().Add(time.Hour)
	require.NoError(t, store.DeleteBefore(context.Background(), future))

	page, err := store.QueryLogs(context.Background(), usecase.AuditFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(0), page.Total)
}
```

- [ ] **Step 3.2: Run tests to confirm they fail (stubs return empty)**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  OPTIMUS_TEST_MYSQL_DSN="root:root@tcp(127.0.0.1:3306)/optimus_test?parseTime=true" \
  go test ./service/system/infra/persistence/... -v
```

Expected: `TestGormAuditStoreQueryLogs*`, `TestGormAuditStoreExportLogs*`, `TestGormAuditStoreDeleteBefore` FAIL. The original `TestGormAuditStoreQueryLogsPreservesActorUsername` should still PASS (uses the preserved `ListAuditLogs` stub).

- [ ] **Step 3.3: Implement `GormAuditStore` with real query logic**

Replace the entire `server/service/system/infra/persistence/audit_store_gorm.go`:

```go
package persistence

import (
	"context"
	"time"

	systemModel "optimus/server/model/system"
	"optimus/server/service/system/usecase"

	"gorm.io/gorm"
)

type GormAuditStore struct {
	db *gorm.DB
}

func NewGormAuditStore(db *gorm.DB) *GormAuditStore {
	return &GormAuditStore{db: db}
}

func (s *GormAuditStore) buildQuery(ctx context.Context, f usecase.AuditFilter) *gorm.DB {
	q := s.db.WithContext(ctx).Model(&systemModel.SystemAuditLog{})
	if f.ActorUserID != nil {
		q = q.Where("actor_user_id = ?", *f.ActorUserID)
	}
	if f.Action != nil {
		q = q.Where("action = ?", *f.Action)
	}
	if f.ResourceType != nil {
		q = q.Where("resource_type = ?", *f.ResourceType)
	}
	if f.Result != nil {
		q = q.Where("result = ?", *f.Result)
	}
	if f.StartTime != nil {
		q = q.Where("created_at >= ?", *f.StartTime)
	}
	if f.EndTime != nil {
		q = q.Where("created_at <= ?", *f.EndTime)
	}
	if f.Keyword != nil && *f.Keyword != "" {
		q = q.Where("MATCH(detail) AGAINST(? IN BOOLEAN MODE)", *f.Keyword)
	}
	return q
}

func (s *GormAuditStore) QueryLogs(ctx context.Context, f usecase.AuditFilter) (usecase.AuditPage, error) {
	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	q := s.buildQuery(ctx, f)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return usecase.AuditPage{}, err
	}

	var rows []systemModel.SystemAuditLog
	err := q.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error
	if err != nil {
		return usecase.AuditPage{}, err
	}

	return usecase.AuditPage{
		Items:    toAuditLogEntries(rows),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *GormAuditStore) ExportLogs(ctx context.Context, f usecase.AuditFilter) ([]usecase.AuditLogEntry, error) {
	q := s.buildQuery(ctx, f)
	var rows []systemModel.SystemAuditLog
	err := q.Order("created_at DESC").Limit(10000).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toAuditLogEntries(rows), nil
}

func (s *GormAuditStore) DeleteBefore(ctx context.Context, before time.Time) error {
	for {
		result := s.db.WithContext(ctx).
			Exec("DELETE FROM system_audit_logs WHERE created_at < ? LIMIT 1000", before)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
	}
}

func toAuditLogEntries(rows []systemModel.SystemAuditLog) []usecase.AuditLogEntry {
	entries := make([]usecase.AuditLogEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, usecase.AuditLogEntry{
			ID:            row.ID,
			ActorUserID:   row.ActorUserID,
			ActorUsername: row.ActorUsername,
			Action:        row.Action,
			ResourceType:  row.ResourceType,
			ResourceID:    row.ResourceID,
			Result:        row.Result,
			IP:            row.IP,
			UserAgent:     row.UserAgent,
			Detail:        row.Detail,
			CreatedAt:     row.CreatedAt,
		})
	}
	return entries
}
```

Note: `ListAuditLogs` is intentionally removed. The test for it in Step 1 also used `ListAuditLogs` — update that test in the next step.

- [ ] **Step 3.4: Update `TestGormAuditStoreQueryLogsPreservesActorUsername` to use `QueryLogs`**

In `server/service/system/infra/persistence/audit_store_gorm_test.go`, replace the `TestGormAuditStoreQueryLogsPreservesActorUsername` test:

```go
func TestGormAuditStoreQueryLogsPreservesActorUsername(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{
		ActorUserID:   42,
		ActorUsername: "alice",
		Action:        "login",
		Result:        "success",
	}))

	store := NewGormAuditStore(db)
	page, err := store.QueryLogs(context.Background(), usecase.AuditFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, "alice", page.Items[0].ActorUsername)
}
```

- [ ] **Step 3.5: Run all persistence tests**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  OPTIMUS_TEST_MYSQL_DSN="root:root@tcp(127.0.0.1:3306)/optimus_test?parseTime=true" \
  go test ./service/system/infra/persistence/... -v
```

Expected: all tests PASS.

- [ ] **Step 3.6: Full build + test suite**

```bash
cd /home/lilin/trae-project/Optimus/server && go build ./... && go test ./...
```

Expected: all pass.

- [ ] **Step 3.7: Commit**

```bash
cd /home/lilin/trae-project/Optimus && \
git add server/service/system/infra/persistence/audit_store_gorm.go \
        server/service/system/infra/persistence/audit_store_gorm_test.go && \
git commit -m "feat(audit): implement dynamic filter/export/purge in GormAuditStore"
```

---

## Task 4: Rewrite HTTP audit handler and update router

**Files:**
- Modify: `server/api/v1/system/audit.go`
- Modify: `server/api/v1/system/handlers_test.go`
- Modify: `server/router/system/enter.go`

---

- [ ] **Step 4.1: Write failing handler tests**

Add to the end of `server/api/v1/system/handlers_test.go`:

```go
func TestAuditHandlerQueryAuditLogsReturnsServiceUnavailableWhenServiceMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/system/audit-logs", nil)

	handler := &AuditHandler{}
	handler.QueryAuditLogs(ctx)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.JSONEq(t, `{"message":"audit service is not available"}`, recorder.Body.String())
}

func TestAuditHandlerExportAuditLogsReturnsServiceUnavailableWhenServiceMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/system/audit-logs/export", nil)

	handler := &AuditHandler{}
	handler.ExportAuditLogs(ctx)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestParseAuditFilterParsesResultAndActionQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		"/system/audit-logs?page=2&pageSize=50&result=success&action=login",
		nil,
	)

	f := parseAuditFilter(ctx)

	require.Equal(t, 2, f.Page)
	require.Equal(t, 50, f.PageSize)
	require.NotNil(t, f.Result)
	require.Equal(t, "success", *f.Result)
	require.NotNil(t, f.Action)
	require.Equal(t, "login", *f.Action)
}

func TestParseAuditFilterClampsInvalidPageSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/system/audit-logs?pageSize=999", nil)

	f := parseAuditFilter(ctx)

	require.Equal(t, 1, f.Page)
	require.Equal(t, 20, f.PageSize)
}

func TestParseAuditFilterParsesTimeRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		"/system/audit-logs?startTime=2026-01-01T00:00:00Z&endTime=2026-12-31T23:59:59Z",
		nil,
	)

	f := parseAuditFilter(ctx)

	require.NotNil(t, f.StartTime)
	require.Equal(t, 2026, f.StartTime.Year())
	require.NotNil(t, f.EndTime)
	require.Equal(t, 12, int(f.EndTime.Month()))
}
```

- [ ] **Step 4.2: Run tests to confirm they fail**

```bash
cd /home/lilin/trae-project/Optimus/server && go test ./api/v1/system/... -run "TestAuditHandler|TestParseAuditFilter" -v
```

Expected: FAIL — `QueryAuditLogs`, `ExportAuditLogs`, `parseAuditFilter` not defined.

- [ ] **Step 4.3: Rewrite `server/api/v1/system/audit.go`**

```go
package system

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	systemUsecase "optimus/server/service/system/usecase"

	"github.com/gin-gonic/gin"
)

func (h *AuditHandler) QueryAuditLogs(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "audit service is not available")
		return
	}
	result, err := h.service.QueryLogs(c.Request.Context(), parseAuditFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to query audit logs"})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *AuditHandler) ExportAuditLogs(c *gin.Context) {
	if h.service == nil {
		writeServiceUnavailable(c, "audit service is not available")
		return
	}
	entries, err := h.service.ExportLogs(c.Request.Context(), parseAuditFilter(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to export audit logs"})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="audit-logs.csv"`)

	w := csv.NewWriter(c.Writer)
	_ = w.Write([]string{"id", "actor_username", "action", "resource_type", "resource_id", "result", "ip", "created_at", "detail"})
	for _, entry := range entries {
		_ = w.Write([]string{
			strconv.FormatUint(uint64(entry.ID), 10),
			entry.ActorUsername,
			entry.Action,
			entry.ResourceType,
			entry.ResourceID,
			entry.Result,
			entry.IP,
			entry.CreatedAt.Format(time.RFC3339),
			entry.Detail,
		})
	}
	w.Flush()
}

func parseAuditFilter(c *gin.Context) systemUsecase.AuditFilter {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	f := systemUsecase.AuditFilter{Page: page, PageSize: pageSize}

	if v := c.Query("actorUserId"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			uid := uint(id)
			f.ActorUserID = &uid
		}
	}
	if v := c.Query("action"); v != "" {
		f.Action = &v
	}
	if v := c.Query("resourceType"); v != "" {
		f.ResourceType = &v
	}
	if v := c.Query("result"); v != "" {
		f.Result = &v
	}
	if v := c.Query("startTime"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.StartTime = &t
		}
	}
	if v := c.Query("endTime"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.EndTime = &t
		}
	}
	if v := c.Query("keyword"); v != "" {
		f.Keyword = &v
	}
	return f
}
```

- [ ] **Step 4.4: Update `server/router/system/enter.go`**

Replace:
```go
privateGroup.GET("/system/audit-logs", handlers.Audit.AuditLogs)
```

With:
```go
privateGroup.GET("/system/audit-logs", handlers.Audit.QueryAuditLogs)
privateGroup.GET("/system/audit-logs/export", handlers.Audit.ExportAuditLogs)
```

- [ ] **Step 4.5: Run the handler tests**

```bash
cd /home/lilin/trae-project/Optimus/server && go test ./api/v1/system/... -run "TestAuditHandler|TestParseAuditFilter" -v
```

Expected: all 5 new tests PASS.

- [ ] **Step 4.6: Full build + test suite**

```bash
cd /home/lilin/trae-project/Optimus/server && go build ./... && go test ./...
```

Expected: all pass.

- [ ] **Step 4.7: Commit**

```bash
cd /home/lilin/trae-project/Optimus && \
git add server/api/v1/system/audit.go \
        server/api/v1/system/handlers_test.go \
        server/router/system/enter.go && \
git commit -m "feat(audit): add filter query and CSV export endpoints"
```

---

## Task 5: Retention config and daily background worker

**Files:**
- Modify: `server/config/system.go`
- Create: `server/initialize/audit_retention.go`
- Modify: `server/initialize/startup.go`

---

- [ ] **Step 5.1: Write the failing test for the retention worker**

Create `server/initialize/audit_retention_test.go`:

```go
package initialize

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAuditRetentionWorkerCallsPurgeAfterTick(t *testing.T) {
	var purgeCalled bool
	var purgedBefore time.Time

	fakePurger := func(ctx context.Context, before time.Time) error {
		purgeCalled = true
		purgedBefore = before
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a tick interval of 0 so the first tick fires immediately in tests.
	go runRetentionLoop(ctx, fakePurger, 30, 0)

	// Give the goroutine time to fire.
	time.Sleep(50 * time.Millisecond)
	cancel()

	require.True(t, purgeCalled)
	// Cutoff should be ~30 days in the past.
	diff := time.Since(purgedBefore)
	require.Greater(t, diff, 29*24*time.Hour)
	require.Less(t, diff, 31*24*time.Hour)
}
```

- [ ] **Step 5.2: Run to confirm it fails**

```bash
cd /home/lilin/trae-project/Optimus/server && go test ./initialize/... -run TestAuditRetentionWorkerCallsPurgeAfterTick -v
```

Expected: FAIL — `runRetentionLoop` not defined.

- [ ] **Step 5.3: Add `AuditLogRetentionDays` to `server/config/system.go`**

```go
package config

type System struct {
	Env                    string `mapstructure:"env"`
	Addr                   int    `mapstructure:"addr"`
	DBType                 string `mapstructure:"db-type"`
	UseRedis               bool   `mapstructure:"use-redis"`
	UseCasbin              bool   `mapstructure:"use-casbin"`
	RouterPrefix           string `mapstructure:"router-prefix"`
	AdminInitialPassword   string `mapstructure:"admin-initial-password"`
	AuditLogRetentionDays  int    `mapstructure:"audit-log-retention-days"`
}
```

- [ ] **Step 5.4: Create `server/initialize/audit_retention.go`**

```go
package initialize

import (
	"context"
	"time"

	"optimus/server/service/system/infra/persistence"
	systemUsecase "optimus/server/service/system/usecase"

	"gorm.io/gorm"
)

func startAuditRetentionWorker(ctx context.Context, db *gorm.DB, retentionDays int) {
	if retentionDays <= 0 {
		retentionDays = 90
	}
	svc := systemUsecase.NewAuditService(persistence.NewGormAuditStore(db))
	go runRetentionLoop(ctx, svc.PurgeBefore, retentionDays, nextDailyTick())
}

func runRetentionLoop(ctx context.Context, purge func(context.Context, time.Time) error, retentionDays int, firstTick time.Duration) {
	timer := time.NewTimer(firstTick)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			cutoff := time.Now().AddDate(0, 0, -retentionDays)
			_ = purge(ctx, cutoff)
			timer.Reset(24 * time.Hour)
		}
	}
}

// nextDailyTick returns the duration until the next 02:00 local time.
func nextDailyTick() time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return time.Until(next)
}
```

- [ ] **Step 5.5: Wire the worker into `server/initialize/startup.go`**

After the migration step, add the goroutine start. The full `Startup()` becomes:

```go
func Startup() {
	db := Gorm()
	if db == nil {
		return
	}

	Redis()
	Casbin()

	if err := systemService.AutoMigrateSystem(db); err != nil {
		logDependencyInitFailure("system", "auto migrate", err)
		return
	}
	if err := systemService.SeedSystemData(db, global.OPTIMUS_CONFIG.System.AdminInitialPassword); err != nil {
		logDependencyInitFailure("system", "seed data", err)
	}

	startAuditRetentionWorker(context.Background(), db, global.OPTIMUS_CONFIG.System.AuditLogRetentionDays)
}
```

Add `"context"` to the import block of `startup.go`.

- [ ] **Step 5.6: Run the retention worker test**

```bash
cd /home/lilin/trae-project/Optimus/server && go test ./initialize/... -run TestAuditRetentionWorkerCallsPurgeAfterTick -v
```

Expected: PASS.

- [ ] **Step 5.7: Full build + test suite**

```bash
cd /home/lilin/trae-project/Optimus/server && go build ./... && go test ./...
```

Expected: all pass.

- [ ] **Step 5.8: Commit**

```bash
cd /home/lilin/trae-project/Optimus && \
git add server/config/system.go \
        server/initialize/audit_retention.go \
        server/initialize/audit_retention_test.go \
        server/initialize/startup.go && \
git commit -m "feat(audit): add configurable retention worker (default 90 days)"
```

---

## Final verification

- [ ] **Step 6.1: Full build and test**

```bash
cd /home/lilin/trae-project/Optimus/server && go build ./... && go test ./...
```

Expected: all pass, no compile warnings.

- [ ] **Step 6.2: Verify new routes are registered**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  go test ./initialize/... -run "TestRouters" -v
```

Expected: existing router tests PASS (they do not call audit routes, so no change needed).
