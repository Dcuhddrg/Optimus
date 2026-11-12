package system

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"
	systemResponse "optimus/server/model/system/response"
	ldapUtil "optimus/server/utils/ldap"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrLDAPDisabled       = errors.New("LDAP is not enabled")
	ErrSyncScopeEmpty     = errors.New("no users found in the sync scope; nothing was changed")
	ErrSyncPreviewInvalid = errors.New("the preview is missing, expired, already executed, started by another administrator, or the LDAP config changed since; preview again")
)

// Execution outcomes and the actions they belong to.
const (
	syncOutcomeDone        = "done"
	syncOutcomeSkipped     = "skipped"
	syncOutcomeError       = "error"
	syncOutcomeUnprocessed = "unprocessed"

	syncActionSeparate    = "separate"
	syncActionBind        = "bind"
	syncActionUpdate      = "update"
	syncActionCreate      = "create"
	syncActionAutoDisable = "autoDisable"
	syncActionDisable     = "disable"

	syncStatusCompleted = "completed"
	syncStatusPartial   = "partial"
)

// Preview snapshots live in memory (design 3.3 D6): 30 minutes, one
// execution, only by the administrator who previewed. A restart drops them.
const syncSnapshotTTL = 30 * time.Minute

var syncNow = time.Now

type snapshotItem struct {
	category      SyncCategory
	accountSideID string
	entrySideID   string
	entryUsername string
}

type syncSnapshot struct {
	actorID       uint
	expiresAt     time.Time
	configID      uint
	configVersion uint64
	used          bool
	// items is keyed by account ID; only rows with an account can be chosen
	// or auto-disabled, so only those are recorded.
	items map[uint]snapshotItem
}

var syncSnapshots = struct {
	sync.Mutex
	byID map[string]*syncSnapshot
}{byID: map[string]*syncSnapshot{}}

func storeSyncSnapshot(snap *syncSnapshot) string {
	syncSnapshots.Lock()
	defer syncSnapshots.Unlock()
	now := syncNow()
	for id, existing := range syncSnapshots.byID {
		if existing.used || now.After(existing.expiresAt) {
			delete(syncSnapshots.byID, id)
		}
	}
	id := uuid.NewString()
	syncSnapshots.byID[id] = snap
	return id
}

// claimSyncSnapshot hands out a snapshot once, to its own administrator.
func claimSyncSnapshot(id string, actorID uint) (*syncSnapshot, error) {
	syncSnapshots.Lock()
	defer syncSnapshots.Unlock()
	snap, ok := syncSnapshots.byID[id]
	if !ok || snap.used || snap.actorID != actorID || syncNow().After(snap.expiresAt) {
		return nil, ErrSyncPreviewInvalid
	}
	snap.used = true
	return snap, nil
}

// PreviewSync reads the whole sync scope and returns the plan without
// writing anything. Any directory error, or an empty scope, aborts.
func (s *LDAPService) PreviewSync(ctx context.Context, actorID uint, meta AuditMeta) (systemResponse.LDAPSyncPreview, error) {
	event := failedAudit(actorID, meta, "ldap.sync.preview", "ldap", 0)
	defer func() { recordAudit(ctx, event) }()
	cfg, err := s.getLDAPConfig(ctx)
	if err != nil {
		event.Detail = auditDetails("config_unavailable", nil)
		return systemResponse.LDAPSyncPreview{}, err
	}
	if !cfg.Enabled {
		event.Detail = auditDetails("ldap_disabled", nil)
		return systemResponse.LDAPSyncPreview{}, ErrLDAPDisabled
	}
	plan, err := computeSyncPlan(ctx, newLDAPClient(), cfg)
	if err != nil {
		event.Detail = auditDetails(syncAbortReason(err), nil)
		return systemResponse.LDAPSyncPreview{}, err
	}
	snap := &syncSnapshot{actorID: actorID, expiresAt: syncNow().Add(syncSnapshotTTL), configID: cfg.ID, configVersion: cfg.Version, items: map[uint]snapshotItem{}}
	for _, item := range plan.Items {
		if item.Account != nil {
			snap.items[item.Account.ID] = snapshotOf(item)
		}
	}
	preview := systemResponse.LDAPSyncPreview{
		PreviewID:       storeSyncSnapshot(snap),
		ExpiresAt:       snap.expiresAt,
		DisablesBlocked: plan.DisablesBlocked,
		ConfigChanged:   plan.ConfigChanged,
		Items:           make([]systemResponse.LDAPSyncItem, 0, len(plan.Items)),
	}
	for _, item := range plan.Items {
		preview.Items = append(preview.Items, syncItemView(item))
	}
	event.Result = "success"
	event.Detail = auditDetailsMap("previewed", map[string]any{"counts": categoryCounts(plan), "disablesBlocked": plan.DisablesBlocked, "configChanged": plan.ConfigChanged})
	return preview, nil
}

