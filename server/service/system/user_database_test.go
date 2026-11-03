package system

import (
	"context"
	"testing"

	systemModel "optimus/server/model/system"

	"github.com/stretchr/testify/require"
)

func TestUserServiceDatabaseWouldHaveAdminCapableUserAfterStatus(t *testing.T) {
	db := openPersistenceTestDB(t)

	adminRole := systemModel.SystemRole{Code: "admin", Name: "Administrator", Status: defaultRoleStatus}
	require.NoError(t, db.Create(&adminRole).Error)

	user := systemModel.SystemUser{UUID: "u1", Username: "alice", Status: defaultUserStatus}
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: user.ID, RoleID: adminRole.ID}).Error)

	store := (&UserService{})

	// Disabling the only admin user should leave no admin-capable user
	ok, err := store.wouldHaveAdminCapableUserAfterStatus(context.Background(), user.ID, "disabled")
	require.NoError(t, err)
	require.False(t, ok)

	// Another enabled admin exists so disabling user is safe
	other := systemModel.SystemUser{UUID: "u2", Username: "bob", Status: defaultUserStatus}
	require.NoError(t, db.Create(&other).Error)
	require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: other.ID, RoleID: adminRole.ID}).Error)

	ok, err = store.wouldHaveAdminCapableUserAfterStatus(context.Background(), user.ID, "disabled")
	require.NoError(t, err)
	require.True(t, ok)
}
