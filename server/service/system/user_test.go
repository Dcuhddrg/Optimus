package system

import (
	"context"
	"encoding/json"
	"testing"

	"optimus/server/global"
	systemModel "optimus/server/model/system"

	"github.com/stretchr/testify/require"
)

func TestUserServiceListUsersOmitsPasswordHash(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, db.Create(&systemModel.SystemUser{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 1}, Username: "admin", DisplayName: "Administrator", Source: "local", Status: "enabled", PasswordHash: "secret-hash"}).Error)
	service := &UserService{}

	result, err := service.ListUsers(context.Background())

	require.NoError(t, err)
	require.Len(t, result.Items, 1)
	require.Equal(t, "admin", result.Items[0].Username)
}

func TestUserServiceDatabaseWouldHaveAdminCapableUserAfterStatus(t *testing.T) {
	db := openTestDB(t)

	adminRole := systemModel.SystemRole{Code: "admin", Name: "Administrator", Status: DefaultRoleStatus}
	require.NoError(t, db.Create(&adminRole).Error)

	user := systemModel.SystemUser{UUID: "u1", Username: "alice", Status: DefaultUserStatus}
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: user.ID, RoleID: adminRole.ID}).Error)

	store := (&UserService{})

	// Disabling the only admin user should leave no admin-capable user
	ok, err := store.wouldHaveAdminCapableUserAfterStatus(context.Background(), user.ID, "disabled")
	require.NoError(t, err)
	require.False(t, ok)

	// Another enabled admin exists so disabling user is safe
	other := systemModel.SystemUser{UUID: "u2", Username: "bob", Status: DefaultUserStatus}
	require.NoError(t, db.Create(&other).Error)
	require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: other.ID, RoleID: adminRole.ID}).Error)

	ok, err = store.wouldHaveAdminCapableUserAfterStatus(context.Background(), user.ID, "disabled")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestUserStatusAuditRecordsSnapshot(t *testing.T) {
	db := openTestDB(t)
	role := systemModel.SystemRole{Code: "admin", Name: "admin", Status: "enabled"}
	admin := systemModel.SystemUser{UUID: "admin-id", Username: "admin", Source: "local", Status: "enabled"}
	target := systemModel.SystemUser{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 7}, UUID: "target-id", Username: "target", Source: "local", Status: "enabled"}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&admin).Error)
	require.NoError(t, db.Create(&target).Error)
	require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: admin.ID, RoleID: role.ID}).Error)
	result, err := (&UserService{}).UpdateUserStatus(context.Background(), 1, 7, "disabled", AuditMeta{ActorUsername: "operator-snapshot"})
	require.NoError(t, err)
	require.Equal(t, "disabled", result.Status)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	require.Equal(t, "operator-snapshot", audit.ActorUsername)
	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
	require.Equal(t, map[string]any{"before": "enabled", "after": "disabled"}, detail["changes"].(map[string]any)["status"])
	require.NotContains(t, audit.Detail, "DO_NOT_LOG_THIS")
}

func TestLastAdminStatusFailureIsAudited(t *testing.T) {
	db := openTestDB(t)
	role := systemModel.SystemRole{Code: "admin", Name: "admin", Status: "enabled"}
	user := systemModel.SystemUser{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 7}, UUID: "only-admin", Username: "admin", Source: "local", Status: "enabled"}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: user.ID, RoleID: role.ID}).Error)
	_, err := (&UserService{}).UpdateUserStatus(context.Background(), 1, 7, "disabled", AuditMeta{})
	require.ErrorIs(t, err, ErrLastAdminCapableUser)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	require.Equal(t, "failure", audit.Result)
	require.Contains(t, audit.Detail, "last_admin_protected")
	require.NoError(t, db.First(&user, 7).Error)
	require.Equal(t, "enabled", user.Status)
}

func TestUserRoleAuditUsesPersistedSnapshot(t *testing.T) {
	db := openTestDB(t)
	roles := []systemModel.SystemRole{
		{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 1}, Code: "admin", Name: "admin", Status: "enabled"},
		{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 2}, Code: "read", Name: "read", Status: "enabled"},
		{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 3}, Code: "write", Name: "write", Status: "enabled"},
	}
	for i := range roles {
		require.NoError(t, db.Create(&roles[i]).Error)
	}
	admin := systemModel.SystemUser{UUID: "admin-id", Username: "admin", Source: "local", Status: "enabled"}
	target := systemModel.SystemUser{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 7}, UUID: "target-id", Username: "target", Source: "local", Status: "enabled"}
	require.NoError(t, db.Create(&admin).Error)
	require.NoError(t, db.Create(&target).Error)
	for _, pair := range []systemModel.SystemUserRole{{UserID: admin.ID, RoleID: 1}, {UserID: target.ID, RoleID: 1}, {UserID: target.ID, RoleID: 2}} {
		require.NoError(t, db.Create(&pair).Error)
	}
	require.NoError(t, (&UserService{}).UpdateUserRoles(context.Background(), 1, 7, []uint{2, 3, 999}, AuditMeta{}))
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
	change := detail["changes"].(map[string]any)["roleIds"].(map[string]any)
	require.Equal(t, []any{float64(1), float64(2)}, change["before"])
	require.Equal(t, []any{float64(2), float64(3)}, change["after"])
	require.Equal(t, []any{float64(3)}, change["added"])
	require.Equal(t, []any{float64(1)}, change["removed"])
}
