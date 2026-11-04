package system

import (
	"context"
	"testing"

	"optimus/server/global"
	systemModel "optimus/server/model/system"

	"github.com/stretchr/testify/require"
)

func TestRBACServiceAllowsRoleWithPermission(t *testing.T) {
	db := openTestDB(t)
	permission := systemModel.SystemPermission{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 10}, Code: "system.users.view", Name: "View users", Type: "api", Method: "GET", Path: "/api/v1/system/users"}
	require.NoError(t, db.Create(&permission).Error)
	require.NoError(t, db.Create(&systemModel.SystemRolePermission{RoleID: 1, PermissionID: 10}).Error)
	service := &RBACService{}

	allowed, err := service.AuthorizeAPI(context.Background(), "GET", "/api/v1/system/users", []uint{1})

	require.NoError(t, err)
	require.True(t, allowed)
}

func TestRBACServiceDatabaseFindAPIPermissionNotFound(t *testing.T) {
	openTestDB(t)

	store := (&RBACService{})
	_, found, err := store.findAPIPermission(context.Background(), "GET", "/nonexistent")

	require.NoError(t, err)
	require.False(t, found)
}

func TestRBACServiceDatabaseFindAPIPermissionFound(t *testing.T) {
	db := openTestDB(t)

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
