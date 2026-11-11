package system

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"optimus/server/config"
	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"
	ldapUtil "optimus/server/utils/ldap"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// fakeDirectory is a scriptable LDAP client keyed by username.
type fakeDirectory struct {
	mu        sync.Mutex
	entries   map[string]LDAPUser
	authErr   error
	testErr   error
	onLookup  func(username string)
	lookupErr error
	lookups   int
}

func newFakeDirectory(entries ...LDAPUser) *fakeDirectory {
	d := &fakeDirectory{entries: map[string]LDAPUser{}}
	for _, e := range entries {
		d.entries[e.Username] = e
	}
	return d
}

func (d *fakeDirectory) set(e LDAPUser) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.entries[e.Username] = e
}

func (d *fakeDirectory) get(username string) (LDAPUser, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	e, ok := d.entries[username]
	if !ok {
		return LDAPUser{}, ldapUtil.ErrUserNotFound
	}
	return e, nil
}

func (d *fakeDirectory) Test(context.Context, LDAPConnectionConfig) error { return d.testErr }

func (d *fakeDirectory) Authenticate(_ context.Context, _ LDAPConnectionConfig, username, _ string) (LDAPUser, error) {
	if d.authErr != nil {
		return LDAPUser{}, d.authErr
	}
	return d.get(username)
}

func (d *fakeDirectory) LookupUsername(_ context.Context, _ LDAPConnectionConfig, username string) (LDAPUser, error) {
	d.mu.Lock()
	d.lookups++
	hook := d.onLookup
	d.mu.Unlock()
	if hook != nil {
		hook(username)
	}
	if d.lookupErr != nil {
		return LDAPUser{}, d.lookupErr
	}
	return d.get(username)
}

var testLDAPConn = LDAPConnectionConfig{UserFilter: "(uid={username})", UsernameAttr: "uid", IDAttr: "entryUUID", DepartmentAttr: "departmentNumber"}

func dirEntry(username, id string) LDAPUser {
	return LDAPUser{DN: "uid=" + username + ",ou=people,dc=optimus,dc=test", Username: username, DisplayName: strings.ToUpper(username), Email: username + "@optimus.test", Department: "ops", ExternalID: id, Version: "v1"}
}

// syncAs runs the in-lock write-back for a login that authenticated as the
// directory's current entry for username.
func syncAs(t *testing.T, dir *fakeDirectory, username string) (systemModel.SystemUser, *ldapLoginFailure) {
	t.Helper()
	verified, err := dir.get(username)
	require.NoError(t, err)
	return (&AuthService{}).syncLDAPUser(context.Background(), dir, testLDAPConn, username, verified)
}

func seedDefaultRole(t *testing.T, db *gorm.DB) systemModel.SystemRole {
	t.Helper()
	role := systemModel.SystemRole{Code: DefaultUserRoleCode, Name: "user"}
	require.NoError(t, db.Create(&role).Error)
	return role
}

func bindExternalID(t *testing.T, db *gorm.DB, userID uint, id string) {
	t.Helper()
	require.NoError(t, db.Model(&systemModel.SystemUser{}).Where("id = ?", userID).Update("external_id", id).Error)
}

