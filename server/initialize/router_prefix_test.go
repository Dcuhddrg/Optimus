package initialize

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"optimus/server/config"
	"optimus/server/global"
	model "optimus/server/model/system"
	"optimus/server/service"
	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Keep permission keys identical to the real seeds while exercising Routers,
// JWT verification and the real RBAC service against an isolated database.
func TestRoutersAuthorizeCanonicalPermissionWithConfiguredPrefix(t *testing.T) {
	for _, tt := range []struct {
		name, prefix, requestPath, permissionPath, method string
		status                                            int
	}{
		{"empty", "", "/api/v1/system/permissions/tree", "/api/v1/system/permissions/tree", http.MethodGet, http.StatusOK},
		{"default", "/api/v1", "/api/v1/system/permissions/tree", "/api/v1/system/permissions/tree", http.MethodGet, http.StatusOK},
		{"custom", "/custom", "/custom/system/permissions/tree", "/api/v1/system/permissions/tree", http.MethodGet, http.StatusOK},
		{"without leading slash", "custom", "/custom/system/permissions/tree", "/api/v1/system/permissions/tree", http.MethodGet, http.StatusOK},
		{"root", "/", "/system/permissions/tree", "/api/v1/system/permissions/tree", http.MethodGet, http.StatusOK},
		{"trailing slash", "/custom/", "/custom/system/permissions/tree", "/api/v1/system/permissions/tree", http.MethodGet, http.StatusOK},
		{"parameter route", "/custom", "/custom/system/users/42", "/api/v1/system/users/:id", http.MethodPut, http.StatusServiceUnavailable},
	} {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			previousConfig, previousDB := global.OPTIMUS_CONFIG, global.OPTIMUS_DB
			previousServices := *service.ServiceGroupApp
			t.Cleanup(func() {
				global.OPTIMUS_CONFIG, global.OPTIMUS_DB = previousConfig, previousDB
				*service.ServiceGroupApp = previousServices
			})
			db, _ := projectScopeRouterDB(t, "", false)
			global.OPTIMUS_CONFIG.System.RouterPrefix = tt.prefix
			global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "prefix-test-key", ExpiresTime: 3600, Issuer: "prefix-test"}
			setTestDatabase(t, db)
			var permission model.SystemPermission
			err := db.Where("method = ? AND path = ?", tt.method, tt.permissionPath).First(&permission).Error
			if err == gorm.ErrRecordNotFound {
				permission = model.SystemPermission{Code: tt.method + " " + tt.permissionPath, Name: tt.permissionPath, Type: "api", Method: tt.method, Path: tt.permissionPath}
				require.NoError(t, db.Create(&permission).Error)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, db.Create(&model.SystemRolePermission{RoleID: 7, PermissionID: permission.ID}).Error)
			var seenArgs []any
			require.NoError(t, db.Callback().Query().After("gorm:query").Register("test:capture_permission", func(tx *gorm.DB) {
				if tx.Statement.Table == "sys_permissions" {
					seenArgs = append([]any(nil), tx.Statement.Vars...)
				}
				if tt.status == http.StatusServiceUnavailable && tx.Statement.Table == "sys_role_permissions" && tx.Error == nil {
					global.OPTIMUS_DB = nil
				}
			}))
			token, err := systemService.IssueJWT(systemService.TokenSubject{ID: 1, RoleIDs: []uint{7}}, global.OPTIMUS_CONFIG.JWT, time.Now())
			require.NoError(t, err)
			request := httptest.NewRequest(tt.method, tt.requestPath, nil)
			request.Header.Set("Authorization", "Bearer "+token)
			recorder := httptest.NewRecorder()
			Routers().ServeHTTP(recorder, request)
			global.OPTIMUS_DB = db
			require.Contains(t, seenArgs, tt.permissionPath)
			require.Contains(t, seenArgs, tt.method)
			require.Equal(t, tt.status, recorder.Code)
		})
	}
}
