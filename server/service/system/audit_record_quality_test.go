package system_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"optimus/server/global"
	model "optimus/server/model/system"
	service "optimus/server/service/system"
)

var auditTestFailure = errors.New("database failed, password=DO_NOT_LOG_THIS")

type qualityUserStore struct {
	fakeUserStore
	auditErr, updateErr, snapshotErr error
	roleIDs                          []uint
	persistedIDs                     []uint
}

func (s *qualityUserStore) WriteAudit(_ context.Context, e service.AuditEvent) error {
	s.audits = append(s.audits, e)
	return s.auditErr
}
func (s *qualityUserStore) UpdateUser(ctx context.Context, u model.SystemUser) (model.SystemUser, error) {
	if s.updateErr != nil {
		return model.SystemUser{}, s.updateErr
	}
	for i, old := range s.users {
		if old.ID == u.ID {
			s.users[i] = u
		}
	}
	return s.fakeUserStore.UpdateUser(ctx, u)
}
func (s *qualityUserStore) RolesForUser(_ context.Context, _ uint) ([]model.SystemRoleSummary, error) {
	if s.snapshotErr != nil {
		return nil, s.snapshotErr
	}
	r := []model.SystemRoleSummary{}
	for _, id := range s.roleIDs {
		r = append(r, model.SystemRoleSummary{ID: id})
	}
	return r, nil
}
func (s *qualityUserStore) ReplaceUserRoles(_ context.Context, _ uint, ids []uint) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.roleIDs = append([]uint{}, ids...)
	if s.persistedIDs != nil {
		s.roleIDs = s.persistedIDs
	}
	return nil
}

type qualityRoleStore struct {
	fakeRoleStore
	auditErr, updateErr, snapshotErr error
	codes                            []string
}

// Preserve existing fixture compatibility with the new read-only snapshot contract.
func (s *fakeRoleStore) PermissionCodesForRole(_ context.Context, _ uint) ([]string, error) {
	return nil, nil
}
func (s *qualityRoleStore) PermissionCodesForRole(_ context.Context, _ uint) ([]string, error) {
	return append([]string{}, s.codes...), s.snapshotErr
}
func (s *qualityRoleStore) WriteAudit(_ context.Context, e service.AuditEvent) error {
	s.audits = append(s.audits, e)
	return s.auditErr
}
func (s *qualityRoleStore) ReplaceRolePermissions(_ context.Context, _ uint, ids []uint) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.savedPermIDs = append([]uint{}, ids...)
	s.codes = []string{}
	for _, id := range ids {
		for code, p := range s.permissionsByCode {
			if p.ID == id {
				s.codes = append(s.codes, code)
			}
		}
	}
	return nil
}

type qualityLDAPStore struct {
	fakeLDAPStore
	auditErr, saveErr, readErr error
}

func (s *qualityLDAPStore) WriteAudit(_ context.Context, e service.AuditEvent) error {
	s.audits = append(s.audits, e)
	return s.auditErr
}
func (s *qualityLDAPStore) SaveLDAPConfig(ctx context.Context, c model.SystemLDAPConfig) (model.SystemLDAPConfig, error) {
	if s.saveErr != nil {
		return model.SystemLDAPConfig{}, s.saveErr
	}
	return s.fakeLDAPStore.SaveLDAPConfig(ctx, c)
}
func (s *qualityLDAPStore) GetLDAPConfig(ctx context.Context) (model.SystemLDAPConfig, error) {
	if s.readErr != nil {
		return model.SystemLDAPConfig{}, s.readErr
	}
	return s.fakeLDAPStore.GetLDAPConfig(ctx)
}

type qualityLDAPClient struct{ err error }

func (c qualityLDAPClient) Test(context.Context, service.LDAPConnectionConfig) error { return c.err }

type qualityAuthStore struct {
	fakeAuthStore
	auditErr error
}

