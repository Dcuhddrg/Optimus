package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"optimus/server/initialize"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRoutersExposeHealthAndPing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := initialize.Routers()

	routes := []string{
		"/api/v1/health",
		"/api/v1/ping",
		"/api/v1/base/ping",
	}

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, route, nil)

			engine.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestPingRouteReturnsExpectedPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := initialize.Routers()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)

	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "\"message\":\"pong\"")
}

func TestBasePingRouteReturnsExpectedPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := initialize.Routers()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/base/ping", nil)

	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "\"message\":\"pong\"")
}

func TestAuthLoginRouteIsPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := initialize.Routers()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/system/auth/login", nil)

	engine.ServeHTTP(w, req)

	require.NotEqual(t, http.StatusNotFound, w.Code)
	require.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMeRouteRequiresJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := initialize.Routers()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/auth/me", nil)

	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSystemAdminRoutesRequireJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := initialize.Routers()

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/system/users"},
		{http.MethodGet, "/api/v1/system/roles"},
		{http.MethodGet, "/api/v1/system/permissions/tree"},
		{http.MethodGet, "/api/v1/system/ldap/config"},
		{http.MethodGet, "/api/v1/system/audit-logs"},
	}

	for _, route := range routes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(route.method, route.path, nil)

			engine.ServeHTTP(w, req)

			require.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}
