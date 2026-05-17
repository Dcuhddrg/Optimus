# SP4 Cloud 前端 + 用户 Tag Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 落地 cloud 前端页面（账号 + 资产），并引入用户 tag 维度（含 environment 软约定），完成多项目隔离的最后一公里。

**Architecture:** 后端给 `cloud_accounts` 加 `tags_json` 列，新增 `PUT /accounts/:id/tags` 与 `tag=key=value` 列表过滤；handler 解析 query 参数转 `TagFilter`，store 用 `JSON_EXTRACT(tags_json, '$.<key>')` 过滤。前端新增 `CloudAccountsPage` 与 `CloudAssetsPage`，复用通用 `TagChip` / `TagEditor` / `TagFilter` 组件，`environment` 在前端享有预设配色与快捷建议但后端不特殊处理。

**Tech Stack:** Go 1.23、GORM v2、Gin（后端）；React 18、TypeScript、Vite、CSS Modules（前端）；MySQL JSON_EXTRACT。

**Spec reference:** `docs/superpowers/specs/2026-05-17-cloud-account-tags-design.md`。SP1+SP2+SP3 已落地（项目隔离 + 项目管理 + ProjectSwitcher）；SideNav 里 cloud 入口当前是"Coming soon"占位，本 plan 把它激活。

---

## File Structure

**后端新增：** 无新文件
**后端修改：**
- `server/model/cloud/account.go` — `CloudAccount.TagsJSON` 列
- `server/service/cloud/usecase/account.go` — `Tags` 字段加入 `AccountRow/View/CreateAccountRequest`；新增 `UpdateAccountTagsRequest`、`TagFilter`、`tag validation`、`UpdateTags` service 方法、`AccountStore.UpdateTags` 与 `Create/List` 签名扩展
- `server/service/cloud/usecase/account_test.go` — fake store + 既有测试更新 + 新增 tag 测试
- `server/service/cloud/usecase/asset.go` — `AssetFilter.Tags []TagFilter`
- `server/service/cloud/usecase/asset_test.go` — 若存在则按签名更新
- `server/service/cloud/infra/persistence/account_store_gorm.go` — tag 列读写 + JSON_EXTRACT 过滤
- `server/service/cloud/infra/persistence/account_store_gorm_test.go` — 集成测试更新 + 新增
- `server/service/cloud/infra/persistence/asset_store_gorm.go` — Query 加 tag 过滤 (JSON_EXTRACT)
- `server/service/cloud/infra/persistence/asset_store_gorm_test.go` — 集成测试新增
- `server/api/v1/cloud/account.go` — list query parse `tag=` + PUT tags handler
- `server/api/v1/cloud/asset.go` — list query parse `tag=`
- `server/router/cloud/enter.go` — `PUT /accounts/:id/tags` 路由
- `server/service/cloud/cloud.go` — `DefaultCloudPermissions` 加 `cloud.accounts.tags.update`

**前端新增：**
- `web/src/types/cloud.ts` — types
- `web/src/api/cloud.ts` — API client
- `web/src/view/cloud/CloudAccountsPage.tsx` + `.module.css`
- `web/src/view/cloud/CloudAssetsPage.tsx` + `.module.css`
- `web/src/components/tag/TagChip.tsx` + `.module.css`
- `web/src/components/tag/TagEditor.tsx` + `.module.css`
- `web/src/components/tag/TagFilter.tsx` + `.module.css`

**前端修改：**
- `web/src/router/index.tsx` — 注册 `/cloud/accounts` 与 `/cloud/assets`
- `web/src/components/nav/SideNav.tsx` — cloud 入口从 "Coming soon" 改成真链接
- `web/src/locale/messages.ts` — cloud / tag / environment 软约定 keys

---

## Conventions

- **Tag key 正则**：`^[a-z][a-z0-9_-]{0,62}$`（后端、前端共用）
- **Tag value**：1–128 字符 trim 后
- **每账号 tag 上限**：32
- **环境预设值（前端显示）**：`prod` / `dev` / `qa` / `orphan`（仅 UI 配色 + 建议；后端不知道）
- **JSON_EXTRACT 写法**：`JSON_EXTRACT(tags_json, '$.<key>')`；MySQL 5.7+ 支持

---

### Task 1: 后端 Tag 数据模型 + 类型契约

**Files:**
- Modify: `server/model/cloud/account.go`
- Modify: `server/service/cloud/usecase/account.go`

- [ ] **Step 1.1: 给 CloudAccount 加 TagsJSON 列**

`server/model/cloud/account.go`：在 `CredentialJSON` 后插入 `TagsJSON`：

```go
type CloudAccount struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	ProjectID      uint           `gorm:"not null;default:0;index" json:"projectId"`
	Name           string         `gorm:"size:128;not null" json:"name"`
	Provider       string         `gorm:"size:32;not null;index" json:"provider"`
	AccountID      string         `gorm:"size:64;not null" json:"accountId"`
	CredentialType string         `gorm:"size:32;not null" json:"credentialType"`
	CredentialJSON string         `gorm:"type:text" json:"-"`
	TagsJSON       string         `gorm:"type:text;not null;default:'{}'" json:"-"`
	Status         string         `gorm:"size:16;not null;index;default:'active'" json:"status"`
	LastSyncedAt   *time.Time     `json:"lastSyncedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}
```

- [ ] **Step 1.2: 在 usecase/account.go 加 Tag 相关类型与校验**

在文件顶部 import 区追加 `"regexp"`、`"strings"`，并新增类型 + 常量 + 校验函数（位置：错误 sentinels 与 `AccountRow` 之间）：

```go
var (
	ErrInvalidTagKey       = errors.New("tag key must match ^[a-z][a-z0-9_-]{0,62}$")
	ErrInvalidTagValue     = errors.New("tag value must be 1-128 chars")
	ErrTooManyTags         = errors.New("a cloud account cannot have more than 32 tags")
	ErrInvalidTagFilter    = errors.New("invalid tag filter; expected key=value, key=*, or __none__:key")
)

const MaxAccountTags = 32

var tagKeyRegex = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,62}$`)

// ValidateTags normalizes and validates a tag map. Returns a cleaned copy.
func ValidateTags(tags map[string]string) (map[string]string, error) {
	if tags == nil {
		return map[string]string{}, nil
	}
	if len(tags) > MaxAccountTags {
		return nil, ErrTooManyTags
	}
	out := make(map[string]string, len(tags))
	for k, v := range tags {
		if !tagKeyRegex.MatchString(k) {
			return nil, ErrInvalidTagKey
		}
		trimmed := strings.TrimSpace(v)
		if trimmed == "" || len(trimmed) > 128 {
			return nil, ErrInvalidTagValue
		}
		out[k] = trimmed
	}
	return out, nil
}

// TagFilter is a single tag filter clause used in list queries.
type TagFilter struct {
	Key    string
	Value  string // "" + Exists=true means "any value, key must exist"
	Exists bool   // true = key must exist; false = key must NOT exist
}

// ParseTagFilter parses one tag filter token. Accepted forms:
//   key=value          → Exists=true, Value=value
//   key=*              → Exists=true, Value=""
//   __none__:key       → Exists=false
func ParseTagFilter(token string) (TagFilter, error) {
	if strings.HasPrefix(token, "__none__:") {
		key := strings.TrimPrefix(token, "__none__:")
		if !tagKeyRegex.MatchString(key) {
			return TagFilter{}, ErrInvalidTagKey
		}
		return TagFilter{Key: key, Exists: false}, nil
	}
	idx := strings.Index(token, "=")
	if idx <= 0 || idx == len(token)-1 {
		return TagFilter{}, ErrInvalidTagFilter
	}
	key := token[:idx]
	value := token[idx+1:]
	if !tagKeyRegex.MatchString(key) {
		return TagFilter{}, ErrInvalidTagKey
	}
	if value == "*" {
		return TagFilter{Key: key, Value: "", Exists: true}, nil
	}
	if len(value) > 128 {
		return TagFilter{}, ErrInvalidTagValue
	}
	return TagFilter{Key: key, Value: value, Exists: true}, nil
}
```

- [ ] **Step 1.3: 给 AccountRow / AccountView / CreateAccountRequest 加 Tags**

继续在 `account.go` 修改：

```go
type AccountRow struct {
	ID             uint
	ProjectID      uint
	Name           string
	Provider       string
	AccountID      string
	CredentialType string
	CredentialJSON string
	Tags           map[string]string
	Status         string
	LastSyncedAt   *time.Time
}

type AccountView struct {
	ID             uint              `json:"id"`
	ProjectID      uint              `json:"projectId"`
	Name           string            `json:"name"`
	Provider       string            `json:"provider"`
	AccountID      string            `json:"accountId"`
	CredentialType string            `json:"credentialType"`
	Credential     string            `json:"credential"`
	Tags           map[string]string `json:"tags"`
	Status         string            `json:"status"`
	LastSyncedAt   *time.Time        `json:"lastSyncedAt"`
}

type CreateAccountRequest struct {
	Name           string            `json:"name" binding:"required"`
	Provider       string            `json:"provider" binding:"required"`
	AccountID      string            `json:"accountId" binding:"required"`
	CredentialType string            `json:"credentialType" binding:"required"`
	Credential     map[string]string `json:"credential" binding:"required"`
	Tags           map[string]string `json:"tags"`
}

type UpdateAccountTagsRequest struct {
	Tags map[string]string `json:"tags" binding:"required"`
}
```

在 `toAccountView` 函数里把 Tags 映射出来（**保证非 nil**）：

```go
func toAccountView(r AccountRow) AccountView {
	tags := r.Tags
	if tags == nil {
		tags = map[string]string{}
	}
	return AccountView{
		ID:             r.ID,
		ProjectID:      r.ProjectID,
		Name:           r.Name,
		Provider:       r.Provider,
		AccountID:      r.AccountID,
		CredentialType: r.CredentialType,
		Credential:     "***",
		Tags:           tags,
		Status:         r.Status,
		LastSyncedAt:   r.LastSyncedAt,
	}
}
```

- [ ] **Step 1.4: 验证编译**

Run: `cd /home/lilin/trae-project/Optimus/server && go build ./...`
Expected: 因为 Task 2 还没改 store/service，account.go 引用了 AccountStore 旧签名仍可编译；但 fake store 和 GORM store 因为 AccountRow 多了 Tags 字段是值拷贝，零值是 nil，仍能编译。如果出现因签名不匹配的失败，**先按 Task 2 改完再 build**。

- [ ] **Step 1.5: 提交**

```bash
git add server/model/cloud/account.go server/service/cloud/usecase/account.go
git commit -m "$(cat <<'EOF'
新增 CloudAccount tag 数据模型与类型契约

Constraint: tag key/value 都有正则与长度限制；每账号最多 32 tag
Confidence: high
Scope-risk: narrow
Tested: go build ./... (依赖 Task 2 完成全签名)
EOF
)"
```

