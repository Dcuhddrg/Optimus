package system

import (
	"context"
	"errors"
	"testing"
	"time"

	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"
	systemResponse "optimus/server/model/system/response"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type syncFixture struct {
	db  *gorm.DB
	dir *fakeDirectory
	svc *LDAPService
	cfg systemModel.SystemLDAPConfig
}

const syncAdmin = uint(1)

func newSyncFixture(t *testing.T, entries ...LDAPUser) syncFixture {
	t.Helper()
	db := openTestDB(t)
	seedDefaultRole(t, db)
	dir := newFakeDirectory(entries...)
	f := syncFixture{db: db, dir: dir, svc: newLDAPServiceForTest(t, dir), cfg: enableLDAP(t, db)}
	return f
}

// markSynced records a completed sync under the current config, so absence
// can disable automatically.
func (f syncFixture) markSynced(t *testing.T) {
	t.Helper()
	require.NoError(t, f.db.Model(&systemModel.SystemLDAPConfig{}).Where("id = ?", f.cfg.ID).Update("last_sync_key", syncConfigKey(f.cfg)).Error)
}

func (f syncFixture) account(t *testing.T, username string, externalID *string) systemModel.SystemUser {
	t.Helper()
	user := systemModel.SystemUser{UUID: "u-" + username, Username: username, DisplayName: username, Source: "ldap", Status: DefaultUserStatus, ExternalID: externalID}
	require.NoError(t, f.db.Create(&user).Error)
	return user
}

func (f syncFixture) reload(t *testing.T, id uint) systemModel.SystemUser {
	t.Helper()
	var user systemModel.SystemUser
	require.NoError(t, f.db.First(&user, id).Error)
	return user
}

func (f syncFixture) preview(t *testing.T) systemResponse.LDAPSyncPreview {
	t.Helper()
	preview, err := f.svc.PreviewSync(context.Background(), syncAdmin, AuditMeta{})
	require.NoError(t, err)
	return preview
}

func (f syncFixture) execute(t *testing.T, req systemRequest.LDAPSyncExecuteRequest) systemResponse.LDAPSyncResult {
	t.Helper()
	result, err := f.svc.ExecuteSync(context.Background(), syncAdmin, req, AuditMeta{})
	require.NoError(t, err)
	return result
}

func previewCategory(preview systemResponse.LDAPSyncPreview, username string) string {
	for _, item := range preview.Items {
		if (item.Account != nil && item.Account.Username == username) || (item.Account == nil && item.Entry != nil && item.Entry.Username == username) {
			return item.Category
		}
	}
	return ""
}

func outcomeOf(result systemResponse.LDAPSyncResult, username string) systemResponse.LDAPSyncOutcome {
	for _, o := range result.Outcomes {
		if o.Username == username {
			return o
		}
	}
	return systemResponse.LDAPSyncOutcome{}
}

// First sync after configuration: everything that is not in the directory
// needs the admin; after it completes, the next sync disables automatically.
func TestLDAPSyncFirstRunNeedsAdminThenAutoDisables(t *testing.T) {
	f := newSyncFixture(t, dirEntry("alice", "id-alice"), dirEntry("robert", "id-bob"), dirEntry("hank", "id-hank"))
	alice := f.account(t, "alice", strp("id-alice"))
	bob := f.account(t, "bob", strp("id-bob"))
	carol := f.account(t, "carol", strp("id-carol"))

	preview := f.preview(t)
	require.True(t, preview.ConfigChanged)
	require.Equal(t, "update", previewCategory(preview, "bob"))
	require.Equal(t, "create", previewCategory(preview, "hank"))
	require.Equal(t, "unconfirmable", previewCategory(preview, "carol"))

	result := f.execute(t, systemRequest.LDAPSyncExecuteRequest{PreviewID: preview.PreviewID, Disable: []uint{carol.ID}})
	require.Equal(t, syncStatusCompleted, result.Status)
	require.Equal(t, "robert", f.reload(t, bob.ID).Username, "renamed, same account")
	require.Equal(t, "disabled", f.reload(t, carol.ID).Status)
	var hank systemModel.SystemUser
	require.NoError(t, f.db.Where("username = ?", "hank").First(&hank).Error)
	require.Equal(t, "id-hank", *hank.ExternalID)
	roles, err := queryUserRoles(context.Background(), hank.ID)
	require.NoError(t, err)
	require.Len(t, roles, 1)

	// Next sync: alice left the directory and is disabled automatically.
	f.dir.remove("alice")
	preview = f.preview(t)
	require.False(t, preview.ConfigChanged)
	require.Equal(t, "autoDisable", previewCategory(preview, "alice"))
	result = f.execute(t, systemRequest.LDAPSyncExecuteRequest{PreviewID: preview.PreviewID})
	require.Equal(t, syncOutcomeDone, outcomeOf(result, "alice").Outcome)
	require.Equal(t, "disabled", f.reload(t, alice.ID).Status)
	var audit systemModel.SystemAuditLog
	require.NoError(t, f.db.Where("action = ? AND resource_id = ?", "user.status.update", resourceID(alice.ID)).First(&audit).Error)
	require.Contains(t, audit.Detail, "ldap_sync_auto_disabled")
}

func TestLDAPSyncPreviewIsOneUseOwnAndFresh(t *testing.T) {
	f := newSyncFixture(t, dirEntry("hank", "id-hank"))
	ctx := context.Background()

	preview := f.preview(t)
	_, err := f.svc.ExecuteSync(ctx, syncAdmin+1, systemRequest.LDAPSyncExecuteRequest{PreviewID: preview.PreviewID}, AuditMeta{})
	require.ErrorIs(t, err, ErrSyncPreviewInvalid, "another administrator")

	preview = f.preview(t)
	f.execute(t, systemRequest.LDAPSyncExecuteRequest{PreviewID: preview.PreviewID})
	_, err = f.svc.ExecuteSync(ctx, syncAdmin, systemRequest.LDAPSyncExecuteRequest{PreviewID: preview.PreviewID}, AuditMeta{})
	require.ErrorIs(t, err, ErrSyncPreviewInvalid, "executed twice")

	preview = f.preview(t)
	saved := syncNow
	syncNow = func() time.Time { return saved().Add(syncSnapshotTTL + time.Minute) }
	_, err = f.svc.ExecuteSync(ctx, syncAdmin, systemRequest.LDAPSyncExecuteRequest{PreviewID: preview.PreviewID}, AuditMeta{})
	syncNow = saved
	require.ErrorIs(t, err, ErrSyncPreviewInvalid, "expired")

	preview = f.preview(t)
	_, err = f.svc.SaveConfig(ctx, syncAdmin, validLDAPRequest(), AuditMeta{})
	require.NoError(t, err)
	_, err = f.svc.ExecuteSync(ctx, syncAdmin, systemRequest.LDAPSyncExecuteRequest{PreviewID: preview.PreviewID}, AuditMeta{})
	require.ErrorIs(t, err, ErrSyncPreviewInvalid, "config changed after preview")
}

// Binding a legacy account keeps its roles; "another person" retires the old
// account under <uid>~<id> and gives the newcomer a fresh account.
func TestLDAPSyncBindAndSeparate(t *testing.T) {
	f := newSyncFixture(t, dirEntry("erin", "id-erin"), dirEntry("gina", "id-new-gina"))
	erin := f.account(t, "erin", nil)
	gina := f.account(t, "gina", strp("id-old-gina"))
	ops := systemModel.SystemRole{Code: "ops", Name: "ops"}
	require.NoError(t, f.db.Create(&ops).Error)
	require.NoError(t, f.db.Create(&systemModel.SystemUserRole{UserID: erin.ID, RoleID: ops.ID}).Error)

	preview := f.preview(t)
	require.Equal(t, "pendingBinding", previewCategory(preview, "erin"))
	require.Equal(t, "conflict", previewCategory(preview, "gina"))
	result := f.execute(t, systemRequest.LDAPSyncExecuteRequest{PreviewID: preview.PreviewID, Bind: []uint{erin.ID}, Separate: []uint{gina.ID}})
	require.Equal(t, syncStatusCompleted, result.Status)

	bound := f.reload(t, erin.ID)
	require.Equal(t, "id-erin", *bound.ExternalID)
	roles, err := queryUserRoles(context.Background(), erin.ID)
	require.NoError(t, err)
	require.Equal(t, "ops", roles[0].Code)

	retired := f.reload(t, gina.ID)
	require.Equal(t, "gina~"+resourceID(gina.ID), retired.Username)
	require.Equal(t, "disabled", retired.Status)
	require.Equal(t, "u-gina", retired.UUID)
	var newcomer systemModel.SystemUser
	require.NoError(t, f.db.Where("username = ?", "gina").First(&newcomer).Error)
	require.Equal(t, "id-new-gina", *newcomer.ExternalID)
	require.NotEqual(t, gina.ID, newcomer.ID)
	for _, action := range []string{"user.ldap.bind", "user.ldap.separate"} {
		var n int64
		require.NoError(t, f.db.Model(&systemModel.SystemAuditLog{}).Where("action = ?", action).Count(&n).Error)
		require.Equal(t, int64(1), n, action)
	}
}

// The admin checked identity A; before execution the uid moved to identity
// B. The row is still "pending binding", but it is no longer the person the
// admin looked at.
func TestLDAPSyncSkipsChoiceWhoseIdentityChanged(t *testing.T) {
	f := newSyncFixture(t, dirEntry("erin", "id-erin"))
	erin := f.account(t, "erin", nil)
	preview := f.preview(t)
	f.dir.set(dirEntry("erin", "id-someone-else"))

	result := f.execute(t, systemRequest.LDAPSyncExecuteRequest{PreviewID: preview.PreviewID, Bind: []uint{erin.ID}})
	require.Equal(t, "identity_changed", outcomeOf(result, "erin").Reason)
	require.Nil(t, f.reload(t, erin.ID).ExternalID)
}

func TestLDAPSyncUnidentifiableEntriesBlockDisables(t *testing.T) {
	f := newSyncFixture(t, dirEntry("hank", "id-hank"), dirEntry("noid", ""))
	f.markSynced(t)
	carol := f.account(t, "carol", strp("id-carol"))
	legacy := f.account(t, "fred", nil)

	preview := f.preview(t)
	require.True(t, preview.DisablesBlocked)
	result := f.execute(t, systemRequest.LDAPSyncExecuteRequest{PreviewID: preview.PreviewID, Disable: []uint{legacy.ID}})
	require.Equal(t, "disables_blocked", outcomeOf(result, "carol").Reason)
	require.Equal(t, "disables_blocked", outcomeOf(result, "fred").Reason)
	require.Equal(t, DefaultUserStatus, f.reload(t, carol.ID).Status)
	require.Equal(t, DefaultUserStatus, f.reload(t, legacy.ID).Status)
	require.Equal(t, syncOutcomeDone, outcomeOf(result, "hank").Outcome, "creates still run")
}

// A re-read error stops the run: earlier people stay committed, the rest are
// untouched, and the next sync is not treated as completed.
func TestLDAPSyncStopsAtRereadErrorAsPartial(t *testing.T) {
	f := newSyncFixture(t, dirEntry("hank", "id-hank"), dirEntry("ivan", "id-ivan"), dirEntry("jill", "id-jill"))
	preview := f.preview(t)
	calls := 0
	f.dir.onLookupID = func(string) {
		calls++
		if calls == 2 {
			f.dir.lookupIDErr = errors.New("connection reset")
		}
	}

	result := f.execute(t, systemRequest.LDAPSyncExecuteRequest{PreviewID: preview.PreviewID})
	require.Equal(t, syncStatusPartial, result.Status)
	counts := map[string]int{}
	for _, o := range result.Outcomes {
		counts[o.Outcome]++
	}
	require.Equal(t, map[string]int{syncOutcomeDone: 1, syncOutcomeError: 1, syncOutcomeUnprocessed: 1}, counts)
	var created int64
	require.NoError(t, f.db.Model(&systemModel.SystemUser{}).Count(&created).Error)
	require.Equal(t, int64(1), created)
	cfg, err := f.svc.getLDAPConfig(context.Background())
	require.NoError(t, err)
	require.Empty(t, cfg.LastSyncKey, "a partial run does not resume auto-disable")
}

func TestLDAPSyncAbortsOnScopeReadFailure(t *testing.T) {
	f := newSyncFixture(t, dirEntry("hank", "id-hank"))
	f.dir.searchErr = errors.New("paged search failed")
	_, err := f.svc.PreviewSync(context.Background(), syncAdmin, AuditMeta{})
	require.Error(t, err)

	f.dir.searchErr = nil
	f.dir.remove("hank")
	_, err = f.svc.PreviewSync(context.Background(), syncAdmin, AuditMeta{})
	require.ErrorIs(t, err, ErrSyncScopeEmpty)
}
