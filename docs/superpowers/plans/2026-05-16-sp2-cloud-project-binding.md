# SP2 Cloud 模块绑定项目 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让既有 cloud 模块按"当前项目 + 用户可访问项目"自动隔离，所有 cloud_accounts / cloud_assets 查询与写入都被项目维度过滤。

**Architecture:** 给 `cloud_accounts` 加 `project_id` 列；新增 `middleware/project_scope.go` 在 JWT 之后解析 `X-Project-Code` header、查表得到用户可访问项目列表，写入 gin ctx；cloud handlers 从 ctx 取出 currentProjectID + accessibleProjectIDs，作为显式参数传入 service；store 层 SQL 加 `WHERE project_id IN (?)` 过滤。startup 时通过 `BackfillCloudProjectScope` 把既有账号划入 `default` 项目，admin 自动成为 default 项目 owner。

**Tech Stack:** Go 1.23、GORM v2、Gin、testify；MySQL（持久化测试需 `OPTIMUS_TEST_MYSQL_DSN`）。

**Spec reference:** `docs/superpowers/specs/2026-05-14-multi-project-cloud-isolation-design.md` §SP2 + §Permission Model + §API Surface。SP1 plan 已落地 `projects` / `user_projects` 表、`ProjectService.ListAccessibleProjectIDs`、`FindMember`、`FindByCode`。

---

## Requirements Summary

- 给 `cloud_accounts` 增加项目归属字段，并把历史账号幂等迁移到 `default` 项目。
- 在私有 API 链路中解析 `X-Project-Code`，生成当前请求的 `CurrentProjectID` 与 `AccessibleProjectIDs`。
- Cloud account 的创建使用当前项目，读取、更新、删除、同步只允许访问用户可见项目内的账号。
- Cloud asset 查询通过 `cloud_assets.account_id -> cloud_accounts.project_id` 关联过滤，不在 `cloud_assets` 冗余 `project_id`。
- 保持既有 cloud API 路径不变，为 SP3/SP4 前端项目切换与 cloud 页面项目化提供后端契约。

## Acceptance Criteria

- `cloud_accounts.project_id` 存在、非空可用；历史 `project_id = 0` 账号在启动回填后绑定到 `default` 项目。
- 未携带 `X-Project-Code` 创建 cloud account 返回 400；携带不可见项目返回 403；携带不存在项目返回 404。
- 用户调用 account list/detail/update/delete/sync 时，只能操作 `AccessibleProjectIDs` 内账号；跨项目账号表现为 not found/forbidden，不泄露数据。
- asset list 只返回可访问项目账号下的资产；空 `AccessibleProjectIDs` 返回空页。
- 现有 cloud 路由路径不变，SP3 只需注入 `X-Project-Code` header 即可切换项目上下文。
- `go build ./...` clean；`go test ./...` PASS（无 MySQL 时持久化集成测试按现有逻辑 SKIP）。

## Risks and Mitigations

- **Startup migration order:** `BackfillCloudProjectScope` 依赖 system admin 与 project 表。Mitigation：只在 `SeedProjectData(db)` 后调用，并让 backfill 幂等。
- **Router middleware blast radius:** `ProjectScope` 注册在 private group，会影响 project/system 路由。Mitigation：中间件在无 header 时只注入 accessible list，不改变非 cloud handler 行为；nil service 不注册。
- **JOIN column ambiguity:** asset 查询 JOIN `cloud_accounts` 后可能出现同名列。Mitigation：Task 5 要求所有 `cloud_assets` 查询列加表前缀。
- **Internal sync bypass:** `GetForSync` 不带项目过滤。Mitigation：`SyncAccount` 入口先调用 filtered `AccountStore.Get` 做访问校验，之后才调用 `GetForSync`。
- **Existing tests compile drift:** 接口签名变化会让 handler/usecase/store 测试一起失败。Mitigation：按 Task 2-7 顺序逐层更新，每层运行对应 build/test。

## Verification Steps

- Unit: `go test ./service/cloud/usecase/... -v`
- Middleware: `go test ./middleware/... -v -run ProjectScope`
- Persistence: `go test ./service/cloud/infra/persistence/...`（无 MySQL 时 SKIP；有 `OPTIMUS_TEST_MYSQL_DSN` 时 PASS）
- Full backend: `go build ./... && go test ./...`
- Manual API smoke after implementation: login 获取 JWT 后，用同一用户分别请求可访问 / 不可访问 `X-Project-Code` 的 `/api/v1/cloud/accounts`、`/api/v1/cloud/assets`、`POST /api/v1/cloud/accounts/:id/sync`。

## File Structure

**新增：**
- `server/middleware/project_scope.go` — `ProjectScope(resolver)` 中间件、`ProjectScopeResolver` 接口
- `server/middleware/project_scope_test.go` — 中间件单元测试
- `server/utils/project_scope.go` — `ProjectScopeContextKey` 常量 + ctx helpers

**修改：**
- `server/model/cloud/account.go` — `CloudAccount` 加 `ProjectID uint`
- `server/service/cloud/cloud.go` — 新增 `BackfillCloudProjectScope(db)` 函数
- `server/service/cloud/usecase/account.go` — `AccountRow` / `AccountStore` 接口 / `CloudAccountService` 方法签名扩展项目维度
- `server/service/cloud/usecase/account_test.go` — fake store + 现有测试更新
- `server/service/cloud/usecase/asset.go` — `AssetFilter` 加 `AccessibleProjectIDs`、`SyncAccount` 加访问校验
- `server/service/cloud/usecase/asset_test.go`（若存在）/ 测试更新
- `server/service/cloud/infra/persistence/account_store_gorm.go` — Get/List/Update/Delete 加 `project_id IN (?)` 过滤
- `server/service/cloud/infra/persistence/account_store_gorm_test.go` — 集成测试更新
- `server/service/cloud/infra/persistence/asset_store_gorm.go` — Query 加 JOIN cloud_accounts 按 project_id 过滤
- `server/service/cloud/infra/persistence/asset_store_gorm_test.go` — 集成测试更新
- `server/api/v1/cloud/handlers.go` — 加 `projectScope(c *gin.Context) utils.ProjectScope` 辅助
- `server/api/v1/cloud/account.go` / `asset.go` — handlers 读取 scope 并传入 service
- `server/initialize/router.go` — 在 privateGroup JWTAuth 之后注册 `middleware.ProjectScope(...)`
- `server/initialize/startup.go` — 调用 `cloudService.BackfillCloudProjectScope(db)` 在 `SeedProjectData` 后

---

### Task 1: CloudAccount 模型加 project_id + Backfill 函数

**Files:**
- Modify: `server/model/cloud/account.go`
- Modify: `server/service/cloud/cloud.go`

