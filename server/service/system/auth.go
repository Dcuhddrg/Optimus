package system

import (
	"context"
	"errors"
	"time"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"
	systemResponse "optimus/server/model/system/response"
	"optimus/server/utils"
	ldapUtil "optimus/server/utils/ldap"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user disabled")

	// LDAP login refusals (design 3.3); the messages reach the login page.
	ErrLDAPAccountConflict   = errors.New("account conflict, contact an administrator")
	ErrLDAPBindingPending    = errors.New("account is awaiting administrator confirmation, contact an administrator")
	ErrLDAPIdentityMissing   = errors.New("directory identity is missing, contact an administrator")
	ErrLDAPIdentityAnomaly   = errors.New("directory identity is inconsistent, contact an administrator")
	ErrLDAPDuplicateUsername = errors.New("the directory has duplicate accounts with this username, contact an administrator")
	ErrLDAPDirectoryChanging = errors.New("the directory is changing, please retry later")
)

type AuthService struct{}

func (s *AuthService) Login(ctx context.Context, req systemRequest.LoginRequest, meta AuditMeta) (systemResponse.LoginResult, error) {
	user, err := s.findUserByUsername(ctx, req.Username)
	if err != nil {
		if req.Username != DefaultAdminUsername {
			return s.loginLDAP(ctx, req, meta)
		}
		s.writeAudit(ctx, loginFailure(req.Username, meta, 0, "user_lookup_failed"))
		return systemResponse.LoginResult{}, ErrInvalidCredentials
	}

	if user.Status != DefaultUserStatus {
		s.writeAudit(ctx, loginFailure(req.Username, meta, user.ID, "user_disabled"))
		return systemResponse.LoginResult{}, ErrUserDisabled
	}

	if user.Source != "local" || user.Username != DefaultAdminUsername || !utils.BcryptCheck(req.Password, user.PasswordHash) {
		if user.Username != DefaultAdminUsername {
			return s.loginLDAP(ctx, req, meta)
		}
		s.writeAudit(ctx, loginFailure(req.Username, meta, user.ID, "invalid_credentials"))
		return systemResponse.LoginResult{}, ErrInvalidCredentials
	}

	result, err := s.issueLoginResult(ctx, user, meta)
	if err != nil {
		s.writeAudit(ctx, loginFailure(req.Username, meta, user.ID, "session_issue_failed"))
	}
	return result, err
}

func (s *AuthService) loginLDAP(ctx context.Context, req systemRequest.LoginRequest, meta AuditMeta) (systemResponse.LoginResult, error) {
	cfg, err := s.getLDAPConfig(ctx)
	if err != nil || !cfg.Enabled {
		s.writeAudit(ctx, loginFailure(req.Username, meta, 0, "ldap_config_unavailable"))
		return systemResponse.LoginResult{}, ErrInvalidCredentials
	}
	conn := ldapConnectionConfig(cfg)
	client := newLDAPClient()
	// Outside the write-back lock: password check plus the identity read
	// before and after the bind (design 3.3, write ordering step 1).
	entry, err := client.Authenticate(ctx, conn, req.Username, req.Password)
	if err != nil {
		reason, userErr := ldapAuthenticationFailure(err)
		s.writeAudit(ctx, loginFailure(req.Username, meta, 0, reason))
		return systemResponse.LoginResult{}, userErr
	}
	user, failure := s.syncLDAPUser(ctx, client, conn, req.Username, entry)
	if failure != nil {
		s.writeAudit(ctx, loginFailure(req.Username, meta, failure.userID, failure.reason))
		return systemResponse.LoginResult{}, failure.err
	}
	if user.Status != DefaultUserStatus {
		s.writeAudit(ctx, loginFailure(req.Username, meta, user.ID, "user_disabled"))
		return systemResponse.LoginResult{}, ErrUserDisabled
	}
	result, err := s.issueLoginResult(ctx, user, meta)
	if err != nil {
		s.writeAudit(ctx, loginFailure(req.Username, meta, user.ID, "session_issue_failed"))
	}
	return result, err
}