func TestLDAPLoginCreatesAccountWithRandomUUIDAndIdentity(t *testing.T) {
	db := openTestDB(t)
	role := seedDefaultRole(t, db)
	dir := newFakeDirectory(dirEntry("dave", "id-dave"))

	user, failure := syncAs(t, dir, "dave")
	require.Nil(t, failure)
	require.NotContains(t, user.UUID, "ldap:")
	require.Len(t, user.UUID, 36)
	require.NotEqual(t, "id-dave", user.UUID)
	require.Equal(t, "id-dave", *user.ExternalID)
	require.Equal(t, "ops", user.Department)
	roles, err := queryUserRoles(context.Background(), user.ID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	require.Equal(t, role.ID, roles[0].ID)
}

// Same identity, new username: a rename. Roles stay; the profile follows the
// directory, and an empty attribute clears the field.
func TestLDAPLoginRenamesAccountWithSameIdentity(t *testing.T) {
	f := newAuthStateFixture(t)
	bindExternalID(t, f.db, f.user.ID, "id-alice")
	renamed := dirEntry("alice2", "id-alice")
	renamed.Email = ""
	dir := newFakeDirectory(renamed)

	user, failure := syncAs(t, dir, "alice2")
	require.Nil(t, failure)
	require.Equal(t, f.user.ID, user.ID)
	require.Equal(t, "alice2", user.Username)
	require.Empty(t, user.Email)
	roles, err := queryUserRoles(context.Background(), user.ID)
	require.NoError(t, err)
	require.Len(t, roles, 1)
	require.Equal(t, f.opsRole.ID, roles[0].ID)
}

func TestLDAPLoginRenameOntoTakenUsernameIsConflict(t *testing.T) {
	f := newAuthStateFixture(t)
	bindExternalID(t, f.db, f.user.ID, "id-alice")
	other := systemModel.SystemUser{UUID: "u-bob", Username: "bob", Source: "local", Status: DefaultUserStatus}
	require.NoError(t, f.db.Create(&other).Error)
	dir := newFakeDirectory(dirEntry("bob", "id-alice"))

	_, failure := syncAs(t, dir, "bob")
	require.NotNil(t, failure)
	require.ErrorIs(t, failure.err, ErrLDAPAccountConflict)
	var stored systemModel.SystemUser
	require.NoError(t, f.db.First(&stored, f.user.ID).Error)
	require.Equal(t, "alice", stored.Username)
}

// Same username, different identity: another person (uid reuse).
func TestLDAPLoginSameUsernameDifferentIdentityIsConflict(t *testing.T) {
	f := newAuthStateFixture(t)
	bindExternalID(t, f.db, f.user.ID, "id-alice")
	dir := newFakeDirectory(dirEntry("alice", "id-new-alice"))

	_, failure := syncAs(t, dir, "alice")
	require.NotNil(t, failure)
	require.ErrorIs(t, failure.err, ErrLDAPAccountConflict)
	require.Equal(t, "ldap_identity_conflict", failure.reason)
	require.Equal(t, f.user.ID, failure.userID)
	var count int64
	require.NoError(t, f.db.Model(&systemModel.SystemUser{}).Count(&count).Error)
	require.Equal(t, int64(1), count, "no account is created for the newcomer")
}

// A legacy LDAP account without ExternalID is never bound by login.
func TestLDAPLoginLegacyAccountAwaitsBinding(t *testing.T) {
	f := newAuthStateFixture(t)
	dir := newFakeDirectory(dirEntry("alice", "id-alice"))

	_, failure := syncAs(t, dir, "alice")
	require.NotNil(t, failure)
	require.ErrorIs(t, failure.err, ErrLDAPBindingPending)
	var stored systemModel.SystemUser
	require.NoError(t, f.db.First(&stored, f.user.ID).Error)
	require.Nil(t, stored.ExternalID)
}

// The username moved to another identity between the password check and the
// in-lock re-read: refuse instead of continuing as the new holder.
func TestLDAPLoginRefusesWhenRereadIdentityDiffers(t *testing.T) {
	db := openTestDB(t)
	seedDefaultRole(t, db)
	dir := newFakeDirectory(dirEntry("erin", "id-erin"))
	verified, err := dir.get("erin")
	require.NoError(t, err)
	dir.set(dirEntry("erin", "id-someone-else"))

	_, failure := (&AuthService{}).syncLDAPUser(context.Background(), dir, testLDAPConn, "erin", verified)
	require.NotNil(t, failure)
	require.ErrorIs(t, failure.err, ErrLDAPDirectoryChanging)
	var count int64
	require.NoError(t, db.Model(&systemModel.SystemUser{}).Count(&count).Error)
	require.Zero(t, count)
}

func TestLDAPLoginKeepsDepartmentWhenNotSynced(t *testing.T) {
	f := newAuthStateFixture(t)
	bindExternalID(t, f.db, f.user.ID, "id-alice")
	require.NoError(t, f.db.Model(&systemModel.SystemUser{}).Where("id = ?", f.user.ID).Update("department", "kept").Error)
	dir := newFakeDirectory(dirEntry("alice", "id-alice"))
	conn := testLDAPConn
	conn.DepartmentAttr = ""

	verified, err := dir.get("alice")
	require.NoError(t, err)
	user, failure := (&AuthService{}).syncLDAPUser(context.Background(), dir, conn, "alice", verified)
	require.Nil(t, failure)
	require.Equal(t, "kept", user.Department)
}

// Login write-backs run one at a time: the second waits for the first to
// finish its re-read and commit.
func TestLDAPLoginWriteBacksAreSerialized(t *testing.T) {
	db := openTestDB(t)
	seedDefaultRole(t, db)
	dir := newFakeDirectory(dirEntry("fay", "id-fay"), dirEntry("gus", "id-gus"))
	inFirst, release := make(chan struct{}), make(chan struct{})
	secondLookup := make(chan struct{}, 1)
	dir.onLookup = func(username string) {
		if username == "fay" {
			close(inFirst)
			<-release
			return
		}
		secondLookup <- struct{}{}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); _, f := syncAs(t, dir, "fay"); require.Nil(t, f) }()
	<-inFirst
	go func() { defer wg.Done(); _, f := syncAs(t, dir, "gus"); require.Nil(t, f) }()
	select {
	case <-secondLookup:
		t.Fatal("second write-back read the directory while the first held the lock")
	case <-time.After(200 * time.Millisecond):
	}
	close(release)
	wg.Wait()
	<-secondLookup
}

