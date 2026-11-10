package initialize

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"optimus/server/config"
	"optimus/server/global"
	model "optimus/server/model/system"
	rootService "optimus/server/service"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestAuditProtocolThroughRealRouter(t *testing.T) {
	for _, prefix := range []string{"/api/v1", "/custom/v2", "/"} {
		t.Run(prefix, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			oldConfig, oldDB, oldServices := global.OPTIMUS_CONFIG, global.OPTIMUS_DB, *rootService.ServiceGroupApp
			t.Cleanup(func() {
				global.OPTIMUS_CONFIG, global.OPTIMUS_DB, *rootService.ServiceGroupApp = oldConfig, oldDB, oldServices
			})
			db, _ := projectScopeRouterDB(t, "", true)
			global.OPTIMUS_CONFIG.System.RouterPrefix = prefix
			global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "protocol-test-key", ExpiresTime: "1h", Issuer: "protocol-test"}
			setTestDatabase(t, db)
			for _, path := range []string{"/system/audit-logs", "/system/audit-logs/export"} {
				grantAPI(t, db, 1, "GET", path)
			}
			entries := make([]model.SystemAuditLog, 42)
			for i := range entries {
				entries[i] = model.SystemAuditLog{Action: "auth.login.failure", Result: "failure"}
			}
			require.NoError(t, db.Create(&entries).Error)
			require.NoError(t, db.Create(&model.SystemAuditLog{ActorUsername: "alice", Action: "auth.logout", Result: "success"}).Error)
			failAudit, missing := false, false
			require.NoError(t, db.Callback().Query().Before("gorm:query").Register("test:fail_audit_query", func(tx *gorm.DB) {
				if failAudit && tx.Statement.Table == "sys_audit_logs" {
					tx.AddError(errors.New("private database error"))
				}
			}))
			token, err := utils.NewJWT().CreateToken(utils.NewJWT().CreateClaims(utils.BaseClaims{ID: 1, Username: "admin", RoleIDs: []uint{1}, RoleCodes: []string{"admin"}}))
			require.NoError(t, err)
			base := strings.TrimRight(prefix, "/")
			require.Len(t, Routers().Routes(), 39) // includes /swagger/*any (#73)
			request := func(path, credential string) *httptest.ResponseRecorder {
				global.OPTIMUS_DB = db
				if missing {
					// The enforcer and the user state are cached, so authorization passes; the handler
					// then finds no database.
					require.NotNil(t, utils.GetCasbin())
					global.OPTIMUS_DB = nil
				}
				w := httptest.NewRecorder()
				req := httptest.NewRequest("GET", base+path, nil)
				if credential != "" {
					req.Header.Set("x-token", credential)
				}
				Routers().ServeHTTP(w, req)
				global.OPTIMUS_DB = db
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
			revokeAPI(t, db, 1, "GET", "/system/audit-logs")
			w := request("/system/audit-logs", token)
			require.Equal(t, 200, w.Code) // GVA permission denial
			require.JSONEq(t, `{"code":7,"data":{},"msg":"permission denied"}`, w.Body.String())
			grantAPI(t, db, 1, "GET", "/system/audit-logs")
			w = request("/system/audit-logs?page=2&pageSize=50&result=failure", token)
			require.Equal(t, 200, w.Code)
			require.JSONEq(t, `{"code":0,"data":{"list":[],"total":42,"page":2,"pageSize":50},"msg":"成功"}`, w.Body.String())
			w = request("/system/audit-logs/export", token)
			require.Equal(t, 200, w.Code)
			require.Contains(t, w.Header().Get("Content-Type"), "text/csv")
			require.Contains(t, w.Body.String(), "alice")
			require.NotContains(t, w.Body.String(), `"code"`)
			failAudit = true
			w = request("/system/audit-logs", token)
			require.Equal(t, 200, w.Code)
			require.JSONEq(t, `{"code":7,"data":{},"msg":"failed to query audit logs"}`, w.Body.String())
			w = request("/system/audit-logs/export", token)
			require.Equal(t, 200, w.Code)
			require.JSONEq(t, `{"code":7,"data":{},"msg":"failed to export audit logs"}`, w.Body.String())
			failAudit, missing = false, true
			w = request("/system/audit-logs", token)
			require.Equal(t, 200, w.Code)
			require.JSONEq(t, `{"code":7,"data":{},"msg":"audit service is not available"}`, w.Body.String())
			missing = false
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
