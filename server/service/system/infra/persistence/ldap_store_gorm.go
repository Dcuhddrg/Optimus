package persistence

import (
	"context"
	"errors"

	"optimus/server/service/system/usecase"
	systemModel "optimus/server/model/system"

	"gorm.io/gorm"
)

type GormLDAPStore struct {
	db *gorm.DB
}

func NewGormLDAPStore(db *gorm.DB) *GormLDAPStore {
	return &GormLDAPStore{db: db}
}

func (s *GormLDAPStore) GetLDAPConfig(ctx context.Context) (systemModel.SystemLDAPConfig, error) {
	var cfg systemModel.SystemLDAPConfig
	err := s.db.WithContext(ctx).Order("id ASC").First(&cfg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return systemModel.SystemLDAPConfig{}, nil
	}
	return cfg, err
}

func (s *GormLDAPStore) SaveLDAPConfig(ctx context.Context, cfg systemModel.SystemLDAPConfig) (systemModel.SystemLDAPConfig, error) {
	if cfg.ID == 0 {
		err := s.db.WithContext(ctx).Create(&cfg).Error
		return cfg, err
	}
	err := s.db.WithContext(ctx).Omit("CreatedAt").Save(&cfg).Error
	return cfg, err
}

func (s *GormLDAPStore) WriteAudit(ctx context.Context, event usecase.AuditEvent) error {
	return NewGormAuthStore(s.db).WriteAudit(ctx, event)
}
