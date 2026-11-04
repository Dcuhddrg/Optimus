package system

import (
	"context"
	"encoding/json"
	"testing"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"

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

	p1 := systemModel.SystemPermission{Code: "content.create", Name: "Create Content", Type: "api"}
	p2 := systemModel.SystemPermission{Code: "content.delete", Name: "Delete Content", Type: "api"}
	require.NoError(t, db.Create(&[]systemModel.SystemPermission{p1, p2}).Error)
	require.NoError(t, db.Where("code = ?", "content.create").First(&p1).Error)
	require.NoError(t, db.Where("code = ?", "content.delete").First(&p2).Error)

	store := (&RoleService{})

	require.NoError(t, store.replaceRolePermissions(context.Background(), role.ID, []uint{p1.ID}))
	var rps []systemModel.SystemRolePermission
	require.NoError(t, db.Where("role_id = ?", role.ID).Find(&rps).Error)
	require.Len(t, rps, 1)

	require.NoError(t, store.replaceRolePermissions(context.Background(), role.ID, []uint{p1.ID, p2.ID}))
	require.NoError(t, db.Where("role_id = ?", role.ID).Find(&rps).Error)
	require.Len(t, rps, 2)
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
	read := systemModel.SystemPermission{Code: "read", Name: "read", Type: "api"}
	write := systemModel.SystemPermission{Code: "write", Name: "write", Type: "api"}
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
