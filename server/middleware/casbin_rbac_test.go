package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// Allow/deny against real casbin_rule policies is covered by the DB-backed
// router tests in initialize; this covers requests without claims.
func TestCasbinHandlerDeniesMissingClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/private", CasbinHandler(), func(c *gin.Context) { c.Status(http.StatusNoContent) })

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/private", nil))

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"code":7,"data":{},"msg":"permission denied"}`, w.Body.String())
}
