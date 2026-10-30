package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"io"
	"net/http/httptest"
	"optimus/server/utils/response"
	"testing"
)

func TestRecoveryPreservesLegacyAndFormatsMigratedFailures(t *testing.T) {
	previousWriter := gin.DefaultErrorWriter
	gin.DefaultErrorWriter = io.Discard
	t.Cleanup(func() { gin.DefaultErrorWriter = previousWriter })
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(Recovery())
	group := engine.Group("/api/v1")
	group.Use(response.EnvelopeForRoutes(group.BasePath(), []response.Route{{Method: "GET", Path: "/system/audit-logs"}, {Method: "GET", Path: "/system/audit-logs/export"}}))
	panicHandler := func(*gin.Context) { panic("private panic details") }
	group.GET("/system/audit-logs", panicHandler)
	group.GET("/legacy", panicHandler)
	group.GET("/system/audit-logs/export", func(c *gin.Context) {
		c.Header("Content-Type", "text/csv")
		c.String(200, "id\n1")
		panic("private panic after partial CSV")
	})
	for _, path := range []string{"/legacy", "/system/audit-logs", "/system/audit-logs/export"} {
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, httptest.NewRequest("GET", "/api/v1"+path, nil))
		if path == "/system/audit-logs/export" {
			require.Equal(t, 200, w.Code) // Headers already sent cannot be rewritten.
			require.Equal(t, "id\n1", w.Body.String())
			continue
		}
		require.Equal(t, 500, w.Code)
		if path == "/legacy" {
			require.Empty(t, w.Body.String())
		} else {
			require.JSONEq(t, `{"code":7,"data":{},"msg":"internal server error"}`, w.Body.String())
			require.NotContains(t, w.Body.String(), "private panic")
		}
	}
}
