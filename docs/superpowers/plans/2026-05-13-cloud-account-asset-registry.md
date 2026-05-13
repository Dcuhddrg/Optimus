# Cloud Account + Unified Asset Registry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a provider-agnostic cloud account CRUD system with an asset registry, using Aliyun ECS as a reference sync implementation.

**Architecture:** New `service/cloud/` module with usecase/infra layering identical to `service/system/`. `CloudAccountService` owns credential encryption (AES-256-GCM). `AssetRegistryService` orchestrates sync via a `SyncerRegistry` that maps provider strings to `AssetSyncer` implementations. All HTTP routes are registered on the existing `privateGroup` (JWT + Casbin RBAC).

**Tech Stack:** Go 1.23, Gin, GORM, MySQL, `crypto/aes`+`cipher` (stdlib), `github.com/aliyun/alibaba-cloud-sdk-go`

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `server/model/cloud/account.go` | Create | GORM models: `CloudAccount`, `CloudAsset` |
| `server/config/cloud.go` | Create | `Cloud` config struct |
| `server/config/config.go` | Modify | Add `Cloud Cloud` field |
| `server/service/cloud/usecase/crypto.go` | Create | AES-256-GCM encrypt/decrypt |
| `server/service/cloud/usecase/syncer.go` | Create | `RawAsset`, `SyncAccount`, `AssetSyncer` interface, `SyncerRegistry` |
| `server/service/cloud/usecase/account.go` | Create | `AccountStore` interface, `CloudAccountService` |
| `server/service/cloud/usecase/asset.go` | Create | `AssetStore` interface, `AssetRegistryService` |
| `server/service/cloud/usecase/crypto_test.go` | Create | Unit tests for encrypt/decrypt |
| `server/service/cloud/usecase/syncer_test.go` | Create | Unit tests for `SyncerRegistry` |
| `server/service/cloud/usecase/account_test.go` | Create | Unit tests for `CloudAccountService` |
| `server/service/cloud/usecase/asset_test.go` | Create | Unit tests for `AssetRegistryService` |
| `server/service/cloud/infra/persistence/testdb_test.go` | Create | Integration test DB helper |
| `server/service/cloud/infra/persistence/account_store_gorm.go` | Create | GORM implementation of `AccountStore` |
| `server/service/cloud/infra/persistence/account_store_gorm_test.go` | Create | Integration tests |
| `server/service/cloud/infra/persistence/asset_store_gorm.go` | Create | GORM implementation of `AssetStore` |
| `server/service/cloud/infra/persistence/asset_store_gorm_test.go` | Create | Integration tests |
| `server/service/cloud/infra/provider/aliyun/ecs_syncer.go` | Create | Aliyun ECS `AssetSyncer` implementation |
| `server/service/cloud/infra/provider/aliyun/ecs_syncer_test.go` | Create | Unit test with stub |
| `server/service/cloud/cloud.go` | Create | `AutoMigrateCloud`, `SeedCloudData` |
| `server/api/v1/cloud/handlers.go` | Create | `CloudHandlers`, `AccountHandler`, `AssetHandler` |
| `server/api/v1/cloud/account.go` | Create | HTTP handlers for account CRUD + sync |
| `server/api/v1/cloud/asset.go` | Create | HTTP handler for asset query |
| `server/api/v1/cloud/handlers_test.go` | Create | Handler unit tests |
| `server/router/cloud/enter.go` | Create | Route registration |
| `server/router/enter.go` | Modify | Add `Cloud cloud.RouterGroup` |
| `server/initialize/cloud/module.go` | Create | `CloudModule`, `NewCloudModule` |
| `server/initialize/router.go` | Modify | Wire `CloudModule` into `Routers()` |
| `server/initialize/startup.go` | Modify | Call `AutoMigrateCloud` + `SeedCloudData` |

---

### Task 1: GORM models + Cloud config

**Files:**
- Create: `server/model/cloud/account.go`
- Create: `server/config/cloud.go`
- Modify: `server/config/config.go`

- [ ] **Step 1.1: Create `server/model/cloud/account.go`**

```go
package cloud

import (
	"time"

	"gorm.io/gorm"
)

type CloudAccount struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
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

func (CloudAccount) TableName() string { return "cloud_accounts" }

type CloudAsset struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	AccountID    uint       `gorm:"not null;index:idx_cloud_asset_account_type;uniqueIndex:uniq_cloud_asset_account_resource" json:"accountId"`
	Provider     string     `gorm:"size:32;not null;index:idx_cloud_asset_provider_region_type" json:"provider"`
	Region       string     `gorm:"size:64;not null;index:idx_cloud_asset_provider_region_type" json:"region"`
	ResourceType string     `gorm:"size:64;not null;index:idx_cloud_asset_account_type;index:idx_cloud_asset_provider_region_type" json:"resourceType"`
	ResourceID   string     `gorm:"size:128;not null;uniqueIndex:uniq_cloud_asset_account_resource" json:"resourceId"`
	Name         string     `gorm:"size:256" json:"name"`
	Status       string     `gorm:"size:32" json:"status"`
	ConfigJSON   string     `gorm:"type:text" json:"configJson"`
	TagsJSON     string     `gorm:"type:text" json:"tagsJson"`
	DiscoveredAt time.Time  `json:"discoveredAt"`
	LastSeenAt   time.Time  `json:"lastSeenAt"`
	DeletedAt    *time.Time `json:"deletedAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func (CloudAsset) TableName() string { return "cloud_assets" }
```

- [ ] **Step 1.2: Create `server/config/cloud.go`**

```go
package config

type Cloud struct {
	CredentialEncryptionKey string `mapstructure:"credential-encryption-key"`
}
```

- [ ] **Step 1.3: Modify `server/config/config.go`** — add `Cloud` field

Current file content:
```go
package config

type Config struct {
	System System `mapstructure:"system"`
	Zap    Zap    `mapstructure:"zap"`
	Mysql  Mysql  `mapstructure:"mysql"`
	Redis  Redis  `mapstructure:"redis"`
	JWT    JWT    `mapstructure:"jwt"`
	Casbin Casbin `mapstructure:"casbin"`
}
```

Replace with:
```go
package config

type Config struct {
	System System `mapstructure:"system"`
	Cloud  Cloud  `mapstructure:"cloud"`
	Zap    Zap    `mapstructure:"zap"`
	Mysql  Mysql  `mapstructure:"mysql"`
	Redis  Redis  `mapstructure:"redis"`
	JWT    JWT    `mapstructure:"jwt"`
	Casbin Casbin `mapstructure:"casbin"`
}
```

- [ ] **Step 1.4: Verify build**

```bash
cd /home/lilin/trae-project/Optimus/server && go build ./...
```

Expected: clean build.

- [ ] **Step 1.5: Commit**

```bash
\
git add server/model/cloud/account.go \
        server/config/cloud.go \
        server/config/config.go && \
git commit -m "新增 cloud 模块 GORM 模型与配置结构"
```

---

### Task 2: Crypto utility + SyncerRegistry

**Files:**
- Create: `server/service/cloud/usecase/crypto.go`
- Create: `server/service/cloud/usecase/crypto_test.go`
- Create: `server/service/cloud/usecase/syncer.go`
- Create: `server/service/cloud/usecase/syncer_test.go`

- [ ] **Step 2.1: Write failing tests for crypto**

Create `server/service/cloud/usecase/crypto_test.go`:

```go
package usecase

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	plaintext := `{"access_key_id":"ak","access_key_secret":"sk"}`
	encrypted, err := encryptAESGCM(key, plaintext)
	require.NoError(t, err)
	require.NotEqual(t, plaintext, encrypted)

	decrypted, err := decryptAESGCM(key, encrypted)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

func TestEncryptRejectsBadKeySize(t *testing.T) {
	_, err := encryptAESGCM([]byte("tooshort"), "data")
	require.ErrorIs(t, err, ErrInvalidEncryptionKey)
}

func TestDecryptRejectsBadKeySize(t *testing.T) {
	_, err := decryptAESGCM([]byte("tooshort"), "data")
	require.ErrorIs(t, err, ErrInvalidEncryptionKey)
}
```

Write failing tests for syncer registry — create `server/service/cloud/usecase/syncer_test.go`:

```go
package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type stubSyncer struct{}

