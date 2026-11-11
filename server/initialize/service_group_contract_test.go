package initialize

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"optimus/server/config"
	"optimus/server/global"
	"optimus/server/service"
	cloudService "optimus/server/service/cloud"
	projectService "optimus/server/service/project"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// GVA services are stateless values: shared dependencies (DB, config, clients)
// are read from global at call time, so API aliases copied at package init stay
// valid. A field here would reintroduce startup assembly.
func TestGroupedServicesAreStateless(t *testing.T) {
	for _, group := range []any{systemService.ServiceGroup{}, cloudService.ServiceGroup{}, projectService.ServiceGroup{}} {
		typ := reflect.TypeOf(group)
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			require.True(t, f.Anonymous, "%s.%s", typ, f.Name)
			require.Zero(t, f.Type.NumField(), "%s.%s must not hold state; read dependencies from global at call time", typ, f.Name)
		}
	}
}

func TestRealRouterReadsCurrentGroupAuthorizer(t *testing.T) {
	saved, cfg := *service.ServiceGroupApp, global.OPTIMUS_CONFIG
	t.Cleanup(func() { *service.ServiceGroupApp = saved; global.OPTIMUS_CONFIG = cfg })
	global.OPTIMUS_CONFIG.System.RouterPrefix = "/api/v1"
	global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "group-contract-key", ExpiresTime: "1h", Issuer: "group-contract"}
	db, _ := projectScopeRouterDB(t, "", true)
	setTestDatabase(t, db)
	token, err := utils.NewJWT().CreateToken(utils.NewJWT().CreateClaims(utils.BaseClaims{ID: 1, RoleIDs: []uint{1}}))
	require.NoError(t, err)
	for _, allowed := range []bool{false, true} {
		if allowed {
			grantAPI(t, db, 1, "GET", "/system/permissions/tree")
		} else {
			revokeAPI(t, db, 1, "GET", "/system/permissions/tree")
		}
		req := httptest.NewRequest("GET", "/api/v1/system/permissions/tree", nil)
		req.Header.Set("x-token", token)
		w := httptest.NewRecorder()
		Routers().ServeHTTP(w, req)
		require.Equal(t, 200, w.Code)
		if allowed {
			require.Contains(t, w.Body.String(), `"code":0`)
		} else {
			require.JSONEq(t, `{"code":7,"data":{},"msg":"permission denied"}`, w.Body.String())
		}
	}
}

// Without a database JWTAuth cannot load the user's current state, so it fails
// closed with a business error instead of reaching the authorizer (design 3.1).
func TestMissingDatabaseFailsClosedBeforeAuthorization(t *testing.T) {
	saved, cfg := *service.ServiceGroupApp, global.OPTIMUS_CONFIG
	t.Cleanup(func() { *service.ServiceGroupApp = saved; global.OPTIMUS_CONFIG = cfg })
	global.OPTIMUS_CONFIG.System.RouterPrefix = "/api/v1"
	global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "missing-db-contract", ExpiresTime: "1h", Issuer: "missing-db-contract"}
	setTestDatabase(t, nil)
	token, err := utils.NewJWT().CreateToken(utils.NewJWT().CreateClaims(utils.BaseClaims{ID: 1, RoleIDs: []uint{1}}))
	require.NoError(t, err)
	req := httptest.NewRequest("GET", "/api/v1/system/permissions/tree", nil)
	req.Header.Set("x-token", token)
	w := httptest.NewRecorder()
	Routers().ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)
	require.JSONEq(t, `{"code":7,"data":{},"msg":"failed to load user state"}`, w.Body.String())
}

func setTestDatabase(t *testing.T, db *gorm.DB) {
	t.Helper()
	previousDB := global.OPTIMUS_DB
	t.Cleanup(func() { global.OPTIMUS_DB = previousDB; utils.ResetCasbin(); systemService.InvalidateAllAuthState() })
	global.OPTIMUS_DB = db
	utils.ResetCasbin()
	systemService.InvalidateAllAuthState()
}
