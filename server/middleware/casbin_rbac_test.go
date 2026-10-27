package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCasbinHandlerAllowsAuthorizedRoute(t *testing.T) {
	engine := rbacTestEngine(t, fakeMiddlewareRBAC{allowed: true})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/users", nil)
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

func TestCasbinHandlerDeniesUnauthorizedRoute(t *testing.T) {
	engine := rbacTestEngine(t, fakeMiddlewareRBAC{allowed: false})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/users", nil)
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestCasbinHandlerDeniesMissingClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/api/v1/system/users", CasbinHandler(fakeMiddlewareRBAC{allowed: true}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/users", nil)
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestCasbinHandlerUsesInjectedAuthorizer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(seedRBACClaims())

	authorizer := fakeMiddlewareRBAC{allowed: true}
	engine.GET("/api/v1/system/users", CasbinHandler(authorizer), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/users", nil)
	engine.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

type fakeMiddlewareRBAC struct {
	allowed bool
}

func (f fakeMiddlewareRBAC) AuthorizeAPI(_ context.Context, method string, path string, roleIDs []uint) (bool, error) {
	return f.allowed && method == http.MethodGet && path == "/api/v1/system/users" && len(roleIDs) == 1 && roleIDs[0] == 1, nil
}

func rbacTestEngine(t *testing.T, authorizer APIAuthorizer) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(seedRBACClaims())
	engine.GET("/api/v1/system/users", CasbinHandler(authorizer), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return engine
}

func seedRBACClaims() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(utils.ClaimsContextKey, &utils.CustomClaims{
			BaseClaims: utils.BaseClaims{
				ID:      1,
				RoleIDs: []uint{1},
			},
		})
	}
}

type prefixMiddlewareRBAC struct {
	method, path string
	roles        []uint
	calls        int
}

func (f *prefixMiddlewareRBAC) AuthorizeAPI(_ context.Context, method, path string, roles []uint) (bool, error) {
	f.method, f.path, f.roles = method, path, roles
	f.calls++
	return method == http.MethodPut && path == "/api/v1/system/users/:id", nil
}

func TestCasbinHandlerWithPrefixMapsOnlyMatchingRoutePrefix(t *testing.T) {
	for _, tt := range []struct {
		name, runtimePrefix, route, request, permission string
		status, calls                                   int
	}{
		{"default", "/api/v1", "/api/v1/system/users/:id", "/api/v1/system/users/42", "/api/v1/system/users/:id", http.StatusOK, 1},
		{"custom", "/custom", "/custom/system/users/:id", "/custom/system/users/42", "/api/v1/system/users/:id", http.StatusOK, 1},
		{"trailing slash", "/custom/", "/custom/system/users/:id", "/custom/system/users/42", "/api/v1/system/users/:id", http.StatusOK, 1},
		{"root", "/", "/system/users/:id", "/system/users/42", "/api/v1/system/users/:id", http.StatusOK, 1},
		{"prefix boundary", "/custom", "/customized/system/users/:id", "/customized/system/users/42", "", http.StatusForbidden, 0},
		{"unrelated prefix", "/custom", "/other/system/users/:id", "/other/system/users/42", "", http.StatusForbidden, 0},
		{"missing prefix", "", "/custom/system/users/:id", "/custom/system/users/42", "", http.StatusForbidden, 0},
		{"unknown permission", "/custom", "/custom/system/unknown", "/custom/system/unknown", "/api/v1/system/unknown", http.StatusForbidden, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			engine := gin.New()
			authorizer := &prefixMiddlewareRBAC{}
			engine.Use(seedRBACClaims(), CasbinHandlerWithPrefix(authorizer, tt.runtimePrefix, "/api/v1"))
			engine.PUT(tt.route, func(c *gin.Context) { c.Status(http.StatusOK) })
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodPut, tt.request, nil))
			require.Equal(t, tt.status, recorder.Code)
			require.Equal(t, tt.calls, authorizer.calls)
			require.Equal(t, tt.permission, authorizer.path)
			if tt.calls != 0 {
				require.Equal(t, http.MethodPut, authorizer.method)
				require.Equal(t, []uint{1}, authorizer.roles)
			}
		})
	}
}

func TestCasbinHandlerWithPrefixDeniesUnmatchedRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	authorizer := &prefixMiddlewareRBAC{}
	engine.Use(seedRBACClaims(), CasbinHandlerWithPrefix(authorizer, "/", "/api/v1"))
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/missing", nil))
	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.Zero(t, authorizer.calls)
}