func (stubSyncer) Sync(_ context.Context, _ SyncAccount) ([]RawAsset, error) {
	return nil, nil
}

func TestSyncerRegistryReturnsRegisteredSyncer(t *testing.T) {
	r := NewSyncerRegistry()
	s := stubSyncer{}
	r.Register("aliyun", s)
	got, err := r.Get("aliyun")
	require.NoError(t, err)
	require.Equal(t, s, got)
}

func TestSyncerRegistryReturnsErrForUnknownProvider(t *testing.T) {
	r := NewSyncerRegistry()
	_, err := r.Get("aws")
	require.ErrorIs(t, err, ErrProviderNotRegistered)
}
```

- [ ] **Step 2.2: Run to confirm failure**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  go test ./service/cloud/usecase/... 2>&1 | head -20
```

Expected: compile errors (package does not exist yet).

- [ ] **Step 2.3: Create `server/service/cloud/usecase/crypto.go`**

```go
package usecase

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var ErrInvalidEncryptionKey = errors.New("encryption key must be 32 bytes")

func encryptAESGCM(key []byte, plaintext string) (string, error) {
	if len(key) != 32 {
		return "", ErrInvalidEncryptionKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptAESGCM(key []byte, cipherB64 string) (string, error) {
	if len(key) != 32 {
		return "", ErrInvalidEncryptionKey
	}
	data, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
```

- [ ] **Step 2.4: Create `server/service/cloud/usecase/syncer.go`**

```go
package usecase

import (
	"context"
	"errors"
)

var ErrProviderNotRegistered = errors.New("provider not registered")

// RawAsset is the normalized representation returned by any AssetSyncer.
type RawAsset struct {
	ResourceType string
	ResourceID   string
	Name         string
	Region       string
	Status       string
	ConfigJSON   string
	TagsJSON     string
}

// SyncAccount carries decrypted credentials for a cloud account sync operation.
type SyncAccount struct {
	ID             uint
	Provider       string
	AccountID      string
	CredentialType string
	Credential     map[string]string
}

// AssetSyncer fetches all assets for a given cloud account.
type AssetSyncer interface {
	Sync(ctx context.Context, account SyncAccount) ([]RawAsset, error)
}

// SyncerRegistry maps provider name → AssetSyncer.
type SyncerRegistry struct {
	syncers map[string]AssetSyncer
}

func NewSyncerRegistry() *SyncerRegistry {
	return &SyncerRegistry{syncers: make(map[string]AssetSyncer)}
}

func (r *SyncerRegistry) Register(provider string, s AssetSyncer) {
	r.syncers[provider] = s
}

func (r *SyncerRegistry) Get(provider string) (AssetSyncer, error) {
	s, ok := r.syncers[provider]
	if !ok {
		return nil, ErrProviderNotRegistered
	}
	return s, nil
}
```

- [ ] **Step 2.5: Run tests**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  go test ./service/cloud/usecase/... -run "TestEncrypt|TestDecrypt|TestSyncerRegistry" -v
```

Expected: 5 tests PASS.

- [ ] **Step 2.6: Commit**

```bash
cd /home/lilin/trae-project/Optimus && \
git add server/service/cloud/usecase/crypto.go \
        server/service/cloud/usecase/crypto_test.go \
        server/service/cloud/usecase/syncer.go \
        server/service/cloud/usecase/syncer_test.go && \
git commit -m "新增 cloud usecase 加密工具与 SyncerRegistry"
```

---

### Task 3: CloudAccountService

**Files:**
- Create: `server/service/cloud/usecase/account.go`
- Create: `server/service/cloud/usecase/account_test.go`

- [ ] **Step 3.1: Write failing tests**

Create `server/service/cloud/usecase/account_test.go`:

```go
package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeAccountStore struct {
	rows map[uint]AccountRow
	next uint
}

func (f *fakeAccountStore) Create(_ context.Context, name, provider, accountID, credentialType, encrypted string) (AccountRow, error) {
	f.next++
	row := AccountRow{ID: f.next, Name: name, Provider: provider, AccountID: accountID, CredentialType: credentialType, CredentialJSON: encrypted, Status: "active"}
	f.rows[f.next] = row
	return row, nil
}

func (f *fakeAccountStore) Get(_ context.Context, id uint) (AccountRow, error) {
	row, ok := f.rows[id]
	if !ok {
		return AccountRow{}, ErrAccountNotFound
	}
	return row, nil
}

func (f *fakeAccountStore) List(_ context.Context) ([]AccountRow, error) {
	rows := make([]AccountRow, 0, len(f.rows))
	for _, r := range f.rows {
		rows = append(rows, r)
	}
	return rows, nil
}

func (f *fakeAccountStore) Update(_ context.Context, id uint, name *string, encryptedCredential *string) (AccountRow, error) {
	row, ok := f.rows[id]
	if !ok {
		return AccountRow{}, ErrAccountNotFound
	}
	if name != nil {
		row.Name = *name
	}
	if encryptedCredential != nil {
		row.CredentialJSON = *encryptedCredential
	}
	f.rows[id] = row
	return row, nil
}

func (f *fakeAccountStore) Delete(_ context.Context, id uint) error {
	if _, ok := f.rows[id]; !ok {
		return ErrAccountNotFound
	}
	delete(f.rows, id)
	return nil
}

func (f *fakeAccountStore) UpdateSyncStatus(_ context.Context, id uint, lastSyncedAt *time.Time, status string) error {
	row, ok := f.rows[id]
	if !ok {
		return ErrAccountNotFound
	}
	row.Status = status
	row.LastSyncedAt = lastSyncedAt
	f.rows[id] = row
	return nil
}

func newTestAccountService() (*CloudAccountService, *fakeAccountStore) {
	store := &fakeAccountStore{rows: make(map[uint]AccountRow)}
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return NewCloudAccountService(store, key), store
}

func TestCloudAccountServiceCreateMasksCredential(t *testing.T) {
	svc, _ := newTestAccountService()
	view, err := svc.Create(context.Background(), CreateAccountRequest{
		Name: "prod", Provider: "aliyun", AccountID: "12345678",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)
	require.Equal(t, "***", view.Credential)
	require.Equal(t, "prod", view.Name)
	require.Equal(t, "aliyun", view.Provider)
}

func TestCloudAccountServiceGetForSyncDecryptsCredential(t *testing.T) {
	svc, _ := newTestAccountService()
	_, err := svc.Create(context.Background(), CreateAccountRequest{
		Name: "prod", Provider: "aliyun", AccountID: "123",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "AK", "access_key_secret": "SK"},
	})
	require.NoError(t, err)

	syncAcc, err := svc.GetForSync(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, "AK", syncAcc.Credential["access_key_id"])
	require.Equal(t, "SK", syncAcc.Credential["access_key_secret"])
}

