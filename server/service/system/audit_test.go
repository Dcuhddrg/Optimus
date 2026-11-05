package system

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"testing"
	"time"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/gorm"
)

func TestAuditServiceQueryLogsNormalizesPagination(t *testing.T) {
	openTestDB(t)
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "login", ActorUsername: "admin"}))

	page, err := (&AuditService{}).QueryLogs(context.Background(), AuditFilter{})
	require.NoError(t, err)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 20, page.PageSize)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "admin", page.Items[0].ActorUsername)
}

func TestAuditServiceDatabaseQueryLogsOrdersNewestFirst(t *testing.T) {
	openTestDB(t)

	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "first", Result: "success"}))
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "second", Result: "success"}))

	store := (&AuditService{})
	page, err := store.QueryLogs(context.Background(), AuditFilter{Page: 1, PageSize: 20})

	require.NoError(t, err)
	require.Equal(t, int64(2), page.Total)
	require.Len(t, page.Items, 2)
	require.True(t, page.Items[0].CreatedAt.After(page.Items[1].CreatedAt) || page.Items[0].CreatedAt.Equal(page.Items[1].CreatedAt),
		"audit logs should be ordered newest first")
}

func TestAuditServiceDatabaseQueryLogsPreservesActorUsername(t *testing.T) {
	openTestDB(t)
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{
		ActorUserID:   42,
		ActorUsername: "alice",
		Action:        "login",
		Result:        "success",
	}))

	store := (&AuditService{})
	page, err := store.QueryLogs(context.Background(), AuditFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, "alice", page.Items[0].ActorUsername)
}

func TestAuditServiceDatabaseQueryLogsPaginates(t *testing.T) {
	openTestDB(t)
	for i := 0; i < 5; i++ {
		require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "x", Result: "success"}))
	}

	store := (&AuditService{})
	page, err := store.QueryLogs(context.Background(), AuditFilter{Page: 1, PageSize: 2})

	require.NoError(t, err)
	require.Equal(t, int64(5), page.Total)
	require.Len(t, page.Items, 2)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 2, page.PageSize)
}

func TestAuditServiceDatabaseQueryLogsFiltersByResult(t *testing.T) {
	openTestDB(t)
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "a", Result: "success"}))
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "b", Result: "failure"}))

	store := (&AuditService{})
	result := "success"
	page, err := store.QueryLogs(context.Background(), AuditFilter{
		Page: 1, PageSize: 20, Result: &result,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)
	require.Equal(t, "success", page.Items[0].Result)
}

func TestAuditServiceDatabaseQueryLogsFiltersByAction(t *testing.T) {
	openTestDB(t)
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "login", Result: "success"}))
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "logout", Result: "success"}))

	store := (&AuditService{})
	action := "login"
	page, err := store.QueryLogs(context.Background(), AuditFilter{
		Page: 1, PageSize: 20, Action: &action,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "login", page.Items[0].Action)
}

func TestAuditServiceDatabaseQueryLogsFiltersByTimeRange(t *testing.T) {
	openTestDB(t)
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{Action: "login", Result: "success"}))

	store := (&AuditService{})
	future := time.Now().Add(time.Hour)
	past := time.Now().Add(-time.Hour)

	pagePast, err := store.QueryLogs(context.Background(), AuditFilter{
		Page: 1, PageSize: 20, EndTime: &past,
	})
	require.NoError(t, err)
	require.Equal(t, int64(0), pagePast.Total)

	pageNow, err := store.QueryLogs(context.Background(), AuditFilter{
		Page: 1, PageSize: 20, StartTime: &past, EndTime: &future,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), pageNow.Total)
}

func TestAuditServiceDatabaseExportLogsReturnsSortedEntries(t *testing.T) {
	openTestDB(t)
	require.NoError(t, writeAuditEvent(context.Background(), AuditEvent{
		ActorUsername: "alice", Action: "login", Result: "success",
	}))

	store := (&AuditService{})
	entries, err := store.ExportLogs(context.Background(), AuditFilter{})

	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "alice", entries[0].ActorUsername)
	require.Equal(t, "login", entries[0].Action)
}

