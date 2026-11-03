package system

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAuditServiceDatabaseQueryLogsOrdersNewestFirst(t *testing.T) {
	openPersistenceTestDB(t)

	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "first", Result: "success"}))
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "second", Result: "success"}))

	store := (&AuditService{})
	page, err := store.QueryLogs(context.Background(), AuditFilter{Page: 1, PageSize: 20})

	require.NoError(t, err)
	require.Equal(t, int64(2), page.Total)
	require.Len(t, page.Items, 2)
	require.True(t, page.Items[0].CreatedAt.After(page.Items[1].CreatedAt) || page.Items[0].CreatedAt.Equal(page.Items[1].CreatedAt),
		"audit logs should be ordered newest first")
}

func TestAuditServiceDatabaseQueryLogsPreservesActorUsername(t *testing.T) {
	openPersistenceTestDB(t)
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{
		ActorUserID:   42,
		ActorUsername: "alice",
		Action:        "login",
		Result:        "success",
	}))

	store := (&AuditService{})
	page, err := store.QueryLogs(context.Background(), AuditFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, "alice", page.Items[0].ActorUsername)
}

func TestAuditServiceDatabaseQueryLogsPaginates(t *testing.T) {
	openPersistenceTestDB(t)
	for i := 0; i < 5; i++ {
		require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "x", Result: "success"}))
	}

	store := (&AuditService{})
	page, err := store.QueryLogs(context.Background(), AuditFilter{Page: 1, PageSize: 2})

	require.NoError(t, err)
	require.Equal(t, int64(5), page.Total)
	require.Len(t, page.Items, 2)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 2, page.PageSize)
}

func TestAuditServiceDatabaseQueryLogsFiltersByResult(t *testing.T) {
	openPersistenceTestDB(t)
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "a", Result: "success"}))
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "b", Result: "failure"}))

	store := (&AuditService{})
	result := "success"
	page, err := store.QueryLogs(context.Background(), AuditFilter{
		Page: 1, PageSize: 20, Result: &result,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "success", page.Items[0].Result)
}

func TestAuditServiceDatabaseQueryLogsFiltersByAction(t *testing.T) {
	openPersistenceTestDB(t)
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "login", Result: "success"}))
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "logout", Result: "success"}))

	store := (&AuditService{})
	action := "login"
	page, err := store.QueryLogs(context.Background(), AuditFilter{
		Page: 1, PageSize: 20, Action: &action,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "login", page.Items[0].Action)
}

func TestAuditServiceDatabaseQueryLogsFiltersByTimeRange(t *testing.T) {
	openPersistenceTestDB(t)
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "login", Result: "success"}))

	store := (&AuditService{})
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	pagePast, err := store.QueryLogs(context.Background(), AuditFilter{
		Page: 1, PageSize: 20, EndTime: &past,
	})
	require.NoError(t, err)
	require.Equal(t, int64(0), pagePast.Total)

	pageNow, err := store.QueryLogs(context.Background(), AuditFilter{
		Page: 1, PageSize: 20, StartTime: &past, EndTime: &future,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), pageNow.Total)
}

func TestAuditServiceDatabaseExportLogsReturnsSortedEntries(t *testing.T) {
	openPersistenceTestDB(t)
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{
		ActorUsername: "alice", Action: "login", Result: "success",
	}))

	store := (&AuditService{})
	entries, err := store.ExportLogs(context.Background(), AuditFilter{})

	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "alice", entries[0].ActorUsername)
	require.Equal(t, "login", entries[0].Action)
}

func TestAuditServiceDatabaseDeleteBefore(t *testing.T) {
	db := openPersistenceTestDB(t)
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "old", Result: "success"}))

	store := (&AuditService{})
	future := time.Now().Add(time.Hour)
	require.NoError(t, store.PurgeBefore(context.Background(), future))

	page, err := store.QueryLogs(context.Background(), AuditFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Equal(t, int64(0), page.Total)

	var retained int64
	require.NoError(t, db.Unscoped().Table("sys_audit_logs").Count(&retained).Error)
	require.Zero(t, retained)
}
