package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"optimus/server/config"
	"optimus/server/global"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/songzhibin97/gkit/cache/local_cache"
	"github.com/stretchr/testify/require"
)

func TestJWTAuthRejectsMissingToken(t *testing.T) {
	engine := jwtTestEngine(t)

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/private", nil))

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.JSONEq(t, `{"code":7,"data":null,"msg":"not logged in or invalid access"}`, w.Body.String())
}

func TestJWTAuthRejectsInvalidTokenAndClearsCookie(t *testing.T) {
	engine := jwtTestEngine(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("x-token", "not-a-jwt")
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Contains(t, w.Body.String(), `"code":7`)
	require.Contains(t, w.Header().Get("Set-Cookie"), "x-token=;")
}

func TestJWTAuthAcceptsHeaderToken(t *testing.T) {
	engine := jwtTestEngine(t)
	token := testToken(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("x-token", token)
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "admin")
	require.Empty(t, w.Header().Get("new-token"), "a fresh token is outside the refresh buffer")
}

func TestJWTAuthFallsBackToTokenCookie(t *testing.T) {
	engine := jwtTestEngine(t)
	token := testToken(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.AddCookie(&http.Cookie{Name: "x-token", Value: token})
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "admin")
}

func TestJWTAuthRejectsRevokedToken(t *testing.T) {
	engine := jwtTestEngine(t)
	token := testToken(t)
	global.BlackCache.SetDefault(token, struct{}{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("x-token", token)
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.JSONEq(t, `{"code":7,"data":null,"msg":"token has been revoked"}`, w.Body.String())
}

func TestJWTAuthRefreshesTokenInsideBuffer(t *testing.T) {
	engine := jwtTestEngine(t)
	// Buffer longer than the lifetime: every valid token is due for refresh.
	global.OPTIMUS_CONFIG.JWT.BufferTime = "2h"
	token := testToken(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("x-token", token)
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	newToken := w.Header().Get("new-token")
	require.NotEmpty(t, newToken)
	require.NotEmpty(t, w.Header().Get("new-expires-at"))
	require.Contains(t, w.Header().Get("Set-Cookie"), "x-token="+newToken)
	claims, err := utils.NewJWT().ParseToken(newToken)
	require.NoError(t, err)
	require.Equal(t, "admin", claims.Username)
}

func jwtTestEngine(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)
	originalConfig, originalCache := global.OPTIMUS_CONFIG, global.BlackCache
	t.Cleanup(func() {
		global.OPTIMUS_CONFIG = originalConfig
		global.BlackCache = originalCache
	})
	global.OPTIMUS_CONFIG.JWT = testMiddlewareJWTConfig()
	global.BlackCache = local_cache.NewCache()
	stubAuthState(t, systemService.UserAuthState{Exists: true, Status: systemService.DefaultUserStatus, RoleIDs: []uint{1}}, nil)

	engine := gin.New()
	engine.GET("/private", JWTAuth(), func(c *gin.Context) {
		claims, ok := c.Get(utils.ClaimsContextKey)
		require.True(t, ok)
		c.JSON(http.StatusOK, gin.H{"username": claims.(*utils.CustomClaims).Username})
	})
	return engine
}

func testToken(t *testing.T) string {
	t.Helper()
	j := utils.NewJWT()
	token, err := j.CreateToken(j.CreateClaims(utils.BaseClaims{ID: 9, UUID: "uuid-9", Username: "admin", RoleCodes: []string{"admin"}}))
	require.NoError(t, err)
	return token
}

func testMiddlewareJWTConfig() config.JWT {
	return config.JWT{
		SigningKey:  "middleware-test-signing-key",
		ExpiresTime: "1h",
		BufferTime:  "5m",
		Issuer:      "optimus-test",
	}
}

func stubAuthState(t *testing.T, state systemService.UserAuthState, err error) {
	t.Helper()
	saved := authStateLoader
	t.Cleanup(func() { authStateLoader = saved })
	authStateLoader = func(context.Context, uint) (systemService.UserAuthState, error) { return state, err }
}

func TestJWTAuthRejectsDisabledUserWith401(t *testing.T) {
	engine := jwtTestEngine(t)
	stubAuthState(t, systemService.UserAuthState{Exists: true, Status: "disabled"}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("x-token", testToken(t))
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.JSONEq(t, `{"code":7,"data":null,"msg":"account is disabled"}`, w.Body.String())
	require.Contains(t, w.Header().Get("Set-Cookie"), "x-token=;")
}

// An external account past its expiry gets 401 on its next request; the
// status stays enabled (user management design 3.6).
func TestJWTAuthRejectsExpiredExternalUserWith401(t *testing.T) {
	engine := jwtTestEngine(t)
	past := time.Now().Add(-time.Minute)
	stubAuthState(t, systemService.UserAuthState{Exists: true, Status: "enabled", PersonType: systemService.PersonExternal, ExpiresAt: &past}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("x-token", testToken(t))
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.JSONEq(t, `{"code":7,"data":null,"msg":"account has expired"}`, w.Body.String())

	future := time.Now().Add(time.Hour)
	stubAuthState(t, systemService.UserAuthState{Exists: true, Status: "enabled", PersonType: systemService.PersonExternal, ExpiresAt: &future}, nil)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("x-token", testToken(t))
	engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code, "extending the expiry restores access")
}

func TestJWTAuthRejectsDeletedUserWith401(t *testing.T) {
	engine := jwtTestEngine(t)
	stubAuthState(t, systemService.UserAuthState{Exists: false}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("x-token", testToken(t))
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthFailsWithoutLoggingOutWhenStateCannotLoad(t *testing.T) {
	engine := jwtTestEngine(t)
	stubAuthState(t, systemService.UserAuthState{}, errors.New("db down"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("x-token", testToken(t))
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"code":7,"data":{},"msg":"failed to load user state"}`, w.Body.String())
}

func TestJWTAuthPassesCurrentRolesNotTokenRoles(t *testing.T) {
	engine := jwtTestEngine(t)
	stubAuthState(t, systemService.UserAuthState{Exists: true, Status: systemService.DefaultUserStatus, RoleIDs: []uint{42}}, nil)
	engine.GET("/roles", JWTAuth(), func(c *gin.Context) {
		roles, _ := c.Get(utils.RoleIDsContextKey)
		c.JSON(http.StatusOK, gin.H{"roles": roles})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	req.Header.Set("x-token", testToken(t))
	engine.ServeHTTP(w, req)

	require.JSONEq(t, `{"roles":[42]}`, w.Body.String())
}