---

### Task 2: AccountStore 接口 + Service tag 方法 + 单元测试

**Files:**
- Modify: `server/service/cloud/usecase/account.go`
- Modify: `server/service/cloud/usecase/account_test.go`

- [ ] **Step 2.1: 扩展 AccountStore 接口签名**

替换 `AccountStore` 接口为：

```go
type AccountStore interface {
	Create(ctx context.Context, projectID uint, row AccountRow) (AccountRow, error)
	Get(ctx context.Context, accessibleProjectIDs []uint, id uint) (AccountRow, error)
	GetForSync(ctx context.Context, id uint) (AccountRow, error)
	List(ctx context.Context, accessibleProjectIDs []uint, tagFilters []TagFilter) ([]AccountRow, error)
	Update(ctx context.Context, accessibleProjectIDs []uint, id uint, name *string, encryptedCredential *string) (AccountRow, error)
	UpdateTags(ctx context.Context, accessibleProjectIDs []uint, id uint, tags map[string]string) (AccountRow, error)
	Delete(ctx context.Context, accessibleProjectIDs []uint, id uint) error
	UpdateSyncStatus(ctx context.Context, id uint, lastSyncedAt *time.Time, status string) error
}
```

> Note: `Create` 改为接收一个 `AccountRow`（已含 Tags），把多参版本收敛掉；调用方在 service 层把 request 转 row。

- [ ] **Step 2.2: 改造 CloudAccountService.Create 与新增 UpdateTags**

替换 `Create` 与新增 `UpdateTags`：

```go
func (s *CloudAccountService) Create(ctx context.Context, projectID uint, req CreateAccountRequest) (AccountView, error) {
	credJSON, err := json.Marshal(req.Credential)
	if err != nil {
		return AccountView{}, err
	}
	encrypted, err := encryptAESGCM(s.encKey, string(credJSON))
	if err != nil {
		return AccountView{}, err
	}
	cleanedTags, err := ValidateTags(req.Tags)
	if err != nil {
		return AccountView{}, err
	}
	row := AccountRow{
		ProjectID:      projectID,
		Name:           req.Name,
		Provider:       req.Provider,
		AccountID:      req.AccountID,
		CredentialType: req.CredentialType,
		CredentialJSON: encrypted,
		Tags:           cleanedTags,
	}
	saved, err := s.store.Create(ctx, projectID, row)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(saved), nil
}

func (s *CloudAccountService) UpdateTags(ctx context.Context, accessibleProjectIDs []uint, id uint, req UpdateAccountTagsRequest) (AccountView, error) {
	cleaned, err := ValidateTags(req.Tags)
	if err != nil {
		return AccountView{}, err
	}
	row, err := s.store.UpdateTags(ctx, accessibleProjectIDs, id, cleaned)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(row), nil
}
```

- [ ] **Step 2.3: 改 List 接受 tagFilters**

```go
func (s *CloudAccountService) List(ctx context.Context, accessibleProjectIDs []uint, tagFilters []TagFilter) ([]AccountView, error) {
	rows, err := s.store.List(ctx, accessibleProjectIDs, tagFilters)
	if err != nil {
		return nil, err
	}
	views := make([]AccountView, len(rows))
	for i, r := range rows {
		views[i] = toAccountView(r)
	}
	return views, nil
}
```

- [ ] **Step 2.4: 更新 fake store + 新增测试**

在 `account_test.go`，替换 fake store 的方法签名与新增字段：

```go
type fakeAccountStore struct {
	mu   sync.Mutex
	rows map[uint]usecase.AccountRow
	next uint
}

func newFakeAccountStore() *fakeAccountStore {
	return &fakeAccountStore{rows: map[uint]usecase.AccountRow{}, next: 100}
}

func (s *fakeAccountStore) Create(_ context.Context, projectID uint, row usecase.AccountRow) (usecase.AccountRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row.ID = s.next
	row.ProjectID = projectID
	row.Status = "active"
	if row.Tags == nil {
		row.Tags = map[string]string{}
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

func (s *fakeAccountStore) GetForSync(_ context.Context, id uint) (usecase.AccountRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[id]
	if !ok {
		return usecase.AccountRow{}, usecase.ErrAccountNotFound
	}
	return row, nil
}

func (s *fakeAccountStore) List(_ context.Context, accessible []uint, filters []usecase.TagFilter) ([]usecase.AccountRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []usecase.AccountRow{}
	for _, r := range s.rows {
		if !containsUint(accessible, r.ProjectID) {
			continue
		}
		if !matchAllTagFilters(r.Tags, filters) {
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

func matchAllTagFilters(tags map[string]string, filters []usecase.TagFilter) bool {
	for _, f := range filters {
		v, exists := tags[f.Key]
		if f.Exists {
			if !exists {
				return false
			}
			if f.Value != "" && v != f.Value {
				return false
			}
		} else {
			if exists {
				return false
			}
		}
	}
	return true
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

func (s *fakeAccountStore) UpdateTags(_ context.Context, accessible []uint, id uint, tags map[string]string) (usecase.AccountRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row, ok := s.rows[id]
	if !ok || !containsUint(accessible, row.ProjectID) {
		return usecase.AccountRow{}, usecase.ErrAccountNotFound
	}
	row.Tags = tags
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

把现有测试里调用 `svc.List(ctx, accessible)` 改为 `svc.List(ctx, accessible, nil)`；调用 `svc.Create(ctx, projectID, req)` 已经匹配新签名，不动；其他 svc 调用类似按新签名更新。

新增三个 tag 专项测试（追加在文件末尾）：

```go
func TestCreate_AcceptsTags(t *testing.T) {
	store := newFakeAccountStore()
	svc, err := usecase.NewCloudAccountService(store, fixed32Key())
	require.NoError(t, err)

	req := sampleCreateRequest()
	req.Tags = map[string]string{"environment": "prod", "team": "sre"}
	view, err := svc.Create(context.Background(), 7, req)
	require.NoError(t, err)
	require.Equal(t, "prod", view.Tags["environment"])
	require.Equal(t, "sre", view.Tags["team"])
}

func TestCreate_RejectsInvalidTagKey(t *testing.T) {
	store := newFakeAccountStore()
	svc, err := usecase.NewCloudAccountService(store, fixed32Key())
	require.NoError(t, err)

	req := sampleCreateRequest()
	req.Tags = map[string]string{"Environment": "prod"}
	_, err = svc.Create(context.Background(), 7, req)
	require.ErrorIs(t, err, usecase.ErrInvalidTagKey)
}

func TestUpdateTags_FiltersByAccessibleProjects(t *testing.T) {
	store := newFakeAccountStore()
	svc, err := usecase.NewCloudAccountService(store, fixed32Key())
	require.NoError(t, err)

	view, err := svc.Create(context.Background(), 7, sampleCreateRequest())
	require.NoError(t, err)

	// Wrong project — should be ErrAccountNotFound
	_, err = svc.UpdateTags(context.Background(), []uint{8}, view.ID,
		usecase.UpdateAccountTagsRequest{Tags: map[string]string{"environment": "dev"}})
	require.ErrorIs(t, err, usecase.ErrAccountNotFound)

	// Correct project
	updated, err := svc.UpdateTags(context.Background(), []uint{7}, view.ID,
		usecase.UpdateAccountTagsRequest{Tags: map[string]string{"environment": "dev"}})
	require.NoError(t, err)
	require.Equal(t, "dev", updated.Tags["environment"])
}

func TestList_FiltersByTag(t *testing.T) {
	store := newFakeAccountStore()
	svc, err := usecase.NewCloudAccountService(store, fixed32Key())
	require.NoError(t, err)

	req1 := sampleCreateRequest()
	req1.Tags = map[string]string{"environment": "prod"}
	_, err = svc.Create(context.Background(), 7, req1)
	require.NoError(t, err)

	req2 := sampleCreateRequest()
	req2.Name = "other"
	req2.Tags = map[string]string{"environment": "dev"}
	_, err = svc.Create(context.Background(), 7, req2)
	require.NoError(t, err)

	views, err := svc.List(context.Background(), []uint{7},
		[]usecase.TagFilter{{Key: "environment", Value: "prod", Exists: true}})
	require.NoError(t, err)
	require.Len(t, views, 1)
	require.Equal(t, "prod", views[0].Tags["environment"])
}
```

- [ ] **Step 2.5: 验证测试 + 编译**

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./service/cloud/usecase/... -v
go build ./service/cloud/usecase/...
```

Expected: 所有测试 PASS（既有 + 4 个新增）。`go build ./...` 在 Task 3/4 完成前仍会失败因为 GORM store 没改，这是预期的。

- [ ] **Step 2.6: 提交**

```bash
git add server/service/cloud/usecase/account.go server/service/cloud/usecase/account_test.go
git commit -m "$(cat <<'EOF'
扩展 CloudAccountService 支持 tag 维度

Constraint: AccountStore.Create 收敛为接收 AccountRow；List 接收 TagFilter
Confidence: high
Scope-risk: medium（影响 store 实现和 handler 调用方）
Tested: go test ./service/cloud/usecase/...
EOF
)"
```

---

### Task 3: GormAccountStore tag 持久化 + JSON_EXTRACT 过滤

**Files:**
- Modify: `server/service/cloud/infra/persistence/account_store_gorm.go`
- Modify: `server/service/cloud/infra/persistence/account_store_gorm_test.go`

- [ ] **Step 3.1: 替换 Create / Get / List / Update / UpdateTags 实现**

`account_store_gorm.go` 完整方法实现（仅展示关键改动函数；其他方法如 Delete / UpdateSyncStatus 保持原状）：

在文件顶部加 `"encoding/json"` import。

