package system

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
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

func TestBuiltinAdminCannotBeDisabledOrLoseAdminRole(t *testing.T) {
	db := openTestDB(t)
	adminRole := systemModel.SystemRole{Code: AdminRoleCode, Name: "admin", Status: "enabled"}
	userRole := systemModel.SystemRole{Code: "user", Name: "user", Status: "enabled"}
	require.NoError(t, db.Create(&adminRole).Error)
	require.NoError(t, db.Create(&userRole).Error)
	builtin := systemModel.SystemUser{UUID: BuiltinAdminUUID, Username: "admin", Source: "local", Status: "enabled"}
	require.NoError(t, db.Create(&builtin).Error)
	require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: builtin.ID, RoleID: adminRole.ID}).Error)

	_, err := (&UserService{}).UpdateUserStatus(context.Background(), 1, builtin.ID, "disabled", AuditMeta{})
	require.ErrorIs(t, err, ErrBuiltinAdminProtected)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	require.Equal(t, "failure", audit.Result)
	require.Contains(t, audit.Detail, "builtin_admin_protected")

	err = (&UserService{}).UpdateUserRoles(context.Background(), 1, builtin.ID, []uint{userRole.ID}, AuditMeta{})
	require.ErrorIs(t, err, ErrBuiltinAdminProtected)
	require.NoError(t, db.Last(&audit).Error)
	require.Contains(t, audit.Detail, "builtin_admin_protected")

	require.NoError(t, db.First(&builtin, builtin.ID).Error)
	require.Equal(t, "enabled", builtin.Status)
	roles, err := queryUserRoles(context.Background(), builtin.ID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	require.Equal(t, AdminRoleCode, roles[0].Code)

	// Adding roles while keeping admin is allowed.
	require.NoError(t, (&UserService{}).UpdateUserRoles(context.Background(), 1, builtin.ID, []uint{adminRole.ID, userRole.ID}, AuditMeta{}))
}

func TestLDAPAdminsCanAllBeDisabledWhileBuiltinAdminRemains(t *testing.T) {
	db := openTestDB(t)
	adminRole := systemModel.SystemRole{Code: AdminRoleCode, Name: "admin", Status: "enabled"}
	require.NoError(t, db.Create(&adminRole).Error)
	builtin := systemModel.SystemUser{UUID: BuiltinAdminUUID, Username: "admin", Source: "local", Status: "enabled"}
	ldapAdmin := systemModel.SystemUser{UUID: "ldap:alice", Username: "alice", Source: "ldap", Status: "enabled"}
	require.NoError(t, db.Create(&builtin).Error)
	require.NoError(t, db.Create(&ldapAdmin).Error)
	require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: builtin.ID, RoleID: adminRole.ID}).Error)
	require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: ldapAdmin.ID, RoleID: adminRole.ID}).Error)

	_, err := (&UserService{}).UpdateUserStatus(context.Background(), 1, ldapAdmin.ID, "disabled", AuditMeta{})
	require.NoError(t, err)
	require.NoError(t, (&UserService{}).UpdateUserRoles(context.Background(), 1, ldapAdmin.ID, []uint{}, AuditMeta{}))
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

// registerBeforeUserUpdate runs fn once, just before the first UPDATE of
// sys_users, to interleave a concurrent write between a read and its write.
func registerBeforeUserUpdate(t *testing.T, db *gorm.DB, fn func()) {
	t.Helper()
	fired := false
	name := "test:before_user_update_" + t.Name()
	require.NoError(t, db.Callback().Update().Before("gorm:update").Register(name, func(tx *gorm.DB) {
		if fired || tx.Statement.Table != "sys_users" {
			return
		}
		fired = true
		fn()
	}))
	t.Cleanup(func() { _ = db.Callback().Update().Remove(name) })
}

func TestProfileUpdateDoesNotRestoreConcurrentDisable(t *testing.T) {
	f := newAuthStateFixture(t)
	registerBeforeUserUpdate(t, f.db, func() {
		_, err := (&UserService{}).UpdateUserStatus(context.Background(), 1, f.user.ID, "disabled", AuditMeta{})
		require.NoError(t, err)
	})

	updated, err := (&UserService{}).UpdateUser(context.Background(), 1, f.user.ID, systemRequest.UpdateUserRequest{DisplayName: "Alice", Email: "alice@example.com"}, AuditMeta{})
	require.NoError(t, err)

	require.Equal(t, "disabled", updated.Status)
	require.Equal(t, "Alice", updated.DisplayName)
	require.False(t, loadState(t, f.user.ID).Enabled())
}

func TestLDAPLoginSyncDoesNotRestoreConcurrentDisable(t *testing.T) {
	f := newAuthStateFixture(t)
	registerBeforeUserUpdate(t, f.db, func() {
		_, err := (&UserService{}).UpdateUserStatus(context.Background(), 1, f.user.ID, "disabled", AuditMeta{})
		require.NoError(t, err)
	})

	synced, err := (&AuthService{}).syncLDAPUser(context.Background(), LDAPUser{Username: f.user.Username, DisplayName: "Alice", Email: "alice@example.com"})
	require.NoError(t, err)
	require.Equal(t, "disabled", synced.Status, "login decides on the re-read status")

	var stored systemModel.SystemUser
	require.NoError(t, f.db.First(&stored, f.user.ID).Error)
	require.Equal(t, "disabled", stored.Status)
	require.Equal(t, "Alice", stored.DisplayName)
	require.False(t, loadState(t, f.user.ID).Enabled())
}

// The status UPDATE commits but re-reading the row fails (request cancelled,
// connection lost): the cache must still be invalidated, or a cached
// "enabled" keeps the disabled user's token working.
func TestStatusWriteInvalidatesCacheEvenIfRereadFails(t *testing.T) {
	f := newAuthStateFixture(t)
	require.True(t, loadState(t, f.user.ID).Enabled()) // cached as enabled
	updated := false
	require.NoError(t, f.db.Callback().Update().After("gorm:update").Register("test:mark_updated", func(tx *gorm.DB) {
		if tx.Statement.Table == "sys_users" {
			updated = true
		}
	}))
	require.NoError(t, f.db.Callback().Query().Before("gorm:query").Register("test:fail_reread", func(tx *gorm.DB) {
		if updated && tx.Statement.Table == "sys_users" {
			tx.AddError(errors.New("connection lost"))
		}
	}))
	t.Cleanup(func() {
		_ = f.db.Callback().Update().Remove("test:mark_updated")
		_ = f.db.Callback().Query().Remove("test:fail_reread")
	})

	_, err := (&UserService{}).UpdateUserStatus(context.Background(), 1, f.user.ID, "disabled", AuditMeta{})
	require.Error(t, err)

	updated = false
	require.False(t, loadState(t, f.user.ID).Enabled())
}
