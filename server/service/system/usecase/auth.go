package usecase

import (
	"context"
	"errors"
	"time"

	systemModel "optimus/server/model/system"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user disabled")
)

type AuthStore interface {
	FindUserByUsername(ctx context.Context, username string) (systemModel.SystemUser, error)
	FindUserByID(ctx context.Context, id uint) (systemModel.SystemUser, error)
	GetLDAPConfig(ctx context.Context) (systemModel.SystemLDAPConfig, error)
	SyncLDAPUser(ctx context.Context, user LDAPUser) (systemModel.SystemUser, error)
	RolesForUser(ctx context.Context, userID uint) ([]SystemRoleSummary, error)
	MenuPermissionCodesForRoles(ctx context.Context, roleIDs []uint) ([]string, error)
	UpdateLastLogin(ctx context.Context, userID uint, at time.Time) error
	WriteAudit(ctx context.Context, event AuditEvent) error
}

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

type AuthService struct {
	store        AuthStore
	tokenIssuer  TokenIssuer
	ldap         LDAPAuthenticator
	now          func() time.Time
}

type LoginRequest struct {
	Username string
	Password string
}

type LoginResult struct {
	Token           string            `json:"token"`
	User            SystemUserSummary `json:"user"`
	Roles           []string          `json:"roles"`
	MenuPermissions []string          `json:"menuPermissions"`
}

type CurrentUserResult struct {
	User            SystemUserSummary `json:"user"`
	Roles           []string          `json:"roles"`
	MenuPermissions []string          `json:"menuPermissions"`
}

type SystemUserSummary struct {
	ID          uint   `json:"id"`
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Source      string `json:"source"`
	Status      string `json:"status"`
}

type SystemRoleSummary struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

func NewAuthService(store AuthStore, tokenIssuer TokenIssuer, ldap LDAPAuthenticator) *AuthService {
	return &AuthService{
		store:       store,
		tokenIssuer: tokenIssuer,
		ldap:        ldap,
		now:         time.Now,
	}
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest, meta AuditMeta) (LoginResult, error) {
	user, err := s.store.FindUserByUsername(ctx, req.Username)
	if err != nil {
		if req.Username != defaultAdminUsername {
			return s.loginLDAP(ctx, req, meta)
		}
		s.writeAudit(ctx, AuditEvent{Action: "auth.login.failure", Result: "failure", IP: meta.IP, UserAgent: meta.UserAgent, Detail: "user lookup failed"})
		return LoginResult{}, ErrInvalidCredentials
	}

	if user.Status != defaultUserStatus {
		s.writeAudit(ctx, AuditEvent{ActorUserID: user.ID, Action: "auth.login.failure", Result: "failure", IP: meta.IP, UserAgent: meta.UserAgent, Detail: "user disabled"})
		return LoginResult{}, ErrUserDisabled
	}

	if user.Source != "local" || user.Username != defaultAdminUsername || !ComparePassword(user.PasswordHash, req.Password) {
		if user.Username != defaultAdminUsername {
			return s.loginLDAP(ctx, req, meta)
		}
		s.writeAudit(ctx, AuditEvent{ActorUserID: user.ID, Action: "auth.login.failure", Result: "failure", IP: meta.IP, UserAgent: meta.UserAgent, Detail: "invalid credentials"})
		return LoginResult{}, ErrInvalidCredentials
	}

	return s.issueLoginResult(ctx, user, meta)
}

func (s *AuthService) loginLDAP(ctx context.Context, req LoginRequest, meta AuditMeta) (LoginResult, error) {
	if s.ldap == nil {
		s.writeAudit(ctx, AuditEvent{Action: "auth.login.failure", Result: "failure", IP: meta.IP, UserAgent: meta.UserAgent, Detail: "ldap client unavailable"})
		return LoginResult{}, ErrInvalidCredentials
	}
	cfg, err := s.store.GetLDAPConfig(ctx)
	if err != nil || !cfg.Enabled {
		s.writeAudit(ctx, AuditEvent{Action: "auth.login.failure", Result: "failure", IP: meta.IP, UserAgent: meta.UserAgent, Detail: "ldap disabled"})
		return LoginResult{}, ErrInvalidCredentials
	}
	ldapUser, err := s.ldap.Authenticate(ctx, ldapConnectionConfig(cfg), req.Username, req.Password)
	if err != nil {
		s.writeAudit(ctx, AuditEvent{Action: "auth.login.failure", Result: "failure", IP: meta.IP, UserAgent: meta.UserAgent, Detail: "ldap authentication failed"})
		return LoginResult{}, ErrInvalidCredentials
	}
	user, err := s.store.SyncLDAPUser(ctx, ldapUser)
	if err != nil {
		return LoginResult{}, err
	}
	if user.Status != defaultUserStatus {
		s.writeAudit(ctx, AuditEvent{ActorUserID: user.ID, Action: "auth.login.failure", Result: "failure", IP: meta.IP, UserAgent: meta.UserAgent, Detail: "user disabled"})
		return LoginResult{}, ErrUserDisabled
	}
	return s.issueLoginResult(ctx, user, meta)
}

func (s *AuthService) issueLoginResult(ctx context.Context, user systemModel.SystemUser, meta AuditMeta) (LoginResult, error) {
	roles, err := s.store.RolesForUser(ctx, user.ID)
	if err != nil {
		return LoginResult{}, err
	}
	roleIDs, roleCodes := summarizeRoles(roles)
	menuPermissions, err := s.store.MenuPermissionCodesForRoles(ctx, roleIDs)
	if err != nil {
		return LoginResult{}, err
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
		return LoginResult{}, err
	}

	_ = s.store.UpdateLastLogin(ctx, user.ID, now)
	s.writeAudit(ctx, AuditEvent{ActorUserID: user.ID, Action: "auth.login.success", Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})

	return LoginResult{
		Token:           token,
		User:            summarizeUser(user),
		Roles:           roleCodes,
		MenuPermissions: menuPermissions,
	}, nil
}

func (s *AuthService) CurrentUser(ctx context.Context, userID uint) (CurrentUserResult, error) {
	user, err := s.store.FindUserByID(ctx, userID)
	if err != nil {
		return CurrentUserResult{}, err
	}
	roles, err := s.store.RolesForUser(ctx, user.ID)
	if err != nil {
		return CurrentUserResult{}, err
	}
	roleIDs, roleCodes := summarizeRoles(roles)
	menuPermissions, err := s.store.MenuPermissionCodesForRoles(ctx, roleIDs)
	if err != nil {
		return CurrentUserResult{}, err
	}

	return CurrentUserResult{
		User:            summarizeUser(user),
		Roles:           roleCodes,
		MenuPermissions: menuPermissions,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, userID uint, meta AuditMeta) error {
	s.writeAudit(ctx, AuditEvent{
		ActorUserID: userID,
		Action:      "auth.logout",
		Result:      "success",
		IP:          meta.IP,
		UserAgent:   meta.UserAgent,
	})
	return nil
}

func (s *AuthService) writeAudit(ctx context.Context, event AuditEvent) {
	_ = s.store.WriteAudit(ctx, event)
}

func summarizeUser(user systemModel.SystemUser) SystemUserSummary {
	return SystemUserSummary{
		ID:          user.ID,
		UUID:        user.UUID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Source:      user.Source,
		Status:      user.Status,
	}
}

func summarizeRoles(roles []SystemRoleSummary) ([]uint, []string) {
	roleIDs := make([]uint, 0, len(roles))
	roleCodes := make([]string, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
		roleCodes = append(roleCodes, role.Code)
	}
	return roleIDs, roleCodes
}
