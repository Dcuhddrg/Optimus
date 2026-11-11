package system

import (
	"errors"
	"os"
	"strconv"
	"testing"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// resetTestDB opens OPTIMUS_TEST_MYSQL_DSN (reserved for tests) and drops the system tables.
func resetTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("OPTIMUS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OPTIMUS_TEST_MYSQL_DSN is not set")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	for _, table := range []string{"sys_role_permissions", "sys_user_roles", "sys_permissions", "sys_roles", "sys_users", "sys_ldap_config", "sys_audit_logs", "jwt_blacklists", "casbin_rule"} {
		require.NoError(t, db.Exec("DROP TABLE IF EXISTS `"+table+"`").Error)
	}
	return db
}

func resetAndMigrateTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := resetTestDB(t)
	require.NoError(t, systemService.AutoMigrateSystem(db))
	previous := global.OPTIMUS_DB
	t.Cleanup(func() { global.OPTIMUS_DB = previous; utils.ResetCasbin() })
	global.OPTIMUS_DB = db
	utils.ResetCasbin()
	return db
}

func TestSeedSystemDataCreatesAdminRolePermissionsAndUserMySQL(t *testing.T) {
	db := resetTestDB(t)

	require.NoError(t, systemService.AutoMigrateSystem(db))
	require.NoError(t, SeedSystemData(db, "admin-secret"))

	var admin systemModel.SystemUser
	require.NoError(t, db.Where("username = ?", "admin").First(&admin).Error)
	require.Equal(t, "local", admin.Source)
	require.True(t, utils.BcryptCheck("admin-secret", admin.PasswordHash))

	var adminRole systemModel.SystemRole
	require.NoError(t, db.Where("code = ?", "admin").First(&adminRole).Error)

	// Menus are granted in sys_role_permissions, APIs as Casbin policies.
	var menus, apis int64
	for _, permission := range systemService.DefaultSystemPermissions() {
		if permission.Type == "api" {
			apis++
		} else {
			menus++
		}
	}
	var menuGrants, policies int64
	require.NoError(t, db.Model(&systemModel.SystemRolePermission{}).Where("role_id = ?", adminRole.ID).Count(&menuGrants).Error)
	require.NoError(t, db.Model(&gormadapter.CasbinRule{}).Where("ptype = ? AND v0 = ?", "p", strconv.Itoa(int(adminRole.ID))).Count(&policies).Error)
	require.Equal(t, menus, menuGrants)
	require.Equal(t, apis, policies)
}

func TestSeedSystemDataIsIdempotentMySQL(t *testing.T) {
	db := resetTestDB(t)

	require.NoError(t, systemService.AutoMigrateSystem(db))
	require.NoError(t, SeedSystemData(db, "admin-secret"))
	require.NoError(t, SeedSystemData(db, "changed-secret"))

	var users int64
	require.NoError(t, db.Model(&systemModel.SystemUser{}).
		Where("username = ?", "admin").
		Count(&users).Error)
	require.Equal(t, int64(1), users)

	var admin systemModel.SystemUser
	require.NoError(t, db.Where("username = ?", "admin").First(&admin).Error)
	require.True(t, utils.BcryptCheck("admin-secret", admin.PasswordHash))
	require.False(t, utils.BcryptCheck("changed-secret", admin.PasswordHash))
}

func TestSeedSystemDataDoesNotOverwriteRolePermissionEditsMySQL(t *testing.T) {
	db := resetTestDB(t)

	require.NoError(t, systemService.AutoMigrateSystem(db))
	require.NoError(t, SeedSystemData(db, "admin-secret"))

	var userRole systemModel.SystemRole
	require.NoError(t, db.Where("code = ?", "user").First(&userRole).Error)
	require.NoError(t, db.Where("role_id = ?", userRole.ID).Delete(&systemModel.SystemRolePermission{}).Error)
	require.NoError(t, SeedSystemData(db, "admin-secret"))

	var rolePermissionCount int64
	require.NoError(t, db.Model(&systemModel.SystemRolePermission{}).
		Where("role_id = ?", userRole.ID).
		Count(&rolePermissionCount).Error)
	require.Equal(t, int64(0), rolePermissionCount)
}

