package system

import (
	"context"
	"errors"
	"fmt"

	"optimus/server/global"
	systemModel "optimus/server/model/system"
	systemRequest "optimus/server/model/system/request"
	systemResponse "optimus/server/model/system/response"
	ldapUtil "optimus/server/utils/ldap"

	"gorm.io/gorm"
)

var (
	ErrInvalidLDAPConfig = errors.New("invalid ldap config")
	// ErrInvalidIdentityAttr rejects an id_attr outside the whitelist.
	ErrInvalidIdentityAttr = errors.New("identity attribute must be one of entryUUID, nsUniqueId, ipaUniqueID")
)

// IdentityAttrLockedError rejects changing id_attr while accounts are bound:
// their identities would be looked up under the wrong attribute.
type IdentityAttrLockedError struct{ Bound int64 }

func (e IdentityAttrLockedError) Error() string {
	return fmt.Sprintf("%d accounts are bound to a directory identity; the identity attribute cannot be changed", e.Bound)
}

const (
	defaultIDAttr         = "entryUUID"
	defaultDepartmentAttr = "departmentNumber"
)

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
	bound, err := countBoundUsers(global.OPTIMUS_DB.WithContext(ctx))
	if err != nil {
		return systemResponse.LDAPConfigResult{}, err
	}
	if cfg.ID == 0 {
		// Nothing saved yet: return the design defaults, so a first save
		// does not turn an absent value into an explicit "not synced".
		cfg.IDAttr = defaultIDAttr
		cfg.DepartmentAttr = defaultDepartmentAttr
	}
	result := ldapConfigResult(cfg)
	result.BoundUsers = bound
	return result, nil
}

func (s *LDAPService) SaveConfig(ctx context.Context, actorID uint, req systemRequest.LDAPConfigRequest, meta AuditMeta) (systemResponse.LDAPConfigResult, error) {
	event := failedAudit(actorID, meta, "ldap.config.update", "ldap", 0)
	defer func() { recordAudit(ctx, event) }()
	// Logins bind identities under this lock; holding it keeps the bound count
	// and the id_attr change consistent.
	ldapWriteBack.Lock()
	defer ldapWriteBack.Unlock()

	// The new row is built from the current one (ID, version, sync key,
	// saved password); without it a save would insert a second row that
	// readers never pick, so a failed read refuses the save.
	current, err := s.getLDAPConfig(ctx)
	if err != nil {
		event.Detail = auditDetails("update_failed", nil)
		return systemResponse.LDAPConfigResult{}, err
	}
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
		IDAttr:          req.IDAttr,
		DepartmentAttr:  current.DepartmentAttr,
		Version:         current.Version + 1,
		LastSyncKey:     current.LastSyncKey,
	}
	if cfg.BindPassword == "" {
		cfg.BindPassword = current.BindPassword
	}
	if cfg.IDAttr == "" {
		cfg.IDAttr = current.IDAttr
	}
	if cfg.IDAttr == "" {
		cfg.IDAttr = defaultIDAttr
	}
	if req.DepartmentAttr != nil {
		cfg.DepartmentAttr = *req.DepartmentAttr
	} else if current.ID == 0 {
		cfg.DepartmentAttr = defaultDepartmentAttr
	}
	if err := validateLDAPConfig(cfg); err != nil {
		event.Detail = auditDetails("invalid_config", nil)
		return systemResponse.LDAPConfigResult{}, err
	}
	bound, err := countBoundUsers(global.OPTIMUS_DB.WithContext(ctx))
	if err != nil {
		event.Detail = auditDetails("update_failed", nil)
		return systemResponse.LDAPConfigResult{}, err
	}
	if bound > 0 && current.IDAttr != cfg.IDAttr {
		event.Detail = auditDetailsMap("identity_attr_locked", map[string]any{"boundUsers": bound})
		return systemResponse.LDAPConfigResult{}, IdentityAttrLockedError{Bound: bound}
	}
	saved, err := s.saveLDAPConfig(ctx, cfg)
	if err != nil {
		event.Detail = auditDetails("update_failed", nil)
		return systemResponse.LDAPConfigResult{}, err
	}
	event.ResourceID = resourceID(saved.ID)
	event.Result = "success"
	event.Detail = auditDetails("updated", map[string]auditChange{
		"enabled":             {Before: current.Enabled, After: saved.Enabled},
		"connectionChanged":   {Before: false, After: current.URL != saved.URL || current.BaseDN != saved.BaseDN || current.BindDN != saved.BindDN},
		"mappingChanged":      {Before: false, After: current.UserFilter != saved.UserFilter || current.UsernameAttr != saved.UsernameAttr || current.DisplayNameAttr != saved.DisplayNameAttr || current.EmailAttr != saved.EmailAttr || current.DepartmentAttr != saved.DepartmentAttr},
		"idAttr":              {Before: current.IDAttr, After: saved.IDAttr},
		"bindPasswordChanged": {Before: false, After: current.BindPassword != saved.BindPassword},
	})
	result := ldapConfigResult(saved)
	result.BoundUsers = bound
	return result, nil
}

