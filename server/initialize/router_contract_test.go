package initialize

import (
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"optimus/server/global"
	cloudRouter "optimus/server/router/cloud"
	projectRouter "optimus/server/router/project"
	systemRouter "optimus/server/router/system"

	"github.com/stretchr/testify/require"
)

func TestModuleRouterGroupsEmbedResponsibilityRouters(t *testing.T) {
	for _, tc := range []struct {
		name   string
		group  any
		fields []string
	}{
		{"system", systemRouter.RouterGroup{}, []string{"BaseRouter", "AuthRouter", "UserRouter", "RoleRouter", "LDAPRouter", "AuditRouter"}},
		{"project", projectRouter.RouterGroup{}, []string{"ProjectRouter", "MemberRouter"}},
		{"cloud", cloudRouter.RouterGroup{}, []string{"AccountRouter", "AssetRouter"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			typ := reflect.TypeOf(tc.group)
			for _, name := range tc.fields {
				f, ok := typ.FieldByName(name)
				require.True(t, ok, name)
				require.True(t, f.Anonymous, name)
			}
		})
	}
}

func TestRegisteredRoutesAndUnauthenticatedContracts(t *testing.T) {
	// Public endpoints bypass authentication; private endpoints use one response contract.
	routes := []struct {
		method, path string
		public       bool
	}{
		{"GET", "/health", true}, {"GET", "/ping", true}, {"GET", "/base/ping", true}, {"GET", "/swagger/*any", true},
		{"POST", "/system/auth/login", true}, {"GET", "/system/auth/me", false}, {"POST", "/system/auth/logout", false},
		{"GET", "/system/users", false}, {"PUT", "/system/users/:id", false}, {"PATCH", "/system/users/:id/status", false}, {"PUT", "/system/users/:id/roles", false}, {"PUT", "/system/users/:id/external", false},
		{"GET", "/system/roles", false}, {"POST", "/system/roles", false}, {"PUT", "/system/roles/:id", false}, {"DELETE", "/system/roles/:id", false}, {"PUT", "/system/roles/:id/permissions", false}, {"GET", "/system/permissions/tree", false},
		{"GET", "/system/ldap/config", false}, {"PUT", "/system/ldap/config", false}, {"POST", "/system/ldap/test", false}, {"POST", "/system/ldap/sync/preview", false}, {"POST", "/system/ldap/sync/execute", false}, {"GET", "/system/audit-logs", false}, {"GET", "/system/audit-logs/export", false},
		{"GET", "/projects", false}, {"POST", "/projects", false}, {"GET", "/projects/:code", false}, {"PUT", "/projects/:code", false}, {"DELETE", "/projects/:code", false},
		{"GET", "/projects/:code/members", false}, {"POST", "/projects/:code/members", false}, {"PATCH", "/projects/:code/members/:userId", false}, {"DELETE", "/projects/:code/members/:userId", false},
		{"GET", "/cloud/accounts", false}, {"GET", "/cloud/accounts/:id", false}, {"POST", "/cloud/accounts", false}, {"PUT", "/cloud/accounts/:id", false}, {"PUT", "/cloud/accounts/:id/tags", false}, {"DELETE", "/cloud/accounts/:id", false}, {"POST", "/cloud/accounts/:id/sync", false}, {"GET", "/cloud/assets", false},
	}
	saved := global.OPTIMUS_CONFIG
	t.Cleanup(func() { global.OPTIMUS_CONFIG = saved })
	for _, prefix := range []string{"/api/v1", "/custom/v2", "/"} {
		t.Run(prefix, func(t *testing.T) {
			global.OPTIMUS_CONFIG.System.RouterPrefix = prefix
			engine := Routers()
			base := strings.TrimRight(prefix, "/")
			expected := []string{}
			for _, route := range routes {
				expected = append(expected, route.method+" "+base+route.path)
			}
			actual := []string{}
			for _, route := range engine.Routes() {
				actual = append(actual, route.Method+" "+route.Path)
			}
			require.ElementsMatch(t, expected, actual)
			for _, route := range routes {
				t.Run(route.method+route.path, func(t *testing.T) {
					path := strings.NewReplacer(":userId", "8", ":id", "7", ":code", "demo", "*any", "index.html").Replace(route.path)
					w := httptest.NewRecorder()
					engine.ServeHTTP(w, httptest.NewRequest(route.method, base+path, nil))
					if route.public {
						if route.method == "POST" {
							require.Equal(t, 200, w.Code)
						} else {
							require.Equal(t, 200, w.Code)
						}
						return
					}
					require.Equal(t, 401, w.Code)
					require.JSONEq(t, `{"code":7,"data":null,"msg":"not logged in or invalid access"}`, w.Body.String())
				})
			}
		})
	}
}