- [ ] **Step 1.1: 改 CloudAccount 加列**

文件: `server/model/cloud/account.go`，在 `Provider` 之前插入 `ProjectID`：

```go
type CloudAccount struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ProjectID      uint           `gorm:"not null;default:0;index" json:"projectId"`
	Name           string         `gorm:"size:128;not null" json:"name"`
	Provider       string         `gorm:"size:32;not null;index" json:"provider"`
	AccountID      string         `gorm:"size:64;not null" json:"accountId"`
	CredentialType string         `gorm:"size:32;not null" json:"credentialType"`
	CredentialJSON string         `gorm:"type:text" json:"-"`
	Status         string         `gorm:"size:16;not null;index;default:'active'" json:"status"`
	LastSyncedAt   *time.Time     `json:"lastSyncedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}
```

- [ ] **Step 1.2: 在 service/cloud/cloud.go 新增 BackfillCloudProjectScope**

文件: `server/service/cloud/cloud.go`，在 `SeedCloudData` 函数后追加（同时新增 imports）：

```go
import (
	"errors"
	"fmt"

	cloudModel "optimus/server/model/cloud"
	projectModel "optimus/server/model/project"
	systemModel "optimus/server/model/system"
	projectUsecase "optimus/server/service/project/usecase"

	"gorm.io/gorm"
)

const defaultProjectCode = "default"

// BackfillCloudProjectScope ensures the default project exists and migrates legacy
// cloud_accounts (project_id = 0) into it. Idempotent — safe to call repeatedly.
func BackfillCloudProjectScope(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	defaultProj, err := ensureDefaultProject(db)
	if err != nil {
		return err
	}
	return db.Model(&cloudModel.CloudAccount{}).
		Where("project_id = 0").
		Update("project_id", defaultProj.ID).Error
}

func ensureDefaultProject(db *gorm.DB) (projectModel.Project, error) {
	var proj projectModel.Project
	err := db.Where("code = ?", defaultProjectCode).First(&proj).Error
	if err == nil {
		return proj, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return projectModel.Project{}, err
	}

	var admin systemModel.SystemUser
	if err := db.Where("username = ?", "admin").First(&admin).Error; err != nil {
		return projectModel.Project{}, fmt.Errorf("cannot find admin user to own default project: %w", err)
	}

	proj = projectModel.Project{
		Name:        "Default",
		Code:        defaultProjectCode,
		Description: "Default project for legacy cloud accounts",
		OwnerID:     admin.ID,
		Status:      "active",
	}
	if err := db.Create(&proj).Error; err != nil {
		return projectModel.Project{}, err
	}

	owner := projectModel.UserProject{
		UserID:      admin.ID,
		ProjectID:   proj.ID,
		ProjectRole: projectUsecase.RoleOwner,
	}
	if err := db.Create(&owner).Error; err != nil {
		return projectModel.Project{}, err
	}
	return proj, nil
}
```

注意：`service/cloud/cloud.go` 此前已 import 部分包，保留原有 import 后追加上述新增项；切勿把不存在的 import 直接覆盖掉。

- [ ] **Step 1.3: 验证编译**

Run: `cd /home/lilin/trae-project/Optimus/server && go build ./...`
Expected: clean.

- [ ] **Step 1.4: 提交**

```bash
git add server/model/cloud/account.go server/service/cloud/cloud.go
git commit -m "$(cat <<'EOF'
CloudAccount 增加 project_id 列并新增默认项目回填函数

Constraint: project_id 默认为 0；BackfillCloudProjectScope 把历史账号划入 default 项目
Confidence: high
Scope-risk: narrow
Tested: go build ./...
EOF
)"
```

---

### Task 2: AccountStore 接口 + CloudAccountService 扩展项目维度

**Files:**
- Modify: `server/service/cloud/usecase/account.go`
- Modify: `server/service/cloud/usecase/account_test.go`

- [ ] **Step 2.1: 改 AccountRow + 新增 ErrProjectScope 错误 + 改接口签名**

文件: `server/service/cloud/usecase/account.go`，更新如下区段。

`AccountRow`：在 `ID` 之后插入 `ProjectID`：

```go
type AccountRow struct {
	ID             uint
	ProjectID      uint
	Name           string
	Provider       string
	AccountID      string
	CredentialType string
	CredentialJSON string
	Status         string
	LastSyncedAt   *time.Time
}
```

`AccountView`：在 `ID` 之后插入 `ProjectID`（API 响应里也带出来）：

```go
type AccountView struct {
	ID             uint       `json:"id"`
	ProjectID      uint       `json:"projectId"`
	Name           string     `json:"name"`
	Provider       string     `json:"provider"`
	AccountID      string     `json:"accountId"`
	CredentialType string     `json:"credentialType"`
	Credential     string     `json:"credential"`
	Status         string     `json:"status"`
	LastSyncedAt   *time.Time `json:"lastSyncedAt"`
}
```

新增错误 sentinel（紧邻 `ErrAccountNotFound`）：

```go
var ErrAccountProjectMismatch = errors.New("cloud account project not accessible")
```

`AccountStore` 接口：把 read/write 方法加项目维度：

```go
type AccountStore interface {
	Create(ctx context.Context, projectID uint, name, provider, accountID, credentialType, encryptedCredential string) (AccountRow, error)
	Get(ctx context.Context, accessibleProjectIDs []uint, id uint) (AccountRow, error)
	List(ctx context.Context, accessibleProjectIDs []uint) ([]AccountRow, error)
	Update(ctx context.Context, accessibleProjectIDs []uint, id uint, name *string, encryptedCredential *string) (AccountRow, error)
	Delete(ctx context.Context, accessibleProjectIDs []uint, id uint) error
	UpdateSyncStatus(ctx context.Context, id uint, lastSyncedAt *time.Time, status string) error
}
```

`CloudAccountService` 方法签名扩展：

```go
func (s *CloudAccountService) Create(ctx context.Context, projectID uint, req CreateAccountRequest) (AccountView, error) {
	// existing body — pass projectID into store.Create:
	credJSON, err := json.Marshal(req.Credential)
	if err != nil {
		return AccountView{}, err
	}
	encrypted, err := encryptAESGCM(s.encKey, string(credJSON))
	if err != nil {
		return AccountView{}, err
	}
	row, err := s.store.Create(ctx, projectID, req.Name, req.Provider, req.AccountID, req.CredentialType, encrypted)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(row), nil
}

func (s *CloudAccountService) Get(ctx context.Context, accessibleProjectIDs []uint, id uint) (AccountView, error) {
	row, err := s.store.Get(ctx, accessibleProjectIDs, id)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(row), nil
}

