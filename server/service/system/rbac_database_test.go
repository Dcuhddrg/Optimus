package system

import (
	"context"
	"testing"

	systemModel "optimus/server/model/system"

	"github.com/stretchr/testify/require"
)

func TestRBACServiceDatabaseFindAPIPermissionNotFound(t *testing.T) {
	openPersistenceTestDB(t)

	store := (&RBACService{})
	_, found, err := store.findAPIPermission(context.Background(), "GET", "/nonexistent")

	require.NoError(t, err)
	require.False(t, found)
}

func TestRBACServiceDatabaseFindAPIPermissionFound(t *testing.T) {
	db := openPersistenceTestDB(t)

	perm := systemModel.SystemPermission{
		Code:   "users.list",
		Name:   "List Users",
		Type:   "api",
		Method: "GET",
		Path:   "/api/v1/system/users",
	}
	require.NoError(t, db.Create(&perm).Error)

	store := (&RBACService{})
	summary, found, err := store.findAPIPermission(context.Background(), "GET", "/api/v1/system/users")

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "users.list", summary.Code)
}
