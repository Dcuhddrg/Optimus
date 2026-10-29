package store

import (
	"context"
	"github.com/stretchr/testify/require"
	model "optimus/server/model/system"
	"testing"
)

func TestPermissionCodesSnapshotUsesSavedAssignments(t *testing.T) {
	db := openPersistenceTestDB(t)
	roles := []model.SystemRole{{Code: "first", Name: "First", Status: "enabled"}, {Code: "second", Name: "Second", Status: "enabled"}}
	require.NoError(t, db.Create(&roles).Error)
	permissions := []model.SystemPermission{{Code: "z.write", Name: "Write", Type: "api"}, {Code: "a.read", Name: "Read", Type: "api"}}
	require.NoError(t, db.Create(&permissions).Error)
	store := NewGormRoleStore(db)
	ctx := context.Background()
	require.NoError(t, store.ReplaceRolePermissions(ctx, roles[0].ID, []uint{permissions[0].ID, permissions[1].ID}))
	codes, err := store.PermissionCodesForRole(ctx, roles[0].ID)
	require.NoError(t, err)
	require.Equal(t, []string{"a.read", "z.write"}, codes)
	other, err := store.PermissionCodesForRole(ctx, roles[1].ID)
	require.NoError(t, err)
	require.Empty(t, other)
	require.NotNil(t, other)
	require.NoError(t, store.ReplaceRolePermissions(ctx, roles[0].ID, []uint{permissions[1].ID}))
	codes, err = store.PermissionCodesForRole(ctx, roles[0].ID)
	require.NoError(t, err)
	require.Equal(t, []string{"a.read"}, codes)
}
