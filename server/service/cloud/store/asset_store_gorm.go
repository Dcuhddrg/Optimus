package store

import (
	"context"
	"strings"
	"time"

	cloudModel "optimus/server/model/cloud"
	cloudService "optimus/server/service/cloud"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormAssetStore struct {
	db *gorm.DB
}

func NewGormAssetStore(db *gorm.DB) *GormAssetStore {
	return &GormAssetStore{db: db}
}

func (s *GormAssetStore) Upsert(ctx context.Context, accountID uint, provider string, assets []cloudService.RawAsset) (cloudModel.SyncResult, error) {
	if len(assets) == 0 {
		return cloudModel.SyncResult{}, nil
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
	res := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "account_id"}, {Name: "resource_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name", "status", "config_json", "tags_json", "last_seen_at", "deleted_at",
			}),
		}).
		CreateInBatches(records, 100)
	if res.Error != nil {
		return cloudModel.SyncResult{}, res.Error
	}
	return cloudModel.SyncResult{Upserted: len(assets)}, nil
}

func (s *GormAssetStore) Query(ctx context.Context, f cloudService.AssetFilter) (cloudModel.AssetPage, error) {
	if len(f.AccessibleProjectIDs) == 0 {
		return cloudModel.AssetPage{Items: []cloudModel.AssetView{}}, nil
	}

	q := s.db.WithContext(ctx).
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

	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	countQ := q.Session(&gorm.Session{})
	var total int64
	if err := countQ.Count(&total).Error; err != nil {
		return cloudModel.AssetPage{}, err
	}
	var models []cloudModel.CloudAsset
	if err := q.Order("cloud_assets.created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&models).Error; err != nil {
		return cloudModel.AssetPage{}, err
	}
	items := make([]cloudModel.AssetView, len(models))
	for i, m := range models {
		items[i] = toAssetView(m)
	}
	return cloudModel.AssetPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func toAssetView(m cloudModel.CloudAsset) cloudModel.AssetView {
	return cloudModel.AssetView{
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
