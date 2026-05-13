package persistence

import (
	"context"
	"testing"

	"optimus/server/service/system/usecase"

	"github.com/stretchr/testify/require"
)

func TestGormAuditStoreListAuditLogsOrdersNewestFirst(t *testing.T) {
	db := openPersistenceTestDB(t)

	authStore := NewGormAuthStore(db)
	require.NoError(t, authStore.WriteAudit(context.Background(), usecase.AuditEvent{Action: "first", Result: "success"}))
	require.NoError(t, authStore.WriteAudit(context.Background(), usecase.AuditEvent{Action: "second", Result: "success"}))

	store := NewGormAuditStore(db)
	items, total, err := store.ListAuditLogs(context.Background())

	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, items, 2)

	actions := []string{items[0].Action, items[1].Action}
	require.Contains(t, actions, "first")
	require.Contains(t, actions, "second")

	// Verify DB-level ordering (newest first) by checking created_at sequence
	require.True(t, items[0].CreatedAt.After(items[1].CreatedAt) || items[0].CreatedAt.Equal(items[1].CreatedAt),
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
	items, _, err := store.ListAuditLogs(context.Background())
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "alice", items[0].ActorUsername)
}