func (s *CloudAccountService) List(ctx context.Context, accessibleProjectIDs []uint) ([]AccountView, error) {
	rows, err := s.store.List(ctx, accessibleProjectIDs)
	if err != nil {
		return nil, err
	}
	views := make([]AccountView, len(rows))
	for i, r := range rows {
		views[i] = toAccountView(r)
	}
	return views, nil
}

func (s *CloudAccountService) Update(ctx context.Context, accessibleProjectIDs []uint, id uint, req UpdateAccountRequest) (AccountView, error) {
	var encryptedPtr *string
	if len(req.Credential) > 0 {
		credJSON, err := json.Marshal(req.Credential)
		if err != nil {
			return AccountView{}, err
		}
		encrypted, err := encryptAESGCM(s.encKey, string(credJSON))
		if err != nil {
			return AccountView{}, err
		}
		encryptedPtr = &encrypted
	}
	row, err := s.store.Update(ctx, accessibleProjectIDs, id, req.Name, encryptedPtr)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(row), nil
}

func (s *CloudAccountService) Delete(ctx context.Context, accessibleProjectIDs []uint, id uint) error {
	return s.store.Delete(ctx, accessibleProjectIDs, id)
}

// UpdateSyncStatus and GetForSync remain unchanged — they are called internally
// by AssetRegistryService after access is already validated.
```

`toAccountView` 增加 ProjectID 映射：

```go
func toAccountView(r AccountRow) AccountView {
	return AccountView{
		ID:             r.ID,
		ProjectID:      r.ProjectID,
		Name:           r.Name,
		Provider:       r.Provider,
		AccountID:      r.AccountID,
		CredentialType: r.CredentialType,
		Credential:     "***",
		Status:         r.Status,
		LastSyncedAt:   r.LastSyncedAt,
	}
}
```

`SyncAccount` 数据结构 (`type SyncAccount struct { ... }`) 若已存在保持不变；`GetForSync` 签名不变（内部调用，service 在 AssetRegistryService.SyncAccount 入口做完访问校验后再调用）。

- [ ] **Step 2.2: 更新 account_test.go 的 fake store 与现有断言**

文件: `server/service/cloud/usecase/account_test.go`。

把 fake store 方法签名按新接口对齐。fake 实现要保存 `projectID` 字段并按 `accessibleProjectIDs` 过滤；新增两个测试：`TestCreate_PersistsProjectID` 和 `TestList_FiltersByAccessibleProjects`。

替换 fake 定义为：

```go
type fakeAccountStore struct {
	mu    sync.Mutex
	rows  map[uint]usecase.AccountRow
	next  uint
}

func newFakeAccountStore() *fakeAccountStore {
	return &fakeAccountStore{rows: map[uint]usecase.AccountRow{}, next: 100}
}

func (s *fakeAccountStore) Create(_ context.Context, projectID uint, name, provider, accountID, credType, enc string) (usecase.AccountRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row := usecase.AccountRow{
		ID:             s.next,
		ProjectID:      projectID,
		Name:           name,
		Provider:       provider,
		AccountID:      accountID,
		CredentialType: credType,
		CredentialJSON: enc,
		Status:         "active",
	}
	s.next++
	s.rows[row.ID] = row
	return row, nil
}

func (s *fakeAccountStore) Get(_ context.Context, accessible []uint, id uint) (usecase.AccountRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[id]
	if !ok || !containsUint(accessible, row.ProjectID) {
		return usecase.AccountRow{}, usecase.ErrAccountNotFound
	}
	return row, nil
}

func (s *fakeAccountStore) List(_ context.Context, accessible []uint) ([]usecase.AccountRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []usecase.AccountRow{}
	for _, r := range s.rows {
		if containsUint(accessible, r.ProjectID) {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *fakeAccountStore) Update(_ context.Context, accessible []uint, id uint, name *string, enc *string) (usecase.AccountRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[id]
	if !ok || !containsUint(accessible, row.ProjectID) {
		return usecase.AccountRow{}, usecase.ErrAccountNotFound
	}
	if name != nil {
		row.Name = *name
	}
	if enc != nil {
		row.CredentialJSON = *enc
	}
	s.rows[id] = row
	return row, nil
}

func (s *fakeAccountStore) Delete(_ context.Context, accessible []uint, id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[id]
	if !ok || !containsUint(accessible, row.ProjectID) {
		return usecase.ErrAccountNotFound
	}
	delete(s.rows, id)
	return nil
}

func (s *fakeAccountStore) UpdateSyncStatus(_ context.Context, id uint, ts *time.Time, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[id]
	if !ok {
		return usecase.ErrAccountNotFound
	}
	row.LastSyncedAt = ts
	row.Status = status
	s.rows[id] = row
	return nil
}

func containsUint(set []uint, v uint) bool {
	for _, x := range set {
		if x == v {
			return true
		}
	}
	return false
}
```

把所有现有调用 `svc.List(ctx)` / `svc.Get(ctx, id)` / `svc.Update(ctx, id, req)` / `svc.Delete(ctx, id)` / `svc.Create(ctx, req)` 改为带 `accessibleProjectIDs` 或 `projectID` 的版本。对全部测试做替换；测试期望项目 ID 为 7（自选），accessible = `[]uint{7}`。

新增两个测试在文件末尾：

```go
func TestCreate_PersistsProjectID(t *testing.T) {
	store := newFakeAccountStore()
	svc, err := usecase.NewCloudAccountService(store, fixed32Key())
	require.NoError(t, err)

	view, err := svc.Create(context.Background(), 7, sampleCreateRequest())
	require.NoError(t, err)
	require.Equal(t, uint(7), view.ProjectID)
}

func TestList_FiltersByAccessibleProjects(t *testing.T) {
	store := newFakeAccountStore()
	svc, err := usecase.NewCloudAccountService(store, fixed32Key())
	require.NoError(t, err)

	_, err = svc.Create(context.Background(), 7, sampleCreateRequest())
	require.NoError(t, err)
	other := sampleCreateRequest()
	other.Name = "other"
	_, err = svc.Create(context.Background(), 8, other)
	require.NoError(t, err)

	views, err := svc.List(context.Background(), []uint{7})
	require.NoError(t, err)
	require.Len(t, views, 1)
	require.Equal(t, uint(7), views[0].ProjectID)
}
```

若现有测试中存在 `sampleCreateRequest` / `fixed32Key` 等辅助则复用，否则参考既有定义补齐（保持与本任务前测试一致即可）。

- [ ] **Step 2.3: 跑测试 + 编译**

Run: `cd /home/lilin/trae-project/Optimus/server && go test ./service/cloud/usecase/... -v`
Expected: PASS — 所有原有测试 + 2 个新增测试。

Run: `cd /home/lilin/trae-project/Optimus/server && go build ./...`
Expected: clean（asset_test.go 也可能编译失败因为它可能调用 service，需要同步更新；如有，按相同模式处理）。

- [ ] **Step 2.4: 提交**

```bash
git add server/service/cloud/usecase/
git commit -m "$(cat <<'EOF'
扩展 CloudAccountService 与 AccountStore 接口接受项目维度

Constraint: Create 接 projectID；其余读写接 accessibleProjectIDs 用于过滤
Confidence: high
Scope-risk: medium（影响全部 cloud handler 调用方）
Tested: go test ./service/cloud/usecase/...
EOF
)"
```

---

### Task 3: GormAccountStore 实施项目过滤

**Files:**
- Modify: `server/service/cloud/infra/persistence/account_store_gorm.go`
- Modify: `server/service/cloud/infra/persistence/account_store_gorm_test.go`

- [ ] **Step 3.1: 改 GORM store 加项目维度**

文件: `server/service/cloud/infra/persistence/account_store_gorm.go`。

`Create` 增加 projectID 参数并在写入时设置：

```go
func (s *GormAccountStore) Create(ctx context.Context, projectID uint, name, provider, accountID, credentialType, encryptedCredential string) (usecase.AccountRow, error) {
	row := cloudModel.CloudAccount{
		ProjectID:      projectID,
		Name:           name,
		Provider:       provider,
		AccountID:      accountID,
		CredentialType: credentialType,
		CredentialJSON: encryptedCredential,
		Status:         "active",
	}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return usecase.AccountRow{}, err
	}
	return toAccountRow(row), nil
}
```

`Get` / `List` / `Update` / `Delete` 加 accessibleProjectIDs 参数并加 `Where("project_id IN ?", accessibleProjectIDs)`：

```go
func (s *GormAccountStore) Get(ctx context.Context, accessibleProjectIDs []uint, id uint) (usecase.AccountRow, error) {
	if len(accessibleProjectIDs) == 0 {
		return usecase.AccountRow{}, usecase.ErrAccountNotFound
	}
	var row cloudModel.CloudAccount
	err := s.db.WithContext(ctx).
		Where("id = ? AND project_id IN ?", id, accessibleProjectIDs).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return usecase.AccountRow{}, usecase.ErrAccountNotFound
	}
	if err != nil {
		return usecase.AccountRow{}, err
	}
	return toAccountRow(row), nil
}

