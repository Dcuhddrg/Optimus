package initialize

import (
	"gorm.io/gorm"
	"optimus/server/config"
	systemService "optimus/server/service/system"
	"optimus/server/service/system/directory"
	persistence "optimus/server/service/system/store"
)

func newSystemServiceGroup(db *gorm.DB, cfg config.Config) systemService.ServiceGroup {
	return systemService.ServiceGroup{
		AuthService:  newSystemAuthService(db, cfg),
		UserService:  newSystemUserService(db),
		RoleService:  newSystemRoleService(db),
		LDAPService:  newSystemLDAPService(db),
		AuditService: newSystemAuditService(db),
		RBAC:         newSystemRBACCapability(db),
		Status:       &systemService.Service{},
	}
}

func newSystemRBACCapability(db *gorm.DB) *systemService.RBACService {
	if db == nil {
		return nil
	}
	return systemService.NewRBACService(persistence.NewGormRBACStore(db))
}

func newSystemAuthService(db *gorm.DB, cfg config.Config) *systemService.AuthService {
	if db == nil {
		return nil
	}
	return systemService.NewAuthService(
		persistence.NewGormAuthStore(db),
		systemService.JWTIssuer{Config: cfg.JWT},
		directory.NewGoLDAPClient(),
	)
}

func newSystemUserService(db *gorm.DB) *systemService.UserService {
	if db == nil {
		return nil
	}
	return systemService.NewUserService(persistence.NewGormUserStore(db))
}

func newSystemRoleService(db *gorm.DB) *systemService.RoleService {
	if db == nil {
		return nil
	}
	return systemService.NewRoleService(persistence.NewGormRoleStore(db))
}

func newSystemLDAPService(db *gorm.DB) *systemService.LDAPService {
	if db == nil {
		return nil
	}
	return systemService.NewLDAPService(
		persistence.NewGormLDAPStore(db),
		directory.NewGoLDAPClient(),
	)
}

func newSystemAuditService(db *gorm.DB) *systemService.AuditService {
	if db == nil {
		return nil
	}
	return systemService.NewAuditService(persistence.NewGormAuditStore(db))
}