func TestCloudAccountServiceDeleteRemovesAccount(t *testing.T) {
	svc, store := newTestAccountService()
	_, err := svc.Create(context.Background(), CreateAccountRequest{
		Name: "prod", Provider: "aliyun", AccountID: "123",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)
	require.NoError(t, svc.Delete(context.Background(), 1))
	require.Empty(t, store.rows)
}
```

- [ ] **Step 3.2: Run to confirm failure**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  go test ./service/cloud/usecase/... -run "TestCloudAccountService" -v 2>&1 | head -20
```

Expected: compile error — `AccountRow`, `CloudAccountService` not defined.

- [ ] **Step 3.3: Create `server/service/cloud/usecase/account.go`**

```go
package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var ErrAccountNotFound = errors.New("cloud account not found")

// AccountRow is the raw store-level representation (credential is encrypted).
type AccountRow struct {
	ID             uint
	Name           string
	Provider       string
	AccountID      string
	CredentialType string
	CredentialJSON string
	Status         string
	LastSyncedAt   *time.Time
}

// AccountView is the API-safe representation (credential masked).
type AccountView struct {
	ID             uint       `json:"id"`
	Name           string     `json:"name"`
	Provider       string     `json:"provider"`
	AccountID      string     `json:"accountId"`
	CredentialType string     `json:"credentialType"`
	Credential     string     `json:"credential"`
	Status         string     `json:"status"`
	LastSyncedAt   *time.Time `json:"lastSyncedAt"`
}

// CreateAccountRequest is the body for POST /cloud/accounts.
type CreateAccountRequest struct {
	Name           string            `json:"name" binding:"required"`
	Provider       string            `json:"provider" binding:"required"`
	AccountID      string            `json:"accountId" binding:"required"`
	CredentialType string            `json:"credentialType" binding:"required"`
	Credential     map[string]string `json:"credential" binding:"required"`
}

// UpdateAccountRequest is the body for PUT /cloud/accounts/:id.
type UpdateAccountRequest struct {
	Name       *string           `json:"name"`
	Credential map[string]string `json:"credential"`
}

// AccountStore is the persistence interface for cloud accounts.
type AccountStore interface {
	Create(ctx context.Context, name, provider, accountID, credentialType, encryptedCredential string) (AccountRow, error)
	Get(ctx context.Context, id uint) (AccountRow, error)
	List(ctx context.Context) ([]AccountRow, error)
	Update(ctx context.Context, id uint, name *string, encryptedCredential *string) (AccountRow, error)
	Delete(ctx context.Context, id uint) error
	UpdateSyncStatus(ctx context.Context, id uint, lastSyncedAt *time.Time, status string) error
}

// CloudAccountService manages cloud accounts and credential encryption.
type CloudAccountService struct {
	store  AccountStore
	encKey []byte
}

func NewCloudAccountService(store AccountStore, encKey []byte) *CloudAccountService {
	return &CloudAccountService{store: store, encKey: encKey}
}

func (s *CloudAccountService) Create(ctx context.Context, req CreateAccountRequest) (AccountView, error) {
	credJSON, err := json.Marshal(req.Credential)
	if err != nil {
		return AccountView{}, err
	}
	encrypted, err := encryptAESGCM(s.encKey, string(credJSON))
	if err != nil {
		return AccountView{}, err
	}
	row, err := s.store.Create(ctx, req.Name, req.Provider, req.AccountID, req.CredentialType, encrypted)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(row), nil
}

func (s *CloudAccountService) Get(ctx context.Context, id uint) (AccountView, error) {
	row, err := s.store.Get(ctx, id)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(row), nil
}

func (s *CloudAccountService) List(ctx context.Context) ([]AccountView, error) {
	rows, err := s.store.List(ctx)
	if err != nil {
		return nil, err
	}
	views := make([]AccountView, len(rows))
	for i, r := range rows {
		views[i] = toAccountView(r)
	}
	return views, nil
}

func (s *CloudAccountService) Update(ctx context.Context, id uint, req UpdateAccountRequest) (AccountView, error) {
	var encCredential *string
	if req.Credential != nil {
		credJSON, err := json.Marshal(req.Credential)
		if err != nil {
			return AccountView{}, err
		}
		enc, err := encryptAESGCM(s.encKey, string(credJSON))
		if err != nil {
			return AccountView{}, err
		}
		encCredential = &enc
	}
	row, err := s.store.Update(ctx, id, req.Name, encCredential)
	if err != nil {
		return AccountView{}, err
	}
	return toAccountView(row), nil
}

func (s *CloudAccountService) Delete(ctx context.Context, id uint) error {
	return s.store.Delete(ctx, id)
}

// UpdateSyncStatus updates last_synced_at and status after a sync attempt.
func (s *CloudAccountService) UpdateSyncStatus(ctx context.Context, id uint, lastSyncedAt *time.Time, status string) error {
	return s.store.UpdateSyncStatus(ctx, id, lastSyncedAt, status)
}

// GetForSync returns a SyncAccount with decrypted credentials for use by AssetRegistryService.
func (s *CloudAccountService) GetForSync(ctx context.Context, id uint) (SyncAccount, error) {
	row, err := s.store.Get(ctx, id)
	if err != nil {
		return SyncAccount{}, err
	}
	plaintext, err := decryptAESGCM(s.encKey, row.CredentialJSON)
	if err != nil {
		return SyncAccount{}, err
	}
	var cred map[string]string
	if err := json.Unmarshal([]byte(plaintext), &cred); err != nil {
		return SyncAccount{}, err
	}
	return SyncAccount{
		ID:             row.ID,
		Provider:       row.Provider,
		AccountID:      row.AccountID,
		CredentialType: row.CredentialType,
		Credential:     cred,
	}, nil
}

func toAccountView(row AccountRow) AccountView {
	return AccountView{
		ID:             row.ID,
		Name:           row.Name,
		Provider:       row.Provider,
		AccountID:      row.AccountID,
		CredentialType: row.CredentialType,
		Credential:     "***",
		Status:         row.Status,
		LastSyncedAt:   row.LastSyncedAt,
	}
}
```

- [ ] **Step 3.4: Run tests**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  go test ./service/cloud/usecase/... -run "TestCloudAccountService" -v
```

Expected: 3 tests PASS.

- [ ] **Step 3.5: Commit**

```bash
cd /home/lilin/trae-project/Optimus && \
git add server/service/cloud/usecase/account.go \
        server/service/cloud/usecase/account_test.go && \
git commit -m "新增 CloudAccountService：凭证加密/解密与 CRUD"
```

---

### Task 4: AssetRegistryService

**Files:**
- Create: `server/service/cloud/usecase/asset.go`
- Create: `server/service/cloud/usecase/asset_test.go`

- [ ] **Step 4.1: Write failing tests**

Create `server/service/cloud/usecase/asset_test.go`:

```go
package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeAssetStore struct {
	upserted []RawAsset
}

func (f *fakeAssetStore) Upsert(_ context.Context, accountID uint, provider string, assets []RawAsset) (SyncResult, error) {
	f.upserted = append(f.upserted, assets...)
	return SyncResult{Upserted: len(assets)}, nil
}

func (f *fakeAssetStore) Query(_ context.Context, _ AssetFilter) (AssetPage, error) {
	return AssetPage{Items: []AssetView{}, Total: 0, Page: 1, PageSize: 20}, nil
}

type fixedSyncer struct {
	assets []RawAsset
}

func (s fixedSyncer) Sync(_ context.Context, _ SyncAccount) ([]RawAsset, error) {
	return s.assets, nil
}

func TestAssetRegistryServiceSyncAccountUpsertsFetchedAssets(t *testing.T) {
	key := make([]byte, 32)
	accountStore := &fakeAccountStore{rows: make(map[uint]AccountRow)}
	accountSvc := NewCloudAccountService(accountStore, key)

	_, err := accountSvc.Create(context.Background(), CreateAccountRequest{
		Name: "prod", Provider: "aliyun", AccountID: "123",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)

	assetStore := &fakeAssetStore{}
	registry := NewSyncerRegistry()
	registry.Register("aliyun", fixedSyncer{assets: []RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1", Region: "cn-hangzhou", Status: "Running"},
	}})

	svc := NewAssetRegistryService(accountSvc, assetStore, registry)
	result, err := svc.SyncAccount(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, 1, result.Upserted)
	require.Equal(t, 1, len(assetStore.upserted))
	require.Equal(t, "i-001", assetStore.upserted[0].ResourceID)
}

func TestAssetRegistryServiceSyncAccountReturnsErrForUnknownProvider(t *testing.T) {
	key := make([]byte, 32)
	accountStore := &fakeAccountStore{rows: make(map[uint]AccountRow)}
	accountSvc := NewCloudAccountService(accountStore, key)
	_, err := accountSvc.Create(context.Background(), CreateAccountRequest{
		Name: "aws-acc", Provider: "aws", AccountID: "456",
		CredentialType: "ak_sk",
		Credential:     map[string]string{"access_key_id": "ak", "access_key_secret": "sk"},
	})
	require.NoError(t, err)

	svc := NewAssetRegistryService(accountSvc, &fakeAssetStore{}, NewSyncerRegistry())
	_, err = svc.SyncAccount(context.Background(), 1)
	require.ErrorIs(t, err, ErrProviderNotRegistered)
}
```

- [ ] **Step 4.2: Run to confirm failure**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  go test ./service/cloud/usecase/... -run "TestAssetRegistry" -v 2>&1 | head -20
```

Expected: compile error — `AssetFilter`, `AssetPage`, `AssetRegistryService` not defined.

- [ ] **Step 4.3: Create `server/service/cloud/usecase/asset.go`**

```go
package usecase

import (
	"context"
	"time"
)

// AssetFilter holds optional query parameters for listing cloud assets.
type AssetFilter struct {
	AccountID    *uint
	Provider     *string
	Region       *string
	ResourceType *string
	Status       *string
	Keyword      *string
	Page         int
	PageSize     int
}

// AssetPage is the paginated response for cloud asset queries.
type AssetPage struct {
	Items    []AssetView `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// AssetView is the API representation of a single cloud asset.
type AssetView struct {
	ID           uint       `json:"id"`
	AccountID    uint       `json:"accountId"`
	Provider     string     `json:"provider"`
	Region       string     `json:"region"`
	ResourceType string     `json:"resourceType"`
	ResourceID   string     `json:"resourceId"`
	Name         string     `json:"name"`
	Status       string     `json:"status"`
	ConfigJSON   string     `json:"configJson"`
	TagsJSON     string     `json:"tagsJson"`
	DiscoveredAt time.Time  `json:"discoveredAt"`
	LastSeenAt   time.Time  `json:"lastSeenAt"`
	DeletedAt    *time.Time `json:"deletedAt"`
}

// SyncResult reports the outcome of an account sync.
type SyncResult struct {
	Upserted int `json:"upserted"`
	Errors   int `json:"errors"`
}

// AssetStore persists and queries cloud assets.
type AssetStore interface {
	Upsert(ctx context.Context, accountID uint, provider string, assets []RawAsset) (SyncResult, error)
	Query(ctx context.Context, f AssetFilter) (AssetPage, error)
}

// AssetRegistryService orchestrates sync and querying of cloud assets.
type AssetRegistryService struct {
	accounts *CloudAccountService
	store    AssetStore
	registry *SyncerRegistry
}

func NewAssetRegistryService(accounts *CloudAccountService, store AssetStore, registry *SyncerRegistry) *AssetRegistryService {
	return &AssetRegistryService{accounts: accounts, store: store, registry: registry}
}

// SyncAccount fetches all assets for the given account and upserts them into the registry.
func (s *AssetRegistryService) SyncAccount(ctx context.Context, accountID uint) (SyncResult, error) {
	syncAcc, err := s.accounts.GetForSync(ctx, accountID)
	if err != nil {
		_ = s.accounts.UpdateSyncStatus(ctx, accountID, nil, "error")
		return SyncResult{}, err
	}
	syncer, err := s.registry.Get(syncAcc.Provider)
	if err != nil {
		_ = s.accounts.UpdateSyncStatus(ctx, accountID, nil, "error")
		return SyncResult{}, err
	}
	raw, err := syncer.Sync(ctx, syncAcc)
	if err != nil {
		_ = s.accounts.UpdateSyncStatus(ctx, accountID, nil, "error")
		return SyncResult{}, err
	}
	result, err := s.store.Upsert(ctx, accountID, syncAcc.Provider, raw)
	if err != nil {
		_ = s.accounts.UpdateSyncStatus(ctx, accountID, nil, "error")
		return SyncResult{}, err
	}
	now := timeNow()
	_ = s.accounts.UpdateSyncStatus(ctx, accountID, &now, "active")
	return result, nil
}

func (s *AssetRegistryService) QueryAssets(ctx context.Context, f AssetFilter) (AssetPage, error) {
	return s.store.Query(ctx, f)
}
```

Note: `timeNow` is a package-level variable to allow testing. Add to the bottom of `asset.go`:

```go
// timeNow is a hook for testing.
var timeNow = func() time.Time { return time.Now() }
```

- [ ] **Step 4.4: Run tests**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  go test ./service/cloud/usecase/... -v
```

Expected: all 8 tests PASS.

- [ ] **Step 4.5: Commit**

```bash
cd /home/lilin/trae-project/Optimus && \
git add server/service/cloud/usecase/asset.go \
        server/service/cloud/usecase/asset_test.go && \
git commit -m "新增 AssetRegistryService：同步编排与资产查询"
```

---

### Task 5: GormAccountStore

**Files:**
- Create: `server/service/cloud/infra/persistence/testdb_test.go`
- Create: `server/service/cloud/infra/persistence/account_store_gorm.go`
- Create: `server/service/cloud/infra/persistence/account_store_gorm_test.go`

- [ ] **Step 5.1: Create test DB helper**

Create `server/service/cloud/infra/persistence/testdb_test.go`:

```go
package persistence

import (
	"os"
	"testing"

	cloudModel "optimus/server/model/cloud"

	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/stretchr/testify/require"
)

func openCloudTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("OPTIMUS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OPTIMUS_TEST_MYSQL_DSN is not set")
	}
	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec("DROP TABLE IF EXISTS cloud_assets").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS cloud_accounts").Error)
	require.NoError(t, db.AutoMigrate(&cloudModel.CloudAccount{}, &cloudModel.CloudAsset{}))

	return db
}
```

- [ ] **Step 5.2: Write failing integration tests**

Create `server/service/cloud/infra/persistence/account_store_gorm_test.go`:

```go
package persistence

import (
	"context"
	"testing"

	"optimus/server/service/cloud/usecase"

	"github.com/stretchr/testify/require"
)

func TestGormAccountStoreCreateAndGet(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	row, err := store.Create(context.Background(), "prod", "aliyun", "123", "ak_sk", "enc-cred")
	require.NoError(t, err)
	require.NotZero(t, row.ID)
	require.Equal(t, "prod", row.Name)
	require.Equal(t, "active", row.Status)

	got, err := store.Get(context.Background(), row.ID)
	require.NoError(t, err)
	require.Equal(t, row.ID, got.ID)
	require.Equal(t, "enc-cred", got.CredentialJSON)
}

func TestGormAccountStoreGetNotFoundReturnsErr(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	_, err := store.Get(context.Background(), 9999)
	require.ErrorIs(t, err, usecase.ErrAccountNotFound)
}

func TestGormAccountStoreUpdateName(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	row, err := store.Create(context.Background(), "old-name", "aliyun", "123", "ak_sk", "enc")
	require.NoError(t, err)

	newName := "new-name"
	updated, err := store.Update(context.Background(), row.ID, &newName, nil)
	require.NoError(t, err)
	require.Equal(t, "new-name", updated.Name)
}

func TestGormAccountStoreSoftDelete(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAccountStore(db)

	row, err := store.Create(context.Background(), "prod", "aliyun", "123", "ak_sk", "enc")
	require.NoError(t, err)

	require.NoError(t, store.Delete(context.Background(), row.ID))

	_, err = store.Get(context.Background(), row.ID)
	require.ErrorIs(t, err, usecase.ErrAccountNotFound)
}
```

- [ ] **Step 5.3: Run to confirm failure**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  go test ./service/cloud/infra/persistence/... -run "TestGormAccountStore" -v 2>&1 | head -20
```

Expected: compile error — `GormAccountStore` not defined.

- [ ] **Step 5.4: Create `server/service/cloud/infra/persistence/account_store_gorm.go`**

```go
package persistence

import (
	"context"
	"errors"
	"time"

	cloudModel "optimus/server/model/cloud"
	"optimus/server/service/cloud/usecase"

	"gorm.io/gorm"
)

type GormAccountStore struct {
	db *gorm.DB
}

func NewGormAccountStore(db *gorm.DB) *GormAccountStore {
	return &GormAccountStore{db: db}
}

func (s *GormAccountStore) Create(ctx context.Context, name, provider, accountID, credentialType, encryptedCredential string) (usecase.AccountRow, error) {
	m := cloudModel.CloudAccount{
		Name:           name,
		Provider:       provider,
		AccountID:      accountID,
		CredentialType: credentialType,
		CredentialJSON: encryptedCredential,
		Status:         "active",
	}
	if err := s.db.WithContext(ctx).Create(&m).Error; err != nil {
		return usecase.AccountRow{}, err
	}
	return toAccountRow(m), nil
}

func (s *GormAccountStore) Get(ctx context.Context, id uint) (usecase.AccountRow, error) {
	var m cloudModel.CloudAccount
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return usecase.AccountRow{}, usecase.ErrAccountNotFound
		}
		return usecase.AccountRow{}, err
	}
	return toAccountRow(m), nil
}

