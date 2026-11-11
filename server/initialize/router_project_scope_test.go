package initialize

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"optimus/server/config"
	"optimus/server/global"
	"optimus/server/model/common/response"
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
	for _, table := range []string{"user_projects", "projects", "sys_role_permissions", "sys_user_roles", "sys_permissions", "sys_roles", "sys_users", "sys_ldap_config", "sys_audit_logs", "sys_data_migrations", "jwt_blacklists", "casbin_rule"} {
		require.NoError(t, db.Exec("DROP TABLE IF EXISTS `"+table+"`").Error)
	}
	require.NoError(t, projectService.AutoMigrateProject(db))
	require.NoError(t, systemService.AutoMigrateSystem(db))
	setTestDatabase(t, db)
	// JWTAuth loads the requester's roles from the database on every request
	// (design 3.1): user 1 holds role 1 (granted below when allowed) and role 7
	// (granted by tests that need a non-admin role).
	for _, id := range []uint{1, 7} {
		require.NoError(t, db.Create(&systemModel.SystemRole{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: id}, Code: "role-" + strconv.Itoa(int(id)), Name: "role"}).Error)
	}
	require.NoError(t, db.Create(&systemModel.SystemUser{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 1}, UUID: "router-test-user", Username: "router-test-user", Source: "local", Status: systemService.DefaultUserStatus}).Error)
	for _, id := range []uint{1, 7} {
		require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: 1, RoleID: id}).Error)
	}
	for _, path := range []string{"/system/permissions/tree", "/projects", "/projects/:code", "/cloud/accounts"} {
		permission := systemModel.SystemPermission{Code: path, Name: path, Type: "api", Method: "GET", Path: path}
		require.NoError(t, db.Create(&permission).Error)
		if allowed {
			grantAPI(t, db, 1, "GET", path)
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
	global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "project-scope-router-test", ExpiresTime: "1h", BufferTime: "5m", Issuer: "project-scope-router-test"}
	setTestDatabase(t, db)
	token, err := utils.NewJWT().CreateToken(utils.NewJWT().CreateClaims(utils.BaseClaims{ID: 1, UUID: "scope-user", Username: "admin", RoleIDs: []uint{1}, RoleCodes: []string{"admin"}}))
	require.NoError(t, err)
	return Routers(), token, calls
}

func TestRoutersLimitProjectScopeToCloudResources(t *testing.T) {
	for _, tc := range []struct {
		name, scenario string
		// GVA: every business result is HTTP 200; code 0 succeeds, code 7 fails.
		projectCode, cloudCode, detailCode int
		checkDetail                        bool
	}{
		{"deleted selected project", "missing", response.SUCCESS, response.ERROR, response.ERROR, true},
		{"inaccessible selected project", "inaccessible", response.SUCCESS, response.ERROR, response.ERROR, true},
		{"removed selected project membership", "removed", response.SUCCESS, response.ERROR, response.ERROR, true},
		{"project resolver unavailable", "unavailable", response.ERROR, response.ERROR, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			engine, token, calls := newScopeTestRouter(t, tc.scenario, true)
			type route struct {
				path string
				code int
			}
			routes := []route{{"/api/v1/system/permissions/tree", response.SUCCESS}, {"/api/v1/projects", tc.projectCode}, {"/api/v1/cloud/accounts", tc.cloudCode}}
			if tc.checkDetail {
				routes = append(routes, route{"/api/v1/projects/stale", tc.detailCode})
			}
			for _, route := range routes {
				t.Run(route.path, func(t *testing.T) {
					before := *calls
					req := httptest.NewRequest(http.MethodGet, route.path, nil)
					req.Header.Set("x-token", token)
					req.Header.Set("X-Project-Code", "stale")
					recorder := httptest.NewRecorder()
					engine.ServeHTTP(recorder, req)
					require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
					var body response.Response
					require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body), recorder.Body.String())
					require.Equal(t, route.code, body.Code, recorder.Body.String())
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
			want := http.StatusUnauthorized // GVA NoAuth
			if !allowed {
				req.Header.Set("x-token", token)
				want = http.StatusOK // GVA permission denial: HTTP 200, code 7
			}
			recorder := httptest.NewRecorder()
			engine.ServeHTTP(recorder, req)
			require.Equal(t, want, recorder.Code)
			require.Contains(t, recorder.Body.String(), `"code":7`)
			require.Zero(t, *calls)
		})
	}
}
