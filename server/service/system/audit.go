package system

import (
	"context"
	"time"

	"optimus/server/global"
	commonRequest "optimus/server/model/common/request"
	systemModel "optimus/server/model/system"

	"gorm.io/gorm"
)

type AuditMeta struct {
	IP            string
	UserAgent     string
	ActorUsername string
}

type AuditEvent struct {
	ActorUserID   uint
	ActorUsername string
	Action        string
	ResourceType  string
	ResourceID    string
	Result        string
	IP            string
	UserAgent     string
	Detail        string
}

type AuditFilter struct {
	ActorUserID  *uint
	Action       *string
	ResourceType *string
	Result       *string
	StartTime    *time.Time
	EndTime      *time.Time
	Keyword      *string
	Page         int
	PageSize     int
}

// AuditOperations is the audit capability shared by HTTP handlers and retention.
type AuditOperations interface {
	QueryLogs(context.Context, AuditFilter) (systemModel.AuditPage, error)
	ExportLogs(context.Context, AuditFilter) ([]systemModel.AuditLogEntry, error)
	PurgeBefore(context.Context, time.Time) error
}

type AuditService struct{}

func (s *AuditService) QueryLogs(ctx context.Context, f AuditFilter) (systemModel.AuditPage, error) {
	pagination := commonRequest.PageInfo{Page: f.Page, PageSize: f.PageSize}
	pagination.Normalize()
	page, pageSize := pagination.Page, pagination.PageSize

	q := s.buildQuery(ctx, f)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return systemModel.AuditPage{}, err
	}

	var rows []systemModel.SystemAuditLog
	err := q.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error
	if err != nil {
		return systemModel.AuditPage{}, err
	}

	return systemModel.AuditPage{
		Items:    toAuditLogEntries(rows),
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *AuditService) ExportLogs(ctx context.Context, f AuditFilter) ([]systemModel.AuditLogEntry, error) {
	q := s.buildQuery(ctx, f)
	var rows []systemModel.SystemAuditLog
	err := q.Order("created_at DESC").Limit(10000).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toAuditLogEntries(rows), nil
}

func (s *AuditService) PurgeBefore(ctx context.Context, before time.Time) error {
	tableName := systemModel.SystemAuditLog{}.TableName()
	for {
		result := global.OPTIMUS_DB.WithContext(ctx).
			Exec("DELETE FROM `"+tableName+"` WHERE created_at < ? LIMIT 1000", before)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
	}
}

func (s *AuditService) buildQuery(ctx context.Context, f AuditFilter) *gorm.DB {
	q := global.OPTIMUS_DB.WithContext(ctx).Model(&systemModel.SystemAuditLog{})
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

func toAuditLogEntries(rows []systemModel.SystemAuditLog) []systemModel.AuditLogEntry {
	entries := make([]systemModel.AuditLogEntry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, systemModel.AuditLogEntry{
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
