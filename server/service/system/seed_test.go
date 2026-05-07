package system

import (
	"os"
	"testing"

	systemModel "optimus/server/model/system"

	"github.com/stretchr/testify/require"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestDefaultSystemRolesAndPermissions(t *testing.T) {
	roles := DefaultSystemRoles()
	require.Contains(t, roleCodes(roles), "admin")
	require.Contains(t, roleCodes(roles), "user")

	permissions := DefaultSystemPermissions()
	require.Contains(t, permissionCodes(permissions), "system.users.view")
	require.Contains(t, permissionCodes(permissions), "system.roles.permissions.update")
	require.Contains(t, permissionCodes(permissions), "system.ldap.update")
	require.Contains(t, permissionCodes(permissions), "system.audit.view")
}

func TestHashPasswordAndComparePassword(t *testing.T) {
	hash, err := HashPassword("admin-secret")
	require.NoError(t, err)
	require.NotEqual(t, "admin-secret", hash)
	require.True(t, ComparePassword(hash, "admin-secret"))
	require.False(t, ComparePassword(hash, "wrong-secret"))
}

func TestSeedSystemDataCreatesAdminRolePermissionsAndUserMySQL(t *testing.T) {
	db := openSeedTestDB(t)

	require.NoError(t, AutoMigrateSystem(db))
	require.NoError(t, SeedSystemData(db, "admin-secret"))

	var admin systemModel.SystemUser
	require.NoError(t, db.Where("username = ?", "admin").First(&admin).Error)
	require.Equal(t, "local", admin.Source)
	require.True(t, ComparePassword(admin.PasswordHash, "admin-secret"))

	var adminRole systemModel.SystemRole
	require.NoError(t, db.Where("code = ?", "admin").First(&adminRole).Error)

	var rolePermissionCount int64
	require.NoError(t, db.Model(&systemModel.SystemRolePermission{}).
		Where("role_id = ?", adminRole.ID).
		Count(&rolePermissionCount).Error)
	require.Equal(t, int64(len(DefaultSystemPermissions())), rolePermissionCount)
}

func TestSeedSystemDataIsIdempotentMySQL(t *testing.T) {
	db := openSeedTestDB(t)

	require.NoError(t, AutoMigrateSystem(db))
	require.NoError(t, SeedSystemData(db, "admin-secret"))
	require.NoError(t, SeedSystemData(db, "changed-secret"))

	var users int64
	require.NoError(t, db.Model(&systemModel.SystemUser{}).
		Where("username = ?", "admin").
		Count(&users).Error)
	require.Equal(t, int64(1), users)

	var admin systemModel.SystemUser
	require.NoError(t, db.Where("username = ?", "admin").First(&admin).Error)
	require.True(t, ComparePassword(admin.PasswordHash, "admin-secret"))
	require.False(t, ComparePassword(admin.PasswordHash, "changed-secret"))
}

func TestSeedSystemDataDoesNotOverwriteRolePermissionEditsMySQL(t *testing.T) {
	db := openSeedTestDB(t)

	require.NoError(t, AutoMigrateSystem(db))
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

func openSeedTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("OPTIMUS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OPTIMUS_TEST_MYSQL_DSN is not set")
	}

	db, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_role_permissions").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_user_roles").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_permissions").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_roles").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_users").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_ldap_config").Error)
	require.NoError(t, db.Exec("DROP TABLE IF EXISTS sys_audit_logs").Error)

	return db
}

func roleCodes(roles []systemModel.SystemRole) []string {
	codes := make([]string, 0, len(roles))
	for _, role := range roles {
		codes = append(codes, role.Code)
	}
	return codes
}

func permissionCodes(permissions []systemModel.SystemPermission) []string {
	codes := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		codes = append(codes, permission.Code)
	}
	return codes
}
