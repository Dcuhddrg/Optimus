package initialize

import (
	systemApi "optimus/server/api/v1/system"
	"optimus/server/config"
	"optimus/server/middleware"
	systemService "optimus/server/service/system"

	"gorm.io/gorm"
)

type SystemCapabilities struct {
	RBAC middleware.APIAuthorizer
}

type SystemModule struct {
	Handlers     *systemApi.SystemHandlers
	Capabilities *SystemCapabilities
	cfg          config.Config
}

func NewSystemModule(db *gorm.DB, cfg config.Config) *SystemModule {
	authService := newSystemAuthService(db, cfg)
	userService := newSystemUserService(db)
	roleService := newSystemRoleService(db)
	ldapService := newSystemLDAPService(db)
	auditService := newSystemAuditService(db)

	return &SystemModule{
		Handlers: systemApi.NewSystemHandlers(systemApi.SystemHandlerDeps{
			Auth:   authService,
			User:   userService,
			Role:   roleService,
			LDAP:   ldapService,
			Audit:  auditService,
			System: &systemService.Service{},
		}),
		Capabilities: &SystemCapabilities{
			RBAC: newSystemRBACCapability(db),
		},
		cfg: cfg,
	}
}

func newSystemRBACCapability(db *gorm.DB) *systemService.RBACService {
	if db == nil {
		return nil
	}

	return systemService.NewRBACService(systemService.NewGormRBACStore(db))
}

func newSystemAuthService(db *gorm.DB, cfg config.Config) *systemService.AuthService {
	if db == nil {
		return nil
	}

	return systemService.NewAuthServiceWithLDAP(
		systemService.NewGormAuthStore(db),
		cfg.JWT,
		systemService.NewGoLDAPClient(),
	)
}

func newSystemUserService(db *gorm.DB) *systemService.UserService {
	if db == nil {
		return nil
	}

	return systemService.NewUserService(systemService.NewGormUserStore(db))
}

func newSystemRoleService(db *gorm.DB) *systemService.RoleService {
	if db == nil {
		return nil
	}

	return systemService.NewRoleService(systemService.NewGormRoleStore(db))
}

func newSystemLDAPService(db *gorm.DB) *systemService.LDAPService {
	if db == nil {
		return nil
	}

	return systemService.NewLDAPService(
		systemService.NewGormLDAPStore(db),
		systemService.NewGoLDAPClient(),
	)
}

func newSystemAuditService(db *gorm.DB) *systemService.AuditService {
	if db == nil {
		return nil
	}

	return systemService.NewAuditService(systemService.NewGormAuditStore(db))
}