func enableLDAP(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Create(&systemModel.SystemLDAPConfig{Enabled: true, URL: "ldap://test", UserFilter: "(uid={username})", UsernameAttr: "uid", IDAttr: "entryUUID"}).Error)
}

func lastLoginReason(t *testing.T, db *gorm.DB) string {
	t.Helper()
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Where("action = ?", "auth.login.failure").Last(&audit).Error)
	var detail map[string]any
	require.NoError(t, json.Unmarshal([]byte(audit.Detail), &detail))
	return detail["reason"].(string)
}

func TestLDAPLoginMapsDirectoryRefusals(t *testing.T) {
	db := openTestDB(t)
	enableLDAP(t, db)
	for _, tc := range []struct {
		err    error
		want   error
		reason string
	}{
		{ldapUtil.ErrDuplicateUsername, ErrLDAPDuplicateUsername, "ldap_duplicate_username"},
		{ldapUtil.ErrMissingIdentity, ErrLDAPIdentityMissing, "ldap_identity_missing"},
		{ldapUtil.ErrDuplicateIdentity, ErrLDAPIdentityAnomaly, "ldap_identity_duplicate"},
		{ldapUtil.ErrMissingVersion, ErrLDAPIdentityAnomaly, "ldap_version_missing"},
		{ldapUtil.ErrEntryChanged, ErrLDAPDirectoryChanging, "ldap_entry_changed"},
		{ldapUtil.ErrInvalidCredentials, ErrInvalidCredentials, "ldap_authentication_failed"},
	} {
		dir := newFakeDirectory()
		dir.authErr = tc.err
		newLDAPServiceForTest(t, dir)
		_, err := (&AuthService{}).Login(context.Background(), systemRequest.LoginRequest{Username: "zoe", Password: "pw"}, AuditMeta{})
		require.ErrorIs(t, err, tc.want, tc.reason)
		require.Equal(t, tc.reason, lastLoginReason(t, db))
	}
}

func TestLDAPLoginEndToEndIssuesSession(t *testing.T) {
	db := openTestDB(t)
	enableLDAP(t, db)
	seedDefaultRole(t, db)
	newLDAPServiceForTest(t, newFakeDirectory(dirEntry("hal", "id-hal")))
	previous := global.OPTIMUS_CONFIG.JWT
	t.Cleanup(func() { global.OPTIMUS_CONFIG.JWT = previous })
	global.OPTIMUS_CONFIG.JWT = config.JWT{SigningKey: "ldap-login-test-key", ExpiresTime: "1h", Issuer: "ldap-login-test"}
	result, err := (&AuthService{}).Login(context.Background(), systemRequest.LoginRequest{Username: "hal", Password: "pw"}, AuditMeta{})
	require.NoError(t, err)
	require.Equal(t, "hal", result.User.Username)
	require.True(t, result.User.Bound)
}

