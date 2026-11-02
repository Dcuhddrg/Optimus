package request_test

import (
	"encoding/json"
	"reflect"
	"testing"

	cloudRequest "optimus/server/model/cloud/request"
	commonRequest "optimus/server/model/common/request"
	systemRequest "optimus/server/model/system/request"

	"github.com/stretchr/testify/require"
)

func TestOperationRequestDTOsPreserveJSONFields(t *testing.T) {
	for _, tc := range []struct {
		value    any
		expected string
	}{
		{systemRequest.UpdateUserStatusRequest{}, `{"status":""}`},
		{systemRequest.UpdateUserRolesRequest{}, `{"roleIds":null}`},
		{systemRequest.UpdateRoleStatusRequest{}, `{"status":""}`},
		{systemRequest.UpdateRolePermissionsRequest{}, `{"permissionCodes":null}`},
		{systemRequest.UpdateUserRolesRequest{RoleIDs: []uint{}}, `{"roleIds":[]}`},
		{systemRequest.UpdateRolePermissionsRequest{PermissionCodes: []string{}}, `{"permissionCodes":[]}`},
	} {
		b, err := json.Marshal(tc.value)
		require.NoError(t, err)
		require.JSONEq(t, tc.expected, string(b))
	}
}

func TestQueryRequestDTOsEmbedSharedPagination(t *testing.T) {
	for _, value := range []any{systemRequest.AuditSearch{}, cloudRequest.AssetSearch{}} {
		field, found := reflect.TypeOf(value).FieldByName("PageInfo")
		require.True(t, found)
		require.True(t, field.Anonymous)
		require.Equal(t, reflect.TypeOf(commonRequest.PageInfo{}), field.Type)
	}
}