func (s *qualityAuthStore) WriteAudit(_ context.Context, e service.AuditEvent) error {
	s.audits = append(s.audits, e)
	return s.auditErr
}

func qualityUser() *qualityUserStore {
	return &qualityUserStore{fakeUserStore: fakeUserStore{users: []model.SystemUser{{ID: 7, Username: "target", Status: "enabled", PasswordHash: "DO_NOT_LOG_THIS"}}, adminCapableAfterStatus: true, adminCapableAfterRoles: true}}
}
func auditDetail(t *testing.T, e service.AuditEvent) map[string]interface{} {
	t.Helper()
	var d map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(e.Detail), &d))
	require.Equal(t, float64(1), d["version"])
	return d
}
func lastAudit(t *testing.T, events []service.AuditEvent) service.AuditEvent {
	t.Helper()
	require.NotEmpty(t, events)
	return events[len(events)-1]
}
func observeQuality(t *testing.T) *observer.ObservedLogs {
	t.Helper()
	core, logs := observer.New(zap.ErrorLevel)
	old := global.OPTIMUS_LOG
	global.OPTIMUS_LOG = zap.New(core)
	t.Cleanup(func() { global.OPTIMUS_LOG = old })
	return logs
}

func TestAuditQualityStatusSnapshotAndActor(t *testing.T) {
	s := qualityUser()
	result, err := service.NewUserService(s).UpdateUserStatus(context.Background(), 1, 7, "disabled", service.AuditMeta{ActorUsername: "operator-snapshot"})
	require.NoError(t, err)
	require.Equal(t, "disabled", result.Status)
	e := lastAudit(t, s.audits)
	require.Equal(t, "operator-snapshot", e.ActorUsername)
	d := auditDetail(t, e)
	require.Equal(t, map[string]interface{}{"before": "enabled", "after": "disabled"}, d["changes"].(map[string]interface{})["status"])
	require.NotContains(t, e.Detail, "DO_NOT_LOG_THIS")
}
func TestAuditQualityWriteFailurePreservesSuccessAndIsObservable(t *testing.T) {
	logs := observeQuality(t)
	s := qualityUser()
	s.auditErr = auditTestFailure
	_, err := service.NewUserService(s).UpdateUserStatus(context.Background(), 1, 7, "disabled", service.AuditMeta{})
	require.NoError(t, err)
	require.Equal(t, "disabled", s.users[0].Status)
	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	require.Equal(t, "audit write failed", entry.Message)
	require.Equal(t, "user.status.update", entry.ContextMap()["action"])
	require.Equal(t, "7", entry.ContextMap()["resource_id"])
	require.NotContains(t, entry.Message+entry.ContextMap()["error_type"].(string), "DO_NOT_LOG_THIS")
	require.NotContains(t, entry.ContextMap(), "detail")
	require.NotContains(t, entry.ContextMap(), "error")
}
func TestAuditQualityProtectedStatusFailureIsRecorded(t *testing.T) {
	s := qualityUser()
	s.adminCapableAfterStatus = false
	_, err := service.NewUserService(s).UpdateUserStatus(context.Background(), 1, 7, "disabled", service.AuditMeta{})
	require.ErrorIs(t, err, service.ErrLastAdminCapableUser)
	e := lastAudit(t, s.audits)
	require.Equal(t, "failure", e.Result)
	require.Equal(t, "last_admin_protected", auditDetail(t, e)["reason"])
	require.Equal(t, "enabled", s.users[0].Status)
}
func TestAuditQualityOriginalFailureSurvivesAuditFailure(t *testing.T) {
	observeQuality(t)
	s := qualityUser()
	s.auditErr = auditTestFailure
	s.updateErr = errors.New("original update failure")
	_, err := service.NewUserService(s).UpdateUserStatus(context.Background(), 1, 7, "disabled", service.AuditMeta{})
	require.ErrorIs(t, err, s.updateErr)
	e := lastAudit(t, s.audits)
	require.Equal(t, "failure", e.Result)
	require.Equal(t, "update_failed", auditDetail(t, e)["reason"])
}
func TestAuditQualityRolesUsePersistedSnapshot(t *testing.T) {
	s := qualityUser()
	s.roleIDs = []uint{2, 1}
	s.persistedIDs = []uint{2, 3}
	require.NoError(t, service.NewUserService(s).UpdateUserRoles(context.Background(), 1, 7, []uint{2, 3, 999}, service.AuditMeta{}))
	change := auditDetail(t, lastAudit(t, s.audits))["changes"].(map[string]interface{})["roleIds"].(map[string]interface{})
	require.Equal(t, []interface{}{float64(1), float64(2)}, change["before"])
	require.Equal(t, []interface{}{float64(2), float64(3)}, change["after"])
	require.Equal(t, []interface{}{float64(3)}, change["added"])
	require.Equal(t, []interface{}{float64(1)}, change["removed"])
}
func TestAuditQualityRoleSnapshotUnavailableDoesNotChangeSuccess(t *testing.T) {
	s := qualityUser()
	s.snapshotErr = auditTestFailure
	require.NoError(t, service.NewUserService(s).UpdateUserRoles(context.Background(), 1, 7, []uint{2}, service.AuditMeta{}))
	e := lastAudit(t, s.audits)
	require.Equal(t, "success", e.Result)
	d := auditDetail(t, e)
	require.Equal(t, "snapshot_unavailable", d["reason"])
	require.Empty(t, d["changes"])
}
func qualityRole() *qualityRoleStore {
	return &qualityRoleStore{fakeRoleStore: fakeRoleStore{roles: []model.SystemRole{{ID: 3, Code: "ops"}}, permissionsByCode: map[string]model.SystemPermission{"read": {ID: 1, Code: "read"}, "write": {ID: 2, Code: "write"}}}, codes: []string{"read"}}
}
func TestAuditQualityPermissionsUseActualCodes(t *testing.T) {
	s := qualityRole()
	require.NoError(t, service.NewRoleService(s).UpdateRolePermissions(context.Background(), 1, 3, []string{"write", "not-a-real-code"}, service.AuditMeta{}))
	d := auditDetail(t, lastAudit(t, s.audits))
	change := d["changes"].(map[string]interface{})["permissionCodes"].(map[string]interface{})
	require.Equal(t, []interface{}{"write"}, change["after"])
	require.Equal(t, []interface{}{"write"}, change["added"])
	require.Equal(t, []interface{}{"read"}, change["removed"])
}
func TestAuditQualityProtectedRoleFailure(t *testing.T) {
	s := qualityRole()
	s.roles[0].Code = "admin"
	err := service.NewRoleService(s).UpdateRolePermissions(context.Background(), 1, 3, []string{"read"}, service.AuditMeta{})
	require.ErrorIs(t, err, service.ErrProtectedAdminRole)
	e := lastAudit(t, s.audits)
	require.Equal(t, "failure", e.Result)
	require.Equal(t, "admin_role_protected", auditDetail(t, e)["reason"])
}
func TestAuditQualityLDAPWhitelistAndPasswordOmission(t *testing.T) {
	s := &qualityLDAPStore{fakeLDAPStore: fakeLDAPStore{config: model.SystemLDAPConfig{ID: 1, Enabled: false, URL: "ldap://old.test", BindPassword: "OLD_SECRET"}}}
	_, err := service.NewLDAPService(s, qualityLDAPClient{}).SaveConfig(context.Background(), 1, model.LDAPConfigRequest{Enabled: true, URL: "ldap://user:URL_SECRET@new.test", BindPassword: "NEW_SECRET"}, service.AuditMeta{})
	require.NoError(t, err)
	e := lastAudit(t, s.audits)
	d := auditDetail(t, e)
	require.Equal(t, "success", e.Result)
	require.Contains(t, d["changes"], "enabled")
	require.Contains(t, d["changes"], "bindPasswordChanged")
	for _, secret := range []string{"OLD_SECRET", "NEW_SECRET", "URL_SECRET", "ldap://"} {
		require.NotContains(t, e.Detail, secret)
	}
}
func TestAuditQualityLDAPTestDoesNotPersistRawError(t *testing.T) {
	s := &qualityLDAPStore{}
	_, err := service.NewLDAPService(s, qualityLDAPClient{err: auditTestFailure}).TestConfig(context.Background(), 1, model.LDAPTestRequest{URL: "ldap://test", BindPassword: "PASSWORD_SECRET", UserFilter: "(uid={username})"}, service.AuditMeta{})
	require.Error(t, err)
	e := lastAudit(t, s.audits)
	require.Equal(t, "failure", e.Result)
	require.Equal(t, "connection_test_failed", auditDetail(t, e)["reason"])
	require.NotContains(t, e.Detail, "DO_NOT_LOG_THIS")
	require.NotContains(t, e.Detail, "PASSWORD_SECRET")
}
func TestAuditQualityFailedLoginHasUnverifiedAttemptNotActor(t *testing.T) {
	s := &qualityAuthStore{}
	_, err := service.NewAuthService(s, fakeTokenIssuer{}, nil).Login(context.Background(), model.LoginRequest{Username: "unknown\naccount", Password: "PASSWORD_SECRET"}, service.AuditMeta{})
	require.ErrorIs(t, err, service.ErrInvalidCredentials)
	e := lastAudit(t, s.audits)
	require.Zero(t, e.ActorUserID)
	require.Empty(t, e.ActorUsername)
	d := auditDetail(t, e)
	require.Equal(t, "unknown account", d["attemptedUsername"])
	require.Empty(t, e.ResourceID)
	require.NotContains(t, e.Detail, "PASSWORD_SECRET")
}
func TestAuditQualityFailedKnownLoginDoesNotClaimAuthenticatedActor(t *testing.T) {
	s := &qualityAuthStore{fakeAuthStore: fakeAuthStore{user: model.SystemUser{ID: 7, Username: "admin", Status: "disabled"}}}
	_, err := service.NewAuthService(s, fakeTokenIssuer{}, nil).Login(context.Background(), model.LoginRequest{Username: "admin", Password: "PASSWORD_SECRET"}, service.AuditMeta{})
	require.ErrorIs(t, err, service.ErrUserDisabled)
	e := lastAudit(t, s.audits)
	require.Zero(t, e.ActorUserID)
	require.Empty(t, e.ActorUsername)
	require.Equal(t, float64(7), auditDetail(t, e)["targetUserId"])
}
func TestAuditQualityLoginAttemptIsBounded(t *testing.T) {
	s := &qualityAuthStore{}
	_, _ = service.NewAuthService(s, fakeTokenIssuer{}, nil).Login(context.Background(), model.LoginRequest{Username: strings.Repeat("界", 300)}, service.AuditMeta{})
	d := auditDetail(t, lastAudit(t, s.audits))
	require.LessOrEqual(t, len([]rune(d["attemptedUsername"].(string))), 128)
}

