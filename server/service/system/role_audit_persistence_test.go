package system_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"optimus/server/global"
	model "optimus/server/model/system"
	service "optimus/server/service/system"
)

func TestProtectedAdminRoleAudit(t *testing.T) {
	db := openExternalPersistenceDB(t)
	require.NoError(t, db.Create(&model.SystemRole{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 3}, Code: "admin", Name: "admin", Status: "enabled"}).Error)
	err := (&service.RoleService{}).UpdateRolePermissions(context.Background(), 1, 3, []string{"read"}, service.AuditMeta{})
	require.ErrorIs(t, err, service.ErrProtectedAdminRole)
	var audit model.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	require.Equal(t, "failure", audit.Result)
	require.Contains(t, audit.Detail, "admin_role_protected")
}

func TestRolePermissionAuditUsesPersistedCodes(t *testing.T) {
	db := openExternalPersistenceDB(t)
	role := model.SystemRole{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 3}, Code: "ops", Name: "ops", Status: "enabled"}
	read := model.SystemPermission{Code: "read", Name: "read", Type: "api"}
	write := model.SystemPermission{Code: "write", Name: "write", Type: "api"}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&read).Error)
	require.NoError(t, db.Create(&write).Error)
	require.NoError(t, db.Create(&model.SystemRolePermission{RoleID: role.ID, PermissionID: read.ID}).Error)
	require.NoError(t, (&service.RoleService{}).UpdateRolePermissions(context.Background(), 1, 3, []string{"write", "not-a-real-code"}, service.AuditMeta{}))
	var audit model.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
	require.Equal(t, float64(1), detail["version"])
	changes := detail["changes"].(map[string]any)["permissionCodes"].(map[string]any)
	require.Equal(t, []any{"write"}, changes["after"])
	require.Equal(t, []any{"write"}, changes["added"])
	require.Equal(t, []any{"read"}, changes["removed"])
}
