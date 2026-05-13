package usecase

import (
	"context"
	"time"
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

type AuditPage struct {
	Items    []AuditLogEntry `json:"items"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

type AuditLogEntry struct {
	ID            uint      `json:"id"`
	ActorUserID   uint      `json:"actorUserId"`
	ActorUsername string    `json:"actorUsername"`
	Action        string    `json:"action"`
	ResourceType  string    `json:"resourceType"`
	ResourceID    string    `json:"resourceId"`
	Result        string    `json:"result"`
	IP            string    `json:"ip"`
	UserAgent     string    `json:"userAgent"`
	Detail        string    `json:"detail"`
	CreatedAt     time.Time `json:"createdAt"`
}

type AuditStore interface {
	QueryLogs(ctx context.Context, f AuditFilter) (AuditPage, error)
	ExportLogs(ctx context.Context, f AuditFilter) ([]AuditLogEntry, error)
	DeleteBefore(ctx context.Context, before time.Time) error
}

type AuditService struct {
	store AuditStore
}

func NewAuditService(store AuditStore) *AuditService {
	return &AuditService{store: store}
}

func (s *AuditService) QueryLogs(ctx context.Context, f AuditFilter) (AuditPage, error) {
	return s.store.QueryLogs(ctx, f)
}

func (s *AuditService) ExportLogs(ctx context.Context, f AuditFilter) ([]AuditLogEntry, error) {
	return s.store.ExportLogs(ctx, f)
}

func (s *AuditService) PurgeBefore(ctx context.Context, before time.Time) error {
	return s.store.DeleteBefore(ctx, before)
}
