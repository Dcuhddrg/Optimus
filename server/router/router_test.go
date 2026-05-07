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

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)

	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
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
