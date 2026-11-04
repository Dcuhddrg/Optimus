package system

import (
	"testing"

	systemModel "optimus/server/model/system"
	"optimus/server/utils"

	"github.com/stretchr/testify/require"
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

func TestSeedSystemDataCreatesAdminRolePermissionsAndUserMySQL(t *testing.T) {
	db := resetTestDB(t)

	require.NoError(t, AutoMigrateSystem(db))
	require.NoError(t, SeedSystemData(db, "admin-secret"))

	var admin systemModel.SystemUser
	require.NoError(t, db.Where("username = ?", "admin").First(&admin).Error)
	require.Equal(t, "local", admin.Source)
	require.True(t, utils.BcryptCheck("admin-secret", admin.PasswordHash))

	var adminRole systemModel.SystemRole
	require.NoError(t, db.Where("code = ?", "admin").First(&adminRole).Error)

	var rolePermissionCount int64
	require.NoError(t, db.Model(&systemModel.SystemRolePermission{}).
		Where("role_id = ?", adminRole.ID).
		Count(&rolePermissionCount).Error)
	require.Equal(t, int64(len(DefaultSystemPermissions())), rolePermissionCount)
}

func TestSeedSystemDataIsIdempotentMySQL(t *testing.T) {
	db := resetTestDB(t)

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
	require.True(t, utils.BcryptCheck("admin-secret", admin.PasswordHash))
	require.False(t, utils.BcryptCheck("changed-secret", admin.PasswordHash))
}

func TestSeedSystemDataDoesNotOverwriteRolePermissionEditsMySQL(t *testing.T) {
	db := resetTestDB(t)

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
