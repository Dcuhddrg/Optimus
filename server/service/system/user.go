package system

import (
	"context"
	"errors"
	"time"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"
	systemResponse "optimus/server/model/system/response"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrUserNotFound = errors.New("user not found")
	// ErrBuiltinAdminProtected: the built-in admin account must stay enabled
	// and keep the admin role, so the system always has an administrator.
	// ErrLDAPProfileReadOnly rejects profile edits of an LDAP account.
	ErrLDAPProfileReadOnly   = errors.New("profile fields of an LDAP account come from the directory and cannot be edited")
	ErrBuiltinAdminProtected = errors.New("built-in admin must stay enabled with the admin role")
	// ErrExternalRoleRestricted: external accounts cannot hold the admin role
	// or approval roles (design 3.6).
	ErrExternalRoleRestricted = errors.New("external accounts cannot hold the admin role or approval roles")
)

type UserService struct{}

func (s *UserService) ListUsers(ctx context.Context) (systemResponse.UserListResult, error) {
	users, err := s.listUsers(ctx)
	if err != nil {
		return systemResponse.UserListResult{}, err
	}

	byID := make(map[uint]systemModel.SystemUser, len(users))
	for _, user := range users {
		byID[user.ID] = user
	}
	now := time.Now()
	owned, err := ownedProjectsByUser(global.OPTIMUS_DB.WithContext(ctx))
	if err != nil {
		return systemResponse.UserListResult{}, err
	}
	items := make([]systemResponse.UserListItem, 0, len(users))
	for _, user := range users {
		roles, err := queryUserRoles(ctx, user.ID)
		if err != nil {
			return systemResponse.UserListResult{}, err
		}
		_, roleCodes := summarizeRoles(roles)
		item := systemResponse.UserListItem{SystemUserSummary: summarizeUser(user), Roles: roleCodes}
		if personTypeOrDefault(user.PersonType) == PersonExternal {
			item.Expired = accountExpired(user.PersonType, user.ExpiresAt, now)
			for _, role := range roles {
				if isRestrictedForExternal(role.Code, role.Approvable) {
					item.RestrictedRoles = append(item.RestrictedRoles, role.Code)
				}
			}
			item.OwnedProjects = owned[user.ID]
			if user.SponsorID != nil {
				sponsor, ok := byID[*user.SponsorID]
				item.Sponsor = sponsor.Username
				item.SponsorInvalid = !ok || sponsor.Status != DefaultUserStatus || personTypeOrDefault(sponsor.PersonType) != PersonEmployee
			}
		}
		items = append(items, item)
	}

	return systemResponse.UserListResult{Items: items, Total: len(items)}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, actorID uint, userID uint, req systemRequest.UpdateUserRequest, meta AuditMeta) (systemResponse.SystemUserSummary, error) {
	user, err := s.findUserByID(ctx, userID)
	if err != nil {
		return systemResponse.SystemUserSummary{}, err
	}
	// LDAP is the source of truth for an LDAP account's profile (design 3.3).
	if user.Source == "ldap" {
		recordAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "user.update", ResourceType: "user", ResourceID: resourceID(userID), Result: "failure", IP: meta.IP, UserAgent: meta.UserAgent, Detail: auditDetails("ldap_profile_read_only", nil)})
		return systemResponse.SystemUserSummary{}, ErrLDAPProfileReadOnly
	}
	// Write only the edited columns: a full-row Save of this earlier read
	// would restore a status disabled in the meantime.
	updated, err := s.updateUserColumns(ctx, user.ID, map[string]any{"display_name": req.DisplayName, "email": req.Email})
	if err != nil {
		return systemResponse.SystemUserSummary{}, err
	}
	recordAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "user.update", ResourceType: "user", ResourceID: resourceID(userID), Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return summarizeUser(updated), nil
}

func (s *UserService) UpdateUserStatus(ctx context.Context, actorID uint, userID uint, status string, meta AuditMeta) (systemResponse.SystemUserSummary, error) {
	event := failedAudit(actorID, meta, "user.status.update", "user", userID)
	defer func() { recordAudit(ctx, event) }()
	user, err := s.findUserByID(ctx, userID)
	if err != nil {
		event.Detail = auditDetails("target_lookup_failed", nil)
		return systemResponse.SystemUserSummary{}, err
	}
	if isBuiltinAdmin(user) && status != DefaultUserStatus {
		event.Detail = auditDetails("builtin_admin_protected", nil)
		return systemResponse.SystemUserSummary{}, ErrBuiltinAdminProtected
	}
	before := user.Status
	if err := s.writeUserColumns(ctx, user.ID, map[string]any{"status": status}); err != nil {
		event.Detail = auditDetails("update_failed", nil)
		return systemResponse.SystemUserSummary{}, err
	}
	// Committed: invalidate before anything else can fail, so a disable
	// takes effect even if re-reading the row below fails.
	InvalidateUserAuthState(userID)
	updated, err := s.findUserByID(ctx, userID)
	if err != nil {
		event.Result = "success"
		event.Detail = auditDetails("snapshot_unavailable", nil)
		return systemResponse.SystemUserSummary{}, err
	}
	event.Result = "success"
	event.Detail = auditDetails("updated", map[string]auditChange{"status": {Before: before, After: updated.Status}})
	return summarizeUser(updated), nil
}

