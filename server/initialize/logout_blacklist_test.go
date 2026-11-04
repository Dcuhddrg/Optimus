package initialize

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"optimus/server/config"
	"optimus/server/global"
	"optimus/server/model/common/response"
	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/songzhibin97/gkit/cache/local_cache"
	"github.com/stretchr/testify/require"
)

// Logout revokes the token like GVA's JsonInBlacklist; the revocation survives
// a restart because LoadAll reloads persisted entries into BlackCache.
func TestLogoutRevokesTokenAcrossRestart(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousConfig, previousCache := global.OPTIMUS_CONFIG, global.BlackCache
	t.Cleanup(func() { global.OPTIMUS_CONFIG = previousConfig; global.BlackCache = previousCache })
	global.OPTIMUS_CONFIG.System.RouterPrefix = "/api/v1"
	global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "logout-test", ExpiresTime: "1h", BufferTime: "5m", Issuer: "logout-test"}
	global.BlackCache = local_cache.NewCache()

	db, _ := projectScopeRouterDB(t, "", true)
	require.NoError(t, db.Create(&systemModel.SystemRole{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 1}, Code: "ops", Name: "ops", Status: "enabled"}).Error)
	hash, err := utils.BcryptHash("admin-secret")
	require.NoError(t, err)
	user := systemModel.SystemUser{UUID: "logout-admin", Username: "admin", DisplayName: "Admin", Source: "local", Status: "enabled", PasswordHash: hash}
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: user.ID, RoleID: 1}).Error)
	logout := systemModel.SystemPermission{Code: "logout", Name: "logout", Type: "api", Method: "POST", Path: "/api/v1/system/auth/logout"}
	require.NoError(t, db.Create(&logout).Error)
	require.NoError(t, db.Create(&systemModel.SystemRolePermission{RoleID: 1, PermissionID: logout.ID}).Error)
	setTestDatabase(t, db)

	call := func(method, path, token, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("x-token", token)
		}
		Routers().ServeHTTP(w, req)
		return w
	}
	w := call(http.MethodPost, "/system/auth/login", "", `{"username":"admin","password":"admin-secret"}`)
	var login struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &login), w.Body.String())
	require.Equal(t, response.SUCCESS, login.Code, w.Body.String())
	require.Contains(t, w.Header().Get("Set-Cookie"), "x-token="+login.Data.Token, "login writes the x-token cookie")
	token := login.Data.Token

	w = call(http.MethodGet, "/system/permissions/tree", token, "")
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"code":0`)

	w = call(http.MethodPost, "/system/auth/logout", token, "")
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"code":0`, w.Body.String())
	require.Contains(t, w.Header().Get("Set-Cookie"), "x-token=;", "logout clears the cookie")

	w = call(http.MethodGet, "/system/permissions/tree", token, "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Contains(t, w.Body.String(), "token has been revoked")

	global.BlackCache = local_cache.NewCache() // simulated restart: memory is empty
	systemService.LoadAll()
	w = call(http.MethodGet, "/system/permissions/tree", token, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "revocation is reloaded from jwt_blacklists")
}