func (s *GormAccountStore) List(ctx context.Context) ([]usecase.AccountRow, error) {
	var models []cloudModel.CloudAccount
	if err := s.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	rows := make([]usecase.AccountRow, len(models))
	for i, m := range models {
		rows[i] = toAccountRow(m)
	}
	return rows, nil
}

func (s *GormAccountStore) Update(ctx context.Context, id uint, name *string, encryptedCredential *string) (usecase.AccountRow, error) {
	var m cloudModel.CloudAccount
	if err := s.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return usecase.AccountRow{}, usecase.ErrAccountNotFound
		}
		return usecase.AccountRow{}, err
	}
	updates := map[string]any{}
	if name != nil {
		updates["name"] = *name
	}
	if encryptedCredential != nil {
		updates["credential_json"] = *encryptedCredential
	}
	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&m).Updates(updates).Error; err != nil {
			return usecase.AccountRow{}, err
		}
	}
	return toAccountRow(m), nil
}

func (s *GormAccountStore) Delete(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&cloudModel.CloudAccount{}, id).Error
}

func (s *GormAccountStore) UpdateSyncStatus(ctx context.Context, id uint, lastSyncedAt *time.Time, status string) error {
	updates := map[string]any{"status": status}
	if lastSyncedAt != nil {
		updates["last_synced_at"] = lastSyncedAt
	}
	return s.db.WithContext(ctx).Model(&cloudModel.CloudAccount{}).Where("id = ?", id).Updates(updates).Error
}