// ldapAuthenticationFailure maps a directory authentication error to the
// audit reason and the error shown to the user.
func ldapAuthenticationFailure(err error) (string, error) {
	switch {
	case errors.Is(err, ldapUtil.ErrDuplicateUsername):
		return "ldap_duplicate_username", ErrLDAPDuplicateUsername
	case errors.Is(err, ldapUtil.ErrMissingIdentity):
		return "ldap_identity_missing", ErrLDAPIdentityMissing
	case errors.Is(err, ldapUtil.ErrDuplicateIdentity):
		return "ldap_identity_duplicate", ErrLDAPIdentityAnomaly
	case errors.Is(err, ldapUtil.ErrMissingVersion):
		return "ldap_version_missing", ErrLDAPIdentityAnomaly
	case errors.Is(err, ldapUtil.ErrEntryChanged):
		return "ldap_entry_changed", ErrLDAPDirectoryChanging
	case errors.Is(err, ldapUtil.ErrInvalidLoginFilter):
		return "ldap_config_invalid", ErrInvalidCredentials
	default:
		return "ldap_authentication_failed", ErrInvalidCredentials
	}
}

func (s *AuthService) issueLoginResult(ctx context.Context, user systemModel.SystemUser, meta AuditMeta) (systemResponse.LoginResult, error) {
	roles, err := queryUserRoles(ctx, user.ID)
	if err != nil {
		return systemResponse.LoginResult{}, err
	}
	roleIDs, roleCodes := summarizeRoles(roles)
	menuPermissions, err := s.menuPermissionCodesForRoles(ctx, roleIDs)
	if err != nil {
		return systemResponse.LoginResult{}, err
	}

	now := time.Now()
	j := utils.NewJWT()
	token, err := j.CreateToken(j.CreateClaims(utils.BaseClaims{
		ID:        user.ID,
		UUID:      user.UUID,
		Username:  user.Username,
		RoleIDs:   roleIDs,
		RoleCodes: roleCodes,
	}))
	if err != nil {
		return systemResponse.LoginResult{}, err
	}

	_ = s.updateLastLogin(ctx, user.ID, now)
	s.writeAudit(ctx, AuditEvent{ActorUserID: user.ID, ActorUsername: user.Username, Action: "auth.login.success", Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})

	return systemResponse.LoginResult{
		Token:           token,
		User:            summarizeUser(user),
		Roles:           roleCodes,
		MenuPermissions: menuPermissions,
	}, nil
}

func (s *AuthService) CurrentUser(ctx context.Context, userID uint) (systemResponse.CurrentUserResult, error) {
	user, err := s.findUserByID(ctx, userID)
	if err != nil {
		return systemResponse.CurrentUserResult{}, err
	}
	roles, err := queryUserRoles(ctx, user.ID)
	if err != nil {
		return systemResponse.CurrentUserResult{}, err
	}
	roleIDs, roleCodes := summarizeRoles(roles)
	menuPermissions, err := s.menuPermissionCodesForRoles(ctx, roleIDs)
	if err != nil {
		return systemResponse.CurrentUserResult{}, err
	}

	return systemResponse.CurrentUserResult{
		User:            summarizeUser(user),
		Roles:           roleCodes,
		MenuPermissions: menuPermissions,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID uint, meta AuditMeta) error {
	s.writeAudit(ctx, AuditEvent{
		ActorUserID:   userID,
		ActorUsername: meta.ActorUsername,
		Action:        "auth.logout",
		Result:        "success",
		IP:            meta.IP,
		UserAgent:     meta.UserAgent,
	})
	return nil
}

func (s *AuthService) writeAudit(ctx context.Context, event AuditEvent) {
	recordAudit(ctx, event)
}

func summarizeUser(user systemModel.SystemUser) systemResponse.SystemUserSummary {
	return systemResponse.SystemUserSummary{
		ID:          user.ID,
		UUID:        user.UUID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Department:  user.Department,
		Bound:       user.ExternalID != nil,
		Source:      user.Source,
		Status:      user.Status,
	}
}

func summarizeRoles(roles []systemModel.SystemRoleSummary) ([]uint, []string) {
	roleIDs := make([]uint, 0, len(roles))
	roleCodes := make([]string, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
		roleCodes = append(roleCodes, role.Code)
	}
	return roleIDs, roleCodes
}

func (s *AuthService) findUserByUsername(ctx context.Context, username string) (systemModel.SystemUser, error) {
	var user systemModel.SystemUser
	err := global.OPTIMUS_DB.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return user, err
}

func (s *AuthService) findUserByID(ctx context.Context, id uint) (systemModel.SystemUser, error) {
	var user systemModel.SystemUser
	err := global.OPTIMUS_DB.WithContext(ctx).Where("id = ?", id).First(&user).Error
	return user, err
}

func (s *AuthService) getLDAPConfig(ctx context.Context) (systemModel.SystemLDAPConfig, error) {
	var cfg systemModel.SystemLDAPConfig
	err := global.OPTIMUS_DB.WithContext(ctx).Order("id ASC").First(&cfg).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return systemModel.SystemLDAPConfig{}, nil
	}
	return cfg, err
}

