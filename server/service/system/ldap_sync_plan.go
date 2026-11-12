package system

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"

	systemModel "optimus/server/model/system"
	systemResponse "optimus/server/model/system/response"
)

// SyncCategory is how a full LDAP sync classifies one entry or account
// (design 3.3 D6).
type SyncCategory string

const (
	SyncCreate         SyncCategory = "create"
	SyncUpdate         SyncCategory = "update"
	SyncAutoDisable    SyncCategory = "autoDisable"
	SyncUnconfirmable  SyncCategory = "unconfirmable"
	SyncOutOfScope     SyncCategory = "outOfScope"
	SyncPendingBinding SyncCategory = "pendingBinding"
	SyncConflict       SyncCategory = "conflict"
	SyncDisabledInfo   SyncCategory = "disabledInfo"
	SyncUnidentifiable SyncCategory = "unidentifiable"
)

// Reasons attached to unconfirmable, unidentifiable and conflict items.
const (
	syncReasonUnbound            = "unbound"
	syncReasonConfigChanged      = "config_changed"
	syncReasonAbsenceCheckFailed = "absence_check_failed"
	syncReasonMissingUsername    = "missing_username"
	syncReasonMissingIdentity    = "missing_identity"
	syncReasonDuplicateUsername  = "duplicate_username"
	syncReasonDuplicateIdentity  = "duplicate_identity"
	syncReasonUsernameReused     = "username_reused"
	syncReasonRenameTaken        = "rename_taken"
	// The username is held by a local account (today only the built-in
	// admin): listed, but never resolved by the sync; the directory entry
	// must be renamed.
	syncReasonLocalAccount = "local_account"
)

// syncAccount is the Optimus side of the comparison.
type syncAccount struct {
	ID          uint
	Username    string
	DisplayName string
	Email       string
	Department  string
	ExternalID  *string
	Source      string
	Status      string
	CreatedAt   time.Time
	Roles       []string
}

func (a syncAccount) externalID() string {
	if a.ExternalID == nil {
		return ""
	}
	return *a.ExternalID
}

type syncFieldChange = systemResponse.LDAPSyncFieldChange

// syncItem is one row of the plan. Account is nil for create and
// unidentifiable; Entry is nil for the absence categories.
type syncItem struct {
	Category SyncCategory
	Reason   string
	Account  *syncAccount
	Entry    *LDAPUser
	Changes  []syncFieldChange
	// Suspicious marks a pending binding whose directory entry was created
	// after the account, or whose creation time is unreadable: the uid was
	// likely reused, so it cannot be confirmed in bulk.
	Suspicious bool
}

// accountSideID and entrySideID are the identities the snapshot records.
func (i syncItem) accountSideID() string {
	if i.Account == nil {
		return ""
	}
	return i.Account.externalID()
}

func (i syncItem) entrySideID() string {
	if i.Entry == nil {
		return ""
	}
	return i.Entry.ExternalID
}

type syncPlan struct {
	Items []syncItem
	// DisablesBlocked is set when an entry could not be identified: an
	// unreadable identity would look like an absent person.
	DisablesBlocked bool
	// ConfigChanged is set when base DN, filter or id_attr changed since the
	// last successful sync; every disable candidate then needs the admin.
	ConfigChanged bool
}

// absenceCheck reports whether an identity exists anywhere in the directory.
type absenceCheck func(externalID string) (bool, error)

