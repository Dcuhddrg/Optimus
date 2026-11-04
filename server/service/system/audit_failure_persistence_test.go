package system_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"gorm.io/gorm"
	"optimus/server/global"
	model "optimus/server/model/system"
	request "optimus/server/model/system/request"
	service "optimus/server/service/system"
)

func TestLDAPSaveFailureSurvivesAuditWriteFailure(t *testing.T) {
	db := openExternalPersistenceDB(t)
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
	_, err := service.NewLDAPServiceForTest(t, &fakeLDAPClientUC{}).SaveConfig(context.Background(), 1, request.LDAPConfigRequest{BindPassword: "SECRET"}, service.AuditMeta{})
	require.ErrorIs(t, err, original)
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "audit write failed", logs.All()[0].Message)
	require.NotContains(t, logs.All()[0].ContextMap(), "error")
}

func TestUserUpdateSurvivesAuditWriteFailure(t *testing.T) {
	db := openExternalPersistenceDB(t)
	role := model.SystemRole{Code: "admin", Name: "admin", Status: "enabled"}
	admin := model.SystemUser{UUID: "admin-id", Username: "admin", Source: "local", Status: "enabled"}
	user := model.SystemUser{UUID: "target-id", Username: "target", Source: "local", Status: "enabled"}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&admin).Error)
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&model.SystemUserRole{UserID: admin.ID, RoleID: role.ID}).Error)
	core, logs := observer.New(zap.ErrorLevel)
	oldLog := global.OPTIMUS_LOG
	global.OPTIMUS_LOG = zap.New(core)
	t.Cleanup(func() { global.OPTIMUS_LOG = oldLog })
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("test:fail_audit", func(tx *gorm.DB) {
		if tx.Statement.Table == "sys_audit_logs" {
			tx.AddError(errors.New("DO_NOT_LOG_THIS"))
		}
	}))
	_, err := (&service.UserService{}).UpdateUserStatus(context.Background(), 1, user.ID, "disabled", service.AuditMeta{})
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
	db := openExternalPersistenceDB(t)
	role := model.SystemRole{Code: "admin", Name: "admin", Status: "enabled"}
	admin := model.SystemUser{UUID: "admin-id", Username: "admin", Source: "local", Status: "enabled"}
	user := model.SystemUser{UUID: "target-id", Username: "target", Source: "local", Status: "enabled"}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&admin).Error)
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&model.SystemUserRole{UserID: admin.ID, RoleID: role.ID}).Error)
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
	var attempted model.SystemAuditLog
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("test:fail_audit", func(tx *gorm.DB) {
		if tx.Statement.Table == "sys_audit_logs" {
			attempted = *tx.Statement.Dest.(*model.SystemAuditLog)
			tx.AddError(errors.New("DO_NOT_LOG_THIS"))
		}
	}))
	_, err := (&service.UserService{}).UpdateUserStatus(context.Background(), 1, user.ID, "disabled", service.AuditMeta{})
	require.ErrorIs(t, err, original)
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "failure", logs.All()[0].ContextMap()["result"])
	require.Contains(t, attempted.Detail, "update_failed")
}

func TestSuccessfulLoginSurvivesAuditWriteFailure(t *testing.T) {
	db := openExternalPersistenceDB(t)
	hash, err := hashTestPassword("LOGIN_SECRET")
	require.NoError(t, err)
	user := model.SystemUser{UUID: "admin-id", Username: "admin", Source: "local", Status: "enabled", PasswordHash: hash}
	require.NoError(t, db.Create(&user).Error)
	core, logs := observer.New(zap.ErrorLevel)
	oldLog := global.OPTIMUS_LOG
	global.OPTIMUS_LOG = zap.New(core)
	t.Cleanup(func() { global.OPTIMUS_LOG = oldLog })
	var attempted model.SystemAuditLog
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register("test:fail_audit", func(tx *gorm.DB) {
		if tx.Statement.Table == "sys_audit_logs" {
			attempted = *tx.Statement.Dest.(*model.SystemAuditLog)
			tx.AddError(errors.New("DO_NOT_LOG_THIS"))
		}
	}))
	result, err := service.NewAuthServiceForTest(t).Login(context.Background(), request.LoginRequest{Username: "admin", Password: "LOGIN_SECRET"}, service.AuditMeta{})
	require.NoError(t, err)
	require.NotEmpty(t, result.Token)
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "auth.login.success", logs.All()[0].ContextMap()["action"])
	require.Equal(t, "admin", attempted.ActorUsername)
}

func TestRoleMutationFailureSurvivesAuditWriteFailure(t *testing.T) {
	db := openExternalPersistenceDB(t)
	role := model.SystemRole{Code: "ops", Name: "ops", Status: "enabled"}
	read := model.SystemPermission{Code: "read", Name: "read", Type: "api"}
	write := model.SystemPermission{Code: "write", Name: "write", Type: "api"}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&read).Error)
	require.NoError(t, db.Create(&write).Error)
	require.NoError(t, db.Create(&model.SystemRolePermission{RoleID: role.ID, PermissionID: read.ID}).Error)
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
	err := (&service.RoleService{}).UpdateRolePermissions(context.Background(), 1, role.ID, []string{"write"}, service.AuditMeta{})
	require.ErrorIs(t, err, original)
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "failure", logs.All()[0].ContextMap()["result"])
	var links []model.SystemRolePermission
	require.NoError(t, db.Where("role_id = ?", role.ID).Find(&links).Error)
	require.Len(t, links, 1)
	require.Equal(t, read.ID, links[0].PermissionID)
}