func TestAuditRecordFallbackDoesNotLeak(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, db.Migrator().DropTable("sys_audit_logs"))
	old := global.OPTIMUS_LOG
	global.OPTIMUS_LOG = nil
	t.Cleanup(func() { global.OPTIMUS_LOG = old })
	out := log.Writer()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(out) })
	recordAudit(context.Background(), AuditEvent{Action: "role.permissions.update", ResourceType: "role", ResourceID: "3", Result: "success", Detail: "SECRET"})
	require.Contains(t, buf.String(), "audit write failed")
	require.Contains(t, buf.String(), "role.permissions.update")
	require.NotContains(t, buf.String(), "SECRET")
}
func TestAuditDetailsDropsNonWhitelistedFields(t *testing.T) {
	raw := auditDetails("updated", map[string]auditChange{"password": {Before: "SECRET", After: "NEW_SECRET"}, "token": {Before: "TOKEN", After: "NEW_TOKEN"}, "status": {Before: "enabled", After: "disabled"}})
	require.NotContains(t, raw, "SECRET")
	require.NotContains(t, raw, "TOKEN")
	require.Contains(t, raw, "status")
}
func TestAuditDetailsEncodingFailureStillProducesValidDetail(t *testing.T) {
	raw := auditDetails("updated", map[string]auditChange{"status": {Before: make(chan int), After: "enabled"}})
	var detail auditRecordDetail
	require.NoError(t, json.Unmarshal([]byte(raw), &detail))
	require.Equal(t, 1, detail.Version)
	require.Equal(t, "detail_unavailable", detail.Reason)
}
func TestAuditSetChangeCanonicalizesAndDoesNotModifyInput(t *testing.T) {
	before := []uint{3, 1, 1}
	after := []uint{4, 3, 4}
	change := auditSetChange(before, after)
	require.Equal(t, []uint{3, 1, 1}, before)
	require.Equal(t, []uint{4, 3, 4}, after)
	require.Equal(t, []uint{1, 3}, change.Before)
	require.Equal(t, []uint{3, 4}, change.After)
	require.Equal(t, []uint{4}, change.Added)
	require.Equal(t, []uint{1}, change.Removed)
}

func TestPermissionCodesSnapshotUsesSavedAssignments(t *testing.T) {
	db := openTestDB(t)
	roles := []systemModel.SystemRole{{Code: "first", Name: "First", Status: "enabled"}, {Code: "second", Name: "Second", Status: "enabled"}}
	require.NoError(t, db.Create(&roles).Error)
	permissions := []systemModel.SystemPermission{{Code: "z.write", Name: "Write", Type: "api", Method: "POST", Path: "/things"}, {Code: "a.read", Name: "Read", Type: "api", Method: "GET", Path: "/things"}}
	require.NoError(t, db.Create(&permissions).Error)
	store := (&RoleService{})
	ctx := context.Background()
	require.NoError(t, store.replaceRolePermissions(ctx, roles[0].ID, permissions))
	codes, err := store.permissionCodesForRole(ctx, roles[0].ID)
	require.NoError(t, err)
	require.Equal(t, []string{"a.read", "z.write"}, codes)
	other, err := store.permissionCodesForRole(ctx, roles[1].ID)
	require.NoError(t, err)
	require.Empty(t, other)
	require.NotNil(t, other)
	require.NoError(t, store.replaceRolePermissions(ctx, roles[0].ID, permissions[1:]))
	codes, err = store.permissionCodesForRole(ctx, roles[0].ID)
	require.NoError(t, err)
	require.Equal(t, []string{"a.read"}, codes)
}

func TestLDAPSaveFailureSurvivesAuditWriteFailure(t *testing.T) {
	db := openTestDB(t)
	core, logs := observer.New(zap.ErrorLevel)
	oldLog := global.OPTIMUS_LOG
	global.OPTIMUS_LOG = zap.New(core)
	t.Cleanup(func() { global.OPTIMUS_LOG = oldLog })
	original := errors.New("original save error")
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("test:fail_writes", func(tx *gorm.DB) {
		switch tx.Statement.Table {
		case "sys_ldap_config":
			tx.AddError(original)
		case "sys_audit_logs":
			tx.AddError(errors.New("DO_NOT_LOG_THIS"))
		}
	}))
	_, err := newLDAPServiceForTest(t, &fakeLDAPClientUC{}).SaveConfig(context.Background(), 1, systemRequest.LDAPConfigRequest{BindPassword: "SECRET"}, AuditMeta{})
	require.ErrorIs(t, err, original)
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "audit write failed", logs.All()[0].Message)
	require.NotContains(t, logs.All()[0].ContextMap(), "error")
}

