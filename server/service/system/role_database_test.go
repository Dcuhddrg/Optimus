package system

import (
	"context"
	"testing"

	systemModel "optimus/server/model/system"

	"github.com/stretchr/testify/require"
)

func TestRoleServiceDatabaseReplaceRolePermissions(t *testing.T) {
	db := openPersistenceTestDB(t)

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
