package initialize

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	api "optimus/server/api/v1/system"
	"optimus/server/config"
	"optimus/server/global"
	model "optimus/server/model/system"
	service "optimus/server/service/system"
)

type roleProtocolStore struct {
	service.RoleStore
	roles            []model.SystemRole
	fail             bool
	savedPermissions []uint
}

type roleProtocolAuthorizer struct{ allowed bool }

func (a roleProtocolAuthorizer) AuthorizeAPI(_ context.Context, _ string, path string, roleIDs []uint) (bool, error) {
	return a.allowed && len(roleIDs) == 1 && roleIDs[0] == 1 &&
		(strings.HasPrefix(path, "/api/v1/system/roles") || path == "/api/v1/system/permissions/tree"), nil
}

func (s *roleProtocolStore) ListRoles(context.Context) ([]model.SystemRole, error) {
	if s.fail {
		return nil, errors.New("private database error")
	}
	return s.roles, nil
}
func (s *roleProtocolStore) FindRoleByID(_ context.Context, id uint) (model.SystemRole, error) {
	if s.fail {
		return model.SystemRole{}, errors.New("private database error")
	}
	for _, role := range s.roles {
		if role.ID == id {
			return role, nil
		}
	}
	return model.SystemRole{}, service.ErrRoleNotFound
}
func (s *roleProtocolStore) CreateRole(_ context.Context, role model.SystemRole) (model.SystemRole, error) {
	if s.fail {
		return model.SystemRole{}, errors.New("private database error")
	}
	role.ID = 3
	return role, nil
}
func (s *roleProtocolStore) UpdateRole(_ context.Context, role model.SystemRole) (model.SystemRole, error) {
	return role, nil
}
func (s *roleProtocolStore) PermissionsByCodes(context.Context, []string) ([]model.SystemPermission, error) {
	return []model.SystemPermission{{ID: 8, Code: "system.roles.view"}}, nil
}
func (s *roleProtocolStore) PermissionCodesForRole(context.Context, uint) ([]string, error) {
	return []string{"system.roles.view"}, nil
}
func (s *roleProtocolStore) ReplaceRolePermissions(_ context.Context, _ uint, ids []uint) error {
	s.savedPermissions = ids
	return nil
}
func (s *roleProtocolStore) WriteAudit(context.Context, service.AuditEvent) error { return nil }

