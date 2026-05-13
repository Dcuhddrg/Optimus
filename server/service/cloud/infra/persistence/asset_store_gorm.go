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
