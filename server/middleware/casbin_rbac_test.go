package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"optimus/server/utils"

	"github.com/casbin/casbin/v3"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// Allow/deny against real casbin_rule policies is covered by the DB-backed
// router tests in initialize; these cover requests without loaded roles and
// the fail-closed state.
func TestCasbinHandlerDeniesMissingClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/private", CasbinHandler(), func(c *gin.Context) { c.Status(http.StatusNoContent) })

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/private", nil))

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"code":7,"data":{},"msg":"permission denied"}`, w.Body.String())
}

func TestCasbinHandlerFailsClosedWhilePolicyIsStale(t *testing.T) {
	gin.SetMode(gin.TestMode)
	saved := utils.BuildCasbinEnforcer
	t.Cleanup(func() { utils.BuildCasbinEnforcer = saved; utils.ResetCasbin() })
	utils.ResetCasbin()
	utils.BuildCasbinEnforcer = func() (*casbin.SyncedEnforcer, error) { return nil, errors.New("db down") }

	engine := gin.New()
	engine.GET("/private", func(c *gin.Context) { c.Set(utils.RoleIDsContextKey, []uint{1}); c.Next() }, CasbinHandler(), func(c *gin.Context) { c.Status(http.StatusNoContent) })

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/private", nil))

	require.Equal(t, http.StatusOK, w.Code, "never-loaded policy must not panic")
	require.JSONEq(t, `{"code":7,"data":{},"msg":"permission system is syncing, please retry later"}`, w.Body.String())
}