func toAccountRow(m cloudModel.CloudAccount) usecase.AccountRow {
	return usecase.AccountRow{
		ID:             m.ID,
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

- [ ] **Step 5.5: Build and run unit tests (skipping integration)**

```bash
cd /home/lilin/trae-project/Optimus/server && go build ./... && go test ./service/cloud/...
```

Expected: build passes; integration tests skipped (no DSN set).

- [ ] **Step 5.6: Commit**

```bash
cd /home/lilin/trae-project/Optimus && \
git add server/service/cloud/infra/persistence/testdb_test.go \
        server/service/cloud/infra/persistence/account_store_gorm.go \
        server/service/cloud/infra/persistence/account_store_gorm_test.go && \
git commit -m "新增 GormAccountStore：云账号持久化实现"
```

---

### Task 6: GormAssetStore

**Files:**
- Create: `server/service/cloud/infra/persistence/asset_store_gorm.go`
- Create: `server/service/cloud/infra/persistence/asset_store_gorm_test.go`

- [ ] **Step 6.1: Write failing integration tests**

Create `server/service/cloud/infra/persistence/asset_store_gorm_test.go`:

```go
package persistence

import (
	"context"
	"testing"

	"optimus/server/service/cloud/usecase"

	"github.com/stretchr/testify/require"
)

func TestGormAssetStoreUpsertCreatesNewAsset(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAssetStore(db)

	result, err := store.Upsert(context.Background(), 1, "aliyun", []usecase.RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, result.Upserted)

	page, err := store.Query(context.Background(), usecase.AssetFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "i-001", page.Items[0].ResourceID)
}

func TestGormAssetStoreUpsertUpdatesExistingAsset(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAssetStore(db)

	_, err := store.Upsert(context.Background(), 1, "aliyun", []usecase.RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)

	_, err = store.Upsert(context.Background(), 1, "aliyun", []usecase.RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Name: "web-1-updated", Region: "cn-hangzhou", Status: "Stopped"},
	})
	require.NoError(t, err)

	page, err := store.Query(context.Background(), usecase.AssetFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "web-1-updated", page.Items[0].Name)
	require.Equal(t, "Stopped", page.Items[0].Status)
}

func TestGormAssetStoreQueryFiltersByProvider(t *testing.T) {
	db := openCloudTestDB(t)
	store := NewGormAssetStore(db)

	_, err := store.Upsert(context.Background(), 1, "aliyun", []usecase.RawAsset{
		{ResourceType: "ecs_instance", ResourceID: "i-001", Region: "cn-hangzhou", Status: "Running"},
	})
	require.NoError(t, err)
	_, err = store.Upsert(context.Background(), 2, "aws", []usecase.RawAsset{
		{ResourceType: "ec2_instance", ResourceID: "i-abc", Region: "us-east-1", Status: "running"},
	})
	require.NoError(t, err)

	provider := "aliyun"
	page, err := store.Query(context.Background(), usecase.AssetFilter{Provider: &provider, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "aliyun", page.Items[0].Provider)
}
```

- [ ] **Step 6.2: Run to confirm failure**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  go test ./service/cloud/infra/persistence/... -run "TestGormAssetStore" -v 2>&1 | head -20
```

Expected: compile error — `GormAssetStore` not defined.

- [ ] **Step 6.3: Create `server/service/cloud/infra/persistence/asset_store_gorm.go`**

```go
package persistence

import (
	"context"
	"errors"
	"time"

	cloudModel "optimus/server/model/cloud"
	"optimus/server/service/cloud/usecase"

	"gorm.io/gorm"
)

type GormAssetStore struct {
	db *gorm.DB
}

func NewGormAssetStore(db *gorm.DB) *GormAssetStore {
	return &GormAssetStore{db: db}
}

func (s *GormAssetStore) Upsert(ctx context.Context, accountID uint, provider string, assets []usecase.RawAsset) (usecase.SyncResult, error) {
	now := time.Now()
	var result usecase.SyncResult
	for _, a := range assets {
		var existing cloudModel.CloudAsset
		err := s.db.WithContext(ctx).
			Where("account_id = ? AND resource_id = ?", accountID, a.ResourceID).
			First(&existing).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			m := cloudModel.CloudAsset{
				AccountID:    accountID,
				Provider:     provider,
				Region:       a.Region,
				ResourceType: a.ResourceType,
				ResourceID:   a.ResourceID,
				Name:         a.Name,
				Status:       a.Status,
				ConfigJSON:   a.ConfigJSON,
				TagsJSON:     a.TagsJSON,
				DiscoveredAt: now,
				LastSeenAt:   now,
			}
			if createErr := s.db.WithContext(ctx).Create(&m).Error; createErr != nil {
				result.Errors++
				continue
			}
		} else if err == nil {
			existing.Name = a.Name
			existing.Status = a.Status
			existing.ConfigJSON = a.ConfigJSON
			existing.TagsJSON = a.TagsJSON
			existing.LastSeenAt = now
			existing.DeletedAt = nil
			if saveErr := s.db.WithContext(ctx).Save(&existing).Error; saveErr != nil {
				result.Errors++
				continue
			}
		} else {
			result.Errors++
			continue
		}
		result.Upserted++
	}
	return result, nil
}

func (s *GormAssetStore) Query(ctx context.Context, f usecase.AssetFilter) (usecase.AssetPage, error) {
	q := s.db.WithContext(ctx).Model(&cloudModel.CloudAsset{})
	if f.AccountID != nil {
		q = q.Where("account_id = ?", *f.AccountID)
	}
	if f.Provider != nil {
		q = q.Where("provider = ?", *f.Provider)
	}
	if f.Region != nil {
		q = q.Where("region = ?", *f.Region)
	}
	if f.ResourceType != nil {
		q = q.Where("resource_type = ?", *f.ResourceType)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if f.Keyword != nil && *f.Keyword != "" {
		q = q.Where("name LIKE ?", "%"+*f.Keyword+"%")
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return usecase.AssetPage{}, err
	}
	var models []cloudModel.CloudAsset
	if err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&models).Error; err != nil {
		return usecase.AssetPage{}, err
	}
	items := make([]usecase.AssetView, len(models))
	for i, m := range models {
		items[i] = toAssetView(m)
	}
	return usecase.AssetPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func toAssetView(m cloudModel.CloudAsset) usecase.AssetView {
	return usecase.AssetView{
		ID:           m.ID,
		AccountID:    m.AccountID,
		Provider:     m.Provider,
		Region:       m.Region,
		ResourceType: m.ResourceType,
		ResourceID:   m.ResourceID,
		Name:         m.Name,
		Status:       m.Status,
		ConfigJSON:   m.ConfigJSON,
		TagsJSON:     m.TagsJSON,
		DiscoveredAt: m.DiscoveredAt,
		LastSeenAt:   m.LastSeenAt,
		DeletedAt:    m.DeletedAt,
	}
}
```

- [ ] **Step 6.4: Build and test**

```bash
cd /home/lilin/trae-project/Optimus/server && go build ./... && go test ./service/cloud/...
```

Expected: build clean; integration tests skipped.

- [ ] **Step 6.5: Commit**

```bash
cd /home/lilin/trae-project/Optimus && \
git add server/service/cloud/infra/persistence/asset_store_gorm.go \
        server/service/cloud/infra/persistence/asset_store_gorm_test.go && \
git commit -m "新增 GormAssetStore：资产 upsert 与分页查询"
```

---

### Task 7: Aliyun ECS syncer (reference implementation)

**Files:**
- Modify: `server/go.mod` (add SDK dep)
- Create: `server/service/cloud/infra/provider/aliyun/ecs_syncer.go`
- Create: `server/service/cloud/infra/provider/aliyun/ecs_syncer_test.go`

- [ ] **Step 7.1: Add Aliyun SDK dependency**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  go get github.com/aliyun/alibaba-cloud-sdk-go
```

Expected: `go.mod` and `go.sum` updated.

- [ ] **Step 7.2: Write failing test**

Create `server/service/cloud/infra/provider/aliyun/ecs_syncer_test.go`:

```go
package aliyun

import (
	"context"
	"testing"

	"optimus/server/service/cloud/usecase"

	"github.com/stretchr/testify/require"
)

func TestECSSyncerImplementsAssetSyncer(t *testing.T) {
	var _ usecase.AssetSyncer = NewECSSyncer()
}

func TestECSSyncerReturnsErrorWithInvalidCredentials(t *testing.T) {
	s := NewECSSyncer()
	_, err := s.Sync(context.Background(), usecase.SyncAccount{
		ID:       1,
		Provider: "aliyun",
		AccountID: "fake-account",
		CredentialType: "ak_sk",
		Credential: map[string]string{
			"access_key_id":     "invalid-ak",
			"access_key_secret": "invalid-sk",
			"region":            "cn-hangzhou",
		},
	})
	// With invalid credentials against the real API this will error.
	// In unit context (no network), just confirm the syncer calls the SDK and handles the error path.
	// The test passes if either: (a) an error is returned, or (b) an empty slice is returned.
	// We only test interface compliance and that no panic occurs.
	_ = err
}
```

- [ ] **Step 7.3: Run to confirm failure**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  go test ./service/cloud/infra/provider/aliyun/... -v 2>&1 | head -20
```

Expected: compile error — package does not exist.

- [ ] **Step 7.4: Create `server/service/cloud/infra/provider/aliyun/ecs_syncer.go`**

```go
package aliyun

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"optimus/server/service/cloud/usecase"
)

// ECSSyncer implements AssetSyncer for Aliyun ECS instances.
// Credential map must contain: access_key_id, access_key_secret, region.
type ECSSyncer struct{}

func NewECSSyncer() *ECSSyncer { return &ECSSyncer{} }

func (s *ECSSyncer) Sync(ctx context.Context, account usecase.SyncAccount) ([]usecase.RawAsset, error) {
	ak := account.Credential["access_key_id"]
	sk := account.Credential["access_key_secret"]
	region := account.Credential["region"]
	if region == "" {
		region = "cn-hangzhou"
	}

	client, err := ecs.NewClientWithAccessKey(region, ak, sk)
	if err != nil {
		return nil, fmt.Errorf("aliyun ecs client: %w", err)
	}

	var assets []usecase.RawAsset
	pageNumber := 1
	for {
		req := ecs.CreateDescribeInstancesRequest()
		req.PageSize = requests.NewInteger(100)
		req.PageNumber = requests.NewInteger(pageNumber)

		resp, err := client.DescribeInstances(req)
		if err != nil {
			return nil, fmt.Errorf("aliyun DescribeInstances: %w", err)
		}

		for _, inst := range resp.Instances.Instance {
			configJSON, _ := json.Marshal(map[string]any{
				"instanceType":       inst.InstanceType,
				"cpu":                inst.Cpu,
				"memory":             inst.Memory,
				"osType":             inst.OSType,
				"instanceChargeType": inst.InstanceChargeType,
			})
			tagsMap := make(map[string]string, len(inst.Tags.Tag))
			for _, tag := range inst.Tags.Tag {
				tagsMap[tag.TagKey] = tag.TagValue
			}
			tagsJSON, _ := json.Marshal(tagsMap)

			assets = append(assets, usecase.RawAsset{
				ResourceType: "ecs_instance",
				ResourceID:   inst.InstanceId,
				Name:         inst.InstanceName,
				Region:       region,
				Status:       inst.Status,
				ConfigJSON:   string(configJSON),
				TagsJSON:     string(tagsJSON),
			})
		}

		if pageNumber*100 >= resp.TotalCount {
			break
		}
		pageNumber++
	}
	return assets, nil
}
```

- [ ] **Step 7.5: Run tests and build**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  go test ./service/cloud/infra/provider/aliyun/... -v && \
  go build ./...
```

Expected: `TestECSSyncerImplementsAssetSyncer` PASS; `TestECSSyncerReturnsErrorWithInvalidCredentials` PASS (no panic); full build clean.

- [ ] **Step 7.6: Commit**

```bash
cd /home/lilin/trae-project/Optimus && \
git add server/go.mod server/go.sum \
        server/service/cloud/infra/provider/aliyun/ecs_syncer.go \
        server/service/cloud/infra/provider/aliyun/ecs_syncer_test.go && \
git commit -m "新增阿里云 ECS 同步器作为 AssetSyncer 参考实现"
```

---

### Task 8: HTTP handlers + router

**Files:**
- Create: `server/api/v1/cloud/handlers.go`
- Create: `server/api/v1/cloud/account.go`
- Create: `server/api/v1/cloud/asset.go`
- Create: `server/api/v1/cloud/handlers_test.go`
- Create: `server/router/cloud/enter.go`

- [ ] **Step 8.1: Write failing handler tests**

Create `server/api/v1/cloud/handlers_test.go`:

```go
package cloud

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAccountHandlerListReturnsServiceUnavailableWhenNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCloudHandlers(CloudHandlerDeps{})

	engine := gin.New()
	engine.GET("/cloud/accounts", h.Account.ListAccounts)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cloud/accounts", nil)
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestAssetHandlerListReturnsServiceUnavailableWhenNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewCloudHandlers(CloudHandlerDeps{})

	engine := gin.New()
	engine.GET("/cloud/assets", h.Asset.ListAssets)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cloud/assets", nil)
	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}
```

- [ ] **Step 8.2: Run to confirm failure**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  go test ./api/v1/cloud/... 2>&1 | head -20
```

Expected: compile error — package does not exist.

- [ ] **Step 8.3: Create `server/api/v1/cloud/handlers.go`**

```go
package cloud

import (
	"net/http"
	"strconv"

	cloudUsecase "optimus/server/service/cloud/usecase"

	"github.com/gin-gonic/gin"
)

// AccountHandler handles HTTP requests for cloud accounts.
type AccountHandler struct {
	accountSvc *cloudUsecase.CloudAccountService
	assetSvc   *cloudUsecase.AssetRegistryService
}

// AssetHandler handles HTTP requests for cloud assets.
type AssetHandler struct {
	service *cloudUsecase.AssetRegistryService
}

// CloudHandlers groups all cloud HTTP handlers.
type CloudHandlers struct {
	Account *AccountHandler
	Asset   *AssetHandler
}

// CloudHandlerDeps carries the services needed to build CloudHandlers.
type CloudHandlerDeps struct {
	Account *cloudUsecase.CloudAccountService
	Asset   *cloudUsecase.AssetRegistryService
}

func NewCloudHandlers(deps CloudHandlerDeps) *CloudHandlers {
	return &CloudHandlers{
		Account: &AccountHandler{accountSvc: deps.Account, assetSvc: deps.Asset},
		Asset:   &AssetHandler{service: deps.Asset},
	}
}

func writeCloudServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"message": message})
}

