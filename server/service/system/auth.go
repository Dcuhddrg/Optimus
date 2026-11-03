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
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user disabled")
)

type TokenSubject struct {
	ID        uint
	UUID      string
	Username  string
	RoleIDs   []uint
	RoleCodes []string
}

type TokenIssuer interface {
	IssueToken(subject TokenSubject) (string, error)
}

type AuthOperations interface {
	Login(context.Context, systemRequest.LoginRequest, AuditMeta) (systemResponse.LoginResult, error)
	CurrentUser(context.Context, uint) (systemResponse.CurrentUserResult, error)
	Logout(context.Context, uint, AuditMeta) error
}

type AuthService struct {
	tokenIssuer TokenIssuer
	ldap        LDAPAuthenticator
	now         func() time.Time
}

func NewAuthService(tokenIssuer TokenIssuer, ldap LDAPAuthenticator) *AuthService {
	return &AuthService{
		tokenIssuer: tokenIssuer,
		ldap:        ldap,
		now:         time.Now,
	}
}

func (s *AuthService) Login(ctx context.Context, req systemRequest.LoginRequest, meta AuditMeta) (systemResponse.LoginResult, error) {
	user, err := s.findUserByUsername(ctx, req.Username)
	if err != nil {
		if req.Username != defaultAdminUsername {
			return s.loginLDAP(ctx, req, meta)
		}
		s.writeAudit(ctx, loginFailure(req.Username, meta, 0, "user_lookup_failed"))
		return systemResponse.LoginResult{}, ErrInvalidCredentials
	}

	if user.Status != defaultUserStatus {
		s.writeAudit(ctx, loginFailure(req.Username, meta, user.ID, "user_disabled"))
		return systemResponse.LoginResult{}, ErrUserDisabled
	}

	if user.Source != "local" || user.Username != defaultAdminUsername || !ComparePassword(user.PasswordHash, req.Password) {
		if user.Username != defaultAdminUsername {
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
	if s.ldap == nil {
		s.writeAudit(ctx, loginFailure(req.Username, meta, 0, "ldap_unavailable"))
		return systemResponse.LoginResult{}, ErrInvalidCredentials
	}
	cfg, err := s.getLDAPConfig(ctx)
	if err != nil || !cfg.Enabled {
		s.writeAudit(ctx, loginFailure(req.Username, meta, 0, "ldap_config_unavailable"))
		return systemResponse.LoginResult{}, ErrInvalidCredentials
	}
	ldapUser, err := s.ldap.Authenticate(ctx, ldapConnectionConfig(cfg), req.Username, req.Password)
	if err != nil {
		s.writeAudit(ctx, loginFailure(req.Username, meta, 0, "ldap_authentication_failed"))
		return systemResponse.LoginResult{}, ErrInvalidCredentials
	}
	user, err := s.syncLDAPUser(ctx, ldapUser)
	if err != nil {
		s.writeAudit(ctx, loginFailure(req.Username, meta, 0, "user_sync_failed"))
		return systemResponse.LoginResult{}, err
	}
	if user.Status != defaultUserStatus {
		s.writeAudit(ctx, loginFailure(req.Username, meta, user.ID, "user_disabled"))
		return systemResponse.LoginResult{}, ErrUserDisabled
	}
	result, err := s.issueLoginResult(ctx, user, meta)
	if err != nil {
		s.writeAudit(ctx, loginFailure(req.Username, meta, user.ID, "session_issue_failed"))
	}
	return result, err
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

	now := s.now()
	token, err := s.tokenIssuer.IssueToken(TokenSubject{
		ID:        user.ID,
		UUID:      user.UUID,
		Username:  user.Username,
		RoleIDs:   roleIDs,
		RoleCodes: roleCodes,
	})
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

func (s *AuthService) syncLDAPUser(ctx context.Context, ldapUser LDAPUser) (systemModel.SystemUser, error) {
	var user systemModel.SystemUser
	err := global.OPTIMUS_DB.WithContext(ctx).Where("username = ?", ldapUser.Username).First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return systemModel.SystemUser{}, err
	}
	if err == gorm.ErrRecordNotFound {
		user = systemModel.SystemUser{
			UUID:     "ldap:" + ldapUser.Username,
			Username: ldapUser.Username,
			Source:   "ldap",
			Status:   defaultUserStatus,
		}
	}
	user.DisplayName = ldapUser.DisplayName
	user.Email = ldapUser.Email
	if err == gorm.ErrRecordNotFound {
		if err := global.OPTIMUS_DB.WithContext(ctx).Create(&user).Error; err != nil {
			return systemModel.SystemUser{}, err
		}
		if err := s.bindDefaultUserRole(ctx, user.ID); err != nil {
			return systemModel.SystemUser{}, err
		}
		return user, nil
	}
	return user, global.OPTIMUS_DB.WithContext(ctx).Save(&user).Error
}

func (s *AuthService) bindDefaultUserRole(ctx context.Context, userID uint) error {
	var role systemModel.SystemRole
	if err := global.OPTIMUS_DB.WithContext(ctx).Where("code = ?", "user").First(&role).Error; err != nil {
		return err
	}
	userRole := systemModel.SystemUserRole{UserID: userID, RoleID: role.ID}
	return global.OPTIMUS_DB.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, role.ID).FirstOrCreate(&userRole).Error
}

func queryUserRoles(ctx context.Context, userID uint) ([]systemModel.SystemRoleSummary, error) {
	var roles []systemModel.SystemRoleSummary
	err := global.OPTIMUS_DB.WithContext(ctx).
		Table((&systemModel.SystemRole{}).TableName()+" AS roles").
		Select("roles.id, roles.code, roles.name").
		Joins("JOIN sys_user_roles user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.status = ?", userID, defaultRoleStatus).
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
