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

func TestUserStatusAuditRecordsSnapshot(t *testing.T) {
	db := openExternalPersistenceDB(t)
	role := model.SystemRole{Code: "admin", Name: "admin", Status: "enabled"}
	admin := model.SystemUser{UUID: "admin-id", Username: "admin", Source: "local", Status: "enabled"}
	target := model.SystemUser{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 7}, UUID: "target-id", Username: "target", Source: "local", Status: "enabled"}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&admin).Error)
	require.NoError(t, db.Create(&target).Error)
	require.NoError(t, db.Create(&model.SystemUserRole{UserID: admin.ID, RoleID: role.ID}).Error)
	result, err := (&service.UserService{}).UpdateUserStatus(context.Background(), 1, 7, "disabled", service.AuditMeta{ActorUsername: "operator-snapshot"})
	require.NoError(t, err)
	require.Equal(t, "disabled", result.Status)
	var audit model.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	require.Equal(t, "operator-snapshot", audit.ActorUsername)
	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
	require.Equal(t, map[string]any{"before": "enabled", "after": "disabled"}, detail["changes"].(map[string]any)["status"])
	require.NotContains(t, audit.Detail, "DO_NOT_LOG_THIS")
}

func TestLastAdminStatusFailureIsAudited(t *testing.T) {
	db := openExternalPersistenceDB(t)
	role := model.SystemRole{Code: "admin", Name: "admin", Status: "enabled"}
	user := model.SystemUser{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 7}, UUID: "only-admin", Username: "admin", Source: "local", Status: "enabled"}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&model.SystemUserRole{UserID: user.ID, RoleID: role.ID}).Error)
	_, err := (&service.UserService{}).UpdateUserStatus(context.Background(), 1, 7, "disabled", service.AuditMeta{})
	require.ErrorIs(t, err, service.ErrLastAdminCapableUser)
	var audit model.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	require.Equal(t, "failure", audit.Result)
	require.Contains(t, audit.Detail, "last_admin_protected")
	require.NoError(t, db.First(&user, 7).Error)
	require.Equal(t, "enabled", user.Status)
}

func TestUserRoleAuditUsesPersistedSnapshot(t *testing.T) {
	db := openExternalPersistenceDB(t)
	roles := []model.SystemRole{
		{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 1}, Code: "admin", Name: "admin", Status: "enabled"},
		{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 2}, Code: "read", Name: "read", Status: "enabled"},
		{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 3}, Code: "write", Name: "write", Status: "enabled"},
	}
	for i := range roles {
		require.NoError(t, db.Create(&roles[i]).Error)
	}
	admin := model.SystemUser{UUID: "admin-id", Username: "admin", Source: "local", Status: "enabled"}
	target := model.SystemUser{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 7}, UUID: "target-id", Username: "target", Source: "local", Status: "enabled"}
	require.NoError(t, db.Create(&admin).Error)
	require.NoError(t, db.Create(&target).Error)
	for _, pair := range []model.SystemUserRole{{UserID: admin.ID, RoleID: 1}, {UserID: target.ID, RoleID: 1}, {UserID: target.ID, RoleID: 2}} {
		require.NoError(t, db.Create(&pair).Error)
	}
	require.NoError(t, (&service.UserService{}).UpdateUserRoles(context.Background(), 1, 7, []uint{2, 3, 999}, service.AuditMeta{}))
	var audit model.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
	change := detail["changes"].(map[string]any)["roleIds"].(map[string]any)
	require.Equal(t, []any{float64(1), float64(2)}, change["before"])
	require.Equal(t, []any{float64(2), float64(3)}, change["after"])
	require.Equal(t, []any{float64(3)}, change["added"])
	require.Equal(t, []any{float64(1)}, change["removed"])
}
