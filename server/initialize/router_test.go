package initialize

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"optimus/server/service"
	"strings"
	"testing"
	"time"

	"optimus/server/config"
	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestRoutersUseGroupedAuthService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	previousServices := *service.ServiceGroupApp
	t.Cleanup(func() {
		*service.ServiceGroupApp = previousServices
	})

	InitializeServices(nil, global.OPTIMUS_CONFIG)
	service.ServiceGroupApp.SystemServiceGroup.AuthService = systemService.NewAuthService(fakeRouterAuthStore{}, fakeRouterTokenIssuer{}, nil)

	engine := Routers()

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/system/auth/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.JSONEq(t, `{"code":7,"data":null,"msg":"username or password is incorrect"}`, recorder.Body.String())
}

func TestRoutersUseGroupedRBACCapability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalConfig := global.OPTIMUS_CONFIG
	previousServices := *service.ServiceGroupApp
	t.Cleanup(func() {
		global.OPTIMUS_CONFIG = originalConfig
		*service.ServiceGroupApp = previousServices
	})

	global.OPTIMUS_CONFIG.JWT = config.JWT{
		SigningKey:  "router-test-signing-key",
		ExpiresTime: 3600,
		BufferTime:  300,
		Issuer:      "optimus-router-test",
	}
	InitializeServices(nil, global.OPTIMUS_CONFIG)
	service.ServiceGroupApp.SystemServiceGroup.RBAC = fakeRouterRBAC{allowed: true}

	engine := Routers()
	token, err := systemService.IssueJWT(systemService.TokenSubject{
		ID:        1,
		UUID:      "router-test-user",
		Username:  "admin",
		RoleIDs:   []uint{1},
		RoleCodes: []string{"admin"},
	}, global.OPTIMUS_CONFIG.JWT, time.Now())
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system/permissions/tree", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), "\"code\":\"system.permissions.view\"")
}

type fakeRouterAuthStore struct{}

type fakeRouterTokenIssuer struct{}

type fakeRouterRBAC struct {
	allowed bool
}

func (f fakeRouterRBAC) AuthorizeAPI(_ context.Context, method string, path string, roleIDs []uint) (bool, error) {
	return f.allowed && method == http.MethodGet && path == "/api/v1/system/permissions/tree" && len(roleIDs) == 1 && roleIDs[0] == 1, nil
}

func (fakeRouterTokenIssuer) IssueToken(systemService.TokenSubject) (string, error) {
	return "", nil
}

func (fakeRouterAuthStore) FindUserByUsername(context.Context, string) (systemModel.SystemUser, error) {
	return systemModel.SystemUser{}, errors.New("not found")
}

func (fakeRouterAuthStore) FindUserByID(context.Context, uint) (systemModel.SystemUser, error) {
	return systemModel.SystemUser{}, errors.New("not implemented")
}

func (fakeRouterAuthStore) GetLDAPConfig(context.Context) (systemModel.SystemLDAPConfig, error) {
	return systemModel.SystemLDAPConfig{}, nil
}

func (fakeRouterAuthStore) SyncLDAPUser(context.Context, systemService.LDAPUser) (systemModel.SystemUser, error) {
	return systemModel.SystemUser{}, errors.New("not implemented")
}

func (fakeRouterAuthStore) RolesForUser(context.Context, uint) ([]systemModel.SystemRoleSummary, error) {
	return nil, errors.New("not implemented")
}

func (fakeRouterAuthStore) MenuPermissionCodesForRoles(context.Context, []uint) ([]string, error) {
	return nil, errors.New("not implemented")
}

func (fakeRouterAuthStore) UpdateLastLogin(context.Context, uint, time.Time) error {
	return nil
}

func (fakeRouterAuthStore) WriteAudit(context.Context, systemService.AuditEvent) error {
	return nil
}
