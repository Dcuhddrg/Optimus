package store

import (
	"context"
	"testing"

	systemModel "optimus/server/model/system"

	"github.com/stretchr/testify/require"
)

func TestGormRBACStoreFindAPIPermissionNotFound(t *testing.T) {
	db := openPersistenceTestDB(t)

	store := NewGormRBACStore(db)
	_, found, err := store.FindAPIPermission(context.Background(), "GET", "/nonexistent")

	require.NoError(t, err)
	require.False(t, found)
}

func TestGormRBACStoreFindAPIPermissionFound(t *testing.T) {
	db := openPersistenceTestDB(t)

	perm := systemModel.SystemPermission{
		Code:   "users.list",
		Name:   "List Users",
		Type:   "api",
		Method: "GET",
		Path:   "/api/v1/system/users",
	}
	require.NoError(t, db.Create(&perm).Error)

	store := NewGormRBACStore(db)
	summary, found, err := store.FindAPIPermission(context.Background(), "GET", "/api/v1/system/users")

	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "users.list", summary.Code)
}