// Databases from before #56 store API grants in sys_role_permissions with
// /api/v1-prefixed paths; startup moves them to casbin_rule, idempotently.
func TestMigrateAPIGrantsToCasbinUpgradesLegacyGrants(t *testing.T) {
	db := resetAndMigrateTestDB(t)
	role := systemModel.SystemRole{Code: "legacy", Name: "legacy", Status: "enabled"}
	require.NoError(t, db.Create(&role).Error)
	api := systemModel.SystemPermission{Code: "legacy.users", Name: "users", Type: "api", Method: "PUT", Path: "/api/v1/system/users/:id"}
	menu := systemModel.SystemPermission{Code: "legacy.menu", Name: "menu", Type: "menu", Path: "/system/users"}
	require.NoError(t, db.Create(&[]systemModel.SystemPermission{api, menu}).Error)
	require.NoError(t, db.Where("code = ?", api.Code).First(&api).Error)
	require.NoError(t, db.Where("code = ?", menu.Code).First(&menu).Error)
	require.NoError(t, db.Create(&[]systemModel.SystemRolePermission{{RoleID: role.ID, PermissionID: api.ID}, {RoleID: role.ID, PermissionID: menu.ID}}).Error)

	for run := 0; run < 2; run++ {
		require.NoError(t, migrateAPIGrantsToCasbin(db))
		var migrated systemModel.SystemPermission
		require.NoError(t, db.First(&migrated, api.ID).Error)
		require.Equal(t, "/system/users/:id", migrated.Path)
		var rows []systemModel.SystemRolePermission
		require.NoError(t, db.Where("role_id = ?", role.ID).Find(&rows).Error)
		require.Len(t, rows, 1, "only the menu grant stays in sys_role_permissions")
		require.Equal(t, menu.ID, rows[0].PermissionID)
		var policies []gormadapter.CasbinRule
		require.NoError(t, db.Where("ptype = ? AND v0 = ?", "p", strconv.Itoa(int(role.ID))).Find(&policies).Error)
		require.Len(t, policies, 1)
		require.Equal(t, "/system/users/:id", policies[0].V1)
		require.Equal(t, "PUT", policies[0].V2)
	}
}

// Rules before #80 allowed disabling the built-in admin while other admins
// existed; startup restores it as the permanent fallback (design 3.4).
func TestSeedSystemDataRepairsBuiltinAdminMySQL(t *testing.T) {
	db := resetTestDB(t)
	require.NoError(t, systemService.AutoMigrateSystem(db))
	require.NoError(t, SeedSystemData(db, "admin-secret"))
	var audits int64
	require.NoError(t, db.Model(&systemModel.SystemAuditLog{}).Where("action = ?", "system.admin.repair").Count(&audits).Error)
	require.Zero(t, audits, "a fresh install is not a repair")

	var admin systemModel.SystemUser
	require.NoError(t, db.Where("uuid = ?", systemService.BuiltinAdminUUID).First(&admin).Error)
	var adminRole systemModel.SystemRole
	require.NoError(t, db.Where("code = ?", "admin").First(&adminRole).Error)
	var usersView systemModel.SystemPermission
	require.NoError(t, db.Where("code = ?", "system.users.menu").First(&usersView).Error)

	require.NoError(t, db.Model(&admin).Update("status", "disabled").Error)
	require.NoError(t, db.Where("user_id = ? AND role_id = ?", admin.ID, adminRole.ID).Delete(&systemModel.SystemUserRole{}).Error)
	require.NoError(t, db.Where("role_id = ? AND permission_id = ?", adminRole.ID, usersView.ID).Delete(&systemModel.SystemRolePermission{}).Error)
	require.NoError(t, db.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", strconv.Itoa(int(adminRole.ID)), "/system/users/:id/roles", "PUT").Delete(&gormadapter.CasbinRule{}).Error)

	require.NoError(t, SeedSystemData(db, "ignored"))

	require.NoError(t, db.First(&admin, admin.ID).Error)
	require.Equal(t, "enabled", admin.Status)
	var binding int64
	require.NoError(t, db.Model(&systemModel.SystemUserRole{}).Where("user_id = ? AND role_id = ?", admin.ID, adminRole.ID).Count(&binding).Error)
	require.Equal(t, int64(1), binding)
	var restoredMenu, restoredAPI int64
	require.NoError(t, db.Model(&systemModel.SystemRolePermission{}).Where("role_id = ? AND permission_id = ?", adminRole.ID, usersView.ID).Count(&restoredMenu).Error)
	require.NoError(t, db.Model(&gormadapter.CasbinRule{}).Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", strconv.Itoa(int(adminRole.ID)), "/system/users/:id/roles", "PUT").Count(&restoredAPI).Error)
	require.Equal(t, int64(1), restoredMenu)
	require.Equal(t, int64(1), restoredAPI)

	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Where("action = ?", "system.admin.repair").First(&audit).Error)
	require.Contains(t, audit.Detail, "re-enabled")
	require.Contains(t, audit.Detail, "admin role restored")
	require.Contains(t, audit.Detail, "system.users.menu")
	require.Contains(t, audit.Detail, "system.users.roles.update")
	require.True(t, utils.BcryptCheck("admin-secret", admin.PasswordHash), "repair keeps the existing password")
}

// Seeds always used the fixed UUID, but a manually edited database may hold a
// local admin with another UUID; it is adopted as the built-in admin.
func TestSeedSystemDataAdoptsLegacyLocalAdminMySQL(t *testing.T) {
	db := resetTestDB(t)
	require.NoError(t, systemService.AutoMigrateSystem(db))
	require.NoError(t, SeedSystemData(db, "admin-secret"))
	require.NoError(t, db.Model(&systemModel.SystemUser{}).Where("uuid = ?", systemService.BuiltinAdminUUID).
		Updates(map[string]any{"uuid": "legacy-admin-uuid", "status": "disabled"}).Error)

	require.NoError(t, SeedSystemData(db, "ignored"))

	var admin systemModel.SystemUser
	require.NoError(t, db.Where("username = ?", "admin").First(&admin).Error)
	require.Equal(t, systemService.BuiltinAdminUUID, admin.UUID)
	require.Equal(t, "enabled", admin.Status)
	var count int64
	require.NoError(t, db.Model(&systemModel.SystemUser{}).Where("username = ?", "admin").Count(&count).Error)
	require.Equal(t, int64(1), count)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Where("action = ?", "system.admin.repair").First(&audit).Error)
	require.Contains(t, audit.Detail, "adopted legacy local admin")
	require.Contains(t, audit.Detail, "re-enabled")
}