type ldapLoginFailure struct {
	reason string
	err    error
	userID uint
}

// syncLDAPUser writes the directory entry back under the LDAP write-back lock
// (design 3.3). It re-reads the entry inside the lock; the identity must
// still be the one whose password was verified, otherwise the login is
// refused instead of continuing as whoever holds the username now. Accounts
// are matched by ExternalID, never by username.
func (s *AuthService) syncLDAPUser(ctx context.Context, client ldapUtil.Client, conn LDAPConnectionConfig, loginName string, verified LDAPUser) (systemModel.SystemUser, *ldapLoginFailure) {
	ldapWriteBack.Lock()
	defer ldapWriteBack.Unlock()

	current, err := client.LookupUsername(ctx, conn, loginName)
	switch {
	case errors.Is(err, ldapUtil.ErrUserNotFound), errors.Is(err, ldapUtil.ErrDuplicateUsername):
		return systemModel.SystemUser{}, &ldapLoginFailure{reason: "ldap_entry_changed", err: ErrLDAPDirectoryChanging}
	case err != nil:
		return systemModel.SystemUser{}, &ldapLoginFailure{reason: "ldap_reread_failed", err: ErrInvalidCredentials}
	case current.ExternalID != verified.ExternalID:
		return systemModel.SystemUser{}, &ldapLoginFailure{reason: "ldap_entry_changed", err: ErrLDAPDirectoryChanging}
	}
	username := current.Username
	if username == "" {
		username = loginName
	}
	profile := map[string]any{"username": username, "display_name": current.DisplayName, "email": current.Email}
	if conn.DepartmentAttr != "" {
		profile["department"] = current.Department
	}

	db := global.OPTIMUS_DB.WithContext(ctx)
	var user systemModel.SystemUser
	err = db.Where("external_id = ?", current.ExternalID).First(&user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return systemModel.SystemUser{}, &ldapLoginFailure{reason: "user_sync_failed", err: err}
	}
	if err == nil {
		// Same identity: a renamed username follows the directory, roles and
		// projects stay. The new name must not belong to another account.
		if user.Username != username {
			var taken int64
			if err := db.Model(&systemModel.SystemUser{}).Where("username = ? AND id <> ?", username, user.ID).Count(&taken).Error; err != nil {
				return systemModel.SystemUser{}, &ldapLoginFailure{reason: "user_sync_failed", err: err, userID: user.ID}
			}
			if taken > 0 {
				return systemModel.SystemUser{}, &ldapLoginFailure{reason: "ldap_identity_conflict", err: ErrLDAPAccountConflict, userID: user.ID}
			}
		}
		// Only the synced profile columns: a full-row Save would restore a
		// status an administrator disabled after the read.
		if err := db.Model(&systemModel.SystemUser{}).Where("id = ?", user.ID).Updates(profile).Error; err != nil {
			return systemModel.SystemUser{}, &ldapLoginFailure{reason: "user_sync_failed", err: err, userID: user.ID}
		}
		// Re-read so login decides on the current status, not the earlier read.
		var updated systemModel.SystemUser
		if err := db.Where("id = ?", user.ID).First(&updated).Error; err != nil {
			return systemModel.SystemUser{}, &ldapLoginFailure{reason: "user_sync_failed", err: err, userID: user.ID}
		}
		return updated, nil
	}

	var byName systemModel.SystemUser
	err = db.Where("username = ?", username).First(&byName).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return systemModel.SystemUser{}, &ldapLoginFailure{reason: "user_sync_failed", err: err}
	}
	if err == nil {
		// A legacy LDAP account is bound only after an admin confirms it is
		// the same person; the username alone does not prove that.
		if byName.Source == "ldap" && byName.ExternalID == nil {
			return systemModel.SystemUser{}, &ldapLoginFailure{reason: "ldap_binding_pending", err: ErrLDAPBindingPending, userID: byName.ID}
		}
		// Same username, different identity: another person.
		return systemModel.SystemUser{}, &ldapLoginFailure{reason: "ldap_identity_conflict", err: ErrLDAPAccountConflict, userID: byName.ID}
	}

	externalID := current.ExternalID
	created := systemModel.SystemUser{
		UUID:        uuid.NewString(),
		Username:    username,
		DisplayName: current.DisplayName,
		Email:       current.Email,
		Department:  current.Department,
		ExternalID:  &externalID,
		Source:      "ldap",
		Status:      DefaultUserStatus,
	}
	if conn.DepartmentAttr == "" {
		created.Department = ""
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&created).Error; err != nil {
			return err
		}
		return bindDefaultUserRole(tx, created.ID)
	})
	if err != nil {
		return systemModel.SystemUser{}, &ldapLoginFailure{reason: "user_sync_failed", err: err}
	}
	InvalidateUserAuthState(created.ID)
	return created, nil
}