func TestUserUpdateSurvivesAuditWriteFailure(t *testing.T) {
	db := openTestDB(t)
	role := systemModel.SystemRole{Code: "admin", Name: "admin", Status: "enabled"}
	admin := systemModel.SystemUser{UUID: "admin-id", Username: "admin", Source: "local", Status: "enabled"}
	user := systemModel.SystemUser{UUID: "target-id", Username: "target", Source: "local", Status: "enabled"}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&admin).Error)
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: admin.ID, RoleID: role.ID}).Error)
	core, logs := observer.New(zap.ErrorLevel)
	oldLog := global.OPTIMUS_LOG
	global.OPTIMUS_LOG = zap.New(core)
	t.Cleanup(func() { global.OPTIMUS_LOG = oldLog })
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("test:fail_audit", func(tx *gorm.DB) {
		if tx.Statement.Table == "sys_audit_logs" {
			tx.AddError(errors.New("DO_NOT_LOG_THIS"))
		}
	}))
	_, err := (&UserService{}).UpdateUserStatus(context.Background(), 1, user.ID, "disabled", AuditMeta{})
	require.NoError(t, err)
	require.NoError(t, db.First(&user, user.ID).Error)
	require.Equal(t, "disabled", user.Status)
	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	require.Equal(t, "audit write failed", entry.Message)
	require.Equal(t, "user.status.update", entry.ContextMap()["action"])
	require.NotContains(t, entry.ContextMap(), "detail")
	require.NotContains(t, entry.ContextMap(), "error")
}

func TestUserUpdateFailureSurvivesAuditWriteFailure(t *testing.T) {
	db := openTestDB(t)
	role := systemModel.SystemRole{Code: "admin", Name: "admin", Status: "enabled"}
	admin := systemModel.SystemUser{UUID: "admin-id", Username: "admin", Source: "local", Status: "enabled"}
	user := systemModel.SystemUser{UUID: "target-id", Username: "target", Source: "local", Status: "enabled"}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&admin).Error)
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: admin.ID, RoleID: role.ID}).Error)
	core, logs := observer.New(zap.ErrorLevel)
	oldLog := global.OPTIMUS_LOG
	global.OPTIMUS_LOG = zap.New(core)
	t.Cleanup(func() { global.OPTIMUS_LOG = oldLog })
	original := errors.New("original update failure")
	require.NoError(t, db.Callback().Update().Before("gorm:update").Register("test:fail_user", func(tx *gorm.DB) {
		if tx.Statement.Table == "sys_users" {
			tx.AddError(original)
		}
	}))
	var attempted systemModel.SystemAuditLog
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("test:fail_audit", func(tx *gorm.DB) {
		if tx.Statement.Table == "sys_audit_logs" {
			attempted = *tx.Statement.Dest.(*systemModel.SystemAuditLog)
			tx.AddError(errors.New("DO_NOT_LOG_THIS"))
		}
	}))
	_, err := (&UserService{}).UpdateUserStatus(context.Background(), 1, user.ID, "disabled", AuditMeta{})
	require.ErrorIs(t, err, original)
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "failure", logs.All()[0].ContextMap()["result"])
	require.Contains(t, attempted.Detail, "update_failed")
}

func TestSuccessfulLoginSurvivesAuditWriteFailure(t *testing.T) {
	db := openTestDB(t)
	hash, err := hashTestPassword("LOGIN_SECRET")
	require.NoError(t, err)
	user := systemModel.SystemUser{UUID: "admin-id", Username: "admin", Source: "local", Status: "enabled", PasswordHash: hash}
	require.NoError(t, db.Create(&user).Error)
	core, logs := observer.New(zap.ErrorLevel)
	oldLog := global.OPTIMUS_LOG
	global.OPTIMUS_LOG = zap.New(core)
	t.Cleanup(func() { global.OPTIMUS_LOG = oldLog })
	var attempted systemModel.SystemAuditLog
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("test:fail_audit", func(tx *gorm.DB) {
		if tx.Statement.Table == "sys_audit_logs" {
			attempted = *tx.Statement.Dest.(*systemModel.SystemAuditLog)
			tx.AddError(errors.New("DO_NOT_LOG_THIS"))
		}
	}))
	result, err := newAuthServiceForTest(t).Login(context.Background(), systemRequest.LoginRequest{Username: "admin", Password: "LOGIN_SECRET"}, AuditMeta{})
	require.NoError(t, err)
	require.NotEmpty(t, result.Token)
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "auth.login.success", logs.All()[0].ContextMap()["action"])
	require.Equal(t, "admin", attempted.ActorUsername)
}

