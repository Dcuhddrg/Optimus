package usecase

import (
	"context"

	systemModel "optimus/server/model/system"
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
