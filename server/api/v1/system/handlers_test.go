package system

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAuthHandlerLoginReturnsServiceUnavailableWhenServiceMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/system/auth/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler := &AuthHandler{}
	handler.Login(ctx)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.JSONEq(t, `{"code":7,"data":{},"msg":"auth service is not available"}`, recorder.Body.String())
}

func TestAuditHandlerQueryAuditLogsReturnsServiceUnavailableWhenServiceMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/system/audit-logs", nil)

	handler := &AuditHandler{}
	handler.QueryAuditLogs(ctx)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.JSONEq(t, `{"code":7,"data":{},"msg":"audit service is not available"}`, recorder.Body.String())
}

func TestAuditHandlerExportAuditLogsReturnsServiceUnavailableWhenServiceMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/system/audit-logs/export", nil)

	handler := &AuditHandler{}
	handler.ExportAuditLogs(ctx)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestParseAuditFilterParsesResultAndActionQueryParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		"/system/audit-logs?page=2&pageSize=50&result=success&action=login",
		nil,
	)

	f := parseAuditFilter(ctx)

	require.Equal(t, 2, f.Page)
	require.Equal(t, 50, f.PageSize)
	require.NotNil(t, f.Result)
	require.Equal(t, "success", *f.Result)
	require.NotNil(t, f.Action)
	require.Equal(t, "login", *f.Action)
}

func TestParseAuditFilterClampsInvalidPageSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/system/audit-logs?pageSize=999", nil)

	f := parseAuditFilter(ctx)

	require.Equal(t, 1, f.Page)
	require.Equal(t, 20, f.PageSize)
}

func TestParseAuditFilterParsesTimeRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodGet,
		"/system/audit-logs?startTime=2026-01-01T00:00:00Z&endTime=2026-12-31T23:59:59Z",
		nil,
	)

	f := parseAuditFilter(ctx)

	require.NotNil(t, f.StartTime)
	require.Equal(t, 2026, f.StartTime.Year())
	require.NotNil(t, f.EndTime)
	require.Equal(t, 12, int(f.EndTime.Month()))
}
