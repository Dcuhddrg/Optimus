package initialize

import (
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"optimus/server/config"
	"testing"
)

func TestSystemServiceGroupWithoutDatabase(t *testing.T) {
	group := newSystemServiceGroup(nil, config.Config{})
	require.Nil(t, group.AuthService)
	require.Nil(t, group.UserService)
	require.Nil(t, group.RoleService)
	require.Nil(t, group.LDAPService)
	require.Nil(t, group.AuditService)
	require.NotNil(t, group.Status)
	require.Nil(t, group.RBAC)
}

func TestSystemServiceGroupBuildsDependenciesWhenDatabaseExists(t *testing.T) {
	group := newSystemServiceGroup(&gorm.DB{}, config.Config{})
	require.NotNil(t, group.AuthService)
	require.NotNil(t, group.UserService)
	require.NotNil(t, group.RoleService)
	require.NotNil(t, group.LDAPService)
	require.NotNil(t, group.AuditService)
	require.NotNil(t, group.Status)
	require.NotNil(t, group.RBAC)
}
