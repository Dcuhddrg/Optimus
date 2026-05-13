package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeAuditStore struct {
	page    AuditPage
	entries []AuditLogEntry
	deleted time.Time
}

func (f *fakeAuditStore) QueryLogs(_ context.Context, _ AuditFilter) (AuditPage, error) {
	return f.page, nil
}

func (f *fakeAuditStore) ExportLogs(_ context.Context, _ AuditFilter) ([]AuditLogEntry, error) {
	return f.entries, nil
}

func (f *fakeAuditStore) DeleteBefore(_ context.Context, before time.Time) error {
	f.deleted = before
	return nil
}

func TestAuditServiceQueryLogsReturnsPageFromStore(t *testing.T) {
	expected := AuditPage{
		Items:    []AuditLogEntry{{ID: 1, Action: "login", ActorUsername: "admin"}},
		Total:    1,
		Page:     1,
		PageSize: 20,
	}
	svc := NewAuditService(&fakeAuditStore{page: expected})
	got, err := svc.QueryLogs(context.Background(), AuditFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, expected, got)
}

func TestAuditServiceExportLogsReturnsEntriesFromStore(t *testing.T) {
	expected := []AuditLogEntry{{ID: 1, Action: "login", ActorUsername: "admin"}}
	svc := NewAuditService(&fakeAuditStore{entries: expected})
	got, err := svc.ExportLogs(context.Background(), AuditFilter{})
	require.NoError(t, err)
	require.Equal(t, expected, got)
}

func TestAuditServicePurgeBeforeCallsDeleteBefore(t *testing.T) {
	store := &fakeAuditStore{}
	svc := NewAuditService(store)
	cutoff := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	require.NoError(t, svc.PurgeBefore(context.Background(), cutoff))
	require.Equal(t, cutoff, store.deleted)
}