func (s *GormAccountStore) List(ctx context.Context, accessibleProjectIDs []uint) ([]usecase.AccountRow, error) {
	if len(accessibleProjectIDs) == 0 {
		return []usecase.AccountRow{}, nil
	}
	var rows []cloudModel.CloudAccount
	err := s.db.WithContext(ctx).
		Where("project_id IN ?", accessibleProjectIDs).
		Order("created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]usecase.AccountRow, len(rows))
	for i, r := range rows {
		out[i] = toAccountRow(r)
	}
	return out, nil
}

func (s *GormAccountStore) Update(ctx context.Context, accessibleProjectIDs []uint, id uint, name *string, encryptedCredential *string) (usecase.AccountRow, error) {
	if len(accessibleProjectIDs) == 0 {
		return usecase.AccountRow{}, usecase.ErrAccountNotFound
	}
	var existing cloudModel.CloudAccount
	err := s.db.WithContext(ctx).
		Where("id = ? AND project_id IN ?", id, accessibleProjectIDs).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return usecase.AccountRow{}, usecase.ErrAccountNotFound
	}
	if err != nil {
		return usecase.AccountRow{}, err
	}
	fields := map[string]any{}
	if name != nil {
		fields["name"] = *name
	}
	if encryptedCredential != nil {
		fields["credential_json"] = *encryptedCredential
	}
	if len(fields) > 0 {
		if err := s.db.WithContext(ctx).Model(&existing).Updates(fields).Error; err != nil {
			return usecase.AccountRow{}, err
		}
	}
	var reloaded cloudModel.CloudAccount
	if err := s.db.WithContext(ctx).First(&reloaded, existing.ID).Error; err != nil {
		return usecase.AccountRow{}, err
	}
	return toAccountRow(reloaded), nil
}

