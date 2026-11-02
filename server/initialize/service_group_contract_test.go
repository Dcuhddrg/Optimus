package initialize

import (
	"net/http/httptest"
	"testing"
	"time"

	api "optimus/server/api/v1"
	"optimus/server/config"
	"optimus/server/global"
	"optimus/server/service"
	systemService "optimus/server/service/system"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestServiceGroupInitializationKeepsAliasAddressesStable(t *testing.T) {
	saved := *service.ServiceGroupApp
	t.Cleanup(func() { *service.ServiceGroupApp = saved })
	root, system := service.ServiceGroupApp, &service.ServiceGroupApp.SystemServiceGroup
	authAPI := &api.ApiGroupApp.SystemApiGroup.AuthApi
	InitializeServices(nil, config.Config{})
	require.Same(t, root, service.ServiceGroupApp)
	require.Same(t, system, &service.ServiceGroupApp.SystemServiceGroup)
	require.Nil(t, system.AuthService)
	require.NotNil(t, system.Status)
	InitializeServices(&gorm.DB{}, config.Config{})
	require.Same(t, system, &service.ServiceGroupApp.SystemServiceGroup)
	require.Same(t, authAPI, &api.ApiGroupApp.SystemApiGroup.AuthApi)
	require.NotNil(t, system.AuthService)
	require.NotNil(t, service.ServiceGroupApp.CloudServiceGroup.CloudAccountService)
	require.NotNil(t, service.ServiceGroupApp.CloudServiceGroup.AssetRegistryService)
	require.NotNil(t, service.ServiceGroupApp.ProjectServiceGroup.ProjectService)
}

func TestRealRouterReadsCurrentGroupAuthorizer(t *testing.T) {
	saved, cfg := *service.ServiceGroupApp, global.OPTIMUS_CONFIG
	t.Cleanup(func() { *service.ServiceGroupApp = saved; global.OPTIMUS_CONFIG = cfg })
	global.OPTIMUS_CONFIG.System.RouterPrefix = "/api/v1"
	global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "group-contract-key", ExpiresTime: 3600, Issuer: "group-contract"}
	InitializeServices(nil, global.OPTIMUS_CONFIG)
	token, err := systemService.IssueJWT(systemService.TokenSubject{ID: 1, RoleIDs: []uint{1}}, global.OPTIMUS_CONFIG.JWT, time.Now())
	require.NoError(t, err)
	for _, allowed := range []bool{false, true} {
		service.ServiceGroupApp.SystemServiceGroup.RBAC = fakeRouterRBAC{allowed: allowed}
		req := httptest.NewRequest("GET", "/api/v1/system/permissions/tree", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		Routers().ServeHTTP(w, req)
		if allowed {
			require.Equal(t, 200, w.Code)
		} else {
			require.Equal(t, 403, w.Code)
		}
	}
}

// Keep the existing typed-nil authorizer recovery behavior in this structural
// batch. Changing missing-database authorization policy is a separate fix.
func TestMissingDatabaseAuthorizerPreservesRecoveryStatus(t *testing.T) {
	saved, cfg := *service.ServiceGroupApp, global.OPTIMUS_CONFIG
	t.Cleanup(func() { *service.ServiceGroupApp = saved; global.OPTIMUS_CONFIG = cfg })
	global.OPTIMUS_CONFIG.System.RouterPrefix = "/api/v1"
	global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "missing-db-contract", ExpiresTime: 3600, Issuer: "missing-db-contract"}
	InitializeServices(nil, global.OPTIMUS_CONFIG)
	token, err := systemService.IssueJWT(systemService.TokenSubject{ID: 1, RoleIDs: []uint{1}}, global.OPTIMUS_CONFIG.JWT, time.Now())
	require.NoError(t, err)
	req := httptest.NewRequest("GET", "/api/v1/system/permissions/tree", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	Routers().ServeHTTP(w, req)
	require.Equal(t, 500, w.Code)
	require.Empty(t, w.Body.String())
}