func snapshotOf(item syncItem) snapshotItem {
	snap := snapshotItem{category: item.Category, accountSideID: item.accountSideID(), entrySideID: item.entrySideID()}
	if item.Entry != nil {
		snap.entryUsername = item.Entry.Username
	}
	return snap
}

func syncAbortReason(err error) string {
	if errors.Is(err, ErrSyncScopeEmpty) {
		return "scope_empty"
	}
	return "directory_read_failed"
}

// computeSyncPlan runs the full read: scope search, accounts and the
// absence check for each disable candidate.
func computeSyncPlan(ctx context.Context, client ldapUtil.Client, cfg systemModel.SystemLDAPConfig) (syncPlan, error) {
	conn := ldapConnectionConfig(cfg)
	entries, err := client.SearchScope(ctx, conn)
	if err != nil {
		return syncPlan{}, err
	}
	if len(entries) == 0 {
		return syncPlan{}, ErrSyncScopeEmpty
	}
	accounts, err := loadSyncAccounts(global.OPTIMUS_DB.WithContext(ctx))
	if err != nil {
		return syncPlan{}, err
	}
	exists := func(id string) (bool, error) { return client.ExistsInDirectory(ctx, conn, id) }
	return buildSyncPlan(entries, accounts, mappingOf(cfg), cfg.LastSyncKey != syncConfigKey(cfg), exists), nil
}

func loadSyncAccounts(db *gorm.DB) ([]syncAccount, error) {
	var users []systemModel.SystemUser
	if err := db.Order("id ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	var grants []struct {
		UserID     uint
		Code       string
		Approvable bool
	}
	if err := db.Table("sys_user_roles AS ur").Select("ur.user_id, r.code, r.approvable").Joins("JOIN sys_roles AS r ON r.id = ur.role_id").Order("r.code").Scan(&grants).Error; err != nil {
		return nil, err
	}
	roles, restricted := map[uint][]string{}, map[uint][]string{}
	for _, g := range grants {
		roles[g.UserID] = append(roles[g.UserID], g.Code)
		if isRestrictedForExternal(g.Code, g.Approvable) {
			restricted[g.UserID] = append(restricted[g.UserID], g.Code)
		}
	}
	var owned []struct {
		UserID uint
		Code   string
	}
	if err := db.Table("user_projects AS up").Select("up.user_id, p.code").Joins("JOIN projects AS p ON p.id = up.project_id").Where("up.project_role = ? AND p.deleted_at IS NULL", "owner").Order("p.code").Scan(&owned).Error; err != nil {
		return nil, err
	}
	projects := map[uint][]string{}
	for _, o := range owned {
		projects[o.UserID] = append(projects[o.UserID], o.Code)
	}
	accounts := make([]syncAccount, 0, len(users))
	for _, u := range users {
		accounts = append(accounts, syncAccount{ID: u.ID, Username: u.Username, DisplayName: u.DisplayName, Email: u.Email, Department: u.Department, ExternalID: u.ExternalID, Source: u.Source, Status: u.Status, CreatedAt: u.CreatedAt, Roles: roles[u.ID],
			PersonType: personTypeOrDefault(u.PersonType), RestrictedRoles: restricted[u.ID], OwnedProjects: projects[u.ID]})
	}
	return accounts, nil
}

// syncTask is one person's work in an execution.
type syncTask struct {
	action string
	item   syncItem
}