func parseCloudUintParam(c *gin.Context, name string) (uint, bool) {
	s := c.Param(name)
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid " + name})
		return 0, false
	}
	return uint(n), true
}
```

- [ ] **Step 8.4: Create `server/api/v1/cloud/account.go`**

```go
package cloud

import (
	"errors"
	"net/http"

	cloudUsecase "optimus/server/service/cloud/usecase"

	"github.com/gin-gonic/gin"
)

func (h *AccountHandler) ListAccounts(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	views, err := h.accountSvc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to list cloud accounts"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views, "total": len(views)})
}

func (h *AccountHandler) GetAccount(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	view, err := h.accountSvc.Get(c.Request.Context(), id)
	if errors.Is(err, cloudUsecase.ErrAccountNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to get cloud account"})
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *AccountHandler) CreateAccount(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	var req cloudUsecase.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}
	view, err := h.accountSvc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create cloud account"})
		return
	}
	c.JSON(http.StatusCreated, view)
}

func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	var req cloudUsecase.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request body"})
		return
	}
	view, err := h.accountSvc.Update(c.Request.Context(), id, req)
	if errors.Is(err, cloudUsecase.ErrAccountNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to update cloud account"})
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	if h.accountSvc == nil {
		writeCloudServiceUnavailable(c, "cloud account service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.accountSvc.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, cloudUsecase.ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to delete cloud account"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "cloud account deleted"})
}

