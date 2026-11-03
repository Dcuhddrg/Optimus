package initialize

import (
	"testing"

	"optimus/server/config"
	systemService "optimus/server/service/system"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestSystemServiceGroupWithoutDatabase(t *testing.T) {
	group := newSystemServiceGroup(nil, config.Config{})
	require.Equal(t, systemService.ServiceGroup{}, group)
}

func TestSystemServiceGroupBuildsDependenciesWhenDatabaseExists(t *testing.T) {
	group := newSystemServiceGroup(&gorm.DB{}, config.Config{})
	require.NotEqual(t, systemService.AuthService{}, group.AuthService)
	require.NotEqual(t, systemService.LDAPService{}, group.LDAPService)
	require.Equal(t, systemService.UserService{}, group.UserService)
	require.Equal(t, systemService.RoleService{}, group.RoleService)
	require.Equal(t, systemService.AuditService{}, group.AuditService)
}
