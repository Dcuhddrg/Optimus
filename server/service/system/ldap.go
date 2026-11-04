package system

import (
	"context"
	"errors"
	"strings"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"
	systemResponse "optimus/server/model/system/response"
	ldapUtil "optimus/server/utils/ldap"

	"gorm.io/gorm"
)

var ErrInvalidLDAPConfig = errors.New("invalid ldap config")

type LDAPService struct{}

type (
	LDAPConnectionConfig = ldapUtil.Config
	LDAPUser             = ldapUtil.User
)

// newLDAPClient builds the directory client per call; tests replace it.
var newLDAPClient = ldapUtil.NewLDAP

func (s *LDAPService) GetConfig(ctx context.Context) (systemResponse.LDAPConfigResult, error) {
	cfg, err := s.getLDAPConfig(ctx)
	if err != nil {
		return systemResponse.LDAPConfigResult{}, err
	}
	return ldapConfigResult(cfg), nil
}

func (s *LDAPService) SaveConfig(ctx context.Context, actorID uint, req systemRequest.LDAPConfigRequest, meta AuditMeta) (systemResponse.LDAPConfigResult, error) {
	event := failedAudit(actorID, meta, "ldap.config.update", "ldap", 0)
	defer func() { recordAudit(ctx, event) }()
	current, currentErr := s.getLDAPConfig(ctx)
	cfg := systemModel.SystemLDAPConfig{
		OPTIMUS_MODEL:   global.OPTIMUS_MODEL{ID: current.ID},
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
	saved, err := s.saveLDAPConfig(ctx, cfg)
	if err != nil {
		event.Detail = auditDetails("update_failed", nil)
		return systemResponse.LDAPConfigResult{}, err
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

func (s *LDAPService) TestConfig(ctx context.Context, actorID uint, req systemRequest.LDAPTestRequest, meta AuditMeta) (systemResponse.LDAPTestResult, error) {
	cfg, err := s.connectionConfig(ctx, req)
	if err != nil {
		event := failedAudit(actorID, meta, "ldap.test.failure", "ldap", 0)
		event.Detail = auditDetails("invalid_config", nil)
		recordAudit(ctx, event)
		return systemResponse.LDAPTestResult{}, err
	}
	if err := newLDAPClient().Test(ctx, cfg); err != nil {
		event := failedAudit(actorID, meta, "ldap.test.failure", "ldap", 0)
		event.Detail = auditDetails("connection_test_failed", nil)
		recordAudit(ctx, event)
		return systemResponse.LDAPTestResult{}, err
	}
	recordAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "ldap.test.success", ResourceType: "ldap", Result: "success", IP: meta.IP, UserAgent: meta.UserAgent, Detail: auditDetails("connection_test_passed", nil)})
	return systemResponse.LDAPTestResult{Success: true, Message: "LDAP connection test passed"}, nil
}

func (s *LDAPService) connectionConfig(ctx context.Context, req systemRequest.LDAPTestRequest) (LDAPConnectionConfig, error) {
	if !strings.Contains(req.UserFilter, "{username}") {
		return LDAPConnectionConfig{}, ErrInvalidLDAPConfig
	}
	password := req.BindPassword
	if req.UseSavedBindPassword {
		current, err := s.getLDAPConfig(ctx)
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

func ldapConfigResult(cfg systemModel.SystemLDAPConfig) systemResponse.LDAPConfigResult {
	return systemResponse.LDAPConfigResult{
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

func (s *LDAPService) getLDAPConfig(ctx context.Context) (systemModel.SystemLDAPConfig, error) {
	var cfg systemModel.SystemLDAPConfig
	err := global.OPTIMUS_DB.WithContext(ctx).Order("id ASC").First(&cfg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return systemModel.SystemLDAPConfig{}, nil
	}
	return cfg, err
}

func (s *LDAPService) saveLDAPConfig(ctx context.Context, cfg systemModel.SystemLDAPConfig) (systemModel.SystemLDAPConfig, error) {
	if cfg.ID == 0 {
		err := global.OPTIMUS_DB.WithContext(ctx).Create(&cfg).Error
		return cfg, err
	}
	err := global.OPTIMUS_DB.WithContext(ctx).Omit("CreatedAt").Save(&cfg).Error
	return cfg, err
}