func TestRoleMutationFailureSurvivesAuditWriteFailure(t *testing.T) {
	db := openTestDB(t)
	role := systemModel.SystemRole{Code: "ops", Name: "ops", Status: "enabled"}
	read := systemModel.SystemPermission{Code: "read", Name: "read", Type: "api", Method: "GET", Path: "/things"}
	write := systemModel.SystemPermission{Code: "write", Name: "write", Type: "api", Method: "POST", Path: "/things"}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&read).Error)
	require.NoError(t, db.Create(&write).Error)
	require.NoError(t, db.Create(&systemModel.SystemRolePermission{RoleID: role.ID, PermissionID: read.ID}).Error)
	core, logs := observer.New(zap.ErrorLevel)
	oldLog := global.OPTIMUS_LOG
	global.OPTIMUS_LOG = zap.New(core)
	t.Cleanup(func() { global.OPTIMUS_LOG = oldLog })
	original := errors.New("original role failure")
	require.NoError(t, db.Callback().Delete().Before("gorm:delete").Register("test:fail_role_permissions", func(tx *gorm.DB) {
		if tx.Statement.Table == "sys_role_permissions" {
			tx.AddError(original)
		}
	}))
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("test:fail_audit", func(tx *gorm.DB) {
		if tx.Statement.Table == "sys_audit_logs" {
			tx.AddError(errors.New("DO_NOT_LOG_THIS"))
		}
	}))
	err := (&RoleService{}).UpdateRolePermissions(context.Background(), 1, role.ID, []string{"write"}, AuditMeta{})
	require.ErrorIs(t, err, original)
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "failure", logs.All()[0].ContextMap()["result"])
	var links []systemModel.SystemRolePermission
	require.NoError(t, db.Where("role_id = ?", role.ID).Find(&links).Error)
	require.Len(t, links, 1)
	require.Equal(t, read.ID, links[0].PermissionID)
}

func TestRolePermissionSnapshotFailuresPreserveMutation(t *testing.T) {
	for _, failAt := range []int{1, 2} {
		t.Run(string(rune('0'+failAt)), func(t *testing.T) {
			db := openTestDB(t)
			role := systemModel.SystemRole{Code: "ops", Name: "ops", Status: "enabled"}
			permission := systemModel.SystemPermission{Code: "write", Name: "write", Type: "api"}
			require.NoError(t, db.Create(&role).Error)
			require.NoError(t, db.Create(&permission).Error)
			reads := 0
			require.NoError(t, db.Callback().Query().Before("gorm:query").Register("test:fail_permission_snapshot", func(tx *gorm.DB) {
				_, pluckingCodes := tx.Statement.Dest.(*[]string)
				if tx.Statement.Table == "sys_permissions" && pluckingCodes {
					reads++
					if reads == failAt {
						tx.AddError(errors.New("snapshot failure"))
					}
				}
			}))
			require.NoError(t, (&RoleService{}).UpdateRolePermissions(context.Background(), 1, role.ID, []string{"write"}, AuditMeta{}))
			require.Equal(t, 2, reads)
			var audit systemModel.SystemAuditLog
			require.NoError(t, db.Last(&audit).Error)
			require.Equal(t, "success", audit.Result)
			var detail map[string]any
			require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
			require.Equal(t, "snapshot_unavailable", detail["reason"])
			require.Empty(t, detail["changes"])
		})
	}
}

func TestUserRoleSnapshotFailuresPreserveMutation(t *testing.T) {
	for _, failAt := range []int{1, 2} {
		t.Run(string(rune('0'+failAt)), func(t *testing.T) {
			db := openTestDB(t)
			adminRole := systemModel.SystemRole{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 1}, Code: "admin", Name: "admin", Status: "enabled"}
			role := systemModel.SystemRole{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 2}, Code: "ops", Name: "ops", Status: "enabled"}
			admin := systemModel.SystemUser{UUID: "admin-id", Username: "admin", Source: "local", Status: "enabled"}
			user := systemModel.SystemUser{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 7}, UUID: "target-id", Username: "target", Source: "local", Status: "enabled"}
			require.NoError(t, db.Create(&adminRole).Error)
			require.NoError(t, db.Create(&role).Error)
			require.NoError(t, db.Create(&admin).Error)
			require.NoError(t, db.Create(&user).Error)
			require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: admin.ID, RoleID: adminRole.ID}).Error)
			reads := 0
			require.NoError(t, db.Callback().Row().Before("gorm:row").Register("test:fail_user_snapshot", func(tx *gorm.DB) {
				if tx.Statement.Table == "roles" {
					reads++
					if reads == failAt {
						tx.AddError(errors.New("snapshot failure"))
					}
				}
			}))
			require.NoError(t, (&UserService{}).UpdateUserRoles(context.Background(), 1, user.ID, []uint{2}, AuditMeta{}))
			require.Equal(t, 2, reads)
			var audit systemModel.SystemAuditLog
			require.NoError(t, db.Last(&audit).Error)
			require.Equal(t, "success", audit.Result)
			var detail map[string]any
			require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
			require.Equal(t, "snapshot_unavailable", detail["reason"])
			require.Empty(t, detail["changes"])
		})
	}
}