func (s *GormAccountStore) Delete(ctx context.Context, accessibleProjectIDs []uint, id uint) error {
	if len(accessibleProjectIDs) == 0 {
		return usecase.ErrAccountNotFound
	}
	res := s.db.WithContext(ctx).
		Where("id = ? AND project_id IN ?", id, accessibleProjectIDs).
		Delete(&cloudModel.CloudAccount{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return usecase.ErrAccountNotFound
	}
	return nil
}
```

`UpdateSyncStatus` 不变（service 已先做访问校验）。

`toAccountRow` 增加 ProjectID：

```go
func toAccountRow(m cloudModel.CloudAccount) usecase.AccountRow {
	return usecase.AccountRow{
		ID:             m.ID,
		ProjectID:      m.ProjectID,
		Name:           m.Name,
		Provider:       m.Provider,
		AccountID:      m.AccountID,
		CredentialType: m.CredentialType,
		CredentialJSON: m.CredentialJSON,
		Status:         m.Status,
		LastSyncedAt:   m.LastSyncedAt,
	}
}
```

- [ ] **Step 3.2: 更新集成测试 + 新增项目过滤断言**

文件: `server/service/cloud/infra/persistence/account_store_gorm_test.go`。

把所有 `store.Create(ctx, ...)` 改为 `store.Create(ctx, projectID, ...)`（用任意合法的 `projectID`，如 1）；所有 `store.Get(ctx, id)` / `List(ctx)` / `Update(ctx, id, ...)` / `Delete(ctx, id)` 改为带 `accessible = []uint{projectID}` 的版本。

新增一个测试：

```go
func TestGet_RejectsAccountInOtherProject(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	row, err := store.Create(context.Background(), 1, "name", "aliyun", "acct", "ak_sk", "enc")
	require.NoError(t, err)

	_, err = store.Get(context.Background(), []uint{2}, row.ID)
	require.ErrorIs(t, err, usecase.ErrAccountNotFound)
}
```

- [ ] **Step 3.3: 测试 + 编译**

Run: `cd /home/lilin/trae-project/Optimus/server && go test ./service/cloud/infra/persistence/...`
Expected: SKIP without MySQL，PASS with MySQL。

Run: `cd /home/lilin/trae-project/Optimus/server && go build ./...`
Expected: clean（asset_store_gorm.go 还未改，可能 build 失败，留给 Task 5 修复；或先把 Task 3 的 build 测试限定到 account 包）。

> 注：本任务结束时 `go build ./...` 在 cloud 包内可能仍有未对齐的调用方（handlers）。这是 Task 4-7 要做的，所以本任务的接受标准缩窄为：account_store_gorm.go 自身可编译（用 `go build ./service/cloud/infra/persistence/...` 验证）+ 持久化测试 PASS。

- [ ] **Step 3.4: 提交**

```bash
git add server/service/cloud/infra/persistence/account_store_gorm.go server/service/cloud/infra/persistence/account_store_gorm_test.go
git commit -m "$(cat <<'EOF'
GormAccountStore 强制按 accessibleProjectIDs 过滤所有读写

Constraint: Get/List/Update/Delete 在 WHERE 中加 project_id IN (?)
Confidence: high
Scope-risk: narrow
Tested: go build ./service/cloud/infra/persistence/... && go test ./service/cloud/infra/persistence/...
EOF
)"
```

---

### Task 4: AssetStore + AssetRegistryService 加项目维度

**Files:**
- Modify: `server/service/cloud/usecase/asset.go`
- Modify: `server/service/cloud/usecase/asset_test.go`（若存在；否则跳过测试更新）

- [ ] **Step 4.1: AssetFilter 加 AccessibleProjectIDs**

文件: `server/service/cloud/usecase/asset.go`。

`AssetFilter` 增加新字段：

```go
type AssetFilter struct {
	AccessibleProjectIDs []uint
	AccountID            *uint
	Provider             *string
	Region               *string
	ResourceType         *string
	Status               *string
	Keyword              *string
	Page                 int
	PageSize             int
}
```

`AssetRegistryService.SyncAccount` 加 accessibleProjectIDs 参数并在调用 `GetForSync` 之前做访问校验：

```go
func (s *AssetRegistryService) SyncAccount(ctx context.Context, accessibleProjectIDs []uint, accountID uint) (SyncResult, error) {
	// 先确认账号属于可访问项目
	if _, err := s.accounts.Get(ctx, accessibleProjectIDs, accountID); err != nil {
		return SyncResult{}, err
	}

	syncAcc, err := s.accounts.GetForSync(ctx, accountID)
	if err != nil {
		statusErr := s.accounts.UpdateSyncStatus(ctx, accountID, nil, "error")
		return SyncResult{}, errors.Join(err, statusErr)
	}
	syncer, err := s.registry.Get(syncAcc.Provider)
	if err != nil {
		statusErr := s.accounts.UpdateSyncStatus(ctx, accountID, nil, "error")
		return SyncResult{}, errors.Join(err, statusErr)
	}
	// 后续 syncer.Sync + store.Upsert 部分保留不变
	// ...existing body continues...
}
```

注意：`SyncAccount` 原本 body 较长（200+ 行根据实际情况）。本步骤只在函数最前面插入访问校验 + 改签名，**保留所有原有逻辑**。

`QueryAssets` 签名不变（filter 字段已经带了 accessibleProjectIDs，handler 传入时填）：

```go
// QueryAssets passes the filter through to the store; the handler is
// responsible for setting filter.AccessibleProjectIDs from the request scope.
func (s *AssetRegistryService) QueryAssets(ctx context.Context, f AssetFilter) (AssetPage, error) {
	if len(f.AccessibleProjectIDs) == 0 {
		return AssetPage{Items: []AssetView{}}, nil
	}
	return s.store.Query(ctx, f)
}
```

- [ ] **Step 4.2: 编译检查（asset_test.go 中调用如有不匹配需同步修复）**

Run: `cd /home/lilin/trae-project/Optimus/server && go build ./service/cloud/usecase/...`
Expected: clean。

如果 `asset_test.go` 调用 `svc.SyncAccount(ctx, id)`，改成 `svc.SyncAccount(ctx, []uint{projectID}, id)` 并在测试 fake 中合理 mock。

- [ ] **Step 4.3: 提交**

```bash
git add server/service/cloud/usecase/asset.go server/service/cloud/usecase/asset_test.go
git commit -m "$(cat <<'EOF'
AssetRegistryService 在 SyncAccount/Query 中按 accessibleProjectIDs 隔离

Constraint: SyncAccount 先用 AccountStore.Get 校验访问；QueryAssets 透传 filter
Confidence: high
Scope-risk: narrow
Tested: go build ./service/cloud/usecase/...
EOF
)"
```

（若 asset_test.go 不存在则只 add asset.go；若 asset_test.go 存在但未修改则不放进 commit。）

---

### Task 5: GormAssetStore JOIN cloud_accounts 按 project_id 过滤

**Files:**
- Modify: `server/service/cloud/infra/persistence/asset_store_gorm.go`
- Modify: `server/service/cloud/infra/persistence/asset_store_gorm_test.go`

- [ ] **Step 5.1: Query 加 JOIN 过滤**

文件: `server/service/cloud/infra/persistence/asset_store_gorm.go`。

在 `Query` 函数顶部应用 accessibleProjectIDs 过滤：

```go
func (s *GormAssetStore) Query(ctx context.Context, f usecase.AssetFilter) (usecase.AssetPage, error) {
	if len(f.AccessibleProjectIDs) == 0 {
		return usecase.AssetPage{Items: []usecase.AssetView{}}, nil
	}

	tx := s.db.WithContext(ctx).
		Model(&cloudModel.CloudAsset{}).
		Joins("JOIN cloud_accounts ON cloud_accounts.id = cloud_assets.account_id").
		Where("cloud_accounts.project_id IN ?", f.AccessibleProjectIDs).
		Where("cloud_accounts.deleted_at IS NULL")

	if f.AccountID != nil {
		tx = tx.Where("cloud_assets.account_id = ?", *f.AccountID)
	}
	if f.Provider != nil {
		tx = tx.Where("cloud_assets.provider = ?", *f.Provider)
	}
	// 其余过滤条件保留原有逻辑，仅把列名加上 `cloud_assets.` 前缀避免歧义
	// ... (Region / ResourceType / Status / Keyword / 分页同样处理)
}
```

**重要**：把原 `Query` 实现里所有未限定列名补齐 `cloud_assets.` 前缀，避免 JOIN 后列名冲突。

`Upsert` 不变（写入时不需要 project 维度，因为 account_id 已隐含了项目归属）。

- [ ] **Step 5.2: 更新 asset_store_gorm_test.go**

所有 `store.Query(ctx, usecase.AssetFilter{...})` 调用补 `AccessibleProjectIDs: []uint{...}` 字段（用测试中的 account 对应的 project ID）。

新增一个测试：

```go
func TestQuery_FiltersByAccessibleProjects(t *testing.T) {
	db := openCloudTestDB(t)
	accountStore := NewGormAccountStore(db)
	assetStore := NewGormAssetStore(db)

	acc1, err := accountStore.Create(context.Background(), 1, "n1", "aliyun", "a1", "ak_sk", "enc")
	require.NoError(t, err)
	acc2, err := accountStore.Create(context.Background(), 2, "n2", "aliyun", "a2", "ak_sk", "enc")
	require.NoError(t, err)

	_, err = assetStore.Upsert(context.Background(), acc1.ID, "aliyun", []usecase.RawAsset{
		{ResourceType: "ecs", ResourceID: "i-1", Name: "in-1"},
	})
	require.NoError(t, err)
	_, err = assetStore.Upsert(context.Background(), acc2.ID, "aliyun", []usecase.RawAsset{
		{ResourceType: "ecs", ResourceID: "i-2", Name: "in-2"},
	})
	require.NoError(t, err)

	page, err := assetStore.Query(context.Background(), usecase.AssetFilter{
		AccessibleProjectIDs: []uint{1},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "in-1", page.Items[0].Name)
}
```

- [ ] **Step 5.3: 编译 + 测试**

Run: `cd /home/lilin/trae-project/Optimus/server && go build ./service/cloud/...`
Expected: clean.

Run: `cd /home/lilin/trae-project/Optimus/server && go test ./service/cloud/infra/persistence/...`
Expected: SKIP / PASS with MySQL。

- [ ] **Step 5.4: 提交**

```bash
git add server/service/cloud/infra/persistence/asset_store_gorm.go server/service/cloud/infra/persistence/asset_store_gorm_test.go
git commit -m "$(cat <<'EOF'
GormAssetStore Query JOIN cloud_accounts 按可访问项目过滤

Constraint: 列名加 cloud_assets. 前缀避免 JOIN 歧义；空 accessibleProjectIDs 直接返回空页
Confidence: high
Scope-risk: narrow
Tested: go build ./service/cloud/... && go test ./service/cloud/infra/persistence/...
EOF
)"
```

---

### Task 6: 项目作用域中间件 + utils helpers

**Files:**
- Create: `server/utils/project_scope.go`
- Create: `server/middleware/project_scope.go`
- Create: `server/middleware/project_scope_test.go`

- [ ] **Step 6.1: 写 utils ctx 常量与 helpers**

文件: `server/utils/project_scope.go`

```go
package utils

