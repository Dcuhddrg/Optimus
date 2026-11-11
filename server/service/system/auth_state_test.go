package system

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	systemModel "optimus/server/model/system"
	"optimus/server/utils"

	"github.com/casbin/casbin/v3"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type authStateFixture struct {
	db        *gorm.DB
	adminRole systemModel.SystemRole
	opsRole   systemModel.SystemRole
	user      systemModel.SystemUser
}

func newAuthStateFixture(t *testing.T) authStateFixture {
	t.Helper()
	db := openTestDB(t)
	f := authStateFixture{
		db:        db,
		adminRole: systemModel.SystemRole{Code: AdminRoleCode, Name: "admin"},
		opsRole:   systemModel.SystemRole{Code: "ops", Name: "ops"},
		user:      systemModel.SystemUser{UUID: "u-alice", Username: "alice", Source: "ldap", Status: DefaultUserStatus},
	}
	require.NoError(t, db.Create(&f.adminRole).Error)
	require.NoError(t, db.Create(&f.opsRole).Error)
	require.NoError(t, db.Create(&f.user).Error)
	require.NoError(t, db.Create(&systemModel.SystemUserRole{UserID: f.user.ID, RoleID: f.opsRole.ID}).Error)
	return f
}

func setAuthStateHooks(t *testing.T, afterLoad, insideStore func(uint)) {
	t.Helper()
	authStateTestHooks.afterLoad = afterLoad
	authStateTestHooks.insideStore = insideStore
	t.Cleanup(func() {
		authStateTestHooks.afterLoad = nil
		authStateTestHooks.insideStore = nil
	})
}

func loadState(t *testing.T, userID uint) UserAuthState {
	t.Helper()
	state, err := (&AuthStateService{}).LoadUserAuthState(context.Background(), userID)
	require.NoError(t, err)
	return state
}

func TestAuthStateChangesApplyOnNextLoadAfterAPIReturns(t *testing.T) {
	f := newAuthStateFixture(t)
	require.Equal(t, []uint{f.opsRole.ID}, loadState(t, f.user.ID).RoleIDs) // now cached

	require.NoError(t, (&UserService{}).UpdateUserRoles(context.Background(), 1, f.user.ID, []uint{}, AuditMeta{}))
	require.Empty(t, loadState(t, f.user.ID).RoleIDs)

	_, err := (&UserService{}).UpdateUserStatus(context.Background(), 1, f.user.ID, "disabled", AuditMeta{})
	require.NoError(t, err)
	require.False(t, loadState(t, f.user.ID).Enabled())
}

// A request that read the old roles and pauses while the admin revokes them
// must not write its stale result into the cache.
func TestAuthStateStaleLoadIsNotWrittenBack(t *testing.T) {
	f := newAuthStateFixture(t)
	revoked := false
	setAuthStateHooks(t, func(userID uint) {
		if userID != f.user.ID || revoked {
			return
		}
		revoked = true
		require.NoError(t, (&UserService{}).UpdateUserRoles(context.Background(), 1, f.user.ID, []uint{}, AuditMeta{}))
	}, nil)

	// The in-flight request may still use what it read: it began before the change.
	require.Equal(t, []uint{f.opsRole.ID}, loadState(t, f.user.ID).RoleIDs)
	// The next request does not see the stale value.
	require.Empty(t, loadState(t, f.user.ID).RoleIDs)
}

// A loader that passed the version check and pauses before assigning holds
// the cache lock, so the revocation's invalidation waits for it and then
// removes the value; after the revoke API returns the cache has no old roles.
func TestAuthStateInvalidationWaitsForInFlightStore(t *testing.T) {
	f := newAuthStateFixture(t)
	entered := make(chan struct{})
	release := make(chan struct{})
	setAuthStateHooks(t, nil, func(userID uint) {
		if userID != f.user.ID {
			return
		}
		close(entered)
		<-release
	})

	loadDone := make(chan UserAuthState)
	go func() {
		state, _ := (&AuthStateService{}).LoadUserAuthState(context.Background(), f.user.ID)
		loadDone <- state
	}()
	<-entered
	authStateTestHooks.insideStore = nil

	revokeDone := make(chan error)
	go func() {
		revokeDone <- (&UserService{}).UpdateUserRoles(context.Background(), 1, f.user.ID, []uint{}, AuditMeta{})
	}()
	select {
	case <-revokeDone:
		t.Fatal("revocation returned while a store held the cache lock")
	case <-time.After(200 * time.Millisecond):
	}
	close(release)
	require.Equal(t, []uint{f.opsRole.ID}, (<-loadDone).RoleIDs)
	require.NoError(t, <-revokeDone)

	require.Empty(t, loadState(t, f.user.ID).RoleIDs)
}

func enforce(t *testing.T, roleID uint, path, method string) bool {
	t.Helper()
	e, ready := (&CasbinService{}).Enforcer()
	require.True(t, ready)
	ok, err := e.Enforce(strconv.Itoa(int(roleID)), path, method)
	require.NoError(t, err)
	return ok
}

func seedUsersViewPermission(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, permission := range DefaultSystemPermissions() {
		require.NoError(t, db.Create(&permission).Error)
	}
}

func TestRolePermissionChangeAppliesToNextDecision(t *testing.T) {
	f := newAuthStateFixture(t)
	seedUsersViewPermission(t, f.db)
	require.False(t, enforce(t, f.opsRole.ID, "/system/users", "GET"))

	require.NoError(t, (&RoleService{}).UpdateRolePermissions(context.Background(), 1, f.opsRole.ID, []string{"system.users.view"}, AuditMeta{}))
	require.True(t, enforce(t, f.opsRole.ID, "/system/users", "GET"))

	require.NoError(t, (&RoleService{}).UpdateRolePermissions(context.Background(), 1, f.opsRole.ID, []string{}, AuditMeta{}))
	require.False(t, enforce(t, f.opsRole.ID, "/system/users", "GET"))
}

