package system

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"
	"optimus/server/utils"

	"github.com/stretchr/testify/require"
)

func TestAuthServiceLoginAdminSuccess(t *testing.T) {
	passwordHash, err := hashTestPassword("admin-secret")
	require.NoError(t, err)
	db := openTestDB(t)
	user := systemModel.SystemUser{
		OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 1},
		UUID:          "admin-uuid",
		Username:      "admin",
		DisplayName:   "Administrator",
		Source:        "local",
		Status:        "enabled",
		PasswordHash:  passwordHash,
	}
	require.NoError(t, db.Create(&user).Error)
	service := newAuthServiceForTest(t)

	result, err := service.Login(context.Background(), systemRequest.LoginRequest{
		Username: "admin",
		Password: "admin-secret",
	}, AuditMeta{IP: "127.0.0.1", UserAgent: "test"})

	require.NoError(t, err)
	require.NotEmpty(t, result.Token)
	require.Equal(t, "admin", result.User.Username)
}

// --- fakes ---

func hashTestPassword(password string) (string, error) {
	return utils.BcryptHash(password)
}

func TestMenuPermissionCodesForRolesOrdersDistinctCodes(t *testing.T) {
	db := openTestDB(t)

	adminRole := systemModel.SystemRole{Code: "admin", Name: "Administrator"}
	opsRole := systemModel.SystemRole{Code: "ops", Name: "Ops"}
	require.NoError(t, db.Create(&[]systemModel.SystemRole{adminRole, opsRole}).Error)
	require.NoError(t, db.Where("code = ?", "admin").First(&adminRole).Error)
	require.NoError(t, db.Where("code = ?", "ops").First(&opsRole).Error)

	first := systemModel.SystemPermission{Code: "system.users.menu", Name: "Users", Type: "menu", Sort: 20}
	second := systemModel.SystemPermission{Code: "dashboard.view", Name: "Dashboard", Type: "menu", Sort: 10}
	apiPerm := systemModel.SystemPermission{Code: "system.users.view", Name: "Users API", Type: "api", Sort: 5}
	require.NoError(t, db.Create(&[]systemModel.SystemPermission{first, second, apiPerm}).Error)

	var permissions []systemModel.SystemPermission
	require.NoError(t, db.Find(&permissions).Error)
	for _, p := range permissions {
		require.NoError(t, db.Create(&systemModel.SystemRolePermission{RoleID: adminRole.ID, PermissionID: p.ID}).Error)
		require.NoError(t, db.Create(&systemModel.SystemRolePermission{RoleID: opsRole.ID, PermissionID: p.ID}).Error)
	}

	codes, err := (&AuthService{}).menuPermissionCodesForRoles(context.Background(), []uint{adminRole.ID, opsRole.ID})

	require.NoError(t, err)
	require.Equal(t, []string{"dashboard.view", "system.users.menu"}, codes)
}

func TestAuthServiceDatabaseSyncLDAPUserCreatesDefaultRole(t *testing.T) {
	db := openTestDB(t)

	userRole := systemModel.SystemRole{Code: "user", Name: "User"}
	require.NoError(t, db.Create(&userRole).Error)

	store := (&AuthService{})
	user, err := store.syncLDAPUser(context.Background(), LDAPUser{
		Username:    "alice",
		DisplayName: "Alice",
		Email:       "alice@example.com",
	})

	require.NoError(t, err)
	require.Equal(t, "alice", user.Username)
	require.Equal(t, "ldap", user.Source)

	var userRoles []systemModel.SystemUserRole
	require.NoError(t, db.Where("user_id = ?", user.ID).Find(&userRoles).Error)
	require.Len(t, userRoles, 1)
	require.Equal(t, userRole.ID, userRoles[0].RoleID)
}

func TestFailedLoginAuditUsesUnverifiedAttempt(t *testing.T) {
	db := openTestDB(t)
	svc := (&AuthService{})
	_, err := svc.Login(context.Background(), systemRequest.LoginRequest{Username: "unknown\naccount", Password: "PASSWORD_SECRET"}, AuditMeta{})
	require.ErrorIs(t, err, ErrInvalidCredentials)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	require.Zero(t, audit.ActorUserID)
	require.Empty(t, audit.ActorUsername)
	require.Empty(t, audit.ResourceID)
	require.NotContains(t, audit.Detail, "PASSWORD_SECRET")
	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
	require.Equal(t, float64(1), detail["version"])
	require.Equal(t, "unknown account", detail["attemptedUsername"])
}

func TestDisabledLoginAuditDoesNotClaimActor(t *testing.T) {
	db := openTestDB(t)
	require.NoError(t, db.Create(&systemModel.SystemUser{OPTIMUS_MODEL: global.OPTIMUS_MODEL{ID: 7}, UUID: "test-user", Username: "admin", Source: "local", Status: "disabled"}).Error)
	_, err := (&AuthService{}).Login(context.Background(), systemRequest.LoginRequest{Username: "admin", Password: "PASSWORD_SECRET"}, AuditMeta{})
	require.ErrorIs(t, err, ErrUserDisabled)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	require.Zero(t, audit.ActorUserID)
	require.Empty(t, audit.ActorUsername)
	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
	require.Equal(t, float64(7), detail["targetUserId"])
}

func TestFailedLoginAuditBoundsAttempt(t *testing.T) {
	db := openTestDB(t)
	_, err := (&AuthService{}).Login(context.Background(), systemRequest.LoginRequest{Username: strings.Repeat("界", 300)}, AuditMeta{})
	require.ErrorIs(t, err, ErrInvalidCredentials)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Last(&audit).Error)
	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
	require.LessOrEqual(t, len([]rune(detail["attemptedUsername"].(string))), 128)
}
