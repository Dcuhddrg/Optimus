package persistence

import (
	"context"

	systemModel "optimus/server/model/system"

	"gorm.io/gorm"
)

type GormAuditStore struct {
	db *gorm.DB
}

func NewGormAuditStore(db *gorm.DB) *GormAuditStore {
	return &GormAuditStore{db: db}
}

func (s *GormAuditStore) ListAuditLogs(ctx context.Context) ([]systemModel.SystemAuditLog, int64, error) {
	var total int64
	if err := s.db.WithContext(ctx).Model(&systemModel.SystemAuditLog{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []systemModel.SystemAuditLog
	err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(100).
		Find(&items).Error
	return items, total, err
}
