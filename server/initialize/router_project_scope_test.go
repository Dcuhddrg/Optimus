package initialize

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"optimus/server/config"
	"optimus/server/global"
	projectModel "optimus/server/model/project"
	systemModel "optimus/server/model/system"
	"optimus/server/service"
	projectService "optimus/server/service/project"
	systemService "optimus/server/service/system"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func projectScopeRouterDB(t *testing.T, scenario string, allowed bool) (*gorm.DB, *int) {
	t.Helper()
	dsn := os.Getenv("OPTIMUS_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("OPTIMUS_TEST_MYSQL_DSN is not set")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	for _, table := range []string{"user_projects", "projects", "sys_role_permissions", "sys_user_roles", "sys_permissions", "sys_roles", "sys_users", "sys_ldap_config", "sys_audit_logs"} {
		require.NoError(t, db.Exec("DROP TABLE IF EXISTS `"+table+"`").Error)
	}
	require.NoError(t, projectService.AutoMigrateProject(db))
	require.NoError(t, systemService.AutoMigrateSystem(db))
	for _, path := range []string{"/api/v1/system/permissions/tree", "/api/v1/projects", "/api/v1/projects/:code", "/api/v1/cloud/accounts"} {
		permission := systemModel.SystemPermission{Code: path, Name: path, Type: "api", Method: "GET", Path: path}
		require.NoError(t, db.Create(&permission).Error)
		if allowed {
			require.NoError(t, db.Create(&systemModel.SystemRolePermission{RoleID: 1, PermissionID: permission.ID}).Error)
		}
	}
	p := projectModel.Project{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 1}, Name: "Accessible", Code: "existing", OwnerID: 1, Status: "active"}
	if scenario == "removed" {
		p.Code = "stale"
	}
	require.NoError(t, db.Create(&p).Error)
	require.NoError(t, db.Create(&projectModel.UserProject{UserID: 1, ProjectID: 1, ProjectRole: "viewer"}).Error)
	if scenario == "inaccessible" {
		require.NoError(t, db.Create(&projectModel.Project{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 2}, Name: "Other", Code: "stale", OwnerID: 2, Status: "active"}).Error)
	}
	calls := 0
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register("test:scope_queries", func(tx *gorm.DB) {
		if tx.Statement.Table != "user_projects" {
			return
		}
		calls++
		if scenario == "unavailable" {
			tx.AddError(errors.New("project store unavailable"))
			return
		}
		if scenario == "removed" {
			if _, ok := tx.Statement.Dest.(*projectModel.UserProject); ok {
				tx.AddError(gorm.ErrRecordNotFound)
			}
		}
	}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	return db, &calls
}

func newScopeTestRouter(t *testing.T, scenario string, allowed bool) (*gin.Engine, string, *int) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, calls := projectScopeRouterDB(t, scenario, allowed)
	oldConfig, oldServices := global.OPTIMUS_CONFIG, *service.ServiceGroupApp
	t.Cleanup(func() { global.OPTIMUS_CONFIG = oldConfig; *service.ServiceGroupApp = oldServices })
	global.OPTIMUS_CONFIG.System.RouterPrefix = "/api/v1"
	global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "project-scope-router-test", ExpiresTime: 3600, BufferTime: 300, Issuer: "project-scope-router-test"}
	setTestDatabase(t, db)
	token, err := utils.NewJWT().CreateToken(utils.NewJWT().CreateClaims(utils.BaseClaims{ID: 1, UUID: "scope-user", Username: "admin", RoleIDs: []uint{1}, RoleCodes: []string{"admin"}}))
	require.NoError(t, err)
	return Routers(), token, calls
}

func TestRoutersLimitProjectScopeToCloudResources(t *testing.T) {
	for _, tc := range []struct {
		name, scenario                           string
		projectStatus, cloudStatus, detailStatus int
	}{
		{"deleted selected project", "missing", http.StatusOK, http.StatusNotFound, http.StatusNotFound},
		{"inaccessible selected project", "inaccessible", http.StatusOK, http.StatusForbidden, http.StatusForbidden},
		{"removed selected project membership", "removed", http.StatusOK, http.StatusForbidden, http.StatusForbidden},
		{"project resolver unavailable", "unavailable", http.StatusInternalServerError, http.StatusInternalServerError, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine, token, calls := newScopeTestRouter(t, tc.scenario, true)
			routes := []struct {
				path   string
				status int
			}{{"/api/v1/system/permissions/tree", http.StatusOK}, {"/api/v1/projects", tc.projectStatus}, {"/api/v1/cloud/accounts", tc.cloudStatus}}
			if tc.detailStatus != 0 {
				routes = append(routes, struct {
					path   string
					status int
				}{"/api/v1/projects/stale", tc.detailStatus})
			}
			for _, route := range routes {
				t.Run(route.path, func(t *testing.T) {
					before := *calls
					req := httptest.NewRequest(http.MethodGet, route.path, nil)
					req.Header.Set("Authorization", "Bearer "+token)
					req.Header.Set("X-Project-Code", "stale")
					recorder := httptest.NewRecorder()
					engine.ServeHTTP(recorder, req)
					require.Equal(t, route.status, recorder.Code, recorder.Body.String())
					if route.path == "/api/v1/system/permissions/tree" {
						require.Equal(t, before, *calls, "system route must not resolve project scope")
					}
				})
			}
		})
	}
}

func TestRoutersKeepAuthenticationBeforeProjectScope(t *testing.T) {
	for _, allowed := range []bool{true, false} {
		t.Run(map[bool]string{true: "missing JWT", false: "RBAC denied"}[allowed], func(t *testing.T) {
			engine, token, calls := newScopeTestRouter(t, "unavailable", allowed)
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
			require.Zero(t, *calls)
		})
	}
}
