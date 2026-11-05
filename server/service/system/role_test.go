package system

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

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

	role := systemModel.SystemRole{Code: "editor", Name: "Editor", Status: defaultRoleStatus}
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
	require.NoError(t, db.Create(&systemModel.SystemRole{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 3}, Code: "admin", Name: "admin", Status: "enabled"}).Error)
	err := (&RoleService{}).UpdateRolePermissions(context.Background(), 1, 3, []string{"read"}, AuditMeta{})
	require.ErrorIs(t, err, ErrProtectedAdminRole)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	require.Equal(t, "failure", audit.Result)
	require.Contains(t, audit.Detail, "admin_role_protected")
}

func TestRolePermissionAuditUsesPersistedCodes(t *testing.T) {
	db := openTestDB(t)
	role := systemModel.SystemRole{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 3}, Code: "ops", Name: "ops", Status: "enabled"}
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
