package initialize

import (
	systemApi "optimus/server/api/v1/system"
	"optimus/server/config"
	systemService "optimus/server/service/system"

	"gorm.io/gorm"
)

type SystemCapabilities struct {
	RBAC *systemService.RBACService
}

type SystemModule struct {
	Handlers     *systemApi.SystemHandlers
	Capabilities *SystemCapabilities
	cfg          config.Config
}

func NewSystemModule(db *gorm.DB, cfg config.Config) *SystemModule {
	return &SystemModule{
		Handlers: systemApi.NewSystemHandlers(),
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
