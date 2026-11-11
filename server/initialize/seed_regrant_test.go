package initialize

import (
	"strconv"
	"testing"

	"optimus/server/global"
	systemModel "optimus/server/model/system"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/stretchr/testify/require"
)

// An administrator revokes cloud and project permissions from the admin role;
// a restart (InitializeData) must not grant them back to every admin holder.
func TestRestartKeepsRevokedModulePermissions(t *testing.T) {
	db, _ := projectScopeRouterDB(t, "", false)
	require.NoError(t, db.Exec("DELETE FROM sys_user_roles").Error)
	require.NoError(t, db.Exec("DELETE FROM sys_users").Error)
	require.NoError(t, db.Exec("DELETE FROM sys_roles").Error)
	previous := global.OPTIMUS_CONFIG
	t.Cleanup(func() { global.OPTIMUS_CONFIG = previous })
	global.OPTIMUS_CONFIG.System.AdminInitialPassword = "admin-secret"
	require.NoError(t, InitializeData(db))

	var adminRole systemModel.SystemRole
	require.NoError(t, db.Where("code = ?", "admin").First(&adminRole).Error)
	sub := strconv.Itoa(int(adminRole.ID))
	var revoked []systemModel.SystemPermission
	require.NoError(t, db.Where("type = ? AND (code LIKE ? OR code LIKE ?)", "api", "cloud.%", "project%").Find(&revoked).Error)
	require.NotEmpty(t, revoked)
	for _, permission := range revoked {
		require.NoError(t, db.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", sub, permission.Path, permission.Method).Delete(&gormadapter.CasbinRule{}).Error)
	}
	var usersView systemModel.SystemPermission
	require.NoError(t, db.Where("code = ?", "system.users.view").First(&usersView).Error)
	require.NoError(t, db.Where("ptype = ? AND v0 = ? AND v1 = ?", "p", sub, usersView.Path).Delete(&gormadapter.CasbinRule{}).Error)

	require.NoError(t, InitializeData(db))

	for _, permission := range revoked {
		var count int64
		require.NoError(t, db.Model(&gormadapter.CasbinRule{}).Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", sub, permission.Path, permission.Method).Count(&count).Error)
		require.Zero(t, count, permission.Code)
	}
	var restored int64
	require.NoError(t, db.Model(&gormadapter.CasbinRule{}).Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", sub, usersView.Path, usersView.Method).Count(&restored).Error)
	require.Equal(t, int64(1), restored, "system.* is always restored")
}
