package initialize

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"optimus/server/config"
	"optimus/server/global"
	systemModel "optimus/server/model/system"
	"optimus/server/service"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRoutersUseGroupedAuthService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	previousServices := *service.ServiceGroupApp
	t.Cleanup(func() {
		*service.ServiceGroupApp = previousServices
	})

	db, _ := projectScopeRouterDB(t, "", true)
	setTestDatabase(t, db)

	engine := Routers()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/system/auth/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	// GVA: a failed login is a business failure, not an unauthenticated request.
	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"code":7,"data":{},"msg":"username or password is incorrect"}`, recorder.Body.String())
}

func TestRouterLoginUsesInitializedAuthDependencies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	previousServices, previousConfig := *service.ServiceGroupApp, global.OPTIMUS_CONFIG
	t.Cleanup(func() { *service.ServiceGroupApp = previousServices; global.OPTIMUS_CONFIG = previousConfig })
	global.OPTIMUS_CONFIG.System.RouterPrefix = "/api/v1"
	global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "router-login-test", ExpiresTime: 3600, Issuer: "router-login-test"}
	db, _ := projectScopeRouterDB(t, "", true)
	hash, err := utils.BcryptHash("admin-secret")
	require.NoError(t, err)
	user := systemModel.SystemUser{UUID: "router-admin", Username: "admin", DisplayName: "Admin", Source: "local", Status: "enabled", PasswordHash: hash}
	require.NoError(t, db.Create(&user).Error)
	setTestDatabase(t, db)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/system/auth/login", strings.NewReader(`{"username":"admin","password":"admin-secret"}`))
	req.Header.Set("Content-Type", "application/json")
	Routers().ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var body struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Zero(t, body.Code)
	require.NotEmpty(t, body.Data.Token)
}

func TestRoutersUseGroupedRBACCapability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalConfig := global.OPTIMUS_CONFIG
	previousServices := *service.ServiceGroupApp
	t.Cleanup(func() {
		global.OPTIMUS_CONFIG = originalConfig
		*service.ServiceGroupApp = previousServices
	})

	global.OPTIMUS_CONFIG.JWT = config.JWT{
		SigningKey:  "router-test-signing-key",
		ExpiresTime: 3600,
		BufferTime:  300,
		Issuer:      "optimus-router-test",
	}
	db, _ := projectScopeRouterDB(t, "", true)
	setTestDatabase(t, db)

	engine := Routers()
	token, err := utils.NewJWT().CreateToken(utils.NewJWT().CreateClaims(utils.BaseClaims{
		ID:        1,
		UUID:      "router-test-user",
		Username:  "admin",
		RoleIDs:   []uint{1},
		RoleCodes: []string{"admin"},
	}))
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/permissions/tree", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), "\"code\":\"system.permissions.view\"")
}