type snapshotFailUserStore struct {
	qualityUserStore
	failAt, reads int
}

func (s *snapshotFailUserStore) RolesForUser(ctx context.Context, id uint) ([]model.SystemRoleSummary, error) {
	s.reads++
	if s.reads == s.failAt {
		return nil, auditTestFailure
	}
	return s.qualityUserStore.RolesForUser(ctx, id)
}

type snapshotFailRoleStore struct {
	qualityRoleStore
	failAt, reads int
}

func (s *snapshotFailRoleStore) PermissionCodesForRole(ctx context.Context, id uint) ([]string, error) {
	s.reads++
	if s.reads == s.failAt {
		return nil, auditTestFailure
	}
	return s.qualityRoleStore.PermissionCodesForRole(ctx, id)
}
func TestAuditQualityBeforeAndAfterSnapshotFailures(t *testing.T) {
	for _, stage := range []int{1, 2} {
		t.Run(string(rune('0'+stage)), func(t *testing.T) {
			user := &snapshotFailUserStore{qualityUserStore: *qualityUser(), failAt: stage}
			require.NoError(t, service.NewUserService(user).UpdateUserRoles(context.Background(), 1, 7, []uint{2}, service.AuditMeta{}))
			ud := auditDetail(t, lastAudit(t, user.audits))
			require.Equal(t, "snapshot_unavailable", ud["reason"])
			require.Empty(t, ud["changes"])
			role := &snapshotFailRoleStore{qualityRoleStore: *qualityRole(), failAt: stage}
			require.NoError(t, service.NewRoleService(role).UpdateRolePermissions(context.Background(), 1, 3, []string{"write"}, service.AuditMeta{}))
			rd := auditDetail(t, lastAudit(t, role.audits))
			require.Equal(t, "snapshot_unavailable", rd["reason"])
			require.Empty(t, rd["changes"])
		})
	}
}
func TestAuditQualityPermissionMutationFailurePreservesOriginalError(t *testing.T) {
	observeQuality(t)
	s := qualityRole()
	s.updateErr = errors.New("original role failure")
	s.auditErr = auditTestFailure
	err := service.NewRoleService(s).UpdateRolePermissions(context.Background(), 1, 3, []string{"write"}, service.AuditMeta{})
	require.ErrorIs(t, err, s.updateErr)
	require.Equal(t, "failure", lastAudit(t, s.audits).Result)
	require.Equal(t, []string{"read"}, s.codes)
}
func TestAuditQualityLDAPSnapshotFailureKeepsSavedResult(t *testing.T) {
	s := &qualityLDAPStore{readErr: auditTestFailure}
	_, err := service.NewLDAPService(s, qualityLDAPClient{}).SaveConfig(context.Background(), 1, model.LDAPConfigRequest{Enabled: true, BindPassword: "SECRET"}, service.AuditMeta{})
	require.NoError(t, err)
	d := auditDetail(t, lastAudit(t, s.audits))
	require.Equal(t, "snapshot_unavailable", d["reason"])
	require.Empty(t, d["changes"])
	require.NotContains(t, lastAudit(t, s.audits).Detail, "SECRET")
}
func TestAuditQualityLDAPSaveFailureAndAuditFailure(t *testing.T) {
	logs := observeQuality(t)
	s := &qualityLDAPStore{saveErr: errors.New("original save error"), auditErr: auditTestFailure}
	_, err := service.NewLDAPService(s, qualityLDAPClient{}).SaveConfig(context.Background(), 1, model.LDAPConfigRequest{BindPassword: "SECRET"}, service.AuditMeta{})
	require.ErrorIs(t, err, s.saveErr)
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "failure", lastAudit(t, s.audits).Result)
}
func TestAuditQualitySuccessfulLoginSurvivesAuditWriteFailure(t *testing.T) {
	logs := observeQuality(t)
	hash, err := hashTestPassword("LOGIN_SECRET")
	require.NoError(t, err)
	s := &qualityAuthStore{fakeAuthStore: *newFakeAuthStore(model.SystemUser{ID: 1, Username: "admin", Source: "local", Status: "enabled", PasswordHash: hash}), auditErr: auditTestFailure}
	result, err := service.NewAuthService(s, fakeTokenIssuer{}, nil).Login(context.Background(), model.LoginRequest{Username: "admin", Password: "LOGIN_SECRET"}, service.AuditMeta{})
	require.NoError(t, err)
	require.NotEmpty(t, result.Token)
	require.Equal(t, 1, logs.Len())
	require.Equal(t, "admin", lastAudit(t, s.audits).ActorUsername)
}