// buildSyncPlan compares the sync scope with the LDAP accounts. accounts
// holds every account, local ones included, so usernames they hold count as
// taken. syncDepartment is false when department_attr is empty.
func buildSyncPlan(entries []LDAPUser, accounts []syncAccount, syncDepartment bool, configChanged bool, exists absenceCheck) syncPlan {
	plan := syncPlan{ConfigChanged: configChanged}

	// Identify entries: both a username and an identity, each unique in scope.
	nameCount, idCount := map[string]int{}, map[string]int{}
	for _, e := range entries {
		if e.Username != "" {
			nameCount[strings.ToLower(e.Username)]++
		}
		if e.ExternalID != "" {
			idCount[e.ExternalID]++
		}
	}
	var valid []LDAPUser
	for i := range entries {
		e := entries[i]
		reason := ""
		switch {
		case e.Username == "":
			reason = syncReasonMissingUsername
		case e.ExternalID == "":
			reason = syncReasonMissingIdentity
		case nameCount[strings.ToLower(e.Username)] > 1:
			reason = syncReasonDuplicateUsername
		case idCount[e.ExternalID] > 1:
			reason = syncReasonDuplicateIdentity
		}
		if reason != "" {
			plan.Items = append(plan.Items, syncItem{Category: SyncUnidentifiable, Reason: reason, Entry: &e})
			plan.DisablesBlocked = true
			continue
		}
		valid = append(valid, e)
	}

	byID, byName := map[string]*syncAccount{}, map[string]*syncAccount{}
	for i := range accounts {
		a := &accounts[i]
		if id := a.externalID(); id != "" {
			byID[id] = a
		}
		byName[strings.ToLower(a.Username)] = a
	}
	matched := map[uint]bool{}

	for i := range valid {
		e := valid[i]
		owner := byID[e.ExternalID]
		holder := byName[strings.ToLower(e.Username)]
		if owner != nil {
			matched[owner.ID] = true
			if holder != nil && holder.ID != owner.ID {
				// Renamed onto a username another account holds.
				matched[holder.ID] = true
				plan.Items = append(plan.Items, syncItem{Category: SyncConflict, Reason: conflictReason(holder, syncReasonRenameTaken), Account: holder, Entry: &e})
				continue
			}
			if owner.Status != DefaultUserStatus {
				plan.Items = append(plan.Items, syncItem{Category: SyncDisabledInfo, Account: owner, Entry: &e})
				continue
			}
			if changes := profileChanges(*owner, e, syncDepartment); len(changes) > 0 {
				plan.Items = append(plan.Items, syncItem{Category: SyncUpdate, Account: owner, Entry: &e, Changes: changes})
			}
			continue
		}
		if holder == nil {
			plan.Items = append(plan.Items, syncItem{Category: SyncCreate, Entry: &e})
			continue
		}
		matched[holder.ID] = true
		if holder.Source == "ldap" && holder.ExternalID == nil {
			plan.Items = append(plan.Items, syncItem{Category: SyncPendingBinding, Account: holder, Entry: &e, Suspicious: e.CreatedAt == nil || e.CreatedAt.After(holder.CreatedAt)})
			continue
		}
		// Same username, another identity (or a local account): another person.
		plan.Items = append(plan.Items, syncItem{Category: SyncConflict, Reason: conflictReason(holder, syncReasonUsernameReused), Account: holder, Entry: &e})
	}

	for i := range accounts {
		a := &accounts[i]
		if a.Source != "ldap" || matched[a.ID] || a.Status != DefaultUserStatus {
			continue
		}
		switch {
		case a.ExternalID == nil:
			plan.Items = append(plan.Items, syncItem{Category: SyncUnconfirmable, Reason: syncReasonUnbound, Account: a})
		case configChanged:
			plan.Items = append(plan.Items, syncItem{Category: SyncUnconfirmable, Reason: syncReasonConfigChanged, Account: a})
		default:
			found, err := exists(*a.ExternalID)
			switch {
			case err != nil:
				plan.Items = append(plan.Items, syncItem{Category: SyncUnconfirmable, Reason: syncReasonAbsenceCheckFailed, Account: a})
			case found:
				plan.Items = append(plan.Items, syncItem{Category: SyncOutOfScope, Account: a})
			default:
				plan.Items = append(plan.Items, syncItem{Category: SyncAutoDisable, Account: a})
			}
		}
	}
	sort.SliceStable(plan.Items, func(i, j int) bool { return plan.Items[i].sortKey() < plan.Items[j].sortKey() })
	return plan
}

func conflictReason(holder *syncAccount, reason string) string {
	if holder.Source != "ldap" {
		return syncReasonLocalAccount
	}
	return reason
}

func (i syncItem) sortKey() string {
	if i.Account != nil {
		return string(i.Category) + "\x00" + strings.ToLower(i.Account.Username)
	}
	return string(i.Category) + "\x00" + strings.ToLower(i.Entry.Username)
}

// profileChanges lists the LDAP-owned fields that differ.
func profileChanges(a syncAccount, e LDAPUser, syncDepartment bool) []syncFieldChange {
	var changes []syncFieldChange
	add := func(field, before, after string) {
		if before != after {
			changes = append(changes, syncFieldChange{Field: field, Before: before, After: after})
		}
	}
	add("username", a.Username, e.Username)
	add("displayName", a.DisplayName, e.DisplayName)
	add("email", a.Email, e.Email)
	if syncDepartment {
		add("department", a.Department, e.Department)
	}
	return changes
}

// syncConfigKey is the part of the config whose change makes the next sync's
// disable candidates need the admin (design 3.3, auto-disable condition 3):
// everything that decides which directory is read and what it shows. The
// server URL and bind account count, since another server or another bind
// account's access rights can hide people who still exist.
func syncConfigKey(cfg systemModel.SystemLDAPConfig) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{cfg.URL, cfg.BindDN, cfg.BaseDN, cfg.UserFilter, cfg.IDAttr}, "\x00")))
	return hex.EncodeToString(sum[:])
}
