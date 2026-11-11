package system

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func strp(s string) *string { return &s }

var planTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func ldapAccount(id uint, username string, externalID *string) syncAccount {
	return syncAccount{ID: id, Username: username, DisplayName: username, ExternalID: externalID, Source: "ldap", Status: DefaultUserStatus, CreatedAt: planTime}
}

func planEntry(username, id string) LDAPUser {
	created := planTime.Add(-time.Hour)
	return LDAPUser{Username: username, DisplayName: username, ExternalID: id, CreatedAt: &created}
}

func noneElsewhere(string) (bool, error) { return false, nil }

func categories(plan syncPlan) map[string]SyncCategory {
	out := map[string]SyncCategory{}
	for _, item := range plan.Items {
		key := ""
		if item.Account != nil {
			key = item.Account.Username
		} else {
			key = "entry:" + item.Entry.Username
		}
		out[key] = item.Category
	}
	return out
}

func itemFor(t *testing.T, plan syncPlan, username string) syncItem {
	t.Helper()
	for _, item := range plan.Items {
		if (item.Account != nil && item.Account.Username == username) || (item.Account == nil && item.Entry.Username == username) {
			return item
		}
	}
	t.Fatalf("no item for %s", username)
	return syncItem{}
}

func TestSyncPlanClassifiesEveryCategory(t *testing.T) {
	disabled := ldapAccount(5, "dora", strp("id-dora"))
	disabled.Status = "disabled"
	accounts := []syncAccount{
		ldapAccount(1, "alice", strp("id-alice")),   // unchanged
		ldapAccount(2, "bob", strp("id-bob")),       // update (rename to robert)
		ldapAccount(3, "carol", strp("id-carol")),   // gone: auto-disable
		ldapAccount(4, "dan", strp("id-dan")),       // moved out of scope
		disabled,                                    // in directory, disabled here
		ldapAccount(6, "erin", nil),                 // legacy, found by username
		ldapAccount(7, "fred", nil),                 // legacy, not found
		ldapAccount(8, "gina", strp("id-old-gina")), // uid reused by another person
		{ID: 9, Username: "admin", Source: "local", Status: DefaultUserStatus},
	}
	entries := []LDAPUser{
		planEntry("alice", "id-alice"),
		planEntry("robert", "id-bob"),
		planEntry("dora", "id-dora"),
		planEntry("erin", "id-erin"),
		planEntry("gina", "id-new-gina"),
		planEntry("hank", "id-hank"),
	}
	entries[1].DisplayName = "bob"
	plan := buildSyncPlan(entries, accounts, true, false, func(id string) (bool, error) { return id == "id-dan", nil })

	require.False(t, plan.DisablesBlocked)
	got := categories(plan)
	require.Equal(t, map[string]SyncCategory{
		"bob":        SyncUpdate,
		"carol":      SyncAutoDisable,
		"dan":        SyncOutOfScope,
		"dora":       SyncDisabledInfo,
		"erin":       SyncPendingBinding,
		"fred":       SyncUnconfirmable,
		"gina":       SyncConflict,
		"entry:hank": SyncCreate,
	}, got, "unchanged alice and local admin produce no row")
	update := itemFor(t, plan, "bob")
	require.Equal(t, []syncFieldChange{{Field: "username", Before: "bob", After: "robert"}}, update.Changes)
	conflict := itemFor(t, plan, "gina")
	require.Equal(t, "id-old-gina", conflict.accountSideID())
	require.Equal(t, "id-new-gina", conflict.entrySideID())
	require.Equal(t, syncReasonUnbound, itemFor(t, plan, "fred").Reason)
}

