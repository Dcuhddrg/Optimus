package system_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"optimus/server/global"
	model "optimus/server/model/system"
	service "optimus/server/service/system"
)

func TestRolePermissionSnapshotFailuresPreserveMutation(t *testing.T) {
	for _, failAt := range []int{1, 2} {
		t.Run(string(rune('0'+failAt)), func(t *testing.T) {
			db := openExternalPersistenceDB(t)
			role := model.SystemRole{Code: "ops", Name: "ops", Status: "enabled"}
			permission := model.SystemPermission{Code: "write", Name: "write", Type: "api"}
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
			require.NoError(t, (&service.RoleService{}).UpdateRolePermissions(context.Background(), 1, role.ID, []string{"write"}, service.AuditMeta{}))
			require.Equal(t, 2, reads)
			var audit model.SystemAuditLog
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
			db := openExternalPersistenceDB(t)
			adminRole := model.SystemRole{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 1}, Code: "admin", Name: "admin", Status: "enabled"}
			role := model.SystemRole{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 2}, Code: "ops", Name: "ops", Status: "enabled"}
			admin := model.SystemUser{UUID: "admin-id", Username: "admin", Source: "local", Status: "enabled"}
			user := model.SystemUser{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 7}, UUID: "target-id", Username: "target", Source: "local", Status: "enabled"}
			require.NoError(t, db.Create(&adminRole).Error)
			require.NoError(t, db.Create(&role).Error)
			require.NoError(t, db.Create(&admin).Error)
			require.NoError(t, db.Create(&user).Error)
			require.NoError(t, db.Create(&model.SystemUserRole{UserID: admin.ID, RoleID: adminRole.ID}).Error)
			reads := 0
			require.NoError(t, db.Callback().Row().Before("gorm:row").Register("test:fail_user_snapshot", func(tx *gorm.DB) {
				if tx.Statement.Table == "roles" {
					reads++
					if reads == failAt {
						tx.AddError(errors.New("snapshot failure"))
					}
				}
			}))
			require.NoError(t, (&service.UserService{}).UpdateUserRoles(context.Background(), 1, user.ID, []uint{2}, service.AuditMeta{}))
			require.Equal(t, 2, reads)
			var audit model.SystemAuditLog
			require.NoError(t, db.Last(&audit).Error)
			require.Equal(t, "success", audit.Result)
			var detail map[string]any
			require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
			require.Equal(t, "snapshot_unavailable", detail["reason"])
			require.Empty(t, detail["changes"])
		})
	}
}