// ExecuteSync re-reads and recomputes, then processes people one at a time,
// each under the LDAP write-back lock (design 3.3, write ordering). It is
// not one transaction: a directory error on a re-read stops the run, people
// already committed stay committed, and the result is "partial".
func (s *LDAPService) ExecuteSync(ctx context.Context, actorID uint, req systemRequest.LDAPSyncExecuteRequest, meta AuditMeta) (systemResponse.LDAPSyncResult, error) {
	event := failedAudit(actorID, meta, "ldap.sync.execute", "ldap", 0)
	defer func() { recordAudit(ctx, event) }()
	snap, err := claimSyncSnapshot(req.PreviewID, actorID)
	if err != nil {
		event.Detail = auditDetails("preview_invalid", nil)
		return systemResponse.LDAPSyncResult{}, err
	}
	cfg, err := s.getLDAPConfig(ctx)
	if err != nil {
		event.Detail = auditDetails("config_unavailable", nil)
		return systemResponse.LDAPSyncResult{}, err
	}
	if !cfg.Enabled || cfg.ID != snap.configID || cfg.Version != snap.configVersion {
		event.Detail = auditDetails("preview_invalid", nil)
		return systemResponse.LDAPSyncResult{}, ErrSyncPreviewInvalid
	}
	// Each person's commit invalidates that person; this catches any path
	// that returns early.
	defer InvalidateAllAuthState()

	client := newLDAPClient()
	plan, err := computeSyncPlan(ctx, client, cfg)
	if err != nil {
		event.Detail = auditDetails(syncAbortReason(err), nil)
		return systemResponse.LDAPSyncResult{}, err
	}

	tasks, outcomes := planSyncTasks(plan, snap, req)
	result := systemResponse.LDAPSyncResult{Status: syncStatusCompleted, Outcomes: outcomes}
	run := &syncRun{ctx: ctx, client: client, cfg: cfg, actorID: actorID, meta: meta}
	for i, task := range tasks {
		outcome, err := run.process(task)
		result.Outcomes = append(result.Outcomes, outcome)
		if err != nil {
			result.Status = syncStatusPartial
			for _, rest := range tasks[i+1:] {
				result.Outcomes = append(result.Outcomes, taskOutcome(rest, syncOutcomeUnprocessed, ""))
			}
			break
		}
	}
	if result.Status == syncStatusCompleted {
		// Auto-disable resumes only after a fully completed sync under this
		// config; the version guard keeps a config saved meanwhile out.
		_ = global.OPTIMUS_DB.WithContext(ctx).Model(&systemModel.SystemLDAPConfig{}).
			Where("id = ? AND version = ?", cfg.ID, cfg.Version).Update("last_sync_key", syncConfigKey(cfg)).Error
	}
	event.Result = "success"
	event.Detail = auditDetailsMap(result.Status, map[string]any{
		"counts":          categoryCounts(plan),
		"outcomes":        outcomeCounts(result.Outcomes),
		"disablesBlocked": plan.DisablesBlocked,
		"configChanged":   plan.ConfigChanged,
		"stopped":         stoppedPeople(result.Outcomes),
	})
	return result, nil
}

