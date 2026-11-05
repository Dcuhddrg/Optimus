package initialize

import (
	"strconv"
	"testing"

	systemService "optimus/server/service/system"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// grantAPI gives roleID a Casbin API policy (path without router prefix) and
// reloads the enforcer, as role permission updates do.
func grantAPI(t *testing.T, db *gorm.DB, roleID uint, method, path string) {
	t.Helper()
	require.NoError(t, systemService.CasbinServiceApp.AddPolicies(db, roleID, []systemService.CasbinInfo{{Path: path, Method: method}}))
	require.NoError(t, systemService.CasbinServiceApp.FreshCasbin())
}

func revokeAPI(t *testing.T, db *gorm.DB, roleID uint, method, path string) {
	t.Helper()
	require.NoError(t, db.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", strconv.Itoa(int(roleID)), path, method).Delete(&gormadapter.CasbinRule{}).Error)
	require.NoError(t, systemService.CasbinServiceApp.FreshCasbin())
}
