package cloud

import (
	"context"
	"errors"
	cloudModel "optimus/server/model/cloud"
	"time"
)

// AssetFilter holds optional query parameters for listing cloud assets.
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

// AssetStore persists and queries cloud assets.
type AssetStore interface {
	Upsert(ctx context.Context, accountID uint, provider string, assets []RawAsset) (cloudModel.SyncResult, error)
	Query(ctx context.Context, f AssetFilter) (cloudModel.AssetPage, error)
}

// AssetRegistryService orchestrates sync and querying of cloud assets.
type AssetRegistryService struct {
	accounts *CloudAccountService
	store    AssetStore
	registry *SyncerRegistry
	now      func() time.Time
}

func NewAssetRegistryService(accounts *CloudAccountService, store AssetStore, registry *SyncerRegistry) *AssetRegistryService {
	return &AssetRegistryService{accounts: accounts, store: store, registry: registry, now: time.Now}
}

// SyncAccount fetches all assets for the given account and upserts them into the registry.
func (s *AssetRegistryService) SyncAccount(ctx context.Context, accessibleProjectIDs []uint, accountID uint) (cloudModel.SyncResult, error) {
	if _, err := s.accounts.Get(ctx, accessibleProjectIDs, accountID); err != nil {
		return cloudModel.SyncResult{}, err
	}

	syncAcc, err := s.accounts.GetForSync(ctx, accountID)
	if err != nil {
		statusErr := s.accounts.UpdateSyncStatus(ctx, accountID, nil, "error")
		return cloudModel.SyncResult{}, errors.Join(err, statusErr)
	}
	syncer, err := s.registry.Get(syncAcc.Provider)
	if err != nil {
		statusErr := s.accounts.UpdateSyncStatus(ctx, accountID, nil, "error")
		return cloudModel.SyncResult{}, errors.Join(err, statusErr)
	}
	raw, err := syncer.Sync(ctx, syncAcc)
	if err != nil {
		statusErr := s.accounts.UpdateSyncStatus(ctx, accountID, nil, "error")
		return cloudModel.SyncResult{}, errors.Join(err, statusErr)
	}
	result, err := s.store.Upsert(ctx, accountID, syncAcc.Provider, raw)
	if err != nil {
		statusErr := s.accounts.UpdateSyncStatus(ctx, accountID, nil, "error")
		return cloudModel.SyncResult{}, errors.Join(err, statusErr)
	}
	now := s.now()
	if statusErr := s.accounts.UpdateSyncStatus(ctx, accountID, &now, "active"); statusErr != nil {
		return result, statusErr
	}
	return result, nil
}

// QueryAssets returns a paginated list of cloud assets matching the filter.
func (s *AssetRegistryService) QueryAssets(ctx context.Context, f AssetFilter) (cloudModel.AssetPage, error) {
	if len(f.AccessibleProjectIDs) == 0 {
		return cloudModel.AssetPage{Items: []cloudModel.AssetView{}, Page: f.Page, PageSize: f.PageSize}, nil
	}
	return s.store.Query(ctx, f)
}
