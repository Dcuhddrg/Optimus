package system

import (
	"context"

	systemModel "optimus/server/model/system"

	"gorm.io/gorm"
)

type AuditMeta struct {
	IP        string
	UserAgent string
}

type AuditEvent struct {
	ActorUserID  uint
	Action       string
	ResourceType string
	ResourceID   string
	Result       string
	IP           string
	UserAgent    string
	Detail       string
}

type AuditLogListResult struct {
	Items []systemModel.SystemAuditLog `json:"items"`
	Total int64                        `json:"total"`
}

type AuditStore interface {
	ListAuditLogs(ctx context.Context) ([]systemModel.SystemAuditLog, int64, error)
}

type AuditService struct {
	store AuditStore
}

func NewAuditService(store AuditStore) *AuditService {
	return &AuditService{store: store}
}

func (s *AuditService) ListAuditLogs(ctx context.Context) (AuditLogListResult, error) {
	items, total, err := s.store.ListAuditLogs(ctx)
	if err != nil {
		return AuditLogListResult{}, err
	}
	return AuditLogListResult{Items: items, Total: total}, nil
}

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
