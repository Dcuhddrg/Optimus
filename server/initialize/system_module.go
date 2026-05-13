package initialize

import (
	systemApi "optimus/server/api/v1/system"
	"optimus/server/config"
	"optimus/server/middleware"
	systemService "optimus/server/service/system"
	"optimus/server/service/system/infra/directory"
	"optimus/server/service/system/infra/identity"
	"optimus/server/service/system/infra/persistence"
	systemUsecase "optimus/server/service/system/usecase"

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

func newSystemRBACCapability(db *gorm.DB) *systemUsecase.RBACService {
	if db == nil {
		return nil
	}
	return systemUsecase.NewRBACService(persistence.NewGormRBACStore(db))
}

func newSystemAuthService(db *gorm.DB, cfg config.Config) *systemUsecase.AuthService {
	if db == nil {
		return nil
	}
	return systemUsecase.NewAuthService(
		persistence.NewGormAuthStore(db),
		identity.JWTIssuer{Config: cfg.JWT},
		directory.NewGoLDAPClient(),
	)
}

func newSystemUserService(db *gorm.DB) *systemUsecase.UserService {
	if db == nil {
		return nil
	}
	return systemUsecase.NewUserService(persistence.NewGormUserStore(db))
}

func newSystemRoleService(db *gorm.DB) *systemUsecase.RoleService {
	if db == nil {
		return nil
	}
	return systemUsecase.NewRoleService(persistence.NewGormRoleStore(db))
}

func newSystemLDAPService(db *gorm.DB) *systemUsecase.LDAPService {
	if db == nil {
		return nil
	}
	return systemUsecase.NewLDAPService(
		persistence.NewGormLDAPStore(db),
		directory.NewGoLDAPClient(),
	)
}

func newSystemAuditService(db *gorm.DB) *systemUsecase.AuditService {
	if db == nil {
		return nil
	}
	return systemUsecase.NewAuditService(persistence.NewGormAuditStore(db))
}