func TestUpdateUserRejectsLDAPProfileEdits(t *testing.T) {
	f := newAuthStateFixture(t)
	_, err := (&UserService{}).UpdateUser(context.Background(), 1, f.user.ID, systemRequest.UpdateUserRequest{DisplayName: "X"}, AuditMeta{})
	require.ErrorIs(t, err, ErrLDAPProfileReadOnly)
}

func validLDAPRequest() systemRequest.LDAPConfigRequest {
	return systemRequest.LDAPConfigRequest{Enabled: true, URL: "ldap://test", BindPassword: "pw", UserFilter: "(uid={username})", UsernameAttr: "uid"}
}

func TestSaveLDAPConfigValidatesFilterAndIdentityAttr(t *testing.T) {
	openTestDB(t)
	svc := newLDAPServiceForTest(t, newFakeDirectory())

	bad := validLDAPRequest()
	bad.UserFilter = "(|(uid={username})(mail={username}))"
	_, err := svc.SaveConfig(context.Background(), 1, bad, AuditMeta{})
	require.ErrorIs(t, err, ldapUtil.ErrInvalidLoginFilter)

	bad = validLDAPRequest()
	bad.IDAttr = "uid"
	_, err = svc.SaveConfig(context.Background(), 1, bad, AuditMeta{})
	require.ErrorIs(t, err, ErrInvalidIdentityAttr)

	saved, err := svc.SaveConfig(context.Background(), 1, validLDAPRequest(), AuditMeta{})
	require.NoError(t, err)
	require.Equal(t, "entryUUID", saved.IDAttr)
	require.Equal(t, "departmentNumber", saved.DepartmentAttr)

	// nil keeps the saved department attribute; "" stops syncing it.
	req := validLDAPRequest()
	empty := ""
	req.DepartmentAttr = &empty
	saved, err = svc.SaveConfig(context.Background(), 1, req, AuditMeta{})
	require.NoError(t, err)
	require.Empty(t, saved.DepartmentAttr)
	saved, err = svc.SaveConfig(context.Background(), 1, validLDAPRequest(), AuditMeta{})
	require.NoError(t, err)
	require.Empty(t, saved.DepartmentAttr)
}

func TestSaveLDAPConfigCannotChangeIdentityAttrWhileBound(t *testing.T) {
	f := newAuthStateFixture(t)
	svc := newLDAPServiceForTest(t, newFakeDirectory())
	_, err := svc.SaveConfig(context.Background(), 1, validLDAPRequest(), AuditMeta{})
	require.NoError(t, err)
	bindExternalID(t, f.db, f.user.ID, "id-alice")

	req := validLDAPRequest()
	req.IDAttr = "nsUniqueId"
	_, err = svc.SaveConfig(context.Background(), 1, req, AuditMeta{})
	var locked IdentityAttrLockedError
	require.ErrorAs(t, err, &locked)
	require.Equal(t, int64(1), locked.Bound)

	req.IDAttr = "entryUUID"
	saved, err := svc.SaveConfig(context.Background(), 1, req, AuditMeta{})
	require.NoError(t, err, "keeping the same attribute is allowed")
	require.Equal(t, int64(1), saved.BoundUsers)
}

func TestLDAPConnectionTestReportsEmptyScope(t *testing.T) {
	db := openTestDB(t)
	dir := newFakeDirectory()
	dir.testErr = ldapUtil.ErrNoUsersInScope
	svc := newLDAPServiceForTest(t, dir)
	_, err := svc.TestConfig(context.Background(), 1, systemRequest.LDAPTestRequest{URL: "ldap://test", BindPassword: "pw", UserFilter: "(uid={username})", UsernameAttr: "uid"}, AuditMeta{})
	require.ErrorIs(t, err, ldapUtil.ErrNoUsersInScope)
	var audit systemModel.SystemAuditLog
	require.NoError(t, db.Where("action = ?", "ldap.test.failure").Last(&audit).Error)
	require.Contains(t, audit.Detail, "no_users_in_scope")
}