// validateLDAPConfig checks the identity attribute and, for an enabled or
// filled-in config, the login filter structure (design 3.3 D3).
func validateLDAPConfig(cfg systemModel.SystemLDAPConfig) error {
	if !ldapUtil.IsIdentityAttribute(cfg.IDAttr) {
		return ErrInvalidIdentityAttr
	}
	if !cfg.Enabled && cfg.UserFilter == "" {
		return nil
	}
	_, err := ldapUtil.ValidateLoginFilter(cfg.UserFilter, cfg.UsernameAttr)
	return err
}

func countBoundUsers(db *gorm.DB) (int64, error) {
	var bound int64
	err := db.Model(&systemModel.SystemUser{}).Where("external_id IS NOT NULL").Count(&bound).Error
	return bound, err
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
		reason := "connection_test_failed"
		switch {
		case errors.Is(err, ldapUtil.ErrNoUsersInScope):
			reason = "no_users_in_scope"
		case errors.Is(err, ldapUtil.ErrSampleMissingIdentity):
			reason = "identity_unreadable"
		}
		event := failedAudit(actorID, meta, "ldap.test.failure", "ldap", 0)
		event.Detail = auditDetails(reason, nil)
		recordAudit(ctx, event)
		return systemResponse.LDAPTestResult{}, err
	}
	recordAudit(ctx, AuditEvent{ActorUserID: actorID, ActorUsername: meta.ActorUsername, Action: "ldap.test.success", ResourceType: "ldap", Result: "success", IP: meta.IP, UserAgent: meta.UserAgent, Detail: auditDetails("connection_test_passed", nil)})
	return systemResponse.LDAPTestResult{Success: true, Message: "LDAP connection test passed"}, nil
}

func (s *LDAPService) connectionConfig(ctx context.Context, req systemRequest.LDAPTestRequest) (LDAPConnectionConfig, error) {
	idAttr := req.IDAttr
	if idAttr == "" {
		idAttr = defaultIDAttr
	}
	if !ldapUtil.IsIdentityAttribute(idAttr) {
		return LDAPConnectionConfig{}, ErrInvalidIdentityAttr
	}
	if _, err := ldapUtil.ValidateLoginFilter(req.UserFilter, req.UsernameAttr); err != nil {
		return LDAPConnectionConfig{}, err
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
		DepartmentAttr:  req.DepartmentAttr,
		IDAttr:          idAttr,
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
		IDAttr:          cfg.IDAttr,
		DepartmentAttr:  cfg.DepartmentAttr,
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
		DepartmentAttr:  cfg.DepartmentAttr,
		IDAttr:          cfg.IDAttr,
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
		department := cfg.DepartmentAttr
		if err := global.OPTIMUS_DB.WithContext(ctx).Create(&cfg).Error; err != nil {
			return cfg, err
		}
		// Create skips a zero value for a column with a default; keep an
		// explicit empty department_attr ("not synced").
		if department == "" {
			cfg.DepartmentAttr = ""
			err := global.OPTIMUS_DB.WithContext(ctx).Model(&cfg).Update("department_attr", "").Error
			return cfg, err
		}
		return cfg, nil
	}
	err := global.OPTIMUS_DB.WithContext(ctx).Omit("CreatedAt").Save(&cfg).Error
	return cfg, err
}