func (s *UserService) UpdateUserRoles(ctx context.Context, actorID uint, userID uint, roleIDs []uint, meta AuditMeta) error {
	event := failedAudit(actorID, meta, "user.roles.update", "user", userID)
	defer func() { recordAudit(ctx, event) }()
	user, err := s.findUserByID(ctx, userID)
	if err != nil {
		event.Detail = auditDetails("target_lookup_failed", nil)
		return err
	}
	if isBuiltinAdmin(user) {
		keepsAdmin, err := includesAdminRole(ctx, roleIDs)
		if err != nil {
			event.Detail = auditDetails("admin_check_failed", nil)
			return err
		}
		if !keepsAdmin {
			event.Detail = auditDetails("builtin_admin_protected", nil)
			return ErrBuiltinAdminProtected
		}
	}
	if personTypeOrDefault(user.PersonType) == PersonExternal {
		restricted, err := includesRestrictedRole(ctx, roleIDs)
		if err != nil {
			event.Detail = auditDetails("update_failed", nil)
			return err
		}
		if restricted {
			event.Detail = auditDetails("external_role_restricted", nil)
			return ErrExternalRoleRestricted
		}
	}
	before, beforeErr := queryUserRoles(ctx, userID)
	if err := s.replaceUserRoles(ctx, userID, roleIDs); err != nil {
		if errors.Is(err, ErrRoleNotFound) {
			event.Detail = auditDetails("role_not_found", nil)
			return err
		}
		event.Detail = auditDetails("update_failed", nil)
		return err
	}
	InvalidateUserAuthState(userID)
	after, afterErr := queryUserRoles(ctx, userID)
	event.Result = "success"
	if beforeErr != nil || afterErr != nil {
		event.Detail = auditDetails("snapshot_unavailable", nil)
	} else {
		beforeIDs, afterIDs := make([]uint, 0, len(before)), make([]uint, 0, len(after))
		for _, role := range before {
			beforeIDs = append(beforeIDs, role.ID)
		}
		for _, role := range after {
			afterIDs = append(afterIDs, role.ID)
		}
		event.Detail = auditDetails("updated", map[string]auditChange{"roleIds": auditSetChange(beforeIDs, afterIDs)})
	}
	return nil
}

func isBuiltinAdmin(user systemModel.SystemUser) bool {
	return user.Source == "local" && user.UUID == BuiltinAdminUUID
}

// includesRestrictedRole reports whether any of roleIDs is the admin role or
// an approval role. Authorization ignores them for external accounts anyway,
// so a concurrent person type or flag change cannot grant them.
func includesRestrictedRole(ctx context.Context, roleIDs []uint) (bool, error) {
	if len(roleIDs) == 0 {
		return false, nil
	}
	var count int64
	err := global.OPTIMUS_DB.WithContext(ctx).Model(&systemModel.SystemRole{}).Where("id IN ? AND (code = ? OR approvable = ?)", roleIDs, AdminRoleCode, true).Count(&count).Error
	return count > 0, err
}

func includesAdminRole(ctx context.Context, roleIDs []uint) (bool, error) {
	var adminRole systemModel.SystemRole
	if err := global.OPTIMUS_DB.WithContext(ctx).Where("code = ?", AdminRoleCode).First(&adminRole).Error; err != nil {
		return false, err
	}
	for _, roleID := range roleIDs {
		if roleID == adminRole.ID {
			return true, nil
		}
	}
	return false, nil
}

func (s *UserService) listUsers(ctx context.Context) ([]systemModel.SystemUser, error) {
	var users []systemModel.SystemUser
	err := global.OPTIMUS_DB.WithContext(ctx).Order("id ASC").Find(&users).Error
	return users, err
}

func (s *UserService) findUserByID(ctx context.Context, id uint) (systemModel.SystemUser, error) {
	var user systemModel.SystemUser
	err := global.OPTIMUS_DB.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return systemModel.SystemUser{}, ErrUserNotFound
	}
	return user, err
}

// updateUserColumns updates only the given columns and returns the row as
// stored afterwards. Never Save a whole user read earlier: concurrent writes
// to other columns (status, roles' holders) would be overwritten.
func (s *UserService) updateUserColumns(ctx context.Context, userID uint, columns map[string]any) (systemModel.SystemUser, error) {
	if err := s.writeUserColumns(ctx, userID, columns); err != nil {
		return systemModel.SystemUser{}, err
	}
	return s.findUserByID(ctx, userID)
}

func (s *UserService) writeUserColumns(ctx context.Context, userID uint, columns map[string]any) error {
	return global.OPTIMUS_DB.WithContext(ctx).Model(&systemModel.SystemUser{}).Where("id = ?", userID).Updates(columns).Error
}