func TestSyncPlanUnidentifiableEntriesBlockDisables(t *testing.T) {
	accounts := []syncAccount{ldapAccount(1, "carol", strp("id-carol"))}
	entries := []LDAPUser{
		planEntry("", "id-x"),
		planEntry("noid", ""),
		planEntry("twin", "id-t1"), planEntry("twin", "id-t2"),
		planEntry("u1", "id-same"), planEntry("u2", "id-same"),
		planEntry("hank", "id-hank"),
	}
	plan := buildSyncPlan(entries, accounts, true, false, noneElsewhere)
	require.True(t, plan.DisablesBlocked)
	reasons := map[string]int{}
	for _, item := range plan.Items {
		if item.Category == SyncUnidentifiable {
			reasons[item.Reason]++
		}
	}
	require.Equal(t, map[string]int{syncReasonMissingUsername: 1, syncReasonMissingIdentity: 1, syncReasonDuplicateUsername: 2, syncReasonDuplicateIdentity: 2}, reasons)
	require.Equal(t, SyncCreate, categories(plan)["entry:hank"], "creates still go ahead")
	require.Equal(t, SyncAutoDisable, categories(plan)["carol"], "listed, but execution skips it")
}

func TestSyncPlanAbsenceNeedsAdminWhenUncertain(t *testing.T) {
	accounts := []syncAccount{ldapAccount(1, "carol", strp("id-carol"))}
	entries := []LDAPUser{planEntry("hank", "id-hank")}

	changed := buildSyncPlan(entries, accounts, true, true, noneElsewhere)
	require.Equal(t, SyncUnconfirmable, categories(changed)["carol"])
	require.Equal(t, syncReasonConfigChanged, itemFor(t, changed, "carol").Reason)

	failing := buildSyncPlan(entries, accounts, true, false, func(string) (bool, error) { return false, errors.New("no root access") })
	require.Equal(t, SyncUnconfirmable, categories(failing)["carol"])
	require.Equal(t, syncReasonAbsenceCheckFailed, itemFor(t, failing, "carol").Reason)
}

// A rename onto a username another account holds is a conflict on that
// account; an identity already owned by an account is never offered for
// binding to a legacy account.
func TestSyncPlanRenameOntoTakenUsernameIsConflict(t *testing.T) {
	accounts := []syncAccount{ldapAccount(1, "bob", strp("id-bob")), ldapAccount(2, "robert", nil)}
	plan := buildSyncPlan([]LDAPUser{planEntry("robert", "id-bob")}, accounts, true, false, noneElsewhere)
	item := itemFor(t, plan, "robert")
	require.Equal(t, SyncConflict, item.Category)
	require.Equal(t, syncReasonRenameTaken, item.Reason)
	require.NotContains(t, categories(plan), "bob", "bob is matched, not a disable candidate")
}

func TestSyncPlanFlagsLikelyReusedUIDForBinding(t *testing.T) {
	accounts := []syncAccount{ldapAccount(1, "old", nil), ldapAccount(2, "later", nil), ldapAccount(3, "unknown", nil)}
	later := planEntry("later", "id-later")
	after := planTime.Add(time.Hour)
	later.CreatedAt = &after
	unknown := planEntry("unknown", "id-unknown")
	unknown.CreatedAt = nil
	plan := buildSyncPlan([]LDAPUser{planEntry("old", "id-old"), later, unknown}, accounts, true, false, noneElsewhere)
	require.False(t, itemFor(t, plan, "old").Suspicious)
	require.True(t, itemFor(t, plan, "later").Suspicious, "entry created after the account")
	require.True(t, itemFor(t, plan, "unknown").Suspicious, "creation time unreadable")
}

func TestSyncPlanDepartmentOnlyWhenSynced(t *testing.T) {
	account := ldapAccount(1, "alice", strp("id-alice"))
	account.Department = "old"
	entry := planEntry("alice", "id-alice")
	entry.Department = "new"
	require.Empty(t, buildSyncPlan([]LDAPUser{entry}, []syncAccount{account}, false, false, noneElsewhere).Items)
	require.Len(t, buildSyncPlan([]LDAPUser{entry}, []syncAccount{account}, true, false, noneElsewhere).Items, 1)
}
