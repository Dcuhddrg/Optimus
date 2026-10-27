package initialize

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	projectAPI "optimus/server/api/v1/project"
	systemAPI "optimus/server/api/v1/system"
	"optimus/server/config"
	"optimus/server/global"
	projectModule "optimus/server/initialize/project"
	projectModel "optimus/server/model/project"
	projectUsecase "optimus/server/service/project/usecase"
	systemService "optimus/server/service/system"
	systemUsecase "optimus/server/service/system/usecase"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRoutersLimitProjectScopeToCloudResources(t *testing.T) {
	cases := []struct {
		name          string
		store         *routerScopeStore
		projectStatus int
		cloudStatus   int
		detailStatus  int
	}{
		{
			name:          "deleted selected project",
			store:         &routerScopeStore{missing: true},
			projectStatus: http.StatusOK, cloudStatus: http.StatusNotFound,
			detailStatus: http.StatusNotFound,
		},
		{
			name:          "inaccessible selected project",
			store:         &routerScopeStore{projectID: 2},
			projectStatus: http.StatusOK, cloudStatus: http.StatusForbidden,
			detailStatus: http.StatusForbidden,
		},
		{
			name:          "removed selected project membership",
			store:         &routerScopeStore{projectID: 1, memberMissing: true},
			projectStatus: http.StatusOK, cloudStatus: http.StatusForbidden,
			detailStatus: http.StatusForbidden,
		},
		{
			name:          "project resolver unavailable",
			store:         &routerScopeStore{lookupErr: errors.New("project store unavailable")},
			projectStatus: http.StatusInternalServerError, cloudStatus: http.StatusInternalServerError,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			engine, token := newRouterWithProjectStore(t, tc.store, true)
			routes := []struct {
				path   string
				status int
			}{
				{"/api/v1/system/permissions/tree", http.StatusOK},
				{"/api/v1/projects", tc.projectStatus},
				{"/api/v1/cloud/accounts", tc.cloudStatus},
			}
			if tc.detailStatus != 0 {
				routes = append(routes, struct {
					path   string
					status int
				}{"/api/v1/projects/stale", tc.detailStatus})
			}
			for _, route := range routes {
				t.Run(route.path, func(t *testing.T) {
					before := tc.store.scopeCalls
					req := httptest.NewRequest(http.MethodGet, route.path, nil)
					req.Header.Set("Authorization", "Bearer "+token)
					req.Header.Set("X-Project-Code", "stale")
					recorder := httptest.NewRecorder()
					engine.ServeHTTP(recorder, req)
					require.Equal(t, route.status, recorder.Code, recorder.Body.String())
					if route.path == "/api/v1/system/permissions/tree" {
						require.Equal(t, before, tc.store.scopeCalls, "system route must not resolve project scope")
					}
				})
			}
		})
	}
}

func TestRoutersKeepAuthenticationBeforeProjectScope(t *testing.T) {
	for _, allowed := range []bool{true, false} {
		t.Run(map[bool]string{true: "missing JWT", false: "RBAC denied"}[allowed], func(t *testing.T) {
			store := &routerScopeStore{lookupErr: errors.New("must not be called")}
			engine, token := newRouterWithProjectStore(t, store, allowed)
			req := httptest.NewRequest(http.MethodGet, "/api/v1/cloud/accounts", nil)
			req.Header.Set("X-Project-Code", "stale")
			want := http.StatusUnauthorized
			if !allowed {
				req.Header.Set("Authorization", "Bearer "+token)
				want = http.StatusForbidden
			}
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, req)
			require.Equal(t, want, recorder.Code)
			require.Zero(t, store.scopeCalls)
		})
	}
}

func newRouterWithProjectStore(t *testing.T, store *routerScopeStore, allowed bool) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	oldConfig, oldDB := global.OPTIMUS_CONFIG, global.OPTIMUS_DB
	oldSystem, oldProject := newSystemModule, newProjectModule
	t.Cleanup(func() {
		global.OPTIMUS_CONFIG, global.OPTIMUS_DB = oldConfig, oldDB
		newSystemModule, newProjectModule = oldSystem, oldProject
	})
	global.OPTIMUS_DB = nil
	global.OPTIMUS_CONFIG.System.RouterPrefix = "/api/v1"
	global.OPTIMUS_CONFIG.JWT = config.JWT{
		SigningKey: "project-scope-router-test", ExpiresTime: 3600,
		BufferTime: 300, Issuer: "project-scope-router-test",
	}
	newSystemModule = func(_ *gorm.DB, cfg config.Config) *SystemModule {
		return &SystemModule{
			Handlers:     systemAPI.NewSystemHandlers(systemAPI.SystemHandlerDeps{}),
			Capabilities: &SystemCapabilities{RBAC: routerScopeAuthorizer{allowed: allowed}},
			cfg:          cfg,
		}
	}
	newProjectModule = func(_ *gorm.DB, _ config.Config) *projectModule.ProjectModule {
		svc := projectUsecase.NewProjectService(store)
		return &projectModule.ProjectModule{
			Service:  svc,
			Handlers: projectAPI.NewProjectHandlers(projectAPI.ProjectHandlerDeps{Service: svc}),
		}
	}
	token, err := systemService.IssueJWT(systemUsecase.TokenSubject{
		ID: 1, UUID: "scope-user", Username: "admin", RoleIDs: []uint{1}, RoleCodes: []string{"admin"},
	}, global.OPTIMUS_CONFIG.JWT, time.Now())
	require.NoError(t, err)
	return Routers(), token
}

type routerScopeStore struct {
	projectUsecase.ProjectStore
	projectID     uint
	missing       bool
	memberMissing bool
	lookupErr     error
	scopeCalls    int
}

type routerScopeAuthorizer struct{ allowed bool }

func (a routerScopeAuthorizer) AuthorizeAPI(_ context.Context, method, path string, roleIDs []uint) (bool, error) {
	if !a.allowed || method != http.MethodGet || len(roleIDs) != 1 || roleIDs[0] != 1 {
		return false, nil
	}
	switch path {
	case "/api/v1/system/permissions/tree", "/api/v1/projects", "/api/v1/projects/:code", "/api/v1/cloud/accounts":
		return true, nil
	default:
		return false, nil
	}
}

func (s *routerScopeStore) ListAccessibleProjectIDs(context.Context, uint) ([]uint, error) {
	s.scopeCalls++
	return []uint{1}, s.lookupErr
}

func (s *routerScopeStore) FindByCode(context.Context, string) (projectModel.Project, error) {
	if s.missing {
		return projectModel.Project{}, projectUsecase.ErrProjectNotFound
	}
	p := projectModel.Project{}
	p.ID, p.Code = s.projectID, "stale"
	return p, nil
}

func (s *routerScopeStore) FindMember(_ context.Context, _ uint, projectID uint) (projectModel.UserProject, error) {
	if s.memberMissing || projectID != 1 {
		return projectModel.UserProject{}, projectUsecase.ErrMemberNotFound
	}
	return projectModel.UserProject{ProjectRole: "viewer"}, nil
}

func (s *routerScopeStore) List(context.Context, []uint) ([]projectModel.Project, error) {
	return []projectModel.Project{}, nil
}
