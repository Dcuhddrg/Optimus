package persistence

import (
	"context"
	"strings"
	"time"

	cloudModel "optimus/server/model/cloud"
	"optimus/server/service/cloud/usecase"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormAssetStore struct {
	db *gorm.DB
}

func NewGormAssetStore(db *gorm.DB) *GormAssetStore {
	return &GormAssetStore{db: db}
}

func (s *GormAssetStore) Upsert(ctx context.Context, accountID uint, provider string, assets []usecase.RawAsset) (usecase.SyncResult, error) {
	if len(assets) == 0 {
		return usecase.SyncResult{}, nil
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
		return usecase.SyncResult{}, res.Error
	}
	return usecase.SyncResult{Upserted: len(assets)}, nil
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
		escaped := strings.NewReplacer(`%`, `\%`, `_`, `\_`).Replace(*f.Keyword)
		q = q.Where("name LIKE ? ESCAPE '\\\\'", "%"+escaped+"%")
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
