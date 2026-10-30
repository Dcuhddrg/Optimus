package initialize

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"net/http/httptest"
	api "optimus/server/api/v1/system"
	"optimus/server/config"
	"optimus/server/global"
	model "optimus/server/model/system"
	service "optimus/server/service/system"
	"strings"
	"testing"
	"time"
)

type auditProtocolStore struct {
	fail   bool
	filter service.AuditFilter
}

type auditProtocolAuthorizer struct{ allowed bool }

func (a auditProtocolAuthorizer) AuthorizeAPI(_ context.Context, method, path string, _ []uint) (bool, error) {
	return a.allowed && method == "GET" && (path == "/api/v1/system/audit-logs" || path == "/api/v1/system/audit-logs/export"), nil
}

func (s *auditProtocolStore) QueryLogs(_ context.Context, f service.AuditFilter) (model.AuditPage, error) {
	s.filter = f
	if s.fail {
		return model.AuditPage{}, errors.New("private database error")
	}
	return model.AuditPage{Items: nil, Total: 42, Page: f.Page, PageSize: f.PageSize}, nil
}
func (s *auditProtocolStore) ExportLogs(context.Context, service.AuditFilter) ([]model.AuditLogEntry, error) {
	if s.fail {
		return nil, errors.New("private database error")
	}
	return []model.AuditLogEntry{{ID: 1, ActorUsername: "alice", Action: "auth.logout", Result: "success"}}, nil
}
func (s *auditProtocolStore) DeleteBefore(context.Context, time.Time) error { return nil }
func TestAuditProtocolThroughRealRouter(t *testing.T) {
	for _, prefix := range []string{"/api/v1", "/custom/v2", "/"} {
		t.Run(prefix, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			oldConfig, oldFactory := global.OPTIMUS_CONFIG, newSystemModule
			t.Cleanup(func() { global.OPTIMUS_CONFIG = oldConfig; newSystemModule = oldFactory })
			global.OPTIMUS_CONFIG.System.RouterPrefix = prefix
			global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "protocol-test-key", ExpiresTime: 3600, Issuer: "protocol-test"}
			store := &auditProtocolStore{}
			allowed := true
			missing := false
			newSystemModule = func(_ *gorm.DB, cfg config.Config) *SystemModule {
				deps := api.SystemHandlerDeps{}
				if !missing {
					deps.Audit = service.NewAuditService(store)
				}
				return &SystemModule{Handlers: api.NewSystemHandlers(deps), Capabilities: &SystemCapabilities{RBAC: auditProtocolAuthorizer{allowed: allowed}}, cfg: cfg}
			}
			token, err := service.IssueJWT(service.TokenSubject{ID: 1, Username: "admin", RoleIDs: []uint{1}, RoleCodes: []string{"admin"}}, global.OPTIMUS_CONFIG.JWT, time.Now())
			require.NoError(t, err)
			base := strings.TrimRight(prefix, "/")
			require.Len(t, Routers().Routes(), 38)
			request := func(path, credential string) *httptest.ResponseRecorder {
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", base+path, nil)
				if credential != "" {
					req.Header.Set("Authorization", "Bearer "+credential)
				}
				Routers().ServeHTTP(w, req)
				return w
			}
			for _, path := range []string{"/system/audit-logs", "/system/audit-logs/export"} {
				w := request(path, "")
				require.Equal(t, 401, w.Code)
				var body map[string]any
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
				require.Equal(t, float64(7), body["code"])
				require.Contains(t, body, "data")
				require.Nil(t, body["data"])
			}
			allowed = false
			w := request("/system/audit-logs", token)
			require.Equal(t, 403, w.Code)
			require.JSONEq(t, `{"code":7,"data":{},"msg":"permission denied"}`, w.Body.String())
			allowed = true
			w = request("/system/audit-logs?page=2&pageSize=50&result=failure", token)
			require.Equal(t, 200, w.Code)
			require.JSONEq(t, `{"code":0,"data":{"list":[],"total":42,"page":2,"pageSize":50},"msg":"success"}`, w.Body.String())
			require.Equal(t, "failure", *store.filter.Result)
			w = request("/system/audit-logs/export", token)
			require.Equal(t, 200, w.Code)
			require.Contains(t, w.Header().Get("Content-Type"), "text/csv")
			require.Contains(t, w.Body.String(), "alice")
			require.NotContains(t, w.Body.String(), `"code"`)
			store.fail = true
			w = request("/system/audit-logs", token)
			require.Equal(t, 500, w.Code)
			require.JSONEq(t, `{"code":7,"data":{},"msg":"failed to query audit logs"}`, w.Body.String())
			w = request("/system/audit-logs/export", token)
			require.Equal(t, 500, w.Code)
			require.JSONEq(t, `{"code":7,"data":{},"msg":"failed to export audit logs"}`, w.Body.String())
			missing = true
			w = request("/system/audit-logs", token)
			require.Equal(t, 503, w.Code)
			require.JSONEq(t, `{"code":7,"data":{},"msg":"audit service is not available"}`, w.Body.String())
			w = request("/system/auth/me", "")
			require.Equal(t, 401, w.Code)
			require.Contains(t, w.Body.String(), `"code":7`)
			require.Contains(t, w.Body.String(), `"data":null`)
			w = request("/health", "")
			require.Equal(t, 200, w.Code)
			require.NotContains(t, w.Body.String(), `"code"`)
		})
	}
}
