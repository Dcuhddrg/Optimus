package initialize

import (
	"context"
	"net/http"
	"net/http/httptest"
	"optimus/server/service"
	"testing"
	"time"

	"optimus/server/config"
	"optimus/server/global"
	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// Keep permission keys identical to the real seeds while exercising Routers,
// JWT verification and the real RBAC service without a database dependency.
type routerPrefixPermissionStore struct {
	method string
	path   string
}

func (s *routerPrefixPermissionStore) FindAPIPermission(_ context.Context, method, path string) (systemService.SystemPermissionSummary, bool, error) {
	s.method, s.path = method, path
	for i, permission := range systemService.DefaultSystemPermissions() {
		if permission.Type == "api" && permission.Method == method && permission.Path == path {
			return systemService.SystemPermissionSummary{ID: uint(i + 1), Code: permission.Code}, true, nil
		}
	}
	return systemService.SystemPermissionSummary{}, false, nil
}

func (*routerPrefixPermissionStore) RoleHasPermission(_ context.Context, roleID, _ uint) (bool, error) {
	return roleID == 7, nil
}

func (s *routerPrefixPermissionStore) AuthorizeAPI(ctx context.Context, method, path string, roleIDs []uint) (bool, error) {
	permission, found, err := s.FindAPIPermission(ctx, method, path)
	if err != nil || !found {
		return false, err
	}
	for _, roleID := range roleIDs {
		allowed, err := s.RoleHasPermission(ctx, roleID, permission.ID)
		if err != nil || allowed {
			return allowed, err
		}
	}
	return false, nil
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
			previousServices := *service.ServiceGroupApp
			t.Cleanup(func() {
				global.OPTIMUS_CONFIG, global.OPTIMUS_DB = previousConfig, previousDB
				*service.ServiceGroupApp = previousServices
			})
			global.OPTIMUS_DB = nil
			global.OPTIMUS_CONFIG.System.RouterPrefix = tt.prefix
			global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "prefix-test-key", ExpiresTime: 3600, Issuer: "prefix-test"}
			store := &routerPrefixPermissionStore{}
			initializeTestServices(t, nil, global.OPTIMUS_CONFIG)
			service.ServiceGroupApp.SystemServiceGroup.RBAC = store
			token, err := systemService.IssueJWT(systemService.TokenSubject{ID: 1, RoleIDs: []uint{7}}, global.OPTIMUS_CONFIG.JWT, time.Now())
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
