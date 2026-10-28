package system

import (
	"context"
	systemModel "optimus/server/model/system"
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

type AuditStore interface {
	QueryLogs(ctx context.Context, f AuditFilter) (systemModel.AuditPage, error)
	ExportLogs(ctx context.Context, f AuditFilter) ([]systemModel.AuditLogEntry, error)
	DeleteBefore(ctx context.Context, before time.Time) error
}

type AuditService struct {
	store AuditStore
}

func NewAuditService(store AuditStore) *AuditService {
	return &AuditService{store: store}
}

func (s *AuditService) QueryLogs(ctx context.Context, f AuditFilter) (systemModel.AuditPage, error) {
	return s.store.QueryLogs(ctx, f)
}

func (s *AuditService) ExportLogs(ctx context.Context, f AuditFilter) ([]systemModel.AuditLogEntry, error) {
	return s.store.ExportLogs(ctx, f)
}

func (s *AuditService) PurgeBefore(ctx context.Context, before time.Time) error {
	return s.store.DeleteBefore(ctx, before)
}