// planSyncTasks turns the recomputed plan and the admin's choices into an
// ordered task list. Destructive actions (disable, bind, separate) must be in
// the snapshot with the same category and identities; create and update
// follow the recomputed plan.
func planSyncTasks(plan syncPlan, snap *syncSnapshot, req systemRequest.LDAPSyncExecuteRequest) ([]syncTask, []systemResponse.LDAPSyncOutcome) {
	byAccount := map[uint]syncItem{}
	for _, item := range plan.Items {
		if item.Account != nil {
			byAccount[item.Account.ID] = item
		}
	}
	var skipped []systemResponse.LDAPSyncOutcome
	var separate, bind, updates, creates, autoDisable, disable []syncTask

	choose := func(action string, ids []uint, allowed ...SyncCategory) []syncTask {
		var tasks []syncTask
		seen := map[uint]bool{}
		for _, id := range ids {
			if seen[id] {
				continue
			}
			seen[id] = true
			recorded, inSnapshot := snap.items[id]
			current, inPlan := byAccount[id]
			username := ""
			if inPlan {
				username = current.Account.Username
			}
			switch {
			case !inSnapshot || !categoryIn(recorded.category, allowed):
				skipped = append(skipped, systemResponse.LDAPSyncOutcome{Action: action, AccountID: id, Username: username, Outcome: syncOutcomeSkipped, Reason: "not_in_preview"})
			case !inPlan || current.Category != recorded.category:
				skipped = append(skipped, systemResponse.LDAPSyncOutcome{Action: action, AccountID: id, Username: username, Outcome: syncOutcomeSkipped, Reason: "category_changed"})
			case snapshotOf(current) != recorded:
				skipped = append(skipped, systemResponse.LDAPSyncOutcome{Action: action, AccountID: id, Username: username, Outcome: syncOutcomeSkipped, Reason: "identity_changed"})
			case action == syncActionSeparate && current.Reason == syncReasonLocalAccount:
				skipped = append(skipped, systemResponse.LDAPSyncOutcome{Action: action, AccountID: id, Username: username, Outcome: syncOutcomeSkipped, Reason: syncReasonLocalAccount})
			default:
				tasks = append(tasks, syncTask{action: action, item: current})
			}
		}
		return tasks
	}
	separate = choose(syncActionSeparate, req.Separate, SyncPendingBinding, SyncConflict)
	bind = choose(syncActionBind, req.Bind, SyncPendingBinding)
	disable = choose(syncActionDisable, req.Disable, SyncOutOfScope, SyncUnconfirmable)

	for _, item := range plan.Items {
		switch item.Category {
		case SyncUpdate:
			updates = append(updates, syncTask{action: syncActionUpdate, item: item})
		case SyncCreate:
			creates = append(creates, syncTask{action: syncActionCreate, item: item})
		case SyncAutoDisable:
			recorded, ok := snap.items[item.Account.ID]
			if !ok || recorded != snapshotOf(item) {
				skipped = append(skipped, taskOutcome(syncTask{action: syncActionAutoDisable, item: item}, syncOutcomeSkipped, "not_in_preview"))
				continue
			}
			autoDisable = append(autoDisable, syncTask{action: syncActionAutoDisable, item: item})
		}
	}
	if plan.DisablesBlocked {
		for _, task := range append(autoDisable, disable...) {
			skipped = append(skipped, taskOutcome(task, syncOutcomeSkipped, "disables_blocked"))
		}
		autoDisable, disable = nil, nil
	}
	// Separations first: they free usernames the other steps may need.
	var tasks []syncTask
	for _, group := range [][]syncTask{separate, bind, updates, creates, autoDisable, disable} {
		tasks = append(tasks, group...)
	}
	return tasks, skipped
}

func categoryIn(c SyncCategory, allowed []SyncCategory) bool {
	for _, a := range allowed {
		if a == c {
			return true
		}
	}
	return false
}

func taskOutcome(task syncTask, outcome, reason string) systemResponse.LDAPSyncOutcome {
	o := systemResponse.LDAPSyncOutcome{Action: task.action, Outcome: outcome, Reason: reason}
	if task.item.Account != nil {
		o.AccountID = task.item.Account.ID
		o.Username = task.item.Account.Username
	} else if task.item.Entry != nil {
		o.Username = task.item.Entry.Username
	}
	return o
}

// syncRun processes tasks for one execution.
type syncRun struct {
	ctx     context.Context
	client  ldapUtil.Client
	cfg     systemModel.SystemLDAPConfig
	actorID uint
	meta    AuditMeta
}

// errSyncStop marks an error that stops the run (a re-read failed or the
// config changed); people not yet processed are left untouched.
type errSyncStop struct{ reason string }

func (e errSyncStop) Error() string { return "sync stopped: " + e.reason }

// process handles one person under the write-back lock: re-read, verify,
// write, commit, invalidate.
func (r *syncRun) process(task syncTask) (systemResponse.LDAPSyncOutcome, error) {
	ldapWriteBack.Lock()
	defer ldapWriteBack.Unlock()

	cfg, err := (&LDAPService{}).getLDAPConfig(r.ctx)
	if err != nil || !cfg.Enabled || cfg.ID != r.cfg.ID || cfg.Version != r.cfg.Version {
		return taskOutcome(task, syncOutcomeError, "config_changed"), errSyncStop{"config_changed"}
	}
	reason, err := r.apply(task)
	if err != nil {
		var stop errSyncStop
		if errors.As(err, &stop) {
			return taskOutcome(task, syncOutcomeError, stop.reason), err
		}
		// A database error: stop as well, the remaining people stay as they are.
		return taskOutcome(task, syncOutcomeError, "write_failed"), errSyncStop{"write_failed"}
	}
	if reason != "" {
		return taskOutcome(task, syncOutcomeSkipped, reason), nil
	}
	return taskOutcome(task, syncOutcomeDone, ""), nil
}

