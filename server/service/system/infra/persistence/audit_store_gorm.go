package persistence

import (
	"context"
	"time"

	systemModel "optimus/server/model/system"
	"optimus/server/service/system/usecase"

	"gorm.io/gorm"
)

type GormAuditStore struct {
	db *gorm.DB
}

func NewGormAuditStore(db *gorm.DB) *GormAuditStore {
	return &GormAuditStore{db: db}
}

func (s *GormAuditStore) buildQuery(ctx context.Context, f usecase.AuditFilter) *gorm.DB {
	q := s.db.WithContext(ctx).Model(&systemModel.SystemAuditLog{})
	if f.ActorUserID != nil {
		q = q.Where("actor_user_id = ?", *f.ActorUserID)
	}
	if f.Action != nil {
		q = q.Where("action = ?", *f.Action)
	}
	if f.ResourceType != nil {
		q = q.Where("resource_type = ?", *f.ResourceType)
	}
	if f.Result != nil {
		q = q.Where("result = ?", *f.Result)
	}
	if f.StartTime != nil {
		q = q.Where("created_at >= ?", *f.StartTime)
	}
	if f.EndTime != nil {
		q = q.Where("created_at <= ?", *f.EndTime)
	}
	if f.Keyword != nil && *f.Keyword != "" {
		q = q.Where("detail LIKE ?", "%"+*f.Keyword+"%")
	}
	return q
}

func (s *GormAuditStore) QueryLogs(ctx context.Context, f usecase.AuditFilter) (usecase.AuditPage, error) {
	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	q := s.buildQuery(ctx, f)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return usecase.AuditPage{}, err
	}

	var rows []systemModel.SystemAuditLog
	err := q.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error
	if err != nil {
		return usecase.AuditPage{}, err
	}

	return usecase.AuditPage{
		Items:    toAuditLogEntries(rows),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *GormAuditStore) ExportLogs(ctx context.Context, f usecase.AuditFilter) ([]usecase.AuditLogEntry, error) {
	q := s.buildQuery(ctx, f)
	var rows []systemModel.SystemAuditLog
	err := q.Order("created_at DESC").Limit(10000).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toAuditLogEntries(rows), nil
}

func (s *GormAuditStore) DeleteBefore(ctx context.Context, before time.Time) error {
	tableName := systemModel.SystemAuditLog{}.TableName()
	for {
		result := s.db.WithContext(ctx).
			Exec("DELETE FROM `"+tableName+"` WHERE created_at < ? LIMIT 1000", before)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
	}
}

func toAuditLogEntries(rows []systemModel.SystemAuditLog) []usecase.AuditLogEntry {
	entries := make([]usecase.AuditLogEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, usecase.AuditLogEntry{
			ID:            row.ID,
			ActorUserID:   row.ActorUserID,
			ActorUsername: row.ActorUsername,
			Action:        row.Action,
			ResourceType:  row.ResourceType,
			ResourceID:    row.ResourceID,
			Result:        row.Result,
			IP:            row.IP,
			UserAgent:     row.UserAgent,
			Detail:        row.Detail,
			CreatedAt:     row.CreatedAt,
		})
	}
	return entries
}
