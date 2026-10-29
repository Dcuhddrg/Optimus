package system

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

func NewLDAPService(store LDAPStore, client LDAPTester) *LDAPService {
	return &LDAPService{store: store, client: client}
}

func (s *LDAPService) GetConfig(ctx context.Context) (systemModel.LDAPConfigResult, error) {
	cfg, err := s.store.GetLDAPConfig(ctx)
	if err != nil {
		return systemModel.LDAPConfigResult{}, err
	}
	return ldapConfigResult(cfg), nil
}

func (s *LDAPService) SaveConfig(ctx context.Context, actorID uint, req systemModel.LDAPConfigRequest, meta AuditMeta) (systemModel.LDAPConfigResult, error) {
	event := failedAudit(actorID, meta, "ldap.config.update", "ldap", 0)
	defer func() { recordAudit(ctx, s.store, event) }()
	current, currentErr := s.store.GetLDAPConfig(ctx)
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
		event.Detail = auditDetails("update_failed", nil)
		return systemModel.LDAPConfigResult{}, err
	}
	event.ResourceID = resourceID(saved.ID)
	event.Result = "success"
	if currentErr != nil {
		event.Detail = auditDetails("snapshot_unavailable", nil)
	} else {
		event.Detail = auditDetails("updated", map[string]auditChange{
			"enabled":             {Before: current.Enabled, After: saved.Enabled},
			"connectionChanged":   {Before: false, After: current.URL != saved.URL || current.BaseDN != saved.BaseDN || current.BindDN != saved.BindDN},
			"mappingChanged":      {Before: false, After: current.UserFilter != saved.UserFilter || current.UsernameAttr != saved.UsernameAttr || current.DisplayNameAttr != saved.DisplayNameAttr || current.EmailAttr != saved.EmailAttr},
			"bindPasswordChanged": {Before: false, After: current.BindPassword != saved.BindPassword},
		})
	}
	return ldapConfigResult(saved), nil
}

func (s *LDAPService) TestConfig(ctx context.Context, actorID uint, req systemModel.LDAPTestRequest, meta AuditMeta) (systemModel.LDAPTestResult, error) {
	cfg, err := s.connectionConfig(ctx, req)
	if err != nil {
		event := failedAudit(actorID, meta, "ldap.test.failure", "ldap", 0)
		event.Detail = auditDetails("invalid_config", nil)
		recordAudit(ctx, s.store, event)
		return systemModel.LDAPTestResult{}, err
	}
	if err := s.client.Test(ctx, cfg); err != nil {
		event := failedAudit(actorID, meta, "ldap.test.failure", "ldap", 0)
		event.Detail = auditDetails("connection_test_failed", nil)
		recordAudit(ctx, s.store, event)
		return systemModel.LDAPTestResult{}, err
	}
	recordAudit(ctx, s.store, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "ldap.test.success", ResourceType: "ldap", Result: "success", IP: meta.IP, UserAgent: meta.UserAgent, Detail: auditDetails("connection_test_passed", nil)})
	return systemModel.LDAPTestResult{Success: true, Message: "LDAP connection test passed"}, nil
}

func (s *LDAPService) connectionConfig(ctx context.Context, req systemModel.LDAPTestRequest) (LDAPConnectionConfig, error) {
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

func ldapConfigResult(cfg systemModel.SystemLDAPConfig) systemModel.LDAPConfigResult {
	return systemModel.LDAPConfigResult{
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
