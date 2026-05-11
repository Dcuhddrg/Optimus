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
	engine.GET("/api/v1/system/users", CasbinHandler(), func(c *gin.Context) {
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

func rbacTestEngine(t *testing.T, authorizer rbacAuthorizer) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	previous := newRBACAuthorizer
	newRBACAuthorizer = func() rbacAuthorizer { return authorizer }
	t.Cleanup(func() {
		newRBACAuthorizer = previous
	})

	engine := gin.New()
	engine.Use(seedRBACClaims())
	engine.GET("/api/v1/system/users", CasbinHandler(), func(c *gin.Context) {
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