// reread looks the identity up in the sync scope. Not found and duplicate
// are answers (the person is skipped); any other error stops the run.
func (r *syncRun) reread(externalID string) (LDAPUser, string, error) {
	entry, err := r.client.LookupExternalID(r.ctx, ldapConnectionConfig(r.cfg), externalID)
	switch {
	case errors.Is(err, ldapUtil.ErrUserNotFound):
		return LDAPUser{}, "entry_changed", nil
	case errors.Is(err, ldapUtil.ErrDuplicateIdentity):
		return LDAPUser{}, "identity_duplicated", nil
	case err != nil:
		return LDAPUser{}, "", errSyncStop{"directory_reread_failed"}
	}
	return entry, "", nil
}

func (r *syncRun) db() *gorm.DB { return global.OPTIMUS_DB.WithContext(r.ctx) }

// apply returns a skip reason, or an error that stops the run.
func (r *syncRun) apply(task syncTask) (string, error) {
	item := task.item
	switch task.action {
	case syncActionCreate:
		entry, skip, err := r.reread(item.Entry.ExternalID)
		if err != nil || skip != "" {
			return skip, err
		}
		if entry.Username != item.Entry.Username {
			return "entry_changed", nil
		}
		taken, err := r.identityOrNameTaken(entry.ExternalID, entry.Username, 0)
		if err != nil || taken {
			return "already_taken", err
		}
		user := newLDAPAccount(entry, r.cfg)
		if err := createLDAPAccount(r.db(), &user); err != nil {
			return "", err
		}
		InvalidateUserAuthState(user.ID)
		return "", nil

	case syncActionUpdate:
		entry, skip, err := r.reread(item.Entry.ExternalID)
		if err != nil || skip != "" {
			return skip, err
		}
		var account systemModel.SystemUser
		if err := r.db().Where("external_id = ?", entry.ExternalID).First(&account).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "account_changed", nil
			}
			return "", err
		}
		if account.ID != item.Account.ID || account.Status != DefaultUserStatus {
			return "account_changed", nil
		}
		taken, err := r.identityOrNameTaken("", entry.Username, account.ID)
		if err != nil || taken {
			return "username_taken", err
		}
		return "", r.db().Model(&systemModel.SystemUser{}).Where("id = ?", account.ID).Updates(ldapProfileColumns(entry, r.cfg)).Error

	case syncActionAutoDisable, syncActionDisable:
		var account systemModel.SystemUser
		if err := r.db().Where("id = ?", item.Account.ID).First(&account).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "account_changed", nil
			}
			return "", err
		}
		if account.Status != DefaultUserStatus || externalIDOf(account) != item.accountSideID() {
			return "account_changed", nil
		}
		if item.Category == SyncAutoDisable || item.Category == SyncOutOfScope {
			// Confirm absence again by the account-side identity.
			found, err := r.client.ExistsInDirectory(r.ctx, ldapConnectionConfig(r.cfg), item.accountSideID())
			if err != nil {
				return "", errSyncStop{"directory_reread_failed"}
			}
			if item.Category == SyncAutoDisable && found {
				return "present_in_directory", nil
			}
			if item.Category == SyncOutOfScope {
				if !found {
					return "category_changed", nil
				}
				if _, err := r.client.LookupExternalID(r.ctx, ldapConnectionConfig(r.cfg), item.accountSideID()); !errors.Is(err, ldapUtil.ErrUserNotFound) {
					if err != nil && !errors.Is(err, ldapUtil.ErrDuplicateIdentity) {
						return "", errSyncStop{"directory_reread_failed"}
					}
					return "back_in_scope", nil
				}
			}
		}
		return "", r.disable(account, task.action)

	case syncActionBind:
		entry, skip, err := r.reread(item.Entry.ExternalID)
		if err != nil || skip != "" {
			return skip, err
		}
		if entry.Username != item.Entry.Username {
			return "entry_changed", nil
		}
		taken, err := r.identityOrNameTaken(entry.ExternalID, "", 0)
		if err != nil || taken {
			return "identity_taken", err
		}
		// The condition keeps an existing binding even if locking changes.
		res := r.db().Model(&systemModel.SystemUser{}).Where("id = ? AND external_id IS NULL", item.Account.ID).Update("external_id", entry.ExternalID)
		if res.Error != nil {
			return "", res.Error
		}
		if res.RowsAffected == 0 {
			return "already_bound", nil
		}
		InvalidateUserAuthState(item.Account.ID)
		r.audit("user.ldap.bind", item.Account.ID, "bound", item, entry)
		return "", nil

	case syncActionSeparate:
		entry, skip, err := r.reread(item.Entry.ExternalID)
		if err != nil || skip != "" {
			return skip, err
		}
		if entry.Username != item.Entry.Username {
			return "entry_changed", nil
		}
		return r.separate(item, entry)
	}
	return "unknown_action", nil
}