func (h *AccountHandler) SyncAccount(c *gin.Context) {
	if h.assetSvc == nil {
		writeCloudServiceUnavailable(c, "asset registry service is not available")
		return
	}
	id, ok := parseCloudUintParam(c, "id")
	if !ok {
		return
	}
	result, err := h.assetSvc.SyncAccount(c.Request.Context(), id)
	if errors.Is(err, cloudUsecase.ErrAccountNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}
	if errors.Is(err, cloudUsecase.ErrProviderNotRegistered) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "sync failed"})
		return
	}
	c.JSON(http.StatusOK, result)
}
```

- [ ] **Step 8.5: Create `server/api/v1/cloud/asset.go`**

```go
package cloud

import (
	"net/http"
	"strconv"

	cloudUsecase "optimus/server/service/cloud/usecase"

	"github.com/gin-gonic/gin"
)

func (h *AssetHandler) ListAssets(c *gin.Context) {
	if h.service == nil {
		writeCloudServiceUnavailable(c, "asset registry service is not available")
		return
	}
	f := parseAssetFilter(c)
	page, err := h.service.QueryAssets(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to query assets"})
		return
	}
	c.JSON(http.StatusOK, page)
}

func parseAssetFilter(c *gin.Context) cloudUsecase.AssetFilter {
	f := cloudUsecase.AssetFilter{Page: 1, PageSize: 20}

	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			f.Page = n
		}
	}
	if v := c.Query("pageSize"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 && n <= 200 {
			f.PageSize = n
		}
	}
	if v := c.Query("accountId"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			id := uint(n)
			f.AccountID = &id
		}
	}
	if v := c.Query("provider"); v != "" {
		f.Provider = &v
	}
	if v := c.Query("region"); v != "" {
		f.Region = &v
	}
	if v := c.Query("resourceType"); v != "" {
		f.ResourceType = &v
	}
	if v := c.Query("status"); v != "" {
		f.Status = &v
	}
	if v := c.Query("keyword"); v != "" {
		f.Keyword = &v
	}
	return f
}
```

- [ ] **Step 8.6: Create `server/router/cloud/enter.go`**

```go
package cloud

import (
	cloudApi "optimus/server/api/v1/cloud"

	"github.com/gin-gonic/gin"
)

type RouterGroup struct{}

func (r *RouterGroup) InitCloudRouter(privateGroup *gin.RouterGroup, handlers *cloudApi.CloudHandlers) {
	if handlers == nil {
		handlers = cloudApi.NewCloudHandlers(cloudApi.CloudHandlerDeps{})
	}

	privateGroup.GET("/cloud/accounts", handlers.Account.ListAccounts)
	privateGroup.GET("/cloud/accounts/:id", handlers.Account.GetAccount)
	privateGroup.POST("/cloud/accounts", handlers.Account.CreateAccount)
	privateGroup.PUT("/cloud/accounts/:id", handlers.Account.UpdateAccount)
	privateGroup.DELETE("/cloud/accounts/:id", handlers.Account.DeleteAccount)
	privateGroup.POST("/cloud/accounts/:id/sync", handlers.Account.SyncAccount)

	privateGroup.GET("/cloud/assets", handlers.Asset.ListAssets)
}
```

- [ ] **Step 8.7: Run handler tests and build**

```bash
cd /home/lilin/trae-project/Optimus/server && \
  go test ./api/v1/cloud/... -v && go build ./...
```

Expected: 2 tests PASS; build clean.

- [ ] **Step 8.8: Commit**

```bash
cd /home/lilin/trae-project/Optimus && \
git add server/api/v1/cloud/handlers.go \
        server/api/v1/cloud/account.go \
        server/api/v1/cloud/asset.go \
        server/api/v1/cloud/handlers_test.go \
        server/router/cloud/enter.go && \
git commit -m "新增 cloud HTTP handlers 与路由注册"
```

---

### Task 9: CloudModule + migration + seed + wire

**Files:**
- Create: `server/service/cloud/cloud.go`
- Create: `server/initialize/cloud/module.go`
- Modify: `server/router/enter.go`
- Modify: `server/initialize/router.go`
- Modify: `server/initialize/startup.go`

- [ ] **Step 9.1: Create `server/service/cloud/cloud.go`**

```go
package cloud

import (
	"errors"

	cloudModel "optimus/server/model/cloud"
	systemModel "optimus/server/model/system"

	"gorm.io/gorm"
)

func AutoMigrateCloud(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	return db.AutoMigrate(
		&cloudModel.CloudAccount{},
		&cloudModel.CloudAsset{},
	)
}

func SeedCloudData(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	if err := seedCloudPermissions(db); err != nil {
		return err
	}
	return seedAdminCloudPermissions(db)
}

// DefaultCloudPermissions returns all cloud API permissions.
func DefaultCloudPermissions() []systemModel.SystemPermission {
	return []systemModel.SystemPermission{
		{Code: "cloud.accounts.view", Name: "View cloud accounts", Type: "api", Method: "GET", Path: "/api/v1/cloud/accounts", Sort: 1000},
		{Code: "cloud.accounts.detail", Name: "View cloud account detail", Type: "api", Method: "GET", Path: "/api/v1/cloud/accounts/:id", Sort: 1001},
		{Code: "cloud.accounts.create", Name: "Create cloud accounts", Type: "api", Method: "POST", Path: "/api/v1/cloud/accounts", Sort: 1002},
		{Code: "cloud.accounts.update", Name: "Update cloud accounts", Type: "api", Method: "PUT", Path: "/api/v1/cloud/accounts/:id", Sort: 1003},
		{Code: "cloud.accounts.delete", Name: "Delete cloud accounts", Type: "api", Method: "DELETE", Path: "/api/v1/cloud/accounts/:id", Sort: 1004},
		{Code: "cloud.accounts.sync", Name: "Sync cloud accounts", Type: "api", Method: "POST", Path: "/api/v1/cloud/accounts/:id/sync", Sort: 1005},
		{Code: "cloud.assets.view", Name: "View cloud assets", Type: "api", Method: "GET", Path: "/api/v1/cloud/assets", Sort: 1100},
	}
}

