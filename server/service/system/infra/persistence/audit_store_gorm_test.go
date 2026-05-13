package persistence

import (
	"context"
	"testing"
	"time"

	"optimus/server/service/system/usecase"

	"github.com/stretchr/testify/require"
)

func TestGormAuditStoreListAuditLogsOrdersNewestFirst(t *testing.T) {
	db := openPersistenceTestDB(t)

	authStore := NewGormAuthStore(db)
	require.NoError(t, authStore.WriteAudit(context.Background(), usecase.AuditEvent{Action: "first", Result: "success"}))
	require.NoError(t, authStore.WriteAudit(context.Background(), usecase.AuditEvent{Action: "second", Result: "success"}))

	store := NewGormAuditStore(db)
	page, err := store.QueryLogs(context.Background(), usecase.AuditFilter{Page: 1, PageSize: 20})

	require.NoError(t, err)
	require.Equal(t, int64(2), page.Total)
	require.Len(t, page.Items, 2)
	require.True(t, page.Items[0].CreatedAt.After(page.Items[1].CreatedAt) || page.Items[0].CreatedAt.Equal(page.Items[1].CreatedAt),
		"audit logs should be ordered newest first")
}

func TestGormAuditStoreQueryLogsPreservesActorUsername(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{
		ActorUserID:   42,
		ActorUsername: "alice",
		Action:        "login",
		Result:        "success",
	}))

	store := NewGormAuditStore(db)
	page, err := store.QueryLogs(context.Background(), usecase.AuditFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, "alice", page.Items[0].ActorUsername)
}

func TestGormAuditStoreQueryLogsPaginates(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	for i := 0; i < 5; i++ {
		require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "x", Result: "success"}))
	}

	store := NewGormAuditStore(db)
	page, err := store.QueryLogs(context.Background(), usecase.AuditFilter{Page: 1, PageSize: 2})

	require.NoError(t, err)
	require.Equal(t, int64(5), page.Total)
	require.Len(t, page.Items, 2)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 2, page.PageSize)
}

func TestGormAuditStoreQueryLogsFiltersByResult(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "a", Result: "success"}))
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "b", Result: "failure"}))

	store := NewGormAuditStore(db)
	result := "success"
	page, err := store.QueryLogs(context.Background(), usecase.AuditFilter{
		Page: 1, PageSize: 20, Result: &result,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "success", page.Items[0].Result)
}

func TestGormAuditStoreQueryLogsFiltersByAction(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "login", Result: "success"}))
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "logout", Result: "success"}))

	store := NewGormAuditStore(db)
	action := "login"
	page, err := store.QueryLogs(context.Background(), usecase.AuditFilter{
		Page: 1, PageSize: 20, Action: &action,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "login", page.Items[0].Action)
}

func TestGormAuditStoreQueryLogsFiltersByTimeRange(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "login", Result: "success"}))

	store := NewGormAuditStore(db)
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	pagePast, err := store.QueryLogs(context.Background(), usecase.AuditFilter{
		Page: 1, PageSize: 20, EndTime: &past,
	})
	require.NoError(t, err)
	require.Equal(t, int64(0), pagePast.Total)

	pageNow, err := store.QueryLogs(context.Background(), usecase.AuditFilter{
		Page: 1, PageSize: 20, StartTime: &past, EndTime: &future,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), pageNow.Total)
}

func TestGormAuditStoreExportLogsReturnsSortedEntries(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{
		ActorUsername: "alice", Action: "login", Result: "success",
	}))

	store := NewGormAuditStore(db)
	entries, err := store.ExportLogs(context.Background(), usecase.AuditFilter{})

	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "alice", entries[0].ActorUsername)
	require.Equal(t, "login", entries[0].Action)
}

func TestGormAuditStoreDeleteBefore(t *testing.T) {
	db := openPersistenceTestDB(t)
	auth := NewGormAuthStore(db)
	require.NoError(t, auth.WriteAudit(context.Background(), usecase.AuditEvent{Action: "old", Result: "success"}))

	store := NewGormAuditStore(db)
	future := time.Now().Add(time.Hour)
	require.NoError(t, store.DeleteBefore(context.Background(), future))

	page, err := store.QueryLogs(context.Background(), usecase.AuditFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(0), page.Total)
}