const ProjectScopeContextKey = "project_scope"

// ProjectScope describes the per-request project isolation context.
type ProjectScope struct {
	CurrentProjectID     uint
	CurrentProjectCode   string
	CurrentProjectRole   string
	AccessibleProjectIDs []uint
}
```

- [ ] **Step 6.2: 写中间件**

文件: `server/middleware/project_scope.go`

```go
package middleware

import (
	"context"
	"net/http"

	"optimus/server/utils"

	"github.com/gin-gonic/gin"
)

// ProjectScopeResolver is the contract the project usecase satisfies for the middleware.
type ProjectScopeResolver interface {
	ListAccessibleProjectIDs(ctx context.Context, userID uint) ([]uint, error)
	ResolveProjectByCode(ctx context.Context, code string) (projectID uint, exists bool, err error)
	GetMemberRole(ctx context.Context, userID, projectID uint) (role string, exists bool, err error)
}

// ProjectScope is the gin middleware that populates utils.ProjectScope on the context.
// It MUST run after JWTAuth (it reads the claims via utils.ClaimsContextKey).
func ProjectScope(resolver ProjectScopeResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, exists := c.Get(utils.ClaimsContextKey)
		if !exists {
			c.Next()
			return
		}
		claims, ok := raw.(*utils.CustomClaims)
		if !ok || claims == nil {
			c.Next()
			return
		}

		accessible, err := resolver.ListAccessibleProjectIDs(c.Request.Context(), claims.BaseClaims.ID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to resolve project scope"})
			return
		}

		scope := utils.ProjectScope{AccessibleProjectIDs: accessible}

		if code := c.GetHeader("X-Project-Code"); code != "" {
			projectID, exists, err := resolver.ResolveProjectByCode(c.Request.Context(), code)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to resolve project"})
				return
			}
			if !exists {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "project not found"})
				return
			}
			if !containsUint(accessible, projectID) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "no access to project"})
				return
			}
			role, exists, err := resolver.GetMemberRole(c.Request.Context(), claims.BaseClaims.ID, projectID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "failed to resolve project role"})
				return
			}
			if !exists {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "no access to project"})
				return
			}
			scope.CurrentProjectID = projectID
			scope.CurrentProjectCode = code
			scope.CurrentProjectRole = role
		}

		c.Set(utils.ProjectScopeContextKey, scope)
		c.Next()
	}
}

func containsUint(set []uint, v uint) bool {
	for _, x := range set {
		if x == v {
			return true
		}
	}
	return false
}
```

- [ ] **Step 6.3: 写中间件单元测试**

文件: `server/middleware/project_scope_test.go`

```go
package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"optimus/server/middleware"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakeResolver struct {
	accessible map[uint][]uint
	codes      map[string]uint // code -> projectID
	roles      map[uint]map[uint]string
}

func (f *fakeResolver) ListAccessibleProjectIDs(_ context.Context, userID uint) ([]uint, error) {
	return f.accessible[userID], nil
}

func (f *fakeResolver) ResolveProjectByCode(_ context.Context, code string) (uint, bool, error) {
	pid, ok := f.codes[code]
	return pid, ok, nil
}

func (f *fakeResolver) GetMemberRole(_ context.Context, userID, projectID uint) (string, bool, error) {
	role, ok := f.roles[userID][projectID]
	return role, ok, nil
}

type erroringResolver struct{}

func (erroringResolver) ListAccessibleProjectIDs(context.Context, uint) ([]uint, error) {
	return nil, errors.New("boom")
}
func (erroringResolver) ResolveProjectByCode(context.Context, string) (uint, bool, error) {
	return 0, false, nil
}
func (erroringResolver) GetMemberRole(context.Context, uint, uint) (string, bool, error) {
	return "", false, nil
}

func setupRouter(resolver middleware.ProjectScopeResolver, withClaims bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if withClaims {
			c.Set(utils.ClaimsContextKey, &utils.CustomClaims{BaseClaims: utils.BaseClaims{ID: 42}})
		}
		c.Next()
	})
	r.Use(middleware.ProjectScope(resolver))
	r.GET("/", func(c *gin.Context) {
		scope, _ := c.Get(utils.ProjectScopeContextKey)
		c.JSON(http.StatusOK, scope)
	})
	return r
}

func TestProjectScope_NoClaimsPassesThrough(t *testing.T) {
	r := setupRouter(&fakeResolver{}, false)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestProjectScope_NoHeaderOnlySetsAccessible(t *testing.T) {
	r := setupRouter(&fakeResolver{accessible: map[uint][]uint{42: []uint{7, 8}}}, true)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "\"AccessibleProjectIDs\":[7,8]")
}

