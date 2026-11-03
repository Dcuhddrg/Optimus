package initialize

import (
	"optimus/server/config"
	systemService "optimus/server/service/system"
	"optimus/server/service/system/directory"

	"gorm.io/gorm"
)

func newSystemServiceGroup(db *gorm.DB, cfg config.Config) systemService.ServiceGroup {
	group := systemService.ServiceGroup{}
	if db == nil {
		return group
	}

	group.AuthService = *systemService.NewAuthService(systemService.JWTIssuer{Config: cfg.JWT}, directory.NewGoLDAPClient())
	group.LDAPService = *systemService.NewLDAPService(directory.NewGoLDAPClient())
	return group
}
