package system

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"
	"optimus/server/utils"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/stretchr/testify/require"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestRoleServiceCreateRoleWritesAudit(t *testing.T) {
	db := openTestDB(t)
	service := &RoleService{}

	role, err := service.CreateRole(context.Background(), 1, systemRequest.RoleRequest{
		Code: "ops",
		Name: "Operations",
	}, AuditMeta{})

	require.NoError(t, err)
	require.Equal(t, "ops", role.Code)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Where("action = ?", "role.create").First(&audit).Error)
}

// Preserve existing fixture compatibility with the new read-only snapshot contract.

func TestRoleServiceDatabaseReplaceRolePermissions(t *testing.T) {
	db := openTestDB(t)

	role := systemModel.SystemRole{Code: "editor", Name: "Editor"}
	require.NoError(t, db.Create(&role).Error)

	p1 := systemModel.SystemPermission{Code: "content.create", Name: "Create Content", Type: "api", Method: "POST", Path: "/content"}
	p2 := systemModel.SystemPermission{Code: "content.delete", Name: "Delete Content", Type: "api", Method: "DELETE", Path: "/content/:id"}
	menu := systemModel.SystemPermission{Code: "content.menu", Name: "Content", Type: "menu", Path: "/content"}
	require.NoError(t, db.Create(&[]systemModel.SystemPermission{p1, p2, menu}).Error)
	require.NoError(t, db.Where("code = ?", "content.create").First(&p1).Error)
	require.NoError(t, db.Where("code = ?", "content.delete").First(&p2).Error)
	require.NoError(t, db.Where("code = ?", "content.menu").First(&menu).Error)

	store := (&RoleService{})
	countRows := func() (menus, policies int64) {
		require.NoError(t, db.Model(&systemModel.SystemRolePermission{}).Where("role_id = ?", role.ID).Count(&menus).Error)
		require.NoError(t, db.Model(&gormadapter.CasbinRule{}).Where("ptype = ? AND v0 = ?", "p", strconv.Itoa(int(role.ID))).Count(&policies).Error)
		return
	}

	// API grants become Casbin policies; menu grants stay in sys_role_permissions.
	require.NoError(t, store.replaceRolePermissions(context.Background(), role.ID, []systemModel.SystemPermission{p1, menu}))
	menus, policies := countRows()
	require.Equal(t, int64(1), menus)
	require.Equal(t, int64(1), policies)

	require.NoError(t, store.replaceRolePermissions(context.Background(), role.ID, []systemModel.SystemPermission{p1, p2}))
	menus, policies = countRows()
	require.Zero(t, menus)
	require.Equal(t, int64(2), policies)
	allowed, err := utils.GetCasbin().Enforce(strconv.Itoa(int(role.ID)), "/content/42", "DELETE")
	require.NoError(t, err)
	require.True(t, allowed, "the enforcer reloads after the update")
}

func TestProtectedAdminRoleAudit(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, db.Create(&systemModel.SystemRole{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 3}, Code: "admin", Name: "admin"}).Error)
	err := (&RoleService{}).UpdateRolePermissions(context.Background(), 1, 3, []string{"read"}, AuditMeta{})
	require.ErrorIs(t, err, ErrProtectedAdminRole)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	require.Equal(t, "failure", audit.Result)
	require.Contains(t, audit.Detail, "admin_role_protected")
}

func TestRolePermissionAuditUsesPersistedCodes(t *testing.T) {
	db := openTestDB(t)
	role := systemModel.SystemRole{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 3}, Code: "ops", Name: "ops"}
	read := systemModel.SystemPermission{Code: "read", Name: "read", Type: "api", Method: "GET", Path: "/things"}
	write := systemModel.SystemPermission{Code: "write", Name: "write", Type: "api", Method: "POST", Path: "/things"}
	require.NoError(t, db.Create(&role).Error)
	require.NoError(t, db.Create(&read).Error)
	require.NoError(t, db.Create(&write).Error)
	require.NoError(t, db.Create(&systemModel.SystemRolePermission{RoleID: role.ID, PermissionID: read.ID}).Error)
	require.NoError(t, (&RoleService{}).UpdateRolePermissions(context.Background(), 1, 3, []string{"write", "not-a-real-code"}, AuditMeta{}))
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
	require.Equal(t, float64(1), detail["version"])
	changes := detail["changes"].(map[string]any)["permissionCodes"].(map[string]any)
	require.Equal(t, []any{"write"}, changes["after"])
	require.Equal(t, []any{"write"}, changes["added"])
	require.Equal(t, []any{"read"}, changes["removed"])
}

