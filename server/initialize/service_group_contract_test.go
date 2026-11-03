package initialize

import (
	"net/http/httptest"
	"testing"
	"time"

	api "optimus/server/api/v1"
	cloudAPI "optimus/server/api/v1/cloud"
	systemAPI "optimus/server/api/v1/system"
	"optimus/server/config"
	"optimus/server/global"
	systemModel "optimus/server/model/system"
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
	auth := &system.AuthService
	initializeTestServices(t, nil, config.Config{})
	require.Same(t, root, service.ServiceGroupApp)
	require.Same(t, system, &service.ServiceGroupApp.SystemServiceGroup)
	require.Equal(t, systemService.AuthService{}, system.AuthService)
	db := &gorm.DB{}
	initializeTestServices(t, db, config.Config{})
	require.Same(t, db, global.OPTIMUS_DB)
	require.Same(t, system, &service.ServiceGroupApp.SystemServiceGroup)
	require.Same(t, authAPI, &api.ApiGroupApp.SystemApiGroup.AuthApi)
	require.Same(t, auth, &system.AuthService)
	require.NotEqual(t, systemService.AuthService{}, system.AuthService)
}

func TestRealRouterReadsCurrentGroupAuthorizer(t *testing.T) {
	saved, cfg := *service.ServiceGroupApp, global.OPTIMUS_CONFIG
	t.Cleanup(func() { *service.ServiceGroupApp = saved; global.OPTIMUS_CONFIG = cfg })
	global.OPTIMUS_CONFIG.System.RouterPrefix = "/api/v1"
	global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "group-contract-key", ExpiresTime: 3600, Issuer: "group-contract"}
	db, _ := projectScopeRouterDB(t, "", true)
	initializeTestServices(t, db, global.OPTIMUS_CONFIG)
	var permission systemModel.SystemPermission
	require.NoError(t, db.Where("method = ? AND path = ?", "GET", "/api/v1/system/permissions/tree").First(&permission).Error)
	token, err := systemService.IssueJWT(systemService.TokenSubject{ID: 1, RoleIDs: []uint{1}}, global.OPTIMUS_CONFIG.JWT, time.Now())
	require.NoError(t, err)
	for _, allowed := range []bool{false, true} {
		if allowed {
			require.NoError(t, db.Create(&systemModel.SystemRolePermission{RoleID: 1, PermissionID: permission.ID}).Error)
		} else {
			require.NoError(t, db.Where("role_id = ? AND permission_id = ?", 1, permission.ID).Delete(&systemModel.SystemRolePermission{}).Error)
		}
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
	initializeTestServices(t, nil, global.OPTIMUS_CONFIG)
	token, err := systemService.IssueJWT(systemService.TokenSubject{ID: 1, RoleIDs: []uint{1}}, global.OPTIMUS_CONFIG.JWT, time.Now())
	require.NoError(t, err)
	req := httptest.NewRequest("GET", "/api/v1/system/permissions/tree", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	Routers().ServeHTTP(w, req)
	require.Equal(t, 500, w.Code)
	require.Empty(t, w.Body.String())
}

func initializeTestServices(t *testing.T, db *gorm.DB, cfg config.Config) {
	t.Helper()
	previousDB := global.OPTIMUS_DB
	previousServices := *service.ServiceGroupApp
	t.Cleanup(func() {
		global.OPTIMUS_DB = previousDB
		*service.ServiceGroupApp = previousServices
		systemAPI.RefreshServiceAliases()
		cloudAPI.RefreshServiceAliases()
	})
	InitializeServices(db, cfg)
}
