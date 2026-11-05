package system

import (
	"testing"

	systemModel "optimus/server/model/system"

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
