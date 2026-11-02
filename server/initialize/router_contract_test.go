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
	// Flags: public endpoints bypass authentication; envelope routes opt into the
	// already-migrated response protocol before JWT. Other routes remain legacy.
	routes := []struct {
		method, path     string
		public, envelope bool
	}{
		{"GET", "/health", true, false}, {"GET", "/ping", true, false}, {"GET", "/base/ping", true, false},
		{"POST", "/system/auth/login", true, true}, {"GET", "/system/auth/me", false, true}, {"POST", "/system/auth/logout", false, true},
		{"GET", "/system/users", false, true}, {"PUT", "/system/users/:id", false, true}, {"PATCH", "/system/users/:id/status", false, true}, {"PUT", "/system/users/:id/roles", false, true},
		{"GET", "/system/roles", false, false}, {"POST", "/system/roles", false, false}, {"PUT", "/system/roles/:id", false, false}, {"PATCH", "/system/roles/:id/status", false, false}, {"PUT", "/system/roles/:id/permissions", false, false}, {"GET", "/system/permissions/tree", false, false},
		{"GET", "/system/ldap/config", false, false}, {"PUT", "/system/ldap/config", false, false}, {"POST", "/system/ldap/test", false, false}, {"GET", "/system/audit-logs", false, true}, {"GET", "/system/audit-logs/export", false, true},
		{"GET", "/projects", false, false}, {"POST", "/projects", false, false}, {"GET", "/projects/:code", false, false}, {"PUT", "/projects/:code", false, false}, {"DELETE", "/projects/:code", false, false},
		{"GET", "/projects/:code/members", false, false}, {"POST", "/projects/:code/members", false, false}, {"PATCH", "/projects/:code/members/:userId", false, false}, {"DELETE", "/projects/:code/members/:userId", false, false},
		{"GET", "/cloud/accounts", false, false}, {"GET", "/cloud/accounts/:id", false, false}, {"POST", "/cloud/accounts", false, false}, {"PUT", "/cloud/accounts/:id", false, false}, {"PUT", "/cloud/accounts/:id/tags", false, false}, {"DELETE", "/cloud/accounts/:id", false, false}, {"POST", "/cloud/accounts/:id/sync", false, false}, {"GET", "/cloud/assets", false, false},
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
					path := strings.NewReplacer(":userId", "8", ":id", "7", ":code", "demo").Replace(route.path)
					w := httptest.NewRecorder()
					engine.ServeHTTP(w, httptest.NewRequest(route.method, base+path, nil))
					if route.public {
						if route.method == "POST" {
							require.Equal(t, 400, w.Code)
						} else {
							require.Equal(t, 200, w.Code)
						}
						return
					}
					require.Equal(t, 401, w.Code)
					if route.envelope {
						require.JSONEq(t, `{"code":7,"data":null,"msg":"authorization header is required"}`, w.Body.String())
					} else {
						require.JSONEq(t, `{"message":"authorization header is required"}`, w.Body.String())
					}
				})
			}
		})
	}
}