// separate handles "this is another person": the old account is disabled and
// renamed to <uid>~<id>, keeping its UUID, roles, projects and audit
// references; the directory identity gets its own account (renamed if it
// already has one, otherwise new, with the default role).
func (r *syncRun) separate(item syncItem, entry LDAPUser) (string, error) {
	var old systemModel.SystemUser
	if err := r.db().Where("id = ?", item.Account.ID).First(&old).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "account_changed", nil
		}
		return "", err
	}
	if externalIDOf(old) != item.accountSideID() || old.Username != item.Account.Username {
		return "account_changed", nil
	}
	// Local accounts, the built-in admin above all, are never retired by a
	// directory entry (design 3.4: it is the fallback administrator).
	if old.Source != "ldap" || isBuiltinAdmin(old) {
		return syncReasonLocalAccount, nil
	}
	var newID uint
	err := r.db().Transaction(func(tx *gorm.DB) error {
		retired := fmt.Sprintf("%s~%d", old.Username, old.ID)
		if err := tx.Model(&systemModel.SystemUser{}).Where("id = ?", old.ID).Updates(map[string]any{"username": retired, "status": "disabled"}).Error; err != nil {
			return err
		}
		var owner systemModel.SystemUser
		err := tx.Where("external_id = ?", entry.ExternalID).First(&owner).Error
		if err == nil {
			newID = owner.ID
			return tx.Model(&systemModel.SystemUser{}).Where("id = ?", owner.ID).Updates(ldapProfileColumns(entry, r.cfg)).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		user := newLDAPAccount(entry, r.cfg)
		if err := createLDAPAccount(tx, &user); err != nil {
			return err
		}
		newID = user.ID
		return nil
	})
	if err != nil {
		return "", err
	}
	InvalidateUserAuthState(old.ID)
	InvalidateUserAuthState(newID)
	r.audit("user.ldap.separate", old.ID, "separated", item, entry)
	recordAudit(r.ctx, AuditEvent{ActorUserID: r.actorID, ActorUsername: r.meta.ActorUsername, Action: "user.status.update", ResourceType: "user", ResourceID: resourceID(old.ID), Result: "success", IP: r.meta.IP, UserAgent: r.meta.UserAgent,
		Detail: auditDetails("ldap_sync_separated", map[string]auditChange{"status": {Before: old.Status, After: "disabled"}})})
	return "", nil
}

func (r *syncRun) disable(account systemModel.SystemUser, action string) error {
	if err := r.db().Model(&systemModel.SystemUser{}).Where("id = ?", account.ID).Update("status", "disabled").Error; err != nil {
		return err
	}
	InvalidateUserAuthState(account.ID)
	reason := "ldap_sync_auto_disabled"
	if action == syncActionDisable {
		reason = "ldap_sync_disabled"
	}
	recordAudit(r.ctx, AuditEvent{ActorUserID: r.actorID, ActorUsername: r.meta.ActorUsername, Action: "user.status.update", ResourceType: "user", ResourceID: resourceID(account.ID), Result: "success", IP: r.meta.IP, UserAgent: r.meta.UserAgent,
		Detail: auditDetails(reason, map[string]auditChange{"status": {Before: account.Status, After: "disabled"}})})
	return nil
}

// identityOrNameTaken reports whether an account other than except already
// holds the identity or the username (either may be empty to skip it).
func (r *syncRun) identityOrNameTaken(externalID, username string, except uint) (bool, error) {
	query := r.db().Model(&systemModel.SystemUser{}).Where("id <> ?", except)
	switch {
	case externalID != "" && username != "":
		query = query.Where("external_id = ? OR username = ?", externalID, username)
	case externalID != "":
		query = query.Where("external_id = ?", externalID)
	default:
		query = query.Where("username = ?", username)
	}
	var n int64
	err := query.Count(&n).Error
	return n > 0, err
}