func TestProjectScope_ValidHeaderSetsCurrent(t *testing.T) {
	resolver := &fakeResolver{
		accessible: map[uint][]uint{42: []uint{7}},
		codes:      map[string]uint{"prod": 7},
		roles:      map[uint]map[uint]string{42: map[uint]string{7: "owner"}},
	}
	r := setupRouter(resolver, true)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Project-Code", "prod")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "\"CurrentProjectID\":7")
	require.Contains(t, rec.Body.String(), "\"CurrentProjectRole\":\"owner\"")
}

func TestProjectScope_UnknownCodeReturns404(t *testing.T) {
	r := setupRouter(&fakeResolver{accessible: map[uint][]uint{42: []uint{7}}}, true)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Project-Code", "missing")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestProjectScope_InaccessibleCodeReturns403(t *testing.T) {
	resolver := &fakeResolver{
		accessible: map[uint][]uint{42: []uint{8}},
		codes:      map[string]uint{"prod": 7},
	}
	r := setupRouter(resolver, true)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Project-Code", "prod")
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestProjectScope_ResolverErrorReturns500(t *testing.T) {
	r := setupRouter(erroringResolver{}, true)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
```

- [ ] **Step 6.4: 测试 + 编译**

Run: `cd /home/lilin/trae-project/Optimus/server && go test ./middleware/... -v -run ProjectScope`
Expected: PASS (6 个测试)。

Run: `cd /home/lilin/trae-project/Optimus/server && go build ./...`
Expected: 可能仍失败（handlers 还未对齐），先确认中间件自身可编译。

- [ ] **Step 6.5: 提交**

```bash
git add server/utils/project_scope.go server/middleware/project_scope.go server/middleware/project_scope_test.go
git commit -m "$(cat <<'EOF'
新增 ProjectScope 中间件：解析 X-Project-Code 并注入项目访问上下文

Constraint: 中间件依赖抽象 ProjectScopeResolver，避免与 project usecase 形成包级依赖
Confidence: high
Scope-risk: narrow
Tested: go test ./middleware/...
EOF
)"
```

---

### Task 7: Cloud Handlers + Module 装配 + Startup Backfill

**Files:**
- Modify: `server/api/v1/cloud/handlers.go`
- Modify: `server/api/v1/cloud/account.go`
- Modify: `server/api/v1/cloud/asset.go`
- Modify: `server/initialize/cloud/module.go`
- Modify: `server/initialize/router.go`
- Modify: `server/initialize/startup.go`
- Create: `server/service/project/usecase/scope_adapter.go`（小适配器，使 ProjectService 满足 ProjectScopeResolver 接口）

- [ ] **Step 7.1: 在 project usecase 中加 Resolver 适配器**

文件: `server/service/project/usecase/scope_adapter.go`

```go
package usecase

import "context"

// ResolveProjectByCode looks up a project by code and returns (id, exists).
// Adapter for middleware.ProjectScopeResolver — kept separate from FindByCode
// to expose a boolean rather than ErrProjectNotFound.
func (s *ProjectService) ResolveProjectByCode(ctx context.Context, code string) (uint, bool, error) {
	p, err := s.store.FindByCode(ctx, code)
	if err != nil {
		if isNotFound(err) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return p.ID, true, nil
}

// GetMemberRole returns the user's role in the given project, if any.
func (s *ProjectService) GetMemberRole(ctx context.Context, userID, projectID uint) (string, bool, error) {
	m, err := s.store.FindMember(ctx, userID, projectID)
	if err != nil {
		if isNotFound(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return m.ProjectRole, true, nil
}

func isNotFound(err error) bool {
	return err == ErrProjectNotFound || err == ErrMemberNotFound
}
```

- [ ] **Step 7.2: 在 handlers.go 加 projectScope 辅助函数**

文件: `server/api/v1/cloud/handlers.go`。在现有 `parseCloudUintParam` 等辅助旁追加：

```go
import (
	// existing imports...
	"optimus/server/utils"
)

func projectScope(c *gin.Context) utils.ProjectScope {
	v, exists := c.Get(utils.ProjectScopeContextKey)
	if !exists {
		return utils.ProjectScope{}
	}
	if scope, ok := v.(utils.ProjectScope); ok {
		return scope
	}
	return utils.ProjectScope{}
}
```

- [ ] **Step 7.3: 改 account.go handlers**

文件: `server/api/v1/cloud/account.go`：

- `ListAccounts` → `h.accountSvc.List(c.Request.Context(), scope.AccessibleProjectIDs)`
- `GetAccount` → `h.accountSvc.Get(c.Request.Context(), scope.AccessibleProjectIDs, id)`
- `CreateAccount`：要求 `scope.CurrentProjectID > 0`，否则 400；调用 `h.accountSvc.Create(c.Request.Context(), scope.CurrentProjectID, req)`
- `UpdateAccount` → `h.accountSvc.Update(c.Request.Context(), scope.AccessibleProjectIDs, id, req)`
- `DeleteAccount` → `h.accountSvc.Delete(c.Request.Context(), scope.AccessibleProjectIDs, id)`
- `SyncAccount` → `h.assetSvc.SyncAccount(c.Request.Context(), scope.AccessibleProjectIDs, id)`

Create 部分模板：

```go
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	scope := projectScope(c)
	if scope.CurrentProjectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "X-Project-Code header is required to create a cloud account"})
		return
	}
	var req cloudUsecase.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid create account payload"})
		return
	}
	view, err := h.accountSvc.Create(c.Request.Context(), scope.CurrentProjectID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create cloud account"})
		return
	}
	c.JSON(http.StatusCreated, view)
}
```

其他 5 个方法按相同模式替换调用，保持当前文件里的错误映射 / status code 不变。

- [ ] **Step 7.4: 改 asset.go handlers**

文件: `server/api/v1/cloud/asset.go`。`ListAssets` 在构建 `AssetFilter` 时塞入 `AccessibleProjectIDs: scope.AccessibleProjectIDs`：

```go
filter := cloudUsecase.AssetFilter{
	AccessibleProjectIDs: scope.AccessibleProjectIDs,
	// existing fields populated from query params
}
page, err := h.service.QueryAssets(c.Request.Context(), filter)
```

- [ ] **Step 7.5: 确认 NewCloudModule 无需修改**

`server/initialize/cloud/module.go` 不动 —— `NewCloudAccountService` / `NewAssetRegistryService` 构造函数参数没变。

- [ ] **Step 7.6: 改 initialize/router.go 注册中间件**

文件: `server/initialize/router.go`。

`server/initialize/router.go` 当前已经创建了 `projectModule := newProjectModule(...)`。`ProjectHandlers` 不公开 service；为了让 router 注册 `ProjectScope`，让 `ProjectModule` 直接暴露 `Service *projectUsecase.ProjectService` 字段。

修改 `server/initialize/project/module.go`：

```go
type ProjectModule struct {
	Handlers *projectApi.ProjectHandlers
	Service  *projectUsecase.ProjectService
}

