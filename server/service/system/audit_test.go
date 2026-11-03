package system

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAuditServiceQueryLogsNormalizesPagination(t *testing.T) {
	openPersistenceTestDB(t)
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "login", ActorUsername: "admin"}))

	page, err := (&AuditService{}).QueryLogs(context.Background(), AuditFilter{})
	require.NoError(t, err)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 20, page.PageSize)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "admin", page.Items[0].ActorUsername)
}