```go
func (s *GormAccountStore) Create(ctx context.Context, projectID uint, row usecase.AccountRow) (usecase.AccountRow, error) {
	tagsJSON, err := marshalTags(row.Tags)
	if err != nil {
		return usecase.AccountRow{}, err
	}
	rec := cloudModel.CloudAccount{
		ProjectID:      projectID,
		Name:           row.Name,
		Provider:       row.Provider,
		AccountID:      row.AccountID,
		CredentialType: row.CredentialType,
		CredentialJSON: row.CredentialJSON,
		TagsJSON:       tagsJSON,
		Status:         "active",
	}
	if err := s.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return usecase.AccountRow{}, err
	}
	return toAccountRow(rec)
}

func (s *GormAccountStore) Get(ctx context.Context, accessibleProjectIDs []uint, id uint) (usecase.AccountRow, error) {
	if len(accessibleProjectIDs) == 0 {
		return usecase.AccountRow{}, usecase.ErrAccountNotFound
	}
	var rec cloudModel.CloudAccount
	err := s.db.WithContext(ctx).
		Where("id = ? AND project_id IN ?", id, accessibleProjectIDs).
		First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return usecase.AccountRow{}, usecase.ErrAccountNotFound
	}
	if err != nil {
		return usecase.AccountRow{}, err
	}
	return toAccountRow(rec)
}

func (s *GormAccountStore) GetForSync(ctx context.Context, id uint) (usecase.AccountRow, error) {
	var rec cloudModel.CloudAccount
	err := s.db.WithContext(ctx).First(&rec, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return usecase.AccountRow{}, usecase.ErrAccountNotFound
	}
	if err != nil {
		return usecase.AccountRow{}, err
	}
	return toAccountRow(rec)
}

func (s *GormAccountStore) List(ctx context.Context, accessibleProjectIDs []uint, filters []usecase.TagFilter) ([]usecase.AccountRow, error) {
	if len(accessibleProjectIDs) == 0 {
		return []usecase.AccountRow{}, nil
	}
	tx := s.db.WithContext(ctx).Where("project_id IN ?", accessibleProjectIDs)
	for _, f := range filters {
		expr, args := tagFilterToSQL(f)
		tx = tx.Where(expr, args...)
	}
	var recs []cloudModel.CloudAccount
	if err := tx.Order("created_at DESC").Find(&recs).Error; err != nil {
		return nil, err
	}
	out := make([]usecase.AccountRow, 0, len(recs))
	for _, rec := range recs {
		r, err := toAccountRow(rec)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (s *GormAccountStore) UpdateTags(ctx context.Context, accessibleProjectIDs []uint, id uint, tags map[string]string) (usecase.AccountRow, error) {
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
	tagsJSON, err := marshalTags(tags)
	if err != nil {
		return usecase.AccountRow{}, err
	}
	if err := s.db.WithContext(ctx).Model(&existing).Update("tags_json", tagsJSON).Error; err != nil {
		return usecase.AccountRow{}, err
	}
	existing.TagsJSON = tagsJSON
	return toAccountRow(existing)
}
```

新增辅助函数（文件末尾）：

```go
func marshalTags(tags map[string]string) (string, error) {
	if tags == nil {
		return "{}", nil
	}
	b, err := json.Marshal(tags)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func unmarshalTags(s string) (map[string]string, error) {
	if s == "" {
		return map[string]string{}, nil
	}
	var out map[string]string
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]string{}
	}
	return out, nil
}

// tagFilterToSQL maps a TagFilter to a (where-expression, args) pair.
// Uses MySQL JSON_EXTRACT; the column is `tags_json`.
func tagFilterToSQL(f usecase.TagFilter) (string, []any) {
	// JSON path must come as a literal string, not a parameter (MySQL limitation).
	path := "$." + f.Key
	if !f.Exists {
		return "JSON_EXTRACT(tags_json, ?) IS NULL", []any{path}
	}
	if f.Value == "" {
		return "JSON_EXTRACT(tags_json, ?) IS NOT NULL", []any{path}
	}
	return "JSON_UNQUOTE(JSON_EXTRACT(tags_json, ?)) = ?", []any{path, f.Value}
}
```

把现有 `toAccountRow` 改成返回 `(AccountRow, error)`，把 TagsJSON 解析进去：

```go
func toAccountRow(m cloudModel.CloudAccount) (usecase.AccountRow, error) {
	tags, err := unmarshalTags(m.TagsJSON)
	if err != nil {
		return usecase.AccountRow{}, err
	}
	return usecase.AccountRow{
		ID:             m.ID,
		ProjectID:      m.ProjectID,
		Name:           m.Name,
		Provider:       m.Provider,
		AccountID:      m.AccountID,
		CredentialType: m.CredentialType,
		CredentialJSON: m.CredentialJSON,
		Tags:           tags,
		Status:         m.Status,
		LastSyncedAt:   m.LastSyncedAt,
	}, nil
}
```

`Update`（凭证轮换/重命名）保留原行为；签名不变。

- [ ] **Step 3.2: 更新 + 新增集成测试**

`account_store_gorm_test.go` 中：
- 既有 `store.Create(ctx, projectID, name, provider, accountID, credType, enc)` 调用改为 `store.Create(ctx, projectID, usecase.AccountRow{Name: name, Provider: provider, AccountID: accountID, CredentialType: credType, CredentialJSON: enc})`。
- 既有 `store.List(ctx, accessible)` 改为 `store.List(ctx, accessible, nil)`。

新增 3 个测试：

```go
func TestUpdateTags_RoundTrip(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	row, err := store.Create(context.Background(), 1, usecase.AccountRow{
		Name: "n", Provider: "aliyun", AccountID: "a", CredentialType: "ak_sk", CredentialJSON: "enc",
	})
	require.NoError(t, err)

	updated, err := store.UpdateTags(context.Background(), []uint{1}, row.ID,
		map[string]string{"environment": "prod", "team": "sre"})
	require.NoError(t, err)
	require.Equal(t, "prod", updated.Tags["environment"])
	require.Equal(t, "sre", updated.Tags["team"])

	got, err := store.Get(context.Background(), []uint{1}, row.ID)
	require.NoError(t, err)
	require.Equal(t, "prod", got.Tags["environment"])
}

func TestList_TagFilterEnvironmentProd(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	r1, err := store.Create(context.Background(), 1, usecase.AccountRow{
		Name: "p1", Provider: "aliyun", AccountID: "a1", CredentialType: "ak_sk", CredentialJSON: "e1",
	})
	require.NoError(t, err)
	_, err = store.UpdateTags(context.Background(), []uint{1}, r1.ID, map[string]string{"environment": "prod"})
	require.NoError(t, err)

	r2, err := store.Create(context.Background(), 1, usecase.AccountRow{
		Name: "p2", Provider: "aliyun", AccountID: "a2", CredentialType: "ak_sk", CredentialJSON: "e2",
	})
	require.NoError(t, err)
	_, err = store.UpdateTags(context.Background(), []uint{1}, r2.ID, map[string]string{"environment": "dev"})
	require.NoError(t, err)

	rows, err := store.List(context.Background(), []uint{1},
		[]usecase.TagFilter{{Key: "environment", Value: "prod", Exists: true}})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "p1", rows[0].Name)
}

func TestList_TagFilterNoneKey(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	r1, err := store.Create(context.Background(), 1, usecase.AccountRow{
		Name: "tagged", Provider: "aliyun", AccountID: "a1", CredentialType: "ak_sk", CredentialJSON: "e1",
	})
	require.NoError(t, err)
	_, err = store.UpdateTags(context.Background(), []uint{1}, r1.ID, map[string]string{"environment": "prod"})
	require.NoError(t, err)

	_, err = store.Create(context.Background(), 1, usecase.AccountRow{
		Name: "untagged", Provider: "aliyun", AccountID: "a2", CredentialType: "ak_sk", CredentialJSON: "e2",
	})
	require.NoError(t, err)

	rows, err := store.List(context.Background(), []uint{1},
		[]usecase.TagFilter{{Key: "environment", Exists: false}})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "untagged", rows[0].Name)
}
```

- [ ] **Step 3.3: 跑测试**

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./service/cloud/infra/persistence/...
```

Expected: SKIP 没有 MySQL；with `OPTIMUS_TEST_MYSQL_DSN` 时 PASS 全部（含 3 个新增）。`go build` 还会因为 handler 没改而失败，那是 Task 5 的事。

- [ ] **Step 3.4: 提交**

```bash
git add server/service/cloud/infra/persistence/account_store_gorm.go server/service/cloud/infra/persistence/account_store_gorm_test.go
git commit -m "$(cat <<'EOF'
GormAccountStore 落地 tag 列读写与 JSON_EXTRACT 过滤

