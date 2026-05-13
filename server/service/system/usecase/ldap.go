package usecase

import (
	"context"
	"errors"
	"strings"

	systemModel "optimus/server/model/system"
)

var ErrInvalidLDAPConfig = errors.New("invalid ldap config")

type LDAPStore interface {
	GetLDAPConfig(ctx context.Context) (systemModel.SystemLDAPConfig, error)
	SaveLDAPConfig(ctx context.Context, cfg systemModel.SystemLDAPConfig) (systemModel.SystemLDAPConfig, error)
	WriteAudit(ctx context.Context, event AuditEvent) error
}

type LDAPTester interface {
	Test(ctx context.Context, cfg LDAPConnectionConfig) error
}

type LDAPAuthenticator interface {
	Authenticate(ctx context.Context, cfg LDAPConnectionConfig, username, password string) (LDAPUser, error)
}

type LDAPService struct {
	store  LDAPStore
	client LDAPTester
}

type LDAPConnectionConfig struct {
	URL             string
	BaseDN          string
	BindDN          string
	BindPassword    string
	UserFilter      string
	UsernameAttr    string
	DisplayNameAttr string
	EmailAttr       string
}

type LDAPUser struct {
	Username    string
	DisplayName string
	Email       string
}

type LDAPConfigRequest struct {
	Enabled         bool   `json:"enabled"`
	URL             string `json:"url"`
	BaseDN          string `json:"baseDn"`
	BindDN          string `json:"bindDn"`
	BindPassword    string `json:"bindPassword"`
	UserFilter      string `json:"userFilter"`
	UsernameAttr    string `json:"usernameAttr"`
	DisplayNameAttr string `json:"displayNameAttr"`
	EmailAttr       string `json:"emailAttr"`
}

type LDAPConfigResult struct {
	Enabled         bool   `json:"enabled"`
	URL             string `json:"url"`
	BaseDN          string `json:"baseDn"`
	BindDN          string `json:"bindDn"`
	BindPassword    string `json:"bindPassword,omitempty"`
	HasBindPassword bool   `json:"hasBindPassword"`
	UserFilter      string `json:"userFilter"`
	UsernameAttr    string `json:"usernameAttr"`
	DisplayNameAttr string `json:"displayNameAttr"`
	EmailAttr       string `json:"emailAttr"`
}

type LDAPTestRequest struct {
	URL                  string `json:"url"`
	BaseDN               string `json:"baseDn"`
	BindDN               string `json:"bindDn"`
	BindPassword         string `json:"bindPassword"`
	UseSavedBindPassword bool   `json:"useSavedBindPassword"`
	UserFilter           string `json:"userFilter"`
	UsernameAttr         string `json:"usernameAttr"`
	DisplayNameAttr      string `json:"displayNameAttr"`
	EmailAttr            string `json:"emailAttr"`
}

type LDAPTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func NewLDAPService(store LDAPStore, client LDAPTester) *LDAPService {
	return &LDAPService{store: store, client: client}
}

func (s *LDAPService) GetConfig(ctx context.Context) (LDAPConfigResult, error) {
	cfg, err := s.store.GetLDAPConfig(ctx)
	if err != nil {
		return LDAPConfigResult{}, err
	}
	return ldapConfigResult(cfg), nil
}

func (s *LDAPService) SaveConfig(ctx context.Context, actorID uint, req LDAPConfigRequest, meta AuditMeta) (LDAPConfigResult, error) {
	current, _ := s.store.GetLDAPConfig(ctx)
	cfg := systemModel.SystemLDAPConfig{
		ID:              current.ID,
		Enabled:         req.Enabled,
		URL:             req.URL,
		BaseDN:          req.BaseDN,
		BindDN:          req.BindDN,
		BindPassword:    req.BindPassword,
		UserFilter:      req.UserFilter,
		UsernameAttr:    req.UsernameAttr,
		DisplayNameAttr: req.DisplayNameAttr,
		EmailAttr:       req.EmailAttr,
	}
	if cfg.BindPassword == "" {
		cfg.BindPassword = current.BindPassword
	}
	saved, err := s.store.SaveLDAPConfig(ctx, cfg)
	if err != nil {
		return LDAPConfigResult{}, err
	}
	_ = s.store.WriteAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "ldap.config.update", ResourceType: "ldap", Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return ldapConfigResult(saved), nil
}

func (s *LDAPService) TestConfig(ctx context.Context, actorID uint, req LDAPTestRequest, meta AuditMeta) (LDAPTestResult, error) {
	cfg, err := s.connectionConfig(ctx, req)
	if err != nil {
		_ = s.store.WriteAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "ldap.test.failure", ResourceType: "ldap", Result: "failure", IP: meta.IP, UserAgent: meta.UserAgent, Detail: err.Error()})
		return LDAPTestResult{}, err
	}
	if err := s.client.Test(ctx, cfg); err != nil {
		_ = s.store.WriteAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "ldap.test.failure", ResourceType: "ldap", Result: "failure", IP: meta.IP, UserAgent: meta.UserAgent, Detail: err.Error()})
		return LDAPTestResult{}, err
	}
	_ = s.store.WriteAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "ldap.test.success", ResourceType: "ldap", Result: "success", IP: meta.IP, UserAgent: meta.UserAgent})
	return LDAPTestResult{Success: true, Message: "LDAP connection test passed"}, nil
}

func (s *LDAPService) connectionConfig(ctx context.Context, req LDAPTestRequest) (LDAPConnectionConfig, error) {
	if !strings.Contains(req.UserFilter, "{username}") {
		return LDAPConnectionConfig{}, ErrInvalidLDAPConfig
	}
	password := req.BindPassword
	if req.UseSavedBindPassword {
		current, err := s.store.GetLDAPConfig(ctx)
		if err != nil {
			return LDAPConnectionConfig{}, err
		}
		password = current.BindPassword
	}
	if password == "" {
		return LDAPConnectionConfig{}, ErrInvalidLDAPConfig
	}
	return LDAPConnectionConfig{
		URL:             req.URL,
		BaseDN:          req.BaseDN,
		BindDN:          req.BindDN,
		BindPassword:    password,
		UserFilter:      req.UserFilter,
		UsernameAttr:    req.UsernameAttr,
		DisplayNameAttr: req.DisplayNameAttr,
		EmailAttr:       req.EmailAttr,
	}, nil
}

func ldapConfigResult(cfg systemModel.SystemLDAPConfig) LDAPConfigResult {
	return LDAPConfigResult{
		Enabled:         cfg.Enabled,
		URL:             cfg.URL,
		BaseDN:          cfg.BaseDN,
		BindDN:          cfg.BindDN,
		HasBindPassword: cfg.BindPassword != "",
		UserFilter:      cfg.UserFilter,
		UsernameAttr:    cfg.UsernameAttr,
		DisplayNameAttr: cfg.DisplayNameAttr,
		EmailAttr:       cfg.EmailAttr,
	}
}

func ldapConnectionConfig(cfg systemModel.SystemLDAPConfig) LDAPConnectionConfig {
	return LDAPConnectionConfig{
		URL:             cfg.URL,
		BaseDN:          cfg.BaseDN,
		BindDN:          cfg.BindDN,
		BindPassword:    cfg.BindPassword,
		UserFilter:      cfg.UserFilter,
		UsernameAttr:    cfg.UsernameAttr,
		DisplayNameAttr: cfg.DisplayNameAttr,
		EmailAttr:       cfg.EmailAttr,
	}
}