// bindDefaultUserRole binds the default role to a newly created account,
// once (design 3.3 D5).
func bindDefaultUserRole(db *gorm.DB, userID uint) error {
	var role systemModel.SystemRole
	if err := db.Where("code = ?", DefaultUserRoleCode).First(&role).Error; err != nil {
		return err
	}
	userRole := systemModel.SystemUserRole{UserID: userID, RoleID: role.ID}
	return db.Where("user_id = ? AND role_id = ?", userID, role.ID).FirstOrCreate(&userRole).Error
}

func queryUserRoles(ctx context.Context, userID uint) ([]systemModel.SystemRoleSummary, error) {
	var roles []systemModel.SystemRoleSummary
	err := global.OPTIMUS_DB.WithContext(ctx).
		Table((&systemModel.SystemRole{}).TableName()+" AS roles").
		Select("roles.id, roles.code, roles.name").
		Joins("JOIN sys_user_roles user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Order("roles.id ASC").
		Scan(&roles).Error
	return roles, err
}

func (s *AuthService) menuPermissionCodesForRoles(ctx context.Context, roleIDs []uint) ([]string, error) {
	if len(roleIDs) == 0 {
		return []string{}, nil
	}

	var codes []string
	err := global.OPTIMUS_DB.WithContext(ctx).
		Table((&systemModel.SystemPermission{}).TableName()+" AS permissions").
		Select("permissions.code").
		Joins("JOIN sys_role_permissions role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id IN ? AND permissions.type = ?", roleIDs, "menu").
		Group("permissions.code").
		Order("MIN(permissions.sort) ASC").
		Pluck("permissions.code", &codes).Error
	return codes, err
}

func (s *AuthService) updateLastLogin(ctx context.Context, userID uint, at time.Time) error {
	return global.OPTIMUS_DB.WithContext(ctx).
		Model(&systemModel.SystemUser{}).
		Where("id = ?", userID).
		Update("last_login_at", at).Error
}