func TestRoleServiceQueriesStartupDatabase(t *testing.T) {
	db, err := gorm.Open(gormmysql.New(gormmysql.Config{
		DSN:                       "unused:unused@tcp(127.0.0.1:3306)/unused",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	oldDB := global.OPTIMUS_DB
	t.Cleanup(func() { global.OPTIMUS_DB = oldDB })
	global.OPTIMUS_DB = db

	var query string
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("test:capture", func(tx *gorm.DB) {
		query = tx.Statement.SQL.String()
	}))

	require.NotPanics(t, func() {
		_, err := (&RoleService{}).ListRoles(context.Background())
		require.NoError(t, err)
	})
	require.Contains(t, query, "sys_roles")
	require.Contains(t, query, "ORDER BY id ASC")
}

func TestDeleteRoleProtectsAdminUserAndRolesInUse(t *testing.T) {
	f := newAuthStateFixture(t)
	userRole := systemModel.SystemRole{Code: DefaultUserRoleCode, Name: "user"}
	require.NoError(t, f.db.Create(&userRole).Error)
	svc := &RoleService{}

	require.ErrorIs(t, svc.DeleteRole(context.Background(), 1, f.adminRole.ID, AuditMeta{}), ErrProtectedRole)
	require.ErrorIs(t, svc.DeleteRole(context.Background(), 1, userRole.ID, AuditMeta{}), ErrProtectedRole)
	require.ErrorIs(t, svc.DeleteRole(context.Background(), 1, f.opsRole.ID, AuditMeta{}), ErrRoleInUse, "alice holds ops")
	var audit systemModel.SystemAuditLog
	require.NoError(t, f.db.Last(&audit).Error)
	require.Equal(t, "failure", audit.Result)
	require.Contains(t, audit.Detail, "role_in_use")
	var count int64
	require.NoError(t, f.db.Model(&systemModel.SystemRole{}).Count(&count).Error)
	require.Equal(t, int64(3), count)
}

func TestDeleteRoleRemovesGrantsAndPolicies(t *testing.T) {
	f := newAuthStateFixture(t)
	seedUsersViewPermission(t, f.db)
	temp := systemModel.SystemRole{Code: "temp", Name: "temp"}
	require.NoError(t, f.db.Create(&temp).Error)
	svc := &RoleService{}
	require.NoError(t, svc.UpdateRolePermissions(context.Background(), 1, temp.ID, []string{"system.users.view", "system.users.menu"}, AuditMeta{}))
	require.True(t, enforce(t, temp.ID, "/system/users", "GET"))

	require.NoError(t, svc.DeleteRole(context.Background(), 1, temp.ID, AuditMeta{}))

	var count int64
	require.NoError(t, f.db.Model(&systemModel.SystemRole{}).Where("id = ?", temp.ID).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, f.db.Model(&systemModel.SystemRolePermission{}).Where("role_id = ?", temp.ID).Count(&count).Error)
	require.Zero(t, count)
	require.NoError(t, f.db.Model(&gormadapter.CasbinRule{}).Where("v0 = ?", strconv.Itoa(int(temp.ID))).Count(&count).Error)
	require.Zero(t, count)
	require.False(t, enforce(t, temp.ID, "/system/users", "GET"), "the enforcer was reloaded")
	var audit systemModel.SystemAuditLog
	require.NoError(t, f.db.Last(&audit).Error)
	require.Equal(t, "role.delete", audit.Action)
	require.Equal(t, "success", audit.Result)
}

func TestApprovableFlagIsSetOnRolesButNeverOnAdmin(t *testing.T) {
	f := newAuthStateFixture(t)
	svc := &RoleService{}
	yes, no := true, false

	updated, err := svc.UpdateRole(context.Background(), 1, f.opsRole.ID, systemRequest.RoleRequest{Name: "ops", Approvable: &yes}, AuditMeta{})
	require.NoError(t, err)
	require.True(t, updated.Approvable)
	// Omitting the flag keeps it.
	updated, err = svc.UpdateRole(context.Background(), 1, f.opsRole.ID, systemRequest.RoleRequest{Name: "Ops"}, AuditMeta{})
	require.NoError(t, err)
	require.True(t, updated.Approvable)
	require.Equal(t, "Ops", updated.Name)
	updated, err = svc.UpdateRole(context.Background(), 1, f.opsRole.ID, systemRequest.RoleRequest{Name: "Ops", Approvable: &no}, AuditMeta{})
	require.NoError(t, err)
	require.False(t, updated.Approvable)

	_, err = svc.UpdateRole(context.Background(), 1, f.adminRole.ID, systemRequest.RoleRequest{Name: "admin", Approvable: &yes}, AuditMeta{})
	require.ErrorIs(t, err, ErrProtectedAdminRole)

	created, err := svc.CreateRole(context.Background(), 1, systemRequest.RoleRequest{Code: "audit-approver", Name: "Audit approver", Approvable: &yes}, AuditMeta{})
	require.NoError(t, err)
	require.True(t, created.Approvable)
}

func TestListRolesWithPermissionsIncludesCodes(t *testing.T) {
	f := newAuthStateFixture(t)
	seedUsersViewPermission(t, f.db)
	require.NoError(t, (&RoleService{}).UpdateRolePermissions(context.Background(), 1, f.opsRole.ID, []string{"system.users.menu", "system.users.view"}, AuditMeta{}))

	items, err := (&RoleService{}).ListRolesWithPermissions(context.Background())
	require.NoError(t, err)
	for _, item := range items {
		if item.ID == f.opsRole.ID {
			require.Equal(t, []string{"system.users.menu", "system.users.view"}, item.PermissionCodes)
			return
		}
	}
	t.Fatal("ops role missing")
}

func setRoleLockHooks(t *testing.T, afterAssign func(uint), afterDelete func(uint)) {
	t.Helper()
	roleLockTestHooks.afterAssignLock = afterAssign
	roleLockTestHooks.afterDeleteLock = afterDelete
	t.Cleanup(func() {
		roleLockTestHooks.afterAssignLock = nil
		roleLockTestHooks.afterDeleteLock = nil
	})
}

// The assignment commits first: the delete waits for it, then sees the holder.
func TestDeleteRoleWaitsForConcurrentAssignment(t *testing.T) {
	f := newAuthStateFixture(t)
	temp := systemModel.SystemRole{Code: "temp", Name: "temp"}
	require.NoError(t, f.db.Create(&temp).Error)
	locked, release := make(chan struct{}), make(chan struct{})
	setRoleLockHooks(t, func(userID uint) {
		if userID == f.user.ID {
			close(locked)
			<-release
		}
	}, nil)

	assigned := make(chan error)
	go func() {
		assigned <- (&UserService{}).UpdateUserRoles(context.Background(), 1, f.user.ID, []uint{f.opsRole.ID, temp.ID}, AuditMeta{})
	}()
	<-locked
	roleLockTestHooks.afterAssignLock = nil
	deleted := make(chan error)
	go func() { deleted <- (&RoleService{}).DeleteRole(context.Background(), 1, temp.ID, AuditMeta{}) }()
	select {
	case err := <-deleted:
		t.Fatalf("delete returned while the assignment held its lock: %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	close(release)
	require.NoError(t, <-assigned)
	require.ErrorIs(t, <-deleted, ErrRoleInUse)

	roles, err := queryUserRoles(context.Background(), f.user.ID)
	require.NoError(t, err)
	require.Len(t, roles, 2, "the new assignment survives")
}

// The delete commits first: the assignment waits, then finds the role gone
// and is rejected, leaving the user's existing roles untouched.
func TestAssignmentWaitsForConcurrentDeleteAndIsRejected(t *testing.T) {
	f := newAuthStateFixture(t)
	temp := systemModel.SystemRole{Code: "temp", Name: "temp"}
	require.NoError(t, f.db.Create(&temp).Error)
	locked, release := make(chan struct{}), make(chan struct{})
	setRoleLockHooks(t, nil, func(roleID uint) {
		if roleID == temp.ID {
			close(locked)
			<-release
		}
	})

	deleted := make(chan error)
	go func() { deleted <- (&RoleService{}).DeleteRole(context.Background(), 1, temp.ID, AuditMeta{}) }()
	<-locked
	roleLockTestHooks.afterDeleteLock = nil
	assigned := make(chan error)
	go func() {
		assigned <- (&UserService{}).UpdateUserRoles(context.Background(), 1, f.user.ID, []uint{f.opsRole.ID, temp.ID}, AuditMeta{})
	}()
	select {
	case err := <-assigned:
		t.Fatalf("assignment returned while the delete held its lock: %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	close(release)
	require.NoError(t, <-deleted)
	require.ErrorIs(t, <-assigned, ErrRoleNotFound)

	var dangling int64
	require.NoError(t, f.db.Model(&systemModel.SystemUserRole{}).Where("role_id = ?", temp.ID).Count(&dangling).Error)
	require.Zero(t, dangling)
	roles, err := queryUserRoles(context.Background(), f.user.ID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	require.Equal(t, f.opsRole.ID, roles[0].ID)
}

// The user role is found by code when new LDAP employees are created; its code
// (and admin's) cannot change.
func TestBuiltinRoleCodesCannotChange(t *testing.T) {
	f := newAuthStateFixture(t)
	userRole := systemModel.SystemRole{Code: DefaultUserRoleCode, Name: "user"}
	require.NoError(t, f.db.Create(&userRole).Error)
	svc := &RoleService{}

	_, err := svc.UpdateRole(context.Background(), 1, userRole.ID, systemRequest.RoleRequest{Code: "staff", Name: "Staff"}, AuditMeta{})
	require.ErrorIs(t, err, ErrProtectedRoleCode)
	_, err = svc.UpdateRole(context.Background(), 1, f.adminRole.ID, systemRequest.RoleRequest{Code: "root", Name: "Root"}, AuditMeta{})
	require.ErrorIs(t, err, ErrProtectedRoleCode)
	renamed, err := svc.UpdateRole(context.Background(), 1, userRole.ID, systemRequest.RoleRequest{Code: DefaultUserRoleCode, Name: "Staff"}, AuditMeta{})
	require.NoError(t, err, "the display name can still change")
	require.Equal(t, DefaultUserRoleCode, renamed.Code)
	require.ErrorIs(t, svc.DeleteRole(context.Background(), 1, userRole.ID, AuditMeta{}), ErrProtectedRole)

	created, err := (&AuthService{}).syncLDAPUser(context.Background(), LDAPUser{Username: "newcomer", DisplayName: "Newcomer"})
	require.NoError(t, err)
	roles, err := queryUserRoles(context.Background(), created.ID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	require.Equal(t, DefaultUserRoleCode, roles[0].Code)
}

// The delete commits first: the permission update waits, then finds the role
// gone and writes no grants or policies for it.
func TestPermissionUpdateWaitsForConcurrentDeleteAndIsRejected(t *testing.T) {
	f := newAuthStateFixture(t)
	seedUsersViewPermission(t, f.db)
	temp := systemModel.SystemRole{Code: "temp", Name: "temp"}
	require.NoError(t, f.db.Create(&temp).Error)
	locked, release := make(chan struct{}), make(chan struct{})
	setRoleLockHooks(t, nil, func(roleID uint) {
		if roleID == temp.ID {
			close(locked)
			<-release
		}
	})

	deleted := make(chan error)
	go func() { deleted <- (&RoleService{}).DeleteRole(context.Background(), 1, temp.ID, AuditMeta{}) }()
	<-locked
	roleLockTestHooks.afterDeleteLock = nil
	updated := make(chan error)
	go func() {
		updated <- (&RoleService{}).UpdateRolePermissions(context.Background(), 1, temp.ID, []string{"system.users.view", "system.users.menu"}, AuditMeta{})
	}()
	select {
	case err := <-updated:
		t.Fatalf("permission update returned while the delete held its lock: %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	close(release)
	require.NoError(t, <-deleted)
	require.ErrorIs(t, <-updated, ErrRoleNotFound)

	var grants, policies int64
	require.NoError(t, f.db.Model(&systemModel.SystemRolePermission{}).Where("role_id = ?", temp.ID).Count(&grants).Error)
	require.NoError(t, f.db.Model(&gormadapter.CasbinRule{}).Where("v0 = ?", strconv.Itoa(int(temp.ID))).Count(&policies).Error)
	require.Zero(t, grants+policies)
}

// The permission update commits first: the delete waits, then removes the
// role together with the grants just written.
func TestDeleteRoleWaitsForConcurrentPermissionUpdate(t *testing.T) {
	f := newAuthStateFixture(t)
	seedUsersViewPermission(t, f.db)
	temp := systemModel.SystemRole{Code: "temp", Name: "temp"}
	require.NoError(t, f.db.Create(&temp).Error)
	locked, release := make(chan struct{}), make(chan struct{})
	roleLockTestHooks.afterPermissionLock = func(roleID uint) {
		if roleID == temp.ID {
			close(locked)
			<-release
		}
	}
	t.Cleanup(func() { roleLockTestHooks.afterPermissionLock = nil })

	updated := make(chan error)
	go func() {
		updated <- (&RoleService{}).UpdateRolePermissions(context.Background(), 1, temp.ID, []string{"system.users.view", "system.users.menu"}, AuditMeta{})
	}()
	<-locked
	roleLockTestHooks.afterPermissionLock = nil
	deleted := make(chan error)
	go func() { deleted <- (&RoleService{}).DeleteRole(context.Background(), 1, temp.ID, AuditMeta{}) }()
	select {
	case err := <-deleted:
		t.Fatalf("delete returned while the permission update held its lock: %v", err)
	case <-time.After(300 * time.Millisecond):
	}
	close(release)
	require.NoError(t, <-updated)
	require.NoError(t, <-deleted)

	var grants, policies int64
	require.NoError(t, f.db.Model(&systemModel.SystemRolePermission{}).Where("role_id = ?", temp.ID).Count(&grants).Error)
	require.NoError(t, f.db.Model(&gormadapter.CasbinRule{}).Where("v0 = ?", strconv.Itoa(int(temp.ID))).Count(&policies).Error)
	require.Zero(t, grants+policies)
	require.False(t, enforce(t, temp.ID, "/system/users", "GET"))
}