Constraint: 使用 MySQL JSON_EXTRACT + JSON_UNQUOTE；__none__ 语义映射到 IS NULL
Confidence: high
Scope-risk: narrow
Tested: go test ./service/cloud/infra/persistence/... (SKIP without DSN)
EOF
)"
```

---

### Task 4: 资产查询按账号 tag 过滤

**Files:**
- Modify: `server/service/cloud/usecase/asset.go`
- Modify: `server/service/cloud/infra/persistence/asset_store_gorm.go`
- Modify: `server/service/cloud/infra/persistence/asset_store_gorm_test.go`

- [ ] **Step 4.1: AssetFilter 加 Tags**

在 `asset.go` 中 `AssetFilter` 结构体里加 `Tags`：

```go
type AssetFilter struct {
	AccessibleProjectIDs []uint
	AccountID            *uint
	Provider             *string
	Region               *string
	ResourceType         *string
	Status               *string
	Keyword              *string
	Tags                 []TagFilter
	Page                 int
	PageSize             int
}
```

- [ ] **Step 4.2: GormAssetStore.Query 接 tag 过滤**

在已有 JOIN 后追加（基于 SP2 已落地的 cloud_accounts JOIN）：

```go
for _, tf := range f.Tags {
	expr, args := tagFilterToSQL(tf)
	// 列名加 cloud_accounts. 前缀
	expr = strings.Replace(expr, "tags_json", "cloud_accounts.tags_json", 1)
	q = q.Where(expr, args...)
}
```

> Note: `tagFilterToSQL` 在 Task 3 的 `account_store_gorm.go` 里定义，但**不**在同一 package。要么在 `asset_store_gorm.go` 复制一份，要么挪到 `usecase/account.go` 暴露为可被两个 store 调用。**做法：在 `asset_store_gorm.go` 复制一份相同 helper**（DRY 小成本，避免跨包依赖）。

在 `asset_store_gorm.go` 文件末尾追加：

```go
func tagFilterToSQL(f usecase.TagFilter) (string, []any) {
	path := "$." + f.Key
	if !f.Exists {
		return "tags_json IS NULL OR JSON_EXTRACT(tags_json, ?) IS NULL", []any{path}
	}
	if f.Value == "" {
		return "JSON_EXTRACT(tags_json, ?) IS NOT NULL", []any{path}
	}
	return "JSON_UNQUOTE(JSON_EXTRACT(tags_json, ?)) = ?", []any{path, f.Value}
}
```

- [ ] **Step 4.3: 新增集成测试**

`asset_store_gorm_test.go` 中新增：

```go
func TestQuery_FiltersByAccountTag(t *testing.T) {
	db := openCloudTestDB(t)
	accountStore := NewGormAccountStore(db)
	assetStore := NewGormAssetStore(db)

	a1, err := accountStore.Create(context.Background(), 1, usecase.AccountRow{
		Name: "prod-acc", Provider: "aliyun", AccountID: "p1", CredentialType: "ak_sk", CredentialJSON: "e1",
	})
	require.NoError(t, err)
	_, err = accountStore.UpdateTags(context.Background(), []uint{1}, a1.ID, map[string]string{"environment": "prod"})
	require.NoError(t, err)

	a2, err := accountStore.Create(context.Background(), 1, usecase.AccountRow{
		Name: "dev-acc", Provider: "aliyun", AccountID: "d1", CredentialType: "ak_sk", CredentialJSON: "e2",
	})
	require.NoError(t, err)
	_, err = accountStore.UpdateTags(context.Background(), []uint{1}, a2.ID, map[string]string{"environment": "dev"})
	require.NoError(t, err)

	_, err = assetStore.Upsert(context.Background(), a1.ID, "aliyun", []usecase.RawAsset{
		{ResourceType: "ecs", ResourceID: "i-1", Name: "prod-vm"},
	})
	require.NoError(t, err)
	_, err = assetStore.Upsert(context.Background(), a2.ID, "aliyun", []usecase.RawAsset{
		{ResourceType: "ecs", ResourceID: "i-2", Name: "dev-vm"},
	})
	require.NoError(t, err)

	page, err := assetStore.Query(context.Background(), usecase.AssetFilter{
		AccessibleProjectIDs: []uint{1},
		Tags:                 []usecase.TagFilter{{Key: "environment", Value: "prod", Exists: true}},
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "prod-vm", page.Items[0].Name)
}
```

- [ ] **Step 4.4: 测试**

```bash
cd /home/lilin/trae-project/Optimus/server
go test ./service/cloud/infra/persistence/...
```

Expected: 全部 PASS / SKIP（视 MySQL 而定）。

- [ ] **Step 4.5: 提交**

```bash
git add server/service/cloud/usecase/asset.go server/service/cloud/infra/persistence/asset_store_gorm.go server/service/cloud/infra/persistence/asset_store_gorm_test.go
git commit -m "$(cat <<'EOF'
资产查询按账号 tag 过滤（继承账号 tag）

Constraint: 复用 SP2 cloud_accounts JOIN；JSON_EXTRACT 列名加表前缀
Confidence: high
Scope-risk: narrow
Tested: go test ./service/cloud/infra/persistence/...
EOF
)"
```

---

### Task 5: HTTP Handlers + 路由 + 权限种子

**Files:**
- Modify: `server/api/v1/cloud/account.go`
- Modify: `server/api/v1/cloud/asset.go`
- Modify: `server/api/v1/cloud/handlers.go`（如需要 helper）
- Modify: `server/router/cloud/enter.go`
- Modify: `server/service/cloud/cloud.go`

- [ ] **Step 5.1: 在 account.go 加 PUT tags + list 解析 tag 参数**

新增辅助函数（同文件顶部）：

```go
func parseTagFilters(c *gin.Context) ([]cloudUsecase.TagFilter, error) {
	tokens := c.QueryArray("tag")
	if len(tokens) == 0 {
		return nil, nil
	}
	out := make([]cloudUsecase.TagFilter, 0, len(tokens))
	for _, tok := range tokens {
		tf, err := cloudUsecase.ParseTagFilter(tok)
		if err != nil {
			return nil, err
		}
		out = append(out, tf)
	}
	return out, nil
}
```

修改 `List` 方法（保留现有 scope 获取逻辑），追加 tag 解析：

```go
func (h *CloudHandlers) List(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "account service is not available")
		return
	}
	scope := projectScope(c)
	tagFilters, err := parseTagFilters(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	views, err := h.accountSvc.List(c.Request.Context(), scope.AccessibleProjectIDs, tagFilters)
	if err != nil {
		writeCloudError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views})
}
```

新增 PUT tags handler：

```go
func (h *CloudHandlers) UpdateAccountTags(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "account service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	var req cloudUsecase.UpdateAccountTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid tags payload"})
		return
	}
	scope := projectScope(c)
	view, err := h.accountSvc.UpdateTags(c.Request.Context(), scope.AccessibleProjectIDs, id, req)
	if err != nil {
		writeCloudError(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}
```

`writeCloudError` 已有；需要补 4 个新 sentinel 映射（追加分支）：

```go
// 在 writeCloudError 的 switch 里添加：
case errors.Is(err, cloudUsecase.ErrInvalidTagKey),
	errors.Is(err, cloudUsecase.ErrInvalidTagValue),
	errors.Is(err, cloudUsecase.ErrTooManyTags),
	errors.Is(err, cloudUsecase.ErrInvalidTagFilter):
	c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	return
```

- [ ] **Step 5.2: 在 asset.go 解析 tag 参数**

在 `parseAssetFilter` 中追加：

```go
tagFilters, err := parseTagFilters(c)
if err != nil {
	return cloudUsecase.AssetFilter{}, err
}
filter.Tags = tagFilters
```

（如果 `parseAssetFilter` 返回签名是 `(filter, error)` 就直接返回 error；如果是 panic / write 自身就在调用处捕获。具体看现有代码，按既有模式对齐。）

- [ ] **Step 5.3: 注册 PUT /accounts/:id/tags 路由**

`server/router/cloud/enter.go` 在 accounts 子组内追加：

```go
accounts.PUT(":id/tags", h.UpdateAccountTags)
```

- [ ] **Step 5.4: 加 cloud.accounts.tags.update 权限种子**

`server/service/cloud/cloud.go` `DefaultCloudPermissions` 追加一行（按现有 Sort 顺序：assets.view 是 805 左右，找一个合适位置插入）：

```go
{Code: "cloud.accounts.tags.update", Name: "Update cloud account tags", Type: "api", Method: "PUT", Path: "/api/v1/cloud/accounts/:id/tags", Sort: 808},
```

- [ ] **Step 5.5: 全量编译 + 测试**

```bash
cd /home/lilin/trae-project/Optimus/server
go build ./...
go test ./...
```

Expected: 全部 PASS / SKIP（视 MySQL 而定）。

- [ ] **Step 5.6: 提交**

```bash
git add server/api/v1/cloud/ server/router/cloud/enter.go server/service/cloud/cloud.go
git commit -m "$(cat <<'EOF'
新增 cloud 账号 tag HTTP 接口与 list tag 过滤

Constraint: PUT /api/v1/cloud/accounts/:id/tags 全量覆盖；GET 列表接受多个 tag=...
Confidence: high
Scope-risk: medium（influences SP4 frontend wiring）
Tested: go build ./... && go test ./...
EOF
)"
```

---

### Task 6: 前端 - types + api + Cloud Accounts 基础页面 + 路由 + 导航激活

**Files:**
- Create: `web/src/types/cloud.ts`
- Create: `web/src/api/cloud.ts`
- Create: `web/src/view/cloud/CloudAccountsPage.tsx` + `.module.css`
- Modify: `web/src/router/index.tsx`
- Modify: `web/src/components/nav/SideNav.tsx`
- Modify: `web/src/locale/messages.ts`

- [ ] **Step 6.1: 写 types**

`web/src/types/cloud.ts`：

```ts
export interface CloudAccount {
  id: number;
  projectId: number;
  name: string;
  provider: string;
  accountId: string;
  credentialType: string;
  credential: string; // always "***" in responses
  tags: Record<string, string>;
  status: string;
  lastSyncedAt: string | null;
}

export interface CreateCloudAccountPayload {
  name: string;
  provider: string;
  accountId: string;
  credentialType: string;
  credential: Record<string, string>;
  tags?: Record<string, string>;
}

export interface UpdateCloudAccountTagsPayload {
  tags: Record<string, string>;
}

export interface CloudAsset {
  id: number;
  accountId: number;
  provider: string;
  region: string;
  resourceType: string;
  resourceId: string;
  name: string;
  status: string;
  configJson: string;
  tagsJson: string;
  discoveredAt: string;
  lastSeenAt: string;
  deletedAt: string | null;
}

export interface CloudAssetPage {
  items: CloudAsset[];
  total: number;
  page: number;
  pageSize: number;
}
```

- [ ] **Step 6.2: 写 api client**

`web/src/api/cloud.ts`：

```ts
import { apiRequest } from './client';
import type {
  CloudAccount,
  CloudAssetPage,
  CreateCloudAccountPayload,
  UpdateCloudAccountTagsPayload,
} from '../types/cloud';
import type { ListResponse } from '../types/system';

function buildTagQuery(tags: string[]): string {
  if (tags.length === 0) {
    return '';
  }
  return '?' + tags.map((t) => `tag=${encodeURIComponent(t)}`).join('&');
}

export function listCloudAccounts(tagFilters: string[] = []) {
  return apiRequest<ListResponse<CloudAccount>>('/cloud/accounts' + buildTagQuery(tagFilters));
}