// audit records a binding decision with both sides as they were and the
// creation-time comparison (design 3.3, migration).
func (r *syncRun) audit(action string, accountID uint, reason string, item syncItem, entry LDAPUser) {
	detail := map[string]any{
		"account":    map[string]any{"username": item.Account.Username, "displayName": item.Account.DisplayName, "email": item.Account.Email, "createdAt": item.Account.CreatedAt, "externalId": item.accountSideID()},
		"entry":      map[string]any{"username": entry.Username, "displayName": entry.DisplayName, "email": entry.Email, "externalId": entry.ExternalID, "createdAt": entry.CreatedAt},
		"category":   item.Category,
		"suspicious": item.Suspicious,
	}
	recordAudit(r.ctx, AuditEvent{ActorUserID: r.actorID, ActorUsername: r.meta.ActorUsername, Action: action, ResourceType: "user", ResourceID: resourceID(accountID), Result: "success", IP: r.meta.IP, UserAgent: r.meta.UserAgent, Detail: auditDetailsMap(reason, detail)})
}

func externalIDOf(u systemModel.SystemUser) string {
	if u.ExternalID == nil {
		return ""
	}
	return *u.ExternalID
}

// newLDAPAccount builds a new LDAP account: a random UUID, never derived
// from the username or the identity (design 3.3).
func newLDAPAccount(entry LDAPUser, cfg systemModel.SystemLDAPConfig) systemModel.SystemUser {
	externalID := entry.ExternalID
	user := systemModel.SystemUser{
		UUID:        uuid.NewString(),
		Username:    entry.Username,
		DisplayName: entry.DisplayName,
		Email:       entry.Email,
		ExternalID:  &externalID,
		Source:      "ldap",
		Status:      DefaultUserStatus,
		PersonType:  personTypeOf(entry, cfg),
	}
	if cfg.DepartmentAttr != "" {
		user.Department = entry.Department
	}
	return user
}

// createLDAPAccount inserts the account and binds the default role once.
// External accounts get no role: a new directory entry must not grant access
// by itself (design 3.6).
func createLDAPAccount(db *gorm.DB, user *systemModel.SystemUser) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if user.PersonType == PersonExternal {
			return nil
		}
		return bindDefaultUserRole(tx, user.ID)
	})
}

// ldapProfileColumns are the LDAP-owned columns written by login and sync.
func ldapProfileColumns(entry LDAPUser, cfg systemModel.SystemLDAPConfig) map[string]any {
	columns := map[string]any{"username": entry.Username, "display_name": entry.DisplayName, "email": entry.Email, "person_type": personTypeOf(entry, cfg)}
	if cfg.DepartmentAttr != "" {
		columns["department"] = entry.Department
	}
	return columns
}

func syncItemView(item syncItem) systemResponse.LDAPSyncItem {
	view := systemResponse.LDAPSyncItem{Category: string(item.Category), Reason: item.Reason, Changes: item.Changes, Suspicious: item.Suspicious, RestrictedRoles: item.RestrictedRoles, OwnedProjects: item.OwnedProjects}
	if a := item.Account; a != nil {
		view.Account = &systemResponse.LDAPSyncAccount{ID: a.ID, Username: a.Username, DisplayName: a.DisplayName, Email: a.Email, Department: a.Department, ExternalID: a.externalID(), Status: a.Status, CreatedAt: a.CreatedAt, Roles: a.Roles}
	}
	if e := item.Entry; e != nil {
		view.Entry = &systemResponse.LDAPSyncEntry{DN: e.DN, Username: e.Username, DisplayName: e.DisplayName, Email: e.Email, Department: e.Department, ExternalID: e.ExternalID, CreatedAt: e.CreatedAt}
	}
	return view
}

func categoryCounts(plan syncPlan) map[string]int {
	counts := map[string]int{}
	for _, item := range plan.Items {
		counts[string(item.Category)]++
	}
	return counts
}

func outcomeCounts(outcomes []systemResponse.LDAPSyncOutcome) map[string]int {
	counts := map[string]int{}
	for _, o := range outcomes {
		counts[o.Action+"."+o.Outcome]++
	}
	return counts
}

// stoppedPeople lists who errored or was not processed in a partial run.
func stoppedPeople(outcomes []systemResponse.LDAPSyncOutcome) []string {
	var people []string
	for _, o := range outcomes {
		if o.Outcome == syncOutcomeError || o.Outcome == syncOutcomeUnprocessed {
			people = append(people, o.Outcome+":"+o.Username+"#"+strconv.Itoa(int(o.AccountID)))
		}
	}
	return people
}
