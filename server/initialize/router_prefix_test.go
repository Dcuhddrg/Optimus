package initialize

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"optimus/server/config"
	"optimus/server/global"
	"optimus/server/model/common/response"
	"optimus/server/service"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// Keep permission keys identical to the real seeds while exercising Routers,
// JWT verification and the real RBAC service against an isolated database.
func TestRoutersAuthorizeCanonicalPermissionWithConfiguredPrefix(t *testing.T) {
	for _, tt := range []struct {
		name, prefix, requestPath, permissionPath, method string
		// GVA: authorized requests are HTTP 200; dropDB makes the handler report
		// its service unavailable as a business failure (code 7).
		code   int
		dropDB bool
	}{
		{"empty", "", "/api/v1/system/permissions/tree", "/system/permissions/tree", http.MethodGet, response.SUCCESS, false},
		{"default", "/api/v1", "/api/v1/system/permissions/tree", "/system/permissions/tree", http.MethodGet, response.SUCCESS, false},
		{"custom", "/custom", "/custom/system/permissions/tree", "/system/permissions/tree", http.MethodGet, response.SUCCESS, false},
		{"without leading slash", "custom", "/custom/system/permissions/tree", "/system/permissions/tree", http.MethodGet, response.SUCCESS, false},
		{"root", "/", "/system/permissions/tree", "/system/permissions/tree", http.MethodGet, response.SUCCESS, false},
		{"trailing slash", "/custom/", "/custom/system/permissions/tree", "/system/permissions/tree", http.MethodGet, response.SUCCESS, false},
		{"parameter route", "/custom", "/custom/system/users/42", "/system/users/:id", http.MethodPut, response.ERROR, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			previousConfig, previousDB := global.OPTIMUS_CONFIG, global.OPTIMUS_DB
			previousServices := *service.ServiceGroupApp
			t.Cleanup(func() {
				global.OPTIMUS_CONFIG, global.OPTIMUS_DB = previousConfig, previousDB
				*service.ServiceGroupApp = previousServices
			})
			db, _ := projectScopeRouterDB(t, "", false)
			global.OPTIMUS_CONFIG.System.RouterPrefix = tt.prefix
			global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "prefix-test-key", ExpiresTime: "1h", Issuer: "prefix-test"}
			setTestDatabase(t, db)
			grantAPI(t, db, 7, tt.method, tt.permissionPath)
			if tt.dropDB {
				// The enforcer is cached, so authorization passes; the handler then
				// reports its database-backed service unavailable.
				require.NotNil(t, utils.GetCasbin())
				global.OPTIMUS_DB = nil
			}
			token, err := utils.NewJWT().CreateToken(utils.NewJWT().CreateClaims(utils.BaseClaims{ID: 1, RoleIDs: []uint{7}}))
			require.NoError(t, err)
			request := httptest.NewRequest(tt.method, tt.requestPath, nil)
			request.Header.Set("x-token", token)
			recorder := httptest.NewRecorder()
			Routers().ServeHTTP(recorder, request)
			global.OPTIMUS_DB = db
			require.Equal(t, http.StatusOK, recorder.Code)
			var body response.Response
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body), recorder.Body.String())
			require.Equal(t, tt.code, body.Code, recorder.Body.String())
		})
	}
}
