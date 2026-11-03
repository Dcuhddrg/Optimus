package cloud

import (
	"context"
	"errors"
	"optimus/server/global"
	cloudModel "optimus/server/model/cloud"
	cloudResponse "optimus/server/model/cloud/response"
	commonRequest "optimus/server/model/common/request"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// AssetRegistryService orchestrates sync and querying of cloud assets.
type AssetRegistryService struct {
	accounts *CloudAccountService
	registry *SyncerRegistry
	now      func() time.Time
}

// Ready reports whether the service was constructed.
func (s *AssetRegistryService) Ready() bool {
	return s != nil && s.now != nil
}

func NewAssetRegistryService(accounts *CloudAccountService, registry *SyncerRegistry) *AssetRegistryService {
	return &AssetRegistryService{accounts: accounts, registry: registry, now: time.Now}
}

// SyncAccount fetches all assets for the given account and upserts them into the registry.
func (s *AssetRegistryService) SyncAccount(ctx context.Context, accessibleProjectIDs []uint, accountID uint) (cloudResponse.SyncResult, error) {
	if _, err := s.accounts.Get(ctx, accessibleProjectIDs, accountID); err != nil {
		return cloudResponse.SyncResult{}, err
	}

	syncAcc, err := s.accounts.GetForSync(ctx, accountID)
	if err != nil {
		statusErr := s.accounts.UpdateSyncStatus(ctx, accountID, nil, "error")
		return cloudResponse.SyncResult{}, errors.Join(err, statusErr)
	}
	syncer, err := s.registry.Get(syncAcc.Provider)
	if err != nil {
		statusErr := s.accounts.UpdateSyncStatus(ctx, accountID, nil, "error")
		return cloudResponse.SyncResult{}, errors.Join(err, statusErr)
	}
	raw, err := syncer.Sync(ctx, syncAcc)
	if err != nil {
		statusErr := s.accounts.UpdateSyncStatus(ctx, accountID, nil, "error")
		return cloudResponse.SyncResult{}, errors.Join(err, statusErr)
	}
	result, err := s.upsertAssets(ctx, accountID, syncAcc.Provider, raw)
	if err != nil {
		statusErr := s.accounts.UpdateSyncStatus(ctx, accountID, nil, "error")
		return cloudResponse.SyncResult{}, errors.Join(err, statusErr)
	}
	now := s.now()
	if statusErr := s.accounts.UpdateSyncStatus(ctx, accountID, &now, "active"); statusErr != nil {
		return result, statusErr
	}
	return result, nil
}

// QueryAssets returns a paginated list of cloud assets matching the filter.
func (s *AssetRegistryService) QueryAssets(ctx context.Context, f AssetFilter) (cloudResponse.AssetPage, error) {
	if len(f.AccessibleProjectIDs) == 0 {
		return cloudResponse.AssetPage{Items: []cloudResponse.AssetView{}, Page: f.Page, PageSize: f.PageSize}, nil
	}

	q := global.OPTIMUS_DB.WithContext(ctx).
		Model(&cloudModel.CloudAsset{}).
		Joins("JOIN cloud_accounts ON cloud_accounts.id = cloud_assets.account_id").
		Where("cloud_accounts.project_id IN ?", f.AccessibleProjectIDs).
		Where("cloud_accounts.deleted_at IS NULL")
	if f.AccountID != nil {
		q = q.Where("cloud_assets.account_id = ?", *f.AccountID)
	}
	if f.Provider != nil {
		q = q.Where("cloud_assets.provider = ?", *f.Provider)
	}
	if f.Region != nil {
		q = q.Where("cloud_assets.region = ?", *f.Region)
	}
	if f.ResourceType != nil {
		q = q.Where("cloud_assets.resource_type = ?", *f.ResourceType)
	}
	if f.Status != nil {
		q = q.Where("cloud_assets.status = ?", *f.Status)
	}
	if f.Keyword != nil && *f.Keyword != "" {
		escaped := strings.NewReplacer(`%`, `\%`, `_`, `\_`).Replace(*f.Keyword)
		q = q.Where("cloud_assets.name LIKE ? ESCAPE '\\\\'", "%"+escaped+"%")
	}
	for _, filter := range f.Tags {
		expr, args := tagFilterToSQL(filter)
		expr = strings.Replace(expr, "tags_json", "cloud_accounts.tags_json", 1)
		q = q.Where(expr, args...)
	}

	pagination := commonRequest.PageInfo{Page: f.Page, PageSize: f.PageSize}
	pagination.Normalize()
	page, pageSize := pagination.Page, pagination.PageSize

	countQ := q.Session(&gorm.Session{})
	var total int64
	if err := countQ.Count(&total).Error; err != nil {
		return cloudResponse.AssetPage{}, err
	}
	var models []cloudModel.CloudAsset
	if err := q.Order("cloud_assets.created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&models).Error; err != nil {
		return cloudResponse.AssetPage{}, err
	}
	items := make([]cloudResponse.AssetView, len(models))
	for i, m := range models {
		items[i] = toAssetView(m)
	}
	return cloudResponse.AssetPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AssetRegistryService) upsertAssets(ctx context.Context, accountID uint, provider string, assets []RawAsset) (cloudResponse.SyncResult, error) {
	if len(assets) == 0 {
		return cloudResponse.SyncResult{}, nil
	}
	now := time.Now()
	records := make([]cloudModel.CloudAsset, len(assets))
	for i, a := range assets {
		records[i] = cloudModel.CloudAsset{
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
	}
	// On conflict (account_id, resource_id already exists):
	//   - discovered_at is intentionally omitted from DoUpdates to preserve the first-seen timestamp.
	//     MySQL ON DUPLICATE KEY UPDATE leaves unlisted columns at their stored values.
	//   - deleted_at IS included: a resource that re-appears after being soft-deleted is un-deleted
	//     by setting deleted_at = NULL (the zero value of *time.Time in the freshly-built record).
	res := global.OPTIMUS_DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "account_id"}, {Name: "resource_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name", "status", "config_json", "tags_json", "last_seen_at", "deleted_at",
			}),
		}).
		CreateInBatches(records, 100)
	if res.Error != nil {
		return cloudResponse.SyncResult{}, res.Error
	}
	return cloudResponse.SyncResult{Upserted: len(assets)}, nil
}

func toAssetView(m cloudModel.CloudAsset) cloudResponse.AssetView {
	return cloudResponse.AssetView{
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
