package initialize

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	systemApi "optimus/server/api/v1/system"
	"optimus/server/config"
	"optimus/server/global"
	systemService "optimus/server/service/system"
	systemUsecase "optimus/server/service/system/usecase"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Keep permission keys identical to the real seeds while exercising Routers,
// JWT verification and the real RBAC service without a database dependency.
type routerPrefixPermissionStore struct {
	method string
	path   string
}

func (s *routerPrefixPermissionStore) FindAPIPermission(_ context.Context, method, path string) (systemUsecase.SystemPermissionSummary, bool, error) {
	s.method, s.path = method, path
	for i, permission := range systemService.DefaultSystemPermissions() {
		if permission.Type == "api" && permission.Method == method && permission.Path == path {
			return systemUsecase.SystemPermissionSummary{ID: uint(i + 1), Code: permission.Code}, true, nil
		}
	}
	return systemUsecase.SystemPermissionSummary{}, false, nil
}

func (*routerPrefixPermissionStore) RoleHasPermission(_ context.Context, roleID, _ uint) (bool, error) {
	return roleID == 7, nil
}

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
			previousBuilder := newSystemModule
			t.Cleanup(func() {
				global.OPTIMUS_CONFIG, global.OPTIMUS_DB = previousConfig, previousDB
				newSystemModule = previousBuilder
			})
			global.OPTIMUS_DB = nil
			global.OPTIMUS_CONFIG.System.RouterPrefix = tt.prefix
			global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "prefix-test-key", ExpiresTime: 3600, Issuer: "prefix-test"}
			store := &routerPrefixPermissionStore{}
			newSystemModule = func(_ *gorm.DB, cfg config.Config) *SystemModule {
				return &SystemModule{
					Handlers:     systemApi.NewSystemHandlers(systemApi.SystemHandlerDeps{}),
					Capabilities: &SystemCapabilities{RBAC: systemUsecase.NewRBACService(store)},
					cfg:          cfg,
				}
			}
			token, err := systemService.IssueJWT(systemUsecase.TokenSubject{ID: 1, RoleIDs: []uint{7}}, global.OPTIMUS_CONFIG.JWT, time.Now())
			require.NoError(t, err)
			request := httptest.NewRequest(tt.method, tt.requestPath, nil)
			request.Header.Set("Authorization", "Bearer "+token)
			recorder := httptest.NewRecorder()
			Routers().ServeHTTP(recorder, request)
			require.Equal(t, tt.permissionPath, store.path)
			require.Equal(t, tt.method, store.method)
			require.Equal(t, tt.status, recorder.Code)
		})
	}
}
