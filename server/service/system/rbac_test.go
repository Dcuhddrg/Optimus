package system_test

import (
	"context"
	"testing"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"

	"github.com/stretchr/testify/require"
)

func TestRBACServiceAllowsRoleWithPermission(t *testing.T) {
	db := openExternalPersistenceDB(t)
	permission := systemModel.SystemPermission{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 10}, Code: "system.users.view", Name: "View users", Type: "api", Method: "GET", Path: "/api/v1/system/users"}
	require.NoError(t, db.Create(&permission).Error)
	require.NoError(t, db.Create(&systemModel.SystemRolePermission{RoleID: 1, PermissionID: 10}).Error)
	service := &systemService.RBACService{}

	allowed, err := service.AuthorizeAPI(context.Background(), "GET", "/api/v1/system/users", []uint{1})

	require.NoError(t, err)
	require.True(t, allowed)
}

type fakeRBACStore struct {
	permissions map[string]systemService.SystemPermissionSummary
	allowed     map[uint]bool
}

func (s *fakeRBACStore) FindAPIPermission(_ context.Context, method, path string) (systemService.SystemPermissionSummary, bool, error) {
	p, ok := s.permissions[method+" "+path]
	return p, ok, nil
}

func (s *fakeRBACStore) RoleHasPermission(_ context.Context, roleID, _ uint) (bool, error) {
	return s.allowed[roleID], nil
}
