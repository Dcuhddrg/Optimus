package persistence

import (
	"context"
	"testing"

	"optimus/server/service/system/usecase"
	systemModel "optimus/server/model/system"

	"github.com/stretchr/testify/require"
)

func TestMenuPermissionCodesForRolesOrdersDistinctCodes(t *testing.T) {
	db := openPersistenceTestDB(t)

	adminRole := systemModel.SystemRole{Code: "admin", Name: "Administrator", Status: enabledRoleStatus}
	opsRole := systemModel.SystemRole{Code: "ops", Name: "Ops", Status: enabledRoleStatus}
	require.NoError(t, db.Create(&[]systemModel.SystemRole{adminRole, opsRole}).Error)
	require.NoError(t, db.Where("code = ?", "admin").First(&adminRole).Error)
	require.NoError(t, db.Where("code = ?", "ops").First(&opsRole).Error)

	first := systemModel.SystemPermission{Code: "system.users.menu", Name: "Users", Type: "menu", Sort: 20}
	second := systemModel.SystemPermission{Code: "dashboard.view", Name: "Dashboard", Type: "menu", Sort: 10}
	apiPerm := systemModel.SystemPermission{Code: "system.users.view", Name: "Users API", Type: "api", Sort: 5}
	require.NoError(t, db.Create(&[]systemModel.SystemPermission{first, second, apiPerm}).Error)

	var permissions []systemModel.SystemPermission
	require.NoError(t, db.Find(&permissions).Error)
	for _, p := range permissions {
		require.NoError(t, db.Create(&systemModel.SystemRolePermission{RoleID: adminRole.ID, PermissionID: p.ID}).Error)
		require.NoError(t, db.Create(&systemModel.SystemRolePermission{RoleID: opsRole.ID, PermissionID: p.ID}).Error)
	}

	codes, err := NewGormAuthStore(db).MenuPermissionCodesForRoles(context.Background(), []uint{adminRole.ID, opsRole.ID})

	require.NoError(t, err)
	require.Equal(t, []string{"dashboard.view", "system.users.menu"}, codes)
}

func TestGormAuthStoreSyncLDAPUserCreatesDefaultRole(t *testing.T) {
	db := openPersistenceTestDB(t)

	userRole := systemModel.SystemRole{Code: "user", Name: "User", Status: enabledRoleStatus}
	require.NoError(t, db.Create(&userRole).Error)

	store := NewGormAuthStore(db)
	user, err := store.SyncLDAPUser(context.Background(), usecase.LDAPUser{
		Username:    "alice",
		DisplayName: "Alice",
		Email:       "alice@example.com",
	})

	require.NoError(t, err)
	require.Equal(t, "alice", user.Username)
	require.Equal(t, "ldap", user.Source)

	var userRoles []systemModel.SystemUserRole
	require.NoError(t, db.Where("user_id = ?", user.ID).Find(&userRoles).Error)
	require.Len(t, userRoles, 1)
	require.Equal(t, userRole.ID, userRoles[0].RoleID)
}
