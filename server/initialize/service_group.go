package initialize

import (
	"gorm.io/gorm"
	cloudAPI "optimus/server/api/v1/cloud"
	systemAPI "optimus/server/api/v1/system"
	"optimus/server/config"
	cloudInit "optimus/server/initialize/cloud"
	projectInit "optimus/server/initialize/project"
	"optimus/server/service"
)

// InitializeServices assembles the only production service graph before routes
// or background workers start. Publish value aliases after all module services
// have been assembled so they contain their initialized dependencies.
func InitializeServices(db *gorm.DB, cfg config.Config) {
	service.ServiceGroupApp.SystemServiceGroup = newSystemServiceGroup(db, cfg)
	service.ServiceGroupApp.CloudServiceGroup = cloudInit.NewServiceGroup(db, cfg)
	service.ServiceGroupApp.ProjectServiceGroup = projectInit.NewServiceGroup(db, cfg)
	systemAPI.RefreshServiceAliases()
	cloudAPI.RefreshServiceAliases()
}