func TestAdminRoleMustKeepEverySystemPermission(t *testing.T) {
	f := newAuthStateFixture(t)
	seedUsersViewPermission(t, f.db)
	all := SystemPermissionCodes()

	err := (&RoleService{}).UpdateRolePermissions(context.Background(), 1, f.adminRole.ID, all[1:], AuditMeta{})
	require.ErrorIs(t, err, ErrProtectedAdminRole)

	require.NoError(t, (&RoleService{}).UpdateRolePermissions(context.Background(), 1, f.adminRole.ID, append(all, "dashboard.view"), AuditMeta{}))
}

// After a failed reload the database holds the revocation while memory holds
// the grant; authorization fails closed until the background retry succeeds.
func TestCasbinReloadFailureFailsClosedThenRecovers(t *testing.T) {
	f := newAuthStateFixture(t)
	seedUsersViewPermission(t, f.db)
	require.NoError(t, (&RoleService{}).UpdateRolePermissions(context.Background(), 1, f.opsRole.ID, []string{"system.users.view"}, AuditMeta{}))
	require.True(t, enforce(t, f.opsRole.ID, "/system/users", "GET"))

	savedBuild, savedDelay := utils.BuildCasbinEnforcer, casbinRetryInitial
	t.Cleanup(func() { utils.BuildCasbinEnforcer = savedBuild; casbinRetryInitial = savedDelay })
	casbinRetryInitial = 20 * time.Millisecond
	failing := make(chan struct{})
	utils.BuildCasbinEnforcer = func() (*casbin.SyncedEnforcer, error) {
		select {
		case <-failing:
			return savedBuild()
		default:
			return nil, errors.New("db unavailable")
		}
	}

	err := (&RoleService{}).UpdateRolePermissions(context.Background(), 1, f.opsRole.ID, []string{}, AuditMeta{})
	require.ErrorIs(t, err, ErrCasbinReloadPending)
	_, ready := (&CasbinService{}).Enforcer()
	require.False(t, ready, "the old policy must not authorize")

	close(failing)
	require.Eventually(t, func() bool {
		_, ready := (&CasbinService{}).Enforcer()
		return ready
	}, 5*time.Second, 10*time.Millisecond)
	require.False(t, enforce(t, f.opsRole.ID, "/system/users", "GET"))
}

// A reload that started first but finishes last must not publish an older
// policy: reloads are serialized and read the database under the lock.
func TestConcurrentCasbinReloadsEndWithLatestPolicy(t *testing.T) {
	f := newAuthStateFixture(t)
	seedUsersViewPermission(t, f.db)
	require.NoError(t, utils.ReloadCasbin())

	savedBuild := utils.BuildCasbinEnforcer
	t.Cleanup(func() { utils.BuildCasbinEnforcer = savedBuild })
	firstRead := make(chan struct{})
	release := make(chan struct{})
	calls := 0
	utils.BuildCasbinEnforcer = func() (*casbin.SyncedEnforcer, error) {
		calls++
		e, err := savedBuild()
		if calls == 1 {
			close(firstRead)
			<-release
		}
		return e, err
	}

	first := make(chan error)
	go func() { first <- utils.ReloadCasbin() }()
	<-firstRead // the first reload read the policy without the grant

	require.NoError(t, CasbinServiceApp.AddPolicies(f.db, f.opsRole.ID, []CasbinInfo{{Path: "/system/users", Method: "GET"}}))
	second := make(chan error)
	go func() { second <- utils.ReloadCasbin() }()

	close(release)
	require.NoError(t, <-first)
	require.NoError(t, <-second)
	require.True(t, enforce(t, f.opsRole.ID, "/system/users", "GET"))
}

// Catalog entries and their first grant commit together, so a failure in
// between leaves nothing behind and the retry grants everything.
func TestSeedPermissionsGrantingNewToAdminRetriesAfterFailure(t *testing.T) {
	f := newAuthStateFixture(t)
	catalog := []systemModel.SystemPermission{
		{Code: "demo.menu", Name: "demo", Type: "menu", Path: "/demo"},
		{Code: "demo.view", Name: "view demo", Type: "api", Method: "GET", Path: "/demo"},
	}
	fail := true
	require.NoError(t, f.db.Callback().Create().Before("gorm:create").Register("test:fail_grant", func(tx *gorm.DB) {
		if fail && tx.Statement.Table == "casbin_rule" {
			tx.AddError(errors.New("grant failed"))
		}
	}))
	t.Cleanup(func() { _ = f.db.Callback().Create().Remove("test:fail_grant") })

	require.Error(t, SeedPermissionsGrantingNewToAdmin(f.db, catalog))
	var count int64
	require.NoError(t, f.db.Model(&systemModel.SystemPermission{}).Where("code LIKE ?", "demo.%").Count(&count).Error)
	require.Zero(t, count, "catalog entries roll back with the failed grant")

	fail = false
	require.NoError(t, SeedPermissionsGrantingNewToAdmin(f.db, catalog))
	var menu systemModel.SystemPermission
	require.NoError(t, f.db.Where("code = ?", "demo.menu").First(&menu).Error)
	require.NoError(t, f.db.Model(&systemModel.SystemRolePermission{}).Where("role_id = ? AND permission_id = ?", f.adminRole.ID, menu.ID).Count(&count).Error)
	require.Equal(t, int64(1), count)
	require.True(t, enforce(t, f.adminRole.ID, "/demo", "GET"))
}