func TestSeedSystemDataRefusesNonLocalAdminMySQL(t *testing.T) {
	db := resetTestDB(t)
	require.NoError(t, systemService.AutoMigrateSystem(db))
	require.NoError(t, db.Create(&systemModel.SystemUser{UUID: "ldap:admin", Username: "admin", Source: "ldap", Status: "enabled"}).Error)

	err := SeedSystemData(db, "admin-secret")

	require.ErrorContains(t, err, "blocks the built-in admin")
	require.ErrorContains(t, err, "UPDATE sys_users SET username")
}

// Permissions revoked from the admin role stay revoked after a restart;
// only system.* is restored.
func TestSeedSystemDataKeepsRevokedNonSystemPermissionsMySQL(t *testing.T) {
	db := resetTestDB(t)
	require.NoError(t, systemService.AutoMigrateSystem(db))
	require.NoError(t, SeedSystemData(db, "admin-secret"))
	var adminRole systemModel.SystemRole
	require.NoError(t, db.Where("code = ?", "admin").First(&adminRole).Error)
	var dashboard systemModel.SystemPermission
	require.NoError(t, db.Where("code = ?", "dashboard.view").First(&dashboard).Error)
	require.NoError(t, db.Where("role_id = ? AND permission_id = ?", adminRole.ID, dashboard.ID).Delete(&systemModel.SystemRolePermission{}).Error)

	require.NoError(t, SeedSystemData(db, "admin-secret"))

	var grants int64
	require.NoError(t, db.Model(&systemModel.SystemRolePermission{}).Where("role_id = ? AND permission_id = ?", adminRole.ID, dashboard.ID).Count(&grants).Error)
	require.Zero(t, grants)
}

// Startup fails between creating the catalog and granting it; the next start
// still grants the non-system permissions to the admin role.
func TestSeedSystemDataGrantsCatalogAfterFailedStartMySQL(t *testing.T) {
	db := resetTestDB(t)
	require.NoError(t, systemService.AutoMigrateSystem(db))
	fail := true
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("test:fail_first_grant", func(tx *gorm.DB) {
		if fail && tx.Statement.Table == "casbin_rule" {
			tx.AddError(errors.New("grant failed"))
		}
	}))
	t.Cleanup(func() { _ = db.Callback().Create().Remove("test:fail_first_grant") })

	require.Error(t, SeedSystemData(db, "admin-secret"))
	fail = false
	require.NoError(t, SeedSystemData(db, "admin-secret"))

	var adminRole systemModel.SystemRole
	require.NoError(t, db.Where("code = ?", "admin").First(&adminRole).Error)
	var dashboard systemModel.SystemPermission
	require.NoError(t, db.Where("code = ?", "dashboard.view").First(&dashboard).Error)
	var grants int64
	require.NoError(t, db.Model(&systemModel.SystemRolePermission{}).Where("role_id = ? AND permission_id = ?", adminRole.ID, dashboard.ID).Count(&grants).Error)
	require.Equal(t, int64(1), grants)
}
