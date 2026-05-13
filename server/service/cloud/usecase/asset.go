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

// QueryAssets returns a paginated list of cloud assets matching the filter.
func (s *AssetRegistryService) QueryAssets(ctx context.Context, f AssetFilter) (AssetPage, error) {
	return s.store.Query(ctx, f)
}

// timeNow is a hook for testing.
var timeNow = func() time.Time { return time.Now() }