func TestRoleProtocolThroughRealRouter(t *testing.T) {
	for _, prefix := range []string{"/api/v1", "/custom/v2", "/"} {
		t.Run(prefix, func(t *testing.T) {
			oldConfig, oldFactory := global.OPTIMUS_CONFIG, newSystemModule
			t.Cleanup(func() { global.OPTIMUS_CONFIG = oldConfig; newSystemModule = oldFactory })
			global.OPTIMUS_CONFIG.System.RouterPrefix = prefix
			global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "role-protocol-key", ExpiresTime: 3600, Issuer: "role-protocol"}
			store := &roleProtocolStore{roles: []model.SystemRole{{ID: 1, Code: "ops", Name: "Operations", Status: "enabled"}, {ID: 2, Code: "admin", Name: "Admin", Status: "enabled"}}}
			allowed, missing := true, false
			newSystemModule = func(_ *gorm.DB, cfg config.Config) *SystemModule {
				deps := api.SystemHandlerDeps{}
				if !missing {
					deps.Role = service.NewRoleService(store)
				}
				return &SystemModule{Handlers: api.NewSystemHandlers(deps), Capabilities: &SystemCapabilities{RBAC: roleProtocolAuthorizer{allowed: allowed}}, cfg: cfg}
			}
			token, err := service.IssueJWT(service.TokenSubject{ID: 1, Username: "operator", RoleIDs: []uint{1}, RoleCodes: []string{"ops"}}, global.OPTIMUS_CONFIG.JWT, time.Now())
			require.NoError(t, err)
			request := func(method, path, body, credential string) *httptest.ResponseRecorder {
				req := httptest.NewRequest(method, strings.TrimRight(prefix, "/")+path, strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				if credential != "" {
					req.Header.Set("Authorization", "Bearer "+credential)
				}
				w := httptest.NewRecorder()
				Routers().ServeHTTP(w, req)
				return w
			}
			routes := []struct{ method, path, body string }{
				{"GET", "/system/roles", ""}, {"POST", "/system/roles", `{"code":"new","name":"New"}`},
				{"PUT", "/system/roles/1", `{"code":"ops","name":"Renamed"}`},
				{"PATCH", "/system/roles/1/status", `{"status":"disabled"}`},
				{"PUT", "/system/roles/1/permissions", `{"permissionCodes":["system.roles.view"]}`},
				{"GET", "/system/permissions/tree", ""},
			}
			for _, route := range routes {
				t.Run(route.method+route.path, func(t *testing.T) {
					w := request(route.method, route.path, route.body, "")
					require.Equal(t, 401, w.Code)
					require.JSONEq(t, `{"code":7,"data":null,"msg":"authorization header is required"}`, w.Body.String())
					allowed = false
					w = request(route.method, route.path, route.body, token)
					require.Equal(t, 403, w.Code)
					require.JSONEq(t, `{"code":7,"data":{},"msg":"permission denied"}`, w.Body.String())
					allowed = true
					w = request(route.method, route.path, route.body, token)
					require.Equal(t, 200, w.Code)
					var envelope struct {
						Code int            `json:"code"`
						Data map[string]any `json:"data"`
						Msg  string         `json:"msg"`
					}
					require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
					require.Equal(t, 0, envelope.Code)
					require.Equal(t, "success", envelope.Msg)
					require.NotNil(t, envelope.Data)
					switch {
					case route.path == "/system/roles" && route.method == "GET":
						require.Len(t, envelope.Data["list"], 2)
						require.Equal(t, float64(2), envelope.Data["total"])
						require.NotContains(t, envelope.Data, "page")
					case route.path == "/system/permissions/tree":
						require.NotEmpty(t, envelope.Data["list"])
						require.NotContains(t, envelope.Data, "total")
					case route.path == "/system/roles/1/permissions":
						require.Empty(t, envelope.Data)
						require.Equal(t, []uint{8}, store.savedPermissions)
					default:
						require.NotZero(t, envelope.Data["id"])
						if route.method == "PATCH" {
							require.Equal(t, "disabled", envelope.Data["status"])
						}
					}
				})
			}
			for _, tc := range []struct {
				method, path, body string
				status             int
				msg                string
			}{
				{"POST", "/system/roles", "{", 400, "invalid role request"},
				{"PUT", "/system/roles/1", "{", 400, "invalid role request"},
				{"PATCH", "/system/roles/1/status", "{", 400, "invalid status request"},
				{"PUT", "/system/roles/1/permissions", "{", 400, "invalid permissions request"},
				{"PUT", "/system/roles/nope", "{}", 400, "invalid id"},
				{"PATCH", "/system/roles/0/status", "{}", 400, "invalid id"},
				{"PUT", "/system/roles/nope/permissions", "{}", 400, "invalid id"},
				{"PUT", "/system/roles/99", "{}", 404, service.ErrRoleNotFound.Error()},
				{"PATCH", "/system/roles/99/status", "{}", 404, service.ErrRoleNotFound.Error()},
				{"PUT", "/system/roles/99/permissions", "{}", 404, service.ErrRoleNotFound.Error()},
				{"PUT", "/system/roles/2", `{"code":"other"}`, 403, service.ErrProtectedAdminRole.Error()},
				{"PATCH", "/system/roles/2/status", `{"status":"disabled"}`, 403, service.ErrProtectedAdminRole.Error()},
				{"PUT", "/system/roles/2/permissions", `{"permissionCodes":[]}`, 403, service.ErrProtectedAdminRole.Error()},
			} {
				w := request(tc.method, tc.path, tc.body, token)
				require.Equal(t, tc.status, w.Code, tc.path)
				expected, _ := json.Marshal(map[string]any{"code": 7, "data": map[string]any{}, "msg": tc.msg})
				require.JSONEq(t, string(expected), w.Body.String())
			}
			store.roles = nil
			w := request("GET", "/system/roles", "", token)
			require.JSONEq(t, `{"code":0,"data":{"list":[],"total":0},"msg":"success"}`, w.Body.String())
			store.fail = true
			for _, route := range routes[:5] {
				w = request(route.method, route.path, route.body, token)
				require.Equal(t, 500, w.Code)
				require.Contains(t, w.Body.String(), `"code":7`)
				require.Contains(t, w.Body.String(), `"data":{}`)
				require.NotContains(t, w.Body.String(), "private database error")
			}
			missing = true
			for _, route := range routes[:5] {
				w = request(route.method, route.path, route.body, token)
				require.Equal(t, 503, w.Code)
				require.JSONEq(t, `{"code":7,"data":{},"msg":"role service is not available"}`, w.Body.String())
			}
			w = request("GET", "/system/ldap/config", "", "")
			require.Equal(t, 401, w.Code)
			require.JSONEq(t, `{"message":"authorization header is required"}`, w.Body.String())
			w = request("GET", "/health", "", "")
			require.Equal(t, 200, w.Code)
			require.NotContains(t, w.Body.String(), `"data"`)
		})
	}
}