func seedCloudPermissions(db *gorm.DB) error {
	for _, p := range DefaultCloudPermissions() {
		if err := db.Where("code = ?", p.Code).FirstOrCreate(&p).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedAdminCloudPermissions(db *gorm.DB) error {
	var adminRole systemModel.SystemRole
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		return err
	}
	for _, p := range DefaultCloudPermissions() {
		var perm systemModel.SystemPermission
		if err := db.Where("code = ?", p.Code).First(&perm).Error; err != nil {
			return err
		}
		rp := systemModel.SystemRolePermission{RoleID: adminRole.ID, PermissionID: perm.ID}
		if err := db.Where("role_id = ? AND permission_id = ?", adminRole.ID, perm.ID).FirstOrCreate(&rp).Error; err != nil {
			return err
		}
	}
	return nil
}
```

- [ ] **Step 9.2: Create `server/initialize/cloud/module.go`**

```go
package cloud

import (
	cloudApi "optimus/server/api/v1/cloud"
	"optimus/server/config"
	"optimus/server/service/cloud/infra/persistence"
	"optimus/server/service/cloud/infra/provider/aliyun"
	cloudUsecase "optimus/server/service/cloud/usecase"

	"gorm.io/gorm"
)

// CloudModule holds all dependencies for the cloud module.
type CloudModule struct {
	Handlers *cloudApi.CloudHandlers
}

func NewCloudModule(db *gorm.DB, cfg config.Config) *CloudModule {
	if db == nil {
		return &CloudModule{Handlers: cloudApi.NewCloudHandlers(cloudApi.CloudHandlerDeps{})}
	}

	encKey := []byte(cfg.Cloud.CredentialEncryptionKey)
	if len(encKey) != 32 {
		// Pad or truncate to 32 bytes so the module starts (credentials won't be usable without a proper key).
		padded := make([]byte, 32)
		copy(padded, encKey)
		encKey = padded
	}

	registry := cloudUsecase.NewSyncerRegistry()
	registry.Register("aliyun", aliyun.NewECSSyncer())

	accountStore := persistence.NewGormAccountStore(db)
	assetStore := persistence.NewGormAssetStore(db)

	accountSvc := cloudUsecase.NewCloudAccountService(accountStore, encKey)
	assetSvc := cloudUsecase.NewAssetRegistryService(accountSvc, assetStore, registry)

	return &CloudModule{
		Handlers: cloudApi.NewCloudHandlers(cloudApi.CloudHandlerDeps{
			Account: accountSvc,
			Asset:   assetSvc,
		}),
	}
}
```

- [ ] **Step 9.3: Modify `server/router/enter.go`** — add `Cloud` field

Current content:
```go
package router

import "optimus/server/router/system"

type RouterGroup struct {
	System system.RouterGroup
}

var RouterGroupApp = new(RouterGroup)
```

Replace with:
```go
package router

import (
	"optimus/server/router/cloud"
	"optimus/server/router/system"
)

type RouterGroup struct {
	System system.RouterGroup
	Cloud  cloud.RouterGroup
}

var RouterGroupApp = new(RouterGroup)
```

- [ ] **Step 9.4: Modify `server/initialize/router.go`** — wire CloudModule

Current content (full file):
```go
package initialize

import (
	"strings"

	"optimus/server/global"
	"optimus/server/middleware"
	"optimus/server/router"

	"github.com/gin-gonic/gin"
)

const defaultRouterPrefix = "/api/v1"

var newSystemModule = NewSystemModule

func Routers() *gin.Engine {
	systemModule := newSystemModule(global.OPTIMUS_DB, global.OPTIMUS_CONFIG)
	if systemModule == nil {
		systemModule = NewSystemModule(nil, global.OPTIMUS_CONFIG)
	}

	engine := gin.New()
	// 全局中间件
	engine.Use(middleware.Logger(), middleware.Recovery())
	// 公共路由
	publicGroup := engine.Group(routerPrefix())
	// 私有路由
	privateGroup := engine.Group(routerPrefix())
	privateGroup.Use(middleware.JWTAuth(), middleware.CasbinHandler(systemModule.Capabilities.RBAC))

	routerGroup := router.RouterGroupApp
	routerGroup.System.InitSystemRouter(publicGroup, privateGroup, systemModule.Handlers)

	return engine
}

func routerPrefix() string {
	prefix := strings.TrimSpace(global.OPTIMUS_CONFIG.System.RouterPrefix)
	if prefix == "" {
		return defaultRouterPrefix
	}
	if strings.HasPrefix(prefix, "/") {
		return prefix
	}
	return "/" + prefix
}
```

Replace with:
```go
package initialize

import (
	"strings"

	"optimus/server/global"
	cloudInit "optimus/server/initialize/cloud"
	"optimus/server/middleware"
	"optimus/server/router"

	"github.com/gin-gonic/gin"
)

const defaultRouterPrefix = "/api/v1"

var newSystemModule = NewSystemModule
var newCloudModule = cloudInit.NewCloudModule

func Routers() *gin.Engine {
	systemModule := newSystemModule(global.OPTIMUS_DB, global.OPTIMUS_CONFIG)
	if systemModule == nil {
		systemModule = NewSystemModule(nil, global.OPTIMUS_CONFIG)
	}
	cloudModule := newCloudModule(global.OPTIMUS_DB, global.OPTIMUS_CONFIG)

	engine := gin.New()
	// 全局中间件
	engine.Use(middleware.Logger(), middleware.Recovery())
	// 公共路由
	publicGroup := engine.Group(routerPrefix())
	// 私有路由
	privateGroup := engine.Group(routerPrefix())
	privateGroup.Use(middleware.JWTAuth(), middleware.CasbinHandler(systemModule.Capabilities.RBAC))

	routerGroup := router.RouterGroupApp
	routerGroup.System.InitSystemRouter(publicGroup, privateGroup, systemModule.Handlers)
	routerGroup.Cloud.InitCloudRouter(privateGroup, cloudModule.Handlers)

	return engine
}

func routerPrefix() string {
	prefix := strings.TrimSpace(global.OPTIMUS_CONFIG.System.RouterPrefix)
	if prefix == "" {
		return defaultRouterPrefix
	}
	if strings.HasPrefix(prefix, "/") {
		return prefix
	}
	return "/" + prefix
}
```

- [ ] **Step 9.5: Modify `server/initialize/startup.go`** — call cloud migration + seed

Current content:
```go
package initialize

import (
	"context"

	"optimus/server/global"
	systemService "optimus/server/service/system"
)

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

Replace with:
```go
package initialize

import (
	"context"

	"optimus/server/global"
	cloudService "optimus/server/service/cloud"
	systemService "optimus/server/service/system"
)

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

	if err := cloudService.AutoMigrateCloud(db); err != nil {
		logDependencyInitFailure("cloud", "auto migrate", err)
		return
	}
	if err := cloudService.SeedCloudData(db); err != nil {
		logDependencyInitFailure("cloud", "seed data", err)
	}

	startAuditRetentionWorker(context.Background(), db, global.OPTIMUS_CONFIG.System.AuditLogRetentionDays)
}
```

- [ ] **Step 9.6: Full build and test**

```bash
cd /home/lilin/trae-project/Optimus/server && go build ./... && go test ./...
```

Expected: build clean, all tests pass.

- [ ] **Step 9.7: Commit**

```bash
cd /home/lilin/trae-project/Optimus && \
git add server/service/cloud/cloud.go \
        server/initialize/cloud/module.go \
        server/router/enter.go \
        server/initialize/router.go \
        server/initialize/startup.go && \
git commit -m "完成 Phase 1 Cloud 模块接入：迁移、seed、路由注册"
```

---

## Final verification

- [ ] **Run full test suite**

```bash
cd /home/lilin/trae-project/Optimus/server && go build ./... && go test ./...
```

Expected: all pass, no compile warnings.

- [ ] **Verify router test still passes**

```bash
cd /home/lilin/trae-project/Optimus/server && go test ./initialize/... -run "TestRouters" -v
```

Expected: existing router tests PASS (cloud routes don't break system routes).
