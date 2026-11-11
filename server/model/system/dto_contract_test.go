package system_test

import (
	"encoding/json"
	"net/http/httptest"
	cloudRequest "optimus/server/model/cloud/request"
	cloudResponse "optimus/server/model/cloud/response"
	projectRequest "optimus/server/model/project/request"
	projectResponse "optimus/server/model/project/response"
	systemRequest "optimus/server/model/system/request"
	systemResponse "optimus/server/model/system/response"
	"strings"
	"testing"

	systemModel "optimus/server/model/system"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPublicDTOJSONContracts(t *testing.T) {
	for _, tt := range []struct {
		name  string
		value any
		want  string
	}{
		{"user", systemResponse.SystemUserSummary{}, `{"id":0,"uuid":"","username":"","displayName":"","email":"","department":"","bound":false,"source":"","status":""}`},
		{"currentUser", systemResponse.CurrentUserResult{}, `{"user":{"id":0,"uuid":"","username":"","displayName":"","email":"","department":"","bound":false,"source":"","status":""},"roles":null,"menuPermissions":null}`},
		{"users", systemResponse.UserListResult{Items: []systemResponse.UserListItem{}}, `{"items":[],"total":0}`},
		{"roles", systemResponse.RoleListResult{Items: []systemModel.SystemRole{}}, `{"items":[],"total":0}`},
		{"ldap", systemResponse.LDAPConfigResult{}, `{"enabled":false,"url":"","baseDn":"","bindDn":"","hasBindPassword":false,"userFilter":"","usernameAttr":"","displayNameAttr":"","emailAttr":"","idAttr":"","departmentAttr":"","boundUsers":0}`},
		{"audit", systemResponse.AuditListData{List: []systemModel.AuditLogEntry{}}, `{"list":[],"total":0,"page":0,"pageSize":0}`},
		{"assetPage", cloudResponse.AssetPage{Items: []cloudResponse.AssetView{}}, `{"items":[],"total":0,"page":0,"pageSize":0}`},
		{"account", cloudResponse.AccountView{}, `{"id":0,"projectId":0,"name":"","provider":"","accountId":"","credentialType":"","credential":"","tags":null,"status":"","lastSyncedAt":null}`},
		{"project", projectResponse.ProjectView{}, `{"id":0,"name":"","code":"","description":"","ownerId":0,"status":"","createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}`},
		{"member", projectResponse.ProjectMember{}, `{"userId":0,"username":"","displayName":"","projectRole":""}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.value)
			require.NoError(t, err)
			require.JSONEq(t, tt.want, string(b))
		})
	}
}

func TestRequestBindingContracts(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"username":"alice","password":"secret"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	var login systemRequest.LoginRequest
	require.NoError(t, c.ShouldBindJSON(&login))
	require.Equal(t, "alice", login.Username)
	require.Equal(t, "secret", login.Password)

	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{}`))
	var account cloudRequest.CreateAccountRequest
	require.Error(t, c.ShouldBindJSON(&account))

	var project projectRequest.UpdateProjectRequest
	require.NoError(t, json.Unmarshal([]byte(`{"name":""}`), &project))
	require.NotNil(t, project.Name)
	require.Equal(t, "", *project.Name)
	require.Nil(t, project.Description)
}