func (s *UserService) replaceUserRoles(ctx context.Context, userID uint, roleIDs []uint) error {
	return global.OPTIMUS_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Share-lock the roles being bound, so a concurrent DeleteRole (which
		// takes an exclusive lock) either waits for this commit and then sees
		// the holder, or commits first and this assignment finds the role gone.
		requested := map[uint]bool{}
		for _, roleID := range roleIDs {
			requested[roleID] = true
		}
		existing := []uint{}
		if len(requested) > 0 {
			if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Model(&systemModel.SystemRole{}).Where("id IN ?", roleIDs).Order("id ASC").Pluck("id", &existing).Error; err != nil {
				return err
			}
		}
		// Any requested role missing (deleted, or never existed): reject the
		// whole assignment and keep the user's current bindings, rather than
		// silently replacing them with fewer roles than the admin asked for.
		if len(existing) != len(requested) {
			return ErrRoleNotFound
		}
		if hook := roleLockTestHooks.afterAssignLock; hook != nil {
			hook(userID)
		}
		if err := tx.Where("user_id = ?", userID).Delete(&systemModel.SystemUserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range existing {
			if err := tx.Create(&systemModel.SystemUserRole{UserID: userID, RoleID: roleID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

var (
	// ErrNotExternal: expiry and sponsor apply to external accounts only.
	ErrNotExternal = errors.New("expiry and sponsor apply to external accounts only")
	// ErrInvalidExpiry: the expiry is not a YYYY-MM-DD date.
	ErrInvalidExpiry = errors.New("expiry must be a date (YYYY-MM-DD)")
	// ErrInvalidSponsor: the sponsor must be an enabled employee.
	ErrInvalidSponsor = errors.New("sponsor must be an enabled employee")
)

// UpdateUserExternal sets an external account's expiry and sponsor (design
// 3.6). Without both, the account has no roles; past the expiry it can no
// longer sign in. The write is conditioned on the account still being
// external, and the user's cached authorization is invalidated.
func (s *UserService) UpdateUserExternal(ctx context.Context, actorID uint, userID uint, req systemRequest.UpdateUserExternalRequest, meta AuditMeta) (systemResponse.SystemUserSummary, error) {
	event := failedAudit(actorID, meta, "user.external.update", "user", userID)
	defer func() { recordAudit(ctx, event) }()
	user, err := s.findUserByID(ctx, userID)
	if err != nil {
		event.Detail = auditDetails("target_lookup_failed", nil)
		return systemResponse.SystemUserSummary{}, err
	}
	if personTypeOrDefault(user.PersonType) != PersonExternal {
		event.Detail = auditDetails("not_external", nil)
		return systemResponse.SystemUserSummary{}, ErrNotExternal
	}
	var expiresAt *time.Time
	if req.ExpiresOn != nil {
		end, err := endOfDay(*req.ExpiresOn)
		if err != nil {
			event.Detail = auditDetails("invalid_expiry", nil)
			return systemResponse.SystemUserSummary{}, ErrInvalidExpiry
		}
		expiresAt = &end
	}
	if req.SponsorID != nil {
		sponsor, err := s.findUserByID(ctx, *req.SponsorID)
		if errors.Is(err, ErrUserNotFound) || (err == nil && (sponsor.Status != DefaultUserStatus || personTypeOrDefault(sponsor.PersonType) != PersonEmployee)) {
			event.Detail = auditDetails("invalid_sponsor", nil)
			return systemResponse.SystemUserSummary{}, ErrInvalidSponsor
		}
		if err != nil {
			event.Detail = auditDetails("update_failed", nil)
			return systemResponse.SystemUserSummary{}, err
		}
	}
	res := global.OPTIMUS_DB.WithContext(ctx).Model(&systemModel.SystemUser{}).Where("id = ? AND person_type = ?", userID, PersonExternal).
		Updates(map[string]any{"expires_at": expiresAt, "sponsor_id": req.SponsorID})
	if res.Error != nil {
		event.Detail = auditDetails("update_failed", nil)
		return systemResponse.SystemUserSummary{}, res.Error
	}
	if res.RowsAffected == 0 {
		// Became an employee meanwhile, or the values were unchanged.
		updated, err := s.findUserByID(ctx, userID)
		if err != nil || personTypeOrDefault(updated.PersonType) != PersonExternal {
			event.Detail = auditDetails("not_external", nil)
			return systemResponse.SystemUserSummary{}, ErrNotExternal
		}
	}
	InvalidateUserAuthState(userID)
	updated, err := s.findUserByID(ctx, userID)
	if err != nil {
		event.Result = "success"
		event.Detail = auditDetails("snapshot_unavailable", nil)
		return systemResponse.SystemUserSummary{}, err
	}
	event.Result = "success"
	event.Detail = auditDetails("updated", map[string]auditChange{
		"expiresOn": {Before: expiresOn(user.ExpiresAt), After: expiresOn(updated.ExpiresAt)},
		"sponsorId": {Before: user.SponsorID, After: updated.SponsorID},
	})
	return summarizeUser(updated), nil
}
