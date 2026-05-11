package initialize

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	systemApi "optimus/server/api/v1/system"
	"optimus/server/config"
	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRoutersSeedSystemCompatibilityHandlersForAuthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	previousBuilder := newSystemModule
	previousHandlers := systemApi.CompatibilityHandlers()
	t.Cleanup(func() {
		newSystemModule = previousBuilder
		systemApi.SetCompatibilityHandlers(previousHandlers)
	})

	newSystemModule = func(db *gorm.DB, cfg config.Config) *SystemModule {
		return &SystemModule{
			Handlers: systemApi.NewSystemHandlers(systemApi.SystemHandlerDeps{
				Auth: systemService.NewAuthService(fakeRouterAuthStore{}, cfg.JWT),
			}),
			Capabilities: &SystemCapabilities{},
			cfg:          cfg,
		}
	}

	engine := Routers()

	require.NotNil(t, systemApi.CompatibilityHandlers())
	require.NotNil(t, systemApi.CompatibilityHandlers().Auth)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/system/auth/login", strings.NewReader(`{"username":"admin","password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")

	engine.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.JSONEq(t, `{"message":"username or password is incorrect"}`, recorder.Body.String())
}

type fakeRouterAuthStore struct{}

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

func (fakeRouterAuthStore) RolesForUser(context.Context, uint) ([]systemService.SystemRoleSummary, error) {
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