func NewProjectModule(db *gorm.DB, _ config.Config) *ProjectModule {
	if db == nil {
		return &ProjectModule{Handlers: projectApi.NewProjectHandlers(projectApi.ProjectHandlerDeps{})}
	}
	store := persistence.NewGormProjectStore(db)
	svc := projectUsecase.NewProjectService(store)
	return &ProjectModule{
		Handlers: projectApi.NewProjectHandlers(projectApi.ProjectHandlerDeps{Service: svc}),
		Service:  svc,
	}
}
```

再修改 `server/initialize/router.go`，在当前这一行之后：

```go
privateGroup.Use(middleware.JWTAuth(), middleware.CasbinHandler(systemModule.Capabilities.RBAC))
```

追加：

```go
if projectModule.Service != nil {
	privateGroup.Use(middleware.ProjectScope(projectModule.Service))
}
```

`*projectUsecase.ProjectService` 已经实现 `ListAccessibleProjectIDs` (SP1) + `ResolveProjectByCode` + `GetMemberRole`（本 Task Step 7.1）三个方法，因此自动满足 `middleware.ProjectScopeResolver` 接口。

- [ ] **Step 7.7: startup.go 调用 BackfillCloudProjectScope**

文件: `server/initialize/startup.go`。在 `projectService.SeedProjectData(db)` 之后追加：

```go
if err := cloudService.BackfillCloudProjectScope(db); err != nil {
	logDependencyInitFailure("cloud", "backfill project scope", err)
}
```

注意：放在 `SeedProjectData` 后是关键，因为 backfill 需要 `projects` / `user_projects` 表已建好 + admin 用户已存在。

- [ ] **Step 7.8: 全量编译 + 测试**

Run: `cd /home/lilin/trae-project/Optimus/server && go build ./...`
Expected: clean.

Run: `cd /home/lilin/trae-project/Optimus/server && go test ./...`
Expected: 全部 PASS（持久化测试 SKIP 若无 MySQL）。

- [ ] **Step 7.9: 提交**

```bash
git add server/api/v1/cloud/ server/initialize/cloud/module.go server/initialize/router.go server/initialize/startup.go server/initialize/project/module.go server/service/project/usecase/scope_adapter.go
git commit -m "$(cat <<'EOF'
装配 ProjectScope 中间件、cloud handlers 接入项目作用域、startup 回填默认项目

Constraint: ProjectService 实现 ProjectScopeResolver 接口；Create cloud account 必须指定 X-Project-Code
Confidence: high
Scope-risk: medium（中间件接入全部 cloud 路由 + handlers 调用方全改）
Tested: go build ./... && go test ./...
EOF
)"
```

---

## Self-Review Checklist

**Spec coverage:**
- ✓ `cloud_accounts` 加 `project_id`（Task 1）
- ✓ 数据迁移：default 项目 + admin owner + UPDATE legacy 账号（Task 1 + Task 7 startup hook）
- ✓ ProjectScope middleware：解析 `X-Project-Code`、注入 `accessibleProjectIDs` + `currentProjectID`（Task 6）
- ✓ AccountStore / AssetStore 签名扩展项目维度（Task 2 + Task 4）
- ✓ Handler 从 ctx 取 currentProjectID（Task 7）
- ✓ 既有 cloud API 路径不变（仅请求需带 `X-Project-Code` header）

**Placeholder scan:** 无 TBD/TODO。所有代码段完整、可拷贝。Task 4 的 `SyncAccount` body 用 "(...existing body continues...)" 是因为函数原 body 很长，实施者保留原有逻辑只在头部插入访问校验 + 改签名；这是 explicit instruction，不是占位。

**Type consistency:**
- `AccountRow` / `AccountView` / `AccountStore` 接口 / `CloudAccountService` 方法在 Task 2 统一带 `ProjectID` / `accessibleProjectIDs`。
- `AssetFilter.AccessibleProjectIDs` 在 Task 4 定义，Task 5 与 Task 7 引用一致。
- `middleware.ProjectScopeResolver` 在 Task 6 定义，Task 7 Step 7.1 让 `*projectUsecase.ProjectService` 自动满足（通过 ResolveProjectByCode / GetMemberRole 加成）。
- `utils.ProjectScope` 与 `utils.ProjectScopeContextKey` 在 Task 6 Step 6.1 定义，Task 6 中间件 + Task 7 handlers 共同消费。

**Scope check:** SP2 是单一子系统（后端 cloud 隔离化），无需进一步拆分。

---

## Out of Scope (deferred to SP3/SP4)

- 前端项目切换器 + 项目管理页（SP3）
- 前端 cloud 页面在切换项目时的过滤（SP4）
- 跨项目迁移账号（spec Open Questions）
- 项目软删除后的可恢复 UI（spec Open Questions）

---

## Run-Through Walkthrough（验证最终联调正确性）

实施完所有 7 个 task 后，一个完整的端到端调用链应是：

1. Client 携带 JWT + `X-Project-Code: default` 调 GET `/api/v1/cloud/accounts`
2. `middleware.JWTAuth` 解析 JWT，写入 `utils.ClaimsContextKey`
3. `middleware.ProjectScope` 读 claims，调 `projectService.ListAccessibleProjectIDs(userID)` 拿到 `[]uint{1}`；读 header，调 `projectService.ResolveProjectByCode("default")` 拿到 `(1, true)`；校验 1 ∈ {1}；调 `GetMemberRole(userID, 1)` 拿到 `("owner", true)`；写入 `utils.ProjectScopeContextKey`
4. `CloudHandlers.List` 调 `projectScope(c)` 取出 scope，调 `h.accountSvc.List(ctx, []uint{1})`
5. `CloudAccountService.List` 调 `s.store.List(ctx, []uint{1})`
6. `GormAccountStore.List` 执行 `SELECT * FROM cloud_accounts WHERE project_id IN (1) AND deleted_at IS NULL ORDER BY created_at DESC`
7. 返回所有属于 default 项目的账号

POST 创建账号同理：handler 要求 `scope.CurrentProjectID > 0`，没 header 就是 400；有 header 就用 `scope.CurrentProjectID` 作为新账号的 project_id 落库。

SyncAccount：先 `accountSvc.Get(ctx, accessibleProjectIDs, id)` 校验访问；通过则进入原有 sync 流程。

QueryAssets：filter 里塞入 accessibleProjectIDs，store JOIN cloud_accounts 过滤。