export function createCloudAccount(payload: CreateCloudAccountPayload) {
  return apiRequest<CloudAccount>('/cloud/accounts', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
}

export function deleteCloudAccount(id: number) {
  return apiRequest<{ message: string }>(`/cloud/accounts/${id}`, { method: 'DELETE' });
}

export function syncCloudAccount(id: number) {
  return apiRequest<{ upserted: number; errors: number }>(`/cloud/accounts/${id}/sync`, {
    method: 'POST',
  });
}

export function updateCloudAccountTags(id: number, payload: UpdateCloudAccountTagsPayload) {
  return apiRequest<CloudAccount>(`/cloud/accounts/${id}/tags`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
}

export function listCloudAssets(tagFilters: string[] = []) {
  return apiRequest<CloudAssetPage>('/cloud/assets' + buildTagQuery(tagFilters));
}
```

- [ ] **Step 6.3: 写 CloudAccountsPage 基础版（无 tag UI，先跑通 CRUD）**

`web/src/view/cloud/CloudAccountsPage.tsx`：

```tsx
import { FormEvent, useEffect, useState } from 'react';

import { createCloudAccount, deleteCloudAccount, listCloudAccounts, syncCloudAccount } from '../../api/cloud';
import { useI18n } from '../../locale/useI18n';
import type { CloudAccount } from '../../types/cloud';
import styles from './CloudAccountsPage.module.css';

interface Draft {
  name: string;
  accountId: string;
  accessKeyId: string;
  accessKeySecret: string;
}

const emptyDraft: Draft = { name: '', accountId: '', accessKeyId: '', accessKeySecret: '' };

export function CloudAccountsPage() {
  const { t } = useI18n();
  const [accounts, setAccounts] = useState<CloudAccount[]>([]);
  const [draft, setDraft] = useState<Draft>(emptyDraft);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<number | null>(null);

  async function load() {
    try {
      const result = await listCloudAccounts();
      setAccounts(result.items);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.accounts.loadFailed'));
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function handleCreate(event: FormEvent) {
    event.preventDefault();
    try {
      await createCloudAccount({
        name: draft.name,
        provider: 'aliyun',
        accountId: draft.accountId,
        credentialType: 'ak_sk',
        credential: { access_key_id: draft.accessKeyId, access_key_secret: draft.accessKeySecret },
      });
      setDraft(emptyDraft);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.accounts.saveFailed'));
    }
  }

  async function handleDelete(account: CloudAccount) {
    if (!window.confirm(t('cloud.accounts.deleteConfirm'))) return;
    setBusyId(account.id);
    try {
      await deleteCloudAccount(account.id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.accounts.deleteFailed'));
    } finally {
      setBusyId(null);
    }
  }

  async function handleSync(account: CloudAccount) {
    setBusyId(account.id);
    try {
      await syncCloudAccount(account.id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.accounts.syncFailed'));
    } finally {
      setBusyId(null);
    }
  }

  return (
    <section className={styles.page}>
      <header className={styles.header}>
        <div>
          <p className="page-eyebrow">{t('cloud.accounts.eyebrow')}</p>
          <h1 className={styles.title}>{t('cloud.accounts.title')}</h1>
          <p className={styles.sub}>{t('cloud.accounts.sub')}</p>
        </div>
      </header>

      {error ? <div className={styles.error}>{error}</div> : null}

      <form className={styles.createBar} onSubmit={(event) => void handleCreate(event)}>
        <label>
          <span>{t('cloud.accounts.name')}</span>
          <input
            value={draft.name}
            onChange={(event) => setDraft((value) => ({ ...value, name: event.target.value }))}
            required
          />
        </label>
        <label>
          <span>{t('cloud.accounts.accountId')}</span>
          <input
            value={draft.accountId}
            onChange={(event) => setDraft((value) => ({ ...value, accountId: event.target.value }))}
            required
          />
        </label>
        <label>
          <span>{t('cloud.accounts.accessKeyId')}</span>
          <input
            value={draft.accessKeyId}
            onChange={(event) => setDraft((value) => ({ ...value, accessKeyId: event.target.value }))}
            required
          />
        </label>
        <label>
          <span>{t('cloud.accounts.accessKeySecret')}</span>
          <input
            type="password"
            value={draft.accessKeySecret}
            onChange={(event) => setDraft((value) => ({ ...value, accessKeySecret: event.target.value }))}
            required
          />
        </label>
        <button type="submit">{t('cloud.accounts.create')}</button>
      </form>

      <div className={styles.list}>
        {accounts.map((account) => (
          <article className={styles.row} key={account.id}>
            <div className={styles.identity}>
              <strong>{account.name}</strong>
              <span>{account.provider} · {account.accountId}</span>
              <em>{account.status}</em>
            </div>
            <div className={styles.actions}>
              <button type="button" onClick={() => void handleSync(account)} disabled={busyId === account.id}>
                {t('cloud.accounts.sync')}
              </button>
              <button
                type="button"
                className={styles.danger}
                onClick={() => void handleDelete(account)}
                disabled={busyId === account.id}
              >
                {t('cloud.accounts.delete')}
              </button>
            </div>
          </article>
        ))}
        {accounts.length === 0 ? <p className={styles.empty}>{t('cloud.accounts.empty')}</p> : null}
      </div>
    </section>
  );
}
```

`web/src/view/cloud/CloudAccountsPage.module.css`：

```css
.page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
}

.title {
  margin: 0;
  color: var(--text-primary);
}

.sub {
  margin: 6px 0 0;
  color: var(--shell-muted);
  font-size: 0.92rem;
}

.error {
  padding: 12px 14px;
  border: 1px solid rgba(154, 52, 18, 0.22);
  border-radius: 8px;
  background: rgba(255, 247, 237, 0.82);
  color: #9a3412;
}

.createBar {
  display: grid;
  grid-template-columns: repeat(2, 1fr) auto;
  gap: 10px;
  padding: 12px;
  border: 1px solid var(--shell-border);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.56);
}

.createBar label {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.createBar label span {
  color: var(--shell-muted);
  font-size: 0.72rem;
  font-weight: 800;
}

.createBar input {
  min-height: 38px;
  border: 1px solid var(--shell-border);
  border-radius: 8px;
  padding: 0 10px;
  background: rgba(255, 255, 255, 0.74);
}

.createBar button {
  min-height: 38px;
  border: 1px solid var(--shell-border);
  border-radius: 8px;
  padding: 0 14px;
  background: var(--shell-control-active-bg);
  color: var(--shell-control-active-fg);
  font-weight: 700;
  cursor: pointer;
  grid-column: 3;
  align-self: end;
}

.list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 10px;
  align-items: center;
  padding: 12px;
  border: 1px solid var(--shell-border);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.56);
}

.identity {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.identity strong {
  color: var(--text-primary);
}

.identity span {
  color: var(--shell-muted);
  font-size: 0.86rem;
}

.identity em {
  color: var(--shell-muted);
  font-style: normal;
  font-size: 0.78rem;
  font-weight: 700;
  text-transform: uppercase;
}

.actions {
  display: flex;
  gap: 8px;
}

.actions button {
  min-height: 36px;
  border: 1px solid var(--shell-border);
  border-radius: 8px;
  padding: 0 12px;
  background: rgba(255, 255, 255, 0.72);
  color: var(--shell-control-active-bg);
  font-weight: 700;
  cursor: pointer;
}

.danger {
  color: #9a3412 !important;
}

.empty {
  color: var(--shell-muted);
  text-align: center;
}
```

- [ ] **Step 6.4: 注册路由**

`web/src/router/index.tsx` 在 `/projects/:code/members` 后追加 cloud 路由（保持 ProjectProvider 包装）：

```tsx
import { CloudAccountsPage } from '../view/cloud/CloudAccountsPage';
// ... 其它已有 imports

{
  path: 'cloud/accounts',
  element: <CloudAccountsPage />,
},
```

- [ ] **Step 6.5: SideNav 激活云账号入口**

`web/src/components/nav/SideNav.tsx` 找到 cloud accounts 那条带 "Coming soon" 的入口（之前 codex 加的占位），把 disabled / Coming soon 标记去掉，改为有效的 `<NavLink to="/cloud/accounts">`。**保持 `hasMenu('cloud.accounts.menu')` 权限门**（SP1/SP2 已种子）。

实施者打开 SideNav.tsx 后定位到使用 `comingSoon` 的位置（grep `comingSoon`），把这一条改成真链接，并删除该条目的 `comingSoon` 后缀。**保留** `cloud.assets.menu` 那条暂时仍 Coming soon（Task 9 才激活）。

- [ ] **Step 6.6: i18n keys**

`web/src/locale/messages.ts` EN 部分加：

```ts
cloud: {
  accounts: {
    eyebrow: 'Cloud accounts',
    title: 'Cloud accounts',
    sub: 'Manage cloud provider credentials for the current project',
    name: 'Name',
    accountId: 'Account ID',
    accessKeyId: 'Access key ID',
    accessKeySecret: 'Access key secret',
    create: 'Create',
    sync: 'Sync',
    delete: 'Delete',
    empty: 'No cloud accounts',
    loadFailed: 'Failed to load cloud accounts',
    saveFailed: 'Failed to save cloud account',
    deleteFailed: 'Failed to delete cloud account',
    deleteConfirm: 'Delete this cloud account?',
    syncFailed: 'Failed to sync cloud account',
  },
},
```

ZH 部分加：

```ts
cloud: {
  accounts: {
    eyebrow: '云账号',
    title: '云账号',
    sub: '管理当前项目下的云厂商账号凭证',
    name: '名称',
    accountId: '账号 ID',
    accessKeyId: 'Access Key ID',
    accessKeySecret: 'Access Key Secret',
    create: '创建',
    sync: '同步',
    delete: '删除',
    empty: '当前项目下暂无云账号',
    loadFailed: '加载云账号失败',
    saveFailed: '保存云账号失败',
    deleteFailed: '删除云账号失败',
    deleteConfirm: '确定删除该云账号？',
    syncFailed: '同步云账号失败',
  },
},
```

- [ ] **Step 6.7: 构建 + 提交**

```bash
cd /home/lilin/trae-project/Optimus/web && npm run build
```

Expected: clean build。

```bash
cd /home/lilin/trae-project/Optimus
git add web/src/types/cloud.ts web/src/api/cloud.ts web/src/view/cloud/ web/src/router/index.tsx web/src/components/nav/SideNav.tsx web/src/locale/messages.ts
git commit -m "$(cat <<'EOF'
新增 Cloud Accounts 页面（基础 CRUD + 同步）

Constraint: tag UI 留给下一 task；激活 SideNav cloud 账号入口
Confidence: high
Scope-risk: narrow
Tested: npm run build
EOF
)"
```

---

### Task 7: TagChip + TagEditor 组件 + 集成到账号页

**Files:**
- Create: `web/src/components/tag/TagChip.tsx` + `.module.css`
- Create: `web/src/components/tag/TagEditor.tsx` + `.module.css`
- Modify: `web/src/view/cloud/CloudAccountsPage.tsx`
- Modify: `web/src/view/cloud/CloudAccountsPage.module.css`
- Modify: `web/src/locale/messages.ts`

- [ ] **Step 7.1: TagChip 组件**

`web/src/components/tag/TagChip.tsx`：

```tsx
import styles from './TagChip.module.css';

interface TagChipProps {
  tagKey: string;
  value: string;
  onRemove?: () => void;
}

const ENV_VALUES: Record<string, string> = {
  prod: styles.envProd,
  dev: styles.envDev,
  qa: styles.envQa,
  orphan: styles.envOrphan,
};

export function TagChip({ tagKey, value, onRemove }: TagChipProps) {
  const envClass = tagKey === 'environment' ? ENV_VALUES[value] : undefined;
  return (
    <span className={[styles.chip, envClass].filter(Boolean).join(' ')}>
      <span className={styles.key}>{tagKey}</span>
      <span className={styles.eq}>=</span>
      <span>{value}</span>
      {onRemove ? (
        <button type="button" className={styles.remove} onClick={onRemove} aria-label="remove tag">
          ×
        </button>
      ) : null}
    </span>
  );
}
```

`web/src/components/tag/TagChip.module.css`：

```css
.chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 0.74rem;
  font-weight: 600;
  background: rgba(110, 93, 81, 0.10);
  color: var(--text-secondary);
  font-family: ui-monospace, SFMono-Regular, monospace;
}

.key { color: var(--shell-muted); }
.eq { color: var(--shell-muted); opacity: 0.6; }

.envProd { background: rgba(179, 71, 61, 0.12); color: #9a3412; }
.envProd .key { color: #9a3412; }
.envDev { background: rgba(47, 95, 89, 0.14); color: #2f5f59; }
.envDev .key { color: #2f5f59; }
.envQa { background: rgba(183, 121, 75, 0.18); color: #8a5530; }
.envQa .key { color: #8a5530; }
.envOrphan { background: rgba(110, 93, 81, 0.20); color: #6e5d51; }

.remove {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 4px;
  border: 0;
  background: rgba(110, 93, 81, 0.18);
  color: var(--shell-muted);
  font-size: 0.7rem;
  cursor: pointer;
}
```

- [ ] **Step 7.2: TagEditor 组件**

`web/src/components/tag/TagEditor.tsx`：

```tsx
import { useState } from 'react';

import { useI18n } from '../../locale/useI18n';
import { TagChip } from './TagChip';
import styles from './TagEditor.module.css';

const TAG_KEY_REGEX = /^[a-z][a-z0-9_-]{0,62}$/;
const MAX_TAGS = 32;
const SUGGESTIONS = ['environment=prod', 'environment=dev', 'environment=qa', 'environment=orphan'];

interface TagEditorProps {
  initialTags: Record<string, string>;
  onSubmit: (tags: Record<string, string>) => Promise<void>;
  onCancel: () => void;
}

export function TagEditor({ initialTags, onSubmit, onCancel }: TagEditorProps) {
  const { t } = useI18n();
  const [tags, setTags] = useState<Record<string, string>>(initialTags);
  const [keyInput, setKeyInput] = useState('');
  const [valueInput, setValueInput] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  function addTag(rawKey: string, rawValue: string) {
    const key = rawKey.trim();
    const value = rawValue.trim();
    if (!TAG_KEY_REGEX.test(key)) {
      setError(t('cloud.tags.invalidKey'));
      return;
    }
    if (value.length === 0 || value.length > 128) {
      setError(t('cloud.tags.invalidValue'));
      return;
    }
    if (!(key in tags) && Object.keys(tags).length >= MAX_TAGS) {
      setError(t('cloud.tags.tooMany'));
      return;
    }
    setError(null);
    setTags((prev) => ({ ...prev, [key]: value }));
    setKeyInput('');
    setValueInput('');
  }

  function removeTag(key: string) {
    setTags((prev) => {
      const next = { ...prev };
      delete next[key];
      return next;
    });
  }

  function applySuggestion(suggestion: string) {
    const [k, v] = suggestion.split('=', 2);
    addTag(k, v);
  }

  async function handleSubmit() {
    setSubmitting(true);
    try {
      await onSubmit(tags);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.tags.saveFailed'));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className={styles.editor}>
      <div className={styles.tagList}>
        {Object.entries(tags).length === 0 ? (
          <span className={styles.empty}>{t('cloud.tags.empty')}</span>
        ) : (
          Object.entries(tags).map(([k, v]) => (
            <TagChip key={k} tagKey={k} value={v} onRemove={() => removeTag(k)} />
          ))
        )}
      </div>

      <div className={styles.addRow}>
        <input
          className={styles.keyInput}
          placeholder={t('cloud.tags.keyPlaceholder')}
          value={keyInput}
          onChange={(e) => setKeyInput(e.target.value)}
        />
        <input
          className={styles.valueInput}
          placeholder={t('cloud.tags.valuePlaceholder')}
          value={valueInput}
          onChange={(e) => setValueInput(e.target.value)}
        />
        <button type="button" onClick={() => addTag(keyInput, valueInput)} disabled={submitting}>
          {t('cloud.tags.add')}
        </button>
      </div>

      {error ? <p className={styles.error}>{error}</p> : null}

      <p className={styles.hint}>{t('cloud.tags.hint')}</p>
      <div className={styles.suggestions}>
        {SUGGESTIONS.map((s) => (
          <button key={s} type="button" className={styles.suggestion} onClick={() => applySuggestion(s)}>
            {s}
          </button>
        ))}
      </div>

      <div className={styles.actions}>
        <button type="button" className={styles.cancel} onClick={onCancel} disabled={submitting}>
          {t('cloud.tags.cancel')}
        </button>
        <button type="button" className={styles.save} onClick={() => void handleSubmit()} disabled={submitting}>
          {t('cloud.tags.save')}
        </button>
      </div>
    </div>
  );
}
```

`web/src/components/tag/TagEditor.module.css`：

```css
.editor {
  background: rgba(255, 251, 247, 0.94);
  border: 1px solid var(--shell-border);
  border-radius: 12px;
  padding: 16px;
  max-width: 480px;
}

.tagList {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  min-height: 24px;
  margin-bottom: 12px;
}

.empty {
  color: var(--shell-muted);
  font-style: italic;
  font-size: 0.84rem;
}

.addRow {
  display: grid;
  grid-template-columns: minmax(7rem, 11rem) 1fr auto;
  gap: 6px;
}

.addRow input {
  min-height: 32px;
  padding: 0 10px;
  border: 1px solid var(--shell-border);
  border-radius: 6px;
  background: white;
  font-family: ui-monospace, SFMono-Regular, monospace;
  font-size: 0.82rem;
}

.addRow button {
  background: var(--shell-control-active-bg);
  color: var(--shell-control-active-fg);
  border: 0;
  border-radius: 6px;
  padding: 0 14px;
  font-weight: 700;
  cursor: pointer;
}

.error {
  margin: 8px 0 0;
  color: #9a3412;
  font-size: 0.78rem;
}

.hint {
  margin: 12px 0 6px;
  color: var(--shell-muted);
  font-size: 0.78rem;
}

.suggestions {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.suggestion {
  padding: 2px 9px;
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.78);
  border: 1px dashed var(--shell-border);
  color: var(--shell-muted);
  font-size: 0.72rem;
  font-family: ui-monospace, SFMono-Regular, monospace;
  cursor: pointer;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 14px;
}

.cancel, .save {
  min-height: 32px;
  border-radius: 6px;
  border: 1px solid var(--shell-border);
  padding: 0 14px;
  font-weight: 700;
  cursor: pointer;
}

.save {
  background: var(--shell-control-active-bg);
  color: var(--shell-control-active-fg);
  border-color: var(--shell-control-active-bg);
}
.cancel {
  background: transparent;
  color: var(--shell-muted);
}
```

- [ ] **Step 7.3: 集成到 CloudAccountsPage**

修改 `CloudAccountsPage.tsx`：
- 在 row 显示 tags（TagChip 列表）
- 行 actions 加 "改 tag" 按钮，点击切换 `editingTagsId` 状态
- editor 用 inline 模式（不是 Modal，简化）：选中行下方展开 `TagEditor`

完整改后版：

```tsx
import { FormEvent, useEffect, useState } from 'react';

import { TagChip } from '../../components/tag/TagChip';
import { TagEditor } from '../../components/tag/TagEditor';
import {
  createCloudAccount,
  deleteCloudAccount,
  listCloudAccounts,
  syncCloudAccount,
  updateCloudAccountTags,
} from '../../api/cloud';
import { useI18n } from '../../locale/useI18n';
import type { CloudAccount } from '../../types/cloud';
import styles from './CloudAccountsPage.module.css';

interface Draft {
  name: string;
  accountId: string;
  accessKeyId: string;
  accessKeySecret: string;
}

const emptyDraft: Draft = { name: '', accountId: '', accessKeyId: '', accessKeySecret: '' };

export function CloudAccountsPage() {
  const { t } = useI18n();
  const [accounts, setAccounts] = useState<CloudAccount[]>([]);
  const [draft, setDraft] = useState<Draft>(emptyDraft);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<number | null>(null);
  const [editingTagsId, setEditingTagsId] = useState<number | null>(null);

  async function load() {
    try {
      const result = await listCloudAccounts();
      setAccounts(result.items);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.accounts.loadFailed'));
    }
  }

  useEffect(() => { void load(); }, []);

  async function handleCreate(event: FormEvent) {
    event.preventDefault();
    try {
      await createCloudAccount({
        name: draft.name, provider: 'aliyun', accountId: draft.accountId, credentialType: 'ak_sk',
        credential: { access_key_id: draft.accessKeyId, access_key_secret: draft.accessKeySecret },
      });
      setDraft(emptyDraft);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.accounts.saveFailed'));
    }
  }

  async function handleDelete(account: CloudAccount) {
    if (!window.confirm(t('cloud.accounts.deleteConfirm'))) return;
    setBusyId(account.id);
    try {
      await deleteCloudAccount(account.id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.accounts.deleteFailed'));
    } finally { setBusyId(null); }
  }

  async function handleSync(account: CloudAccount) {
    setBusyId(account.id);
    try {
      await syncCloudAccount(account.id);
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.accounts.syncFailed'));
    } finally { setBusyId(null); }
  }

  async function handleTagsSubmit(account: CloudAccount, tags: Record<string, string>) {
    await updateCloudAccountTags(account.id, { tags });
    setEditingTagsId(null);
    await load();
  }

  return (
    <section className={styles.page}>
      <header className={styles.header}>
        <div>
          <p className="page-eyebrow">{t('cloud.accounts.eyebrow')}</p>
          <h1 className={styles.title}>{t('cloud.accounts.title')}</h1>
          <p className={styles.sub}>{t('cloud.accounts.sub')}</p>
        </div>
      </header>

      {error ? <div className={styles.error}>{error}</div> : null}

      <form className={styles.createBar} onSubmit={(e) => void handleCreate(e)}>
        <label><span>{t('cloud.accounts.name')}</span>
          <input value={draft.name} onChange={(e) => setDraft((v) => ({ ...v, name: e.target.value }))} required />
        </label>
        <label><span>{t('cloud.accounts.accountId')}</span>
          <input value={draft.accountId} onChange={(e) => setDraft((v) => ({ ...v, accountId: e.target.value }))} required />
        </label>
        <label><span>{t('cloud.accounts.accessKeyId')}</span>
          <input value={draft.accessKeyId} onChange={(e) => setDraft((v) => ({ ...v, accessKeyId: e.target.value }))} required />
        </label>
        <label><span>{t('cloud.accounts.accessKeySecret')}</span>
          <input type="password" value={draft.accessKeySecret} onChange={(e) => setDraft((v) => ({ ...v, accessKeySecret: e.target.value }))} required />
        </label>
        <button type="submit">{t('cloud.accounts.create')}</button>
      </form>

      <div className={styles.list}>
        {accounts.map((account) => (
          <article className={styles.rowWrapper} key={account.id}>
            <div className={styles.row}>
              <div className={styles.identity}>
                <strong>{account.name}</strong>
                <span>{account.provider} · {account.accountId}</span>
                <em>{account.status}</em>
              </div>
              <div className={styles.tags}>
                {Object.entries(account.tags).length === 0 ? (
                  <span className={styles.untagged}>{t('cloud.accounts.untagged')}</span>
                ) : (
                  Object.entries(account.tags).map(([k, v]) => <TagChip key={k} tagKey={k} value={v} />)
                )}
              </div>
              <div className={styles.actions}>
                <button type="button" onClick={() => setEditingTagsId(editingTagsId === account.id ? null : account.id)}>
                  {editingTagsId === account.id ? t('cloud.accounts.closeTags') : t('cloud.accounts.editTags')}
                </button>
                <button type="button" onClick={() => void handleSync(account)} disabled={busyId === account.id}>
                  {t('cloud.accounts.sync')}
                </button>
                <button type="button" className={styles.danger} onClick={() => void handleDelete(account)} disabled={busyId === account.id}>
                  {t('cloud.accounts.delete')}
                </button>
              </div>
            </div>
            {editingTagsId === account.id ? (
              <div className={styles.editorRow}>
                <TagEditor
                  initialTags={account.tags}
                  onSubmit={(tags) => handleTagsSubmit(account, tags)}
                  onCancel={() => setEditingTagsId(null)}
                />
              </div>
            ) : null}
          </article>
        ))}
        {accounts.length === 0 ? <p className={styles.empty}>{t('cloud.accounts.empty')}</p> : null}
      </div>
    </section>
  );
}
```

修改 `CloudAccountsPage.module.css`：在 `.row` 上方加：

```css
.rowWrapper {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.row {
  display: grid;
  grid-template-columns: minmax(12rem, 1fr) minmax(12rem, 2fr) auto;
  gap: 14px;
  align-items: center;
  padding: 12px;
  border: 1px solid var(--shell-border);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.56);
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.untagged {
  padding: 2px 9px;
  border: 1px dashed var(--shell-border);
  border-radius: 6px;
  color: var(--shell-muted);
  font-style: italic;
  font-size: 0.78rem;
}

.editorRow {
  padding-left: 16px;
}
```

（删除旧的单参数 `.row { grid-template-columns: 1fr auto; ... }` 那块——本步骤用上面的三列覆盖）

- [ ] **Step 7.4: i18n keys 追加**

在 `cloud.accounts` 下追加：

EN：
```ts
untagged: 'No tags',
editTags: 'Edit tags',
closeTags: 'Close',
```

ZH：
```ts
untagged: '未打 tag',
editTags: '改 tag',
closeTags: '关闭',
```

新增顶级 `cloud.tags` 段：

EN：
```ts
cloud: {
  ...
  tags: {
    empty: 'No tags yet',
    keyPlaceholder: 'key',
    valuePlaceholder: 'value',
    add: 'Add',
    save: 'Save',
    cancel: 'Cancel',
    invalidKey: 'Tag key must start with a lowercase letter; a-z, 0-9, _, - only',
    invalidValue: 'Tag value must be 1-128 chars',
    tooMany: 'Cannot add more than 32 tags',
    saveFailed: 'Failed to save tags',
    hint: 'Tip: use `environment` to classify by deployment environment',
  },
},
```

ZH：
```ts
cloud: {
  ...
  tags: {
    empty: '还没有 tag',
    keyPlaceholder: '键',
    valuePlaceholder: '值',
    add: '添加',
    save: '保存',
    cancel: '取消',
    invalidKey: 'tag 键以小写字母开头，仅允许 a-z、0-9、_、-',
    invalidValue: 'tag 值长度 1-128 个字符',
    tooMany: '最多 32 个 tag',
    saveFailed: '保存 tag 失败',
    hint: '建议使用 `environment` 作为环境分类的 key',
  },
},
```

- [ ] **Step 7.5: 构建 + 提交**

```bash
cd /home/lilin/trae-project/Optimus/web && npm run build
```

```bash
cd /home/lilin/trae-project/Optimus
git add web/src/components/tag/ web/src/view/cloud/ web/src/locale/messages.ts
git commit -m "$(cat <<'EOF'
新增 TagChip / TagEditor 组件并集成到云账号页面

Constraint: environment 在 TagChip 内享有预设配色；editor 校验 key/value/数量
Confidence: high
Scope-risk: narrow
Tested: npm run build
EOF
)"
```

---

### Task 8: TagFilter 组件 + 集成到账号页

**Files:**
- Create: `web/src/components/tag/TagFilter.tsx` + `.module.css`
- Modify: `web/src/view/cloud/CloudAccountsPage.tsx`
- Modify: `web/src/locale/messages.ts`

- [ ] **Step 8.1: TagFilter 组件**

`web/src/components/tag/TagFilter.tsx`：

```tsx
import { useState } from 'react';

import { useI18n } from '../../locale/useI18n';
import { TagChip } from './TagChip';
import styles from './TagFilter.module.css';

interface TagFilterProps {
  active: string[]; // raw tokens, e.g. ['environment=prod', 'team=sre']
  onChange: (tokens: string[]) => void;
  environmentCounts?: Record<string, number>;
  totalCount?: number;
}

const ENV_VALUES = ['prod', 'dev', 'qa', 'orphan'] as const;

export function TagFilter({ active, onChange, environmentCounts, totalCount }: TagFilterProps) {
  const { t } = useI18n();
  const [input, setInput] = useState('');

  function isActiveEnv(value: string | null): boolean {
    const target = value === null ? '__none__:environment' : `environment=${value}`;
    return active.includes(target);
  }

  function toggleEnvChip(value: string | null) {
    const target = value === null ? '__none__:environment' : `environment=${value}`;
    // Single-select among environment chips: drop existing environment-related tokens, then add if not active
    const cleared = active.filter((t) => !t.startsWith('environment=') && !t.startsWith('__none__:environment'));
    if (active.includes(target)) {
      onChange(cleared); // toggle off
    } else {
      onChange([...cleared, target]);
    }
  }

  function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    const trimmed = input.trim();
    if (!trimmed) return;
    if (active.includes(trimmed)) {
      setInput('');
      return;
    }
    onChange([...active, trimmed]);
    setInput('');
  }

  function removeToken(token: string) {
    onChange(active.filter((t) => t !== token));
  }

  const customTokens = active.filter((t) => !t.startsWith('environment=') && !t.startsWith('__none__:environment'));

  return (
    <div className={styles.filter}>
      <div className={styles.chipRow}>
        <span className={styles.label}>{t('cloud.tags.environmentLabel')}：</span>
        <button
          type="button"
          className={[styles.chip, !ENV_VALUES.some((v) => isActiveEnv(v)) && !isActiveEnv(null) ? styles.chipActive : ''].join(' ')}
          onClick={() => onChange(customTokens)}
        >
          {t('cloud.tags.all')} {totalCount !== undefined ? <strong>{totalCount}</strong> : null}
        </button>
        {ENV_VALUES.map((value) => (
          <button
            key={value}
            type="button"
            className={[styles.chip, isActiveEnv(value) ? styles.chipActive : ''].join(' ')}
            onClick={() => toggleEnvChip(value)}
          >
            <TagChip tagKey="env" value={value} />
            {environmentCounts && environmentCounts[value] !== undefined ? <strong>{environmentCounts[value]}</strong> : null}
          </button>
        ))}
        <button
          type="button"
          className={[styles.chip, isActiveEnv(null) ? styles.chipActive : ''].join(' ')}
          onClick={() => toggleEnvChip(null)}
        >
          {t('cloud.tags.untagged')}
          {environmentCounts && environmentCounts.__none__ !== undefined ? <strong>{environmentCounts.__none__}</strong> : null}
        </button>
      </div>

      <form className={styles.inputRow} onSubmit={handleSubmit}>
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder={t('cloud.tags.filterPlaceholder')}
          aria-label={t('cloud.tags.filterPlaceholder')}
        />
        <button type="submit">{t('cloud.tags.addFilter')}</button>
        {customTokens.map((token) => (
          <button key={token} type="button" className={styles.activeToken} onClick={() => removeToken(token)}>
            {token} ×
          </button>
        ))}
      </form>
    </div>
  );
}
```

`web/src/components/tag/TagFilter.module.css`：

```css
.filter {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.chipRow {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.label {
  color: var(--shell-muted);
  font-size: 0.78rem;
  font-weight: 700;
  margin-right: 4px;
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 10px;
  border-radius: 8px;
  border: 1px solid var(--shell-border);
  background: rgba(255, 255, 255, 0.6);
  color: var(--shell-muted);
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
}

.chip strong {
  color: var(--text-primary);
}

.chipActive {
  background: rgba(47, 95, 89, 0.10);
  border-color: rgba(47, 95, 89, 0.30);
  color: var(--shell-control-active-bg);
}

.inputRow {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  align-items: center;
}

.inputRow input {
  flex: 1 1 14rem;
  min-height: 32px;
  padding: 0 10px;
  border: 1px solid var(--shell-border);
  border-radius: 6px;
  background: rgba(255, 255, 255, 0.74);
  font-family: ui-monospace, SFMono-Regular, monospace;
  font-size: 0.84rem;
}

.inputRow button {
  background: var(--shell-control-active-bg);
  color: var(--shell-control-active-fg);
  border: 0;
  border-radius: 6px;
  padding: 0 12px;
  min-height: 32px;
  font-weight: 700;
  cursor: pointer;
}

.activeToken {
  background: rgba(255, 255, 255, 0.78);
  border: 1px dashed var(--shell-border);
  border-radius: 6px;
  padding: 2px 9px;
  font-family: ui-monospace, SFMono-Regular, monospace;
  font-size: 0.78rem;
  color: var(--shell-muted);
  cursor: pointer;
}
```

- [ ] **Step 8.2: 集成到 CloudAccountsPage**

修改 `CloudAccountsPage.tsx`：
- 加 state `tagFilters: string[]`
- 把 `listCloudAccounts()` 改为 `listCloudAccounts(tagFilters)`
- `useEffect` 依赖 `tagFilters` 重新拉取
- 在 PageHeader 下、createBar 上方插入 `<TagFilter active={tagFilters} onChange={setTagFilters} totalCount={accounts.length} />`
  - environmentCounts 留空（Phase 2 计算，本任务先不接入）

具体改动节选：

```tsx
import { TagFilter } from '../../components/tag/TagFilter';
// ...

export function CloudAccountsPage() {
  // ...
  const [tagFilters, setTagFilters] = useState<string[]>([]);

  async function load() {
    try {
      const result = await listCloudAccounts(tagFilters);
      setAccounts(result.items);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.accounts.loadFailed'));
    }
  }

  useEffect(() => { void load(); }, [tagFilters]);

  // 在 header 后插入：
  <TagFilter active={tagFilters} onChange={setTagFilters} totalCount={accounts.length} />
}
```

- [ ] **Step 8.3: i18n keys 追加 cloud.tags**

EN：
```ts
environmentLabel: 'environment',
all: 'All',
filterPlaceholder: 'filter by tag, e.g. environment=prod team=sre',
addFilter: 'Add filter',
```

ZH：
```ts
environmentLabel: '环境',
all: '全部',
filterPlaceholder: '按 tag 过滤，如 environment=prod team=sre',
addFilter: '加入',
```

- [ ] **Step 8.4: 构建 + 提交**

```bash
cd /home/lilin/trae-project/Optimus/web && npm run build
```

```bash
cd /home/lilin/trae-project/Optimus
git add web/src/components/tag/TagFilter.tsx web/src/components/tag/TagFilter.module.css web/src/view/cloud/CloudAccountsPage.tsx web/src/locale/messages.ts
git commit -m "$(cat <<'EOF'
新增 TagFilter 组件，云账号页支持按 tag 过滤

Constraint: environment 4 个预设芯片单选互斥；自定义 tag 通过输入框追加
Confidence: high
Scope-risk: narrow
Tested: npm run build
EOF
)"
```

---

### Task 9: Cloud Assets 页面 + SideNav 资产入口激活

**Files:**
- Create: `web/src/view/cloud/CloudAssetsPage.tsx` + `.module.css`
- Modify: `web/src/router/index.tsx`
- Modify: `web/src/components/nav/SideNav.tsx`
- Modify: `web/src/locale/messages.ts`

- [ ] **Step 9.1: CloudAssetsPage 实现**

`web/src/view/cloud/CloudAssetsPage.tsx`：

```tsx
import { useEffect, useState } from 'react';

import { listCloudAccounts, listCloudAssets } from '../../api/cloud';
import { TagChip } from '../../components/tag/TagChip';
import { TagFilter } from '../../components/tag/TagFilter';
import { useI18n } from '../../locale/useI18n';
import type { CloudAccount, CloudAsset } from '../../types/cloud';
import styles from './CloudAssetsPage.module.css';

export function CloudAssetsPage() {
  const { t } = useI18n();
  const [assets, setAssets] = useState<CloudAsset[]>([]);
  const [total, setTotal] = useState(0);
  const [accountsById, setAccountsById] = useState<Record<number, CloudAccount>>({});
  const [tagFilters, setTagFilters] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);

  async function load() {
    try {
      const [assetsResult, accountsResult] = await Promise.all([
        listCloudAssets(tagFilters),
        listCloudAccounts(),
      ]);
      setAssets(assetsResult.items);
      setTotal(assetsResult.total);
      const indexed: Record<number, CloudAccount> = {};
      for (const acc of accountsResult.items) indexed[acc.id] = acc;
      setAccountsById(indexed);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('cloud.assets.loadFailed'));
    }
  }

  useEffect(() => { void load(); }, [tagFilters]);

  return (
    <section className={styles.page}>
      <header className={styles.header}>
        <div>
          <p className="page-eyebrow">{t('cloud.assets.eyebrow')}</p>
          <h1 className={styles.title}>{t('cloud.assets.title')}</h1>
          <p className={styles.sub}>{t('cloud.assets.sub')}</p>
        </div>
      </header>

      {error ? <div className={styles.error}>{error}</div> : null}

      <TagFilter active={tagFilters} onChange={setTagFilters} totalCount={total} />

      <div className={styles.list}>
        <div className={styles.tableHeader} aria-hidden="true">
          <span>{t('cloud.assets.name')}</span>
          <span>{t('cloud.assets.resourceId')}</span>
          <span>{t('cloud.assets.type')}</span>
          <span>{t('cloud.assets.account')}</span>
          <span>{t('cloud.assets.inheritedTags')}</span>
        </div>
        {assets.map((asset) => {
          const account = accountsById[asset.accountId];
          const accountTags = account?.tags ?? {};
          return (
            <article className={styles.row} key={asset.id}>
              <strong>{asset.name}</strong>
              <code>{asset.resourceId}</code>
              <span>{asset.resourceType}</span>
              <span>{account?.name ?? `#${asset.accountId}`}</span>
              <div className={styles.tags}>
                {Object.entries(accountTags).length === 0 ? (
                  <span className={styles.untagged}>{t('cloud.accounts.untagged')}</span>
                ) : (
                  Object.entries(accountTags).map(([k, v]) => <TagChip key={k} tagKey={k} value={v} />)
                )}
              </div>
            </article>
          );
        })}
        {assets.length === 0 ? <p className={styles.empty}>{t('cloud.assets.empty')}</p> : null}
      </div>
    </section>
  );
}
```

`web/src/view/cloud/CloudAssetsPage.module.css`：

```css
.page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 16px;
}

.title {
  margin: 0;
  color: var(--text-primary);
}

.sub {
  margin: 6px 0 0;
  color: var(--shell-muted);
  font-size: 0.92rem;
}

.error {
  padding: 12px 14px;
  border: 1px solid rgba(154, 52, 18, 0.22);
  border-radius: 8px;
  background: rgba(255, 247, 237, 0.82);
  color: #9a3412;
}

.list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.tableHeader,
.row {
  display: grid;
  grid-template-columns: minmax(8rem, 1fr) minmax(9rem, 1.2fr) minmax(7rem, 1fr) minmax(7rem, 1fr) minmax(12rem, 2fr);
  gap: 12px;
  align-items: center;
  padding: 10px 12px;
}

.tableHeader {
  color: var(--shell-muted);
  font-size: 0.72rem;
  font-weight: 800;
  text-transform: uppercase;
}

.row {
  border: 1px solid var(--shell-border);
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.56);
}

.row code {
  color: var(--shell-muted);
  font-family: ui-monospace, SFMono-Regular, monospace;
  font-size: 0.82rem;
}

.tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.untagged {
  padding: 2px 9px;
  border: 1px dashed var(--shell-border);
  border-radius: 6px;
  color: var(--shell-muted);
  font-style: italic;
  font-size: 0.78rem;
}

.empty {
  color: var(--shell-muted);
  text-align: center;
}
```

- [ ] **Step 9.2: 注册路由 + 激活 SideNav**

`web/src/router/index.tsx` 在 `cloud/accounts` 后追加：

```tsx
import { CloudAssetsPage } from '../view/cloud/CloudAssetsPage';
// ...
{
  path: 'cloud/assets',
  element: <CloudAssetsPage />,
},
```

`SideNav.tsx`：把 cloud assets 那条从 Coming soon 改成真链接 `to="/cloud/assets"`。

- [ ] **Step 9.3: i18n cloud.assets**

EN：
```ts
assets: {
  eyebrow: 'Cloud assets',
  title: 'Cloud assets',
  sub: 'Inherits tags from the owning cloud account',
  name: 'Name',
  resourceId: 'Resource ID',
  type: 'Type',
  account: 'Account',
  inheritedTags: 'Inherited tags',
  loadFailed: 'Failed to load assets',
  empty: 'No cloud assets yet',
},
```

ZH：
```ts
assets: {
  eyebrow: '云资产',
  title: '云资产',
  sub: '从所属云账号继承 tag',
  name: '名称',
  resourceId: '资源 ID',
  type: '类型',
  account: '账号',
  inheritedTags: '继承 tag',
  loadFailed: '加载资产失败',
  empty: '当前项目下暂无云资产',
},
```

放到 `cloud:` 段下、`accounts:` 之后。

- [ ] **Step 9.4: 构建 + 提交**

```bash
cd /home/lilin/trae-project/Optimus/web && npm run build
```

```bash
cd /home/lilin/trae-project/Optimus
git add web/src/view/cloud/CloudAssetsPage.tsx web/src/view/cloud/CloudAssetsPage.module.css web/src/router/index.tsx web/src/components/nav/SideNav.tsx web/src/locale/messages.ts
git commit -m "$(cat <<'EOF'
新增 Cloud Assets 页面（按 tag 过滤 + 账号继承显示）

Constraint: 资产展示从所属账号继承的 tag；过滤器与账号页同语法
Confidence: high
Scope-risk: narrow
Tested: npm run build
EOF
)"
```

---

## Self-Review Checklist

**Spec coverage：**
- ✅ `cloud_accounts.tags_json` 列（Task 1）
- ✅ Tag 校验 regex + 32 限制（Task 1 ValidateTags）
- ✅ AccountView.tags 暴露（Task 1）
- ✅ POST /accounts 接受 tags（Task 1+2+5）
- ✅ PUT /accounts/:id/tags 全量覆盖（Task 5）
- ✅ GET /accounts?tag=… 过滤（Task 3+5）
- ✅ `__none__:key` 语义（Task 1 ParseTagFilter + Task 3 tagFilterToSQL）
- ✅ environment 软约定前端配色 + 建议（Task 7 TagChip + TagEditor）
- ✅ 资产继承（Task 4+9）
- ✅ 前端三组件 TagChip / TagEditor / TagFilter（Task 7+8）
- ✅ 云账号页 + 云资产页（Task 6+9）
- ✅ SideNav 激活 cloud 入口（Task 6+9）

**Placeholder scan**：无 TBD/TODO；少量 "按既有模式对齐"（Task 5.2 parseAssetFilter 集成）是因为现有代码具体形式实施者读一下就能办，不构成 placeholder。

**Type consistency**：
- `AccountStore.Create` 在 Task 2 改签名为接收 AccountRow，Task 3 实现一致；Task 5 handler 调用一致。
- `TagFilter` 在 Task 1 定义；Task 3 SQL 映射、Task 5 handler 解析、前端 (Task 8) 自由格式 string 三处使用语义一致。
- `cloud.tags.*` i18n keys 在 Task 7/8 都引用，Task 7 添加基础组（empty/keyPlaceholder/…），Task 8 追加环境过滤组（environmentLabel/all/…），无重复。

---

## Out of Scope

- 资产级 user tag（资产仅继承）
- Tag 历史/审计
- Tag 模板（强制 environment 这类策略）
- 大资产量分页 UI（首版表格直接展示，分页交互留给后续）
- 移动端响应式优化
- 测试基础设施（保持 web/ 当前 0 测试现状；SP3 已留作 tech debt）

---

## Run-Through Walkthrough

实施完所有 9 个 task 后，端到端调用链：

1. 用户登录后切到 prod 项目，访问 `/cloud/accounts`
2. 前端 `listCloudAccounts([])` → GET `/api/v1/cloud/accounts` 带 `X-Project-Code: prod`
3. 后端 ProjectScope 中间件注入 accessibleProjectIDs，handler 调用 `service.List(ctx, accessible, nil)`
4. GormAccountStore 查 `WHERE project_id IN (?)`，返回带 tag 的 rows
5. 前端列表显示每账号的 `tags` chip，每条 `environment=prod` 享有红色配色
6. 用户点 "改 tag" → 展开 TagEditor → 编辑 → 点保存
7. 前端 PUT `/cloud/accounts/:id/tags` → 后端 ValidateTags → 写库 → 返回新 AccountView
8. 列表刷新，新 tag 即时显示
9. 用户在过滤栏点 "env=prod" 芯片 → 前端 tagFilters 改为 `['environment=prod']` → 重拉列表
10. 后端 SQL 加上 `JSON_UNQUOTE(JSON_EXTRACT(tags_json, '$.environment')) = 'prod'`
11. 用户切到 `/cloud/assets`，资产页通过 Promise.all 拉资产 + 全部账号，渲染时按 accountId 查表展示继承 tag

---

## 备注

本 plan 后端先行（Task 1-5），前端后接（Task 6-9）。在 Task 5 完成时后端 API 已可用；在 Task 6 完成时云账号页面已上线但无 tag UI；Task 7-8 完成后云账号页完整；Task 9 完成后云资产页上线。每个 task 都是一个独立 commit，便于回滚。
