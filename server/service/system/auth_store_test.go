package system

import (
	"context"
	"testing"

	systemModel "optimus/server/model/system"

	"github.com/stretchr/testify/require"
)

func TestMenuPermissionCodesForRolesOrdersDistinctCodesMySQL(t *testing.T) {
	db := openSeedTestDB(t)

	require.NoError(t, AutoMigrateSystem(db))

	adminRole := systemModel.SystemRole{Code: "admin", Name: "Administrator", Status: defaultRoleStatus}
	opsRole := systemModel.SystemRole{Code: "ops", Name: "Ops", Status: defaultRoleStatus}
	require.NoError(t, db.Create(&[]systemModel.SystemRole{adminRole, opsRole}).Error)
	require.NoError(t, db.Where("code = ?", "admin").First(&adminRole).Error)
	require.NoError(t, db.Where("code = ?", "ops").First(&opsRole).Error)

	first := systemModel.SystemPermission{Code: "system.users.menu", Name: "Users", Type: "menu", Sort: 20}
	second := systemModel.SystemPermission{Code: "dashboard.view", Name: "Dashboard", Type: "menu", Sort: 10}
	apiPermission := systemModel.SystemPermission{Code: "system.users.view", Name: "Users API", Type: "api", Sort: 5}
	require.NoError(t, db.Create(&[]systemModel.SystemPermission{first, second, apiPermission}).Error)

	var permissions []systemModel.SystemPermission
	require.NoError(t, db.Find(&permissions).Error)
	for _, permission := range permissions {
		require.NoError(t, db.Create(&systemModel.SystemRolePermission{RoleID: adminRole.ID, PermissionID: permission.ID}).Error)
		require.NoError(t, db.Create(&systemModel.SystemRolePermission{RoleID: opsRole.ID, PermissionID: permission.ID}).Error)
	}

	codes, err := NewGormAuthStore(db).MenuPermissionCodesForRoles(context.Background(), []uint{adminRole.ID, opsRole.ID})

	require.NoError(t, err)
	require.Equal(t, []string{"dashboard.view", "system.users.menu"}, codes)
}
