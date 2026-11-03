package initialize

import (
	"gorm.io/gorm"
	"optimus/server/config"
	"optimus/server/global"
	cloudInit "optimus/server/initialize/cloud"
	projectInit "optimus/server/initialize/project"
	"optimus/server/service"
)

// InitializeServices assembles the only production service graph before routes
// or background workers start. Assign module contents, never replace root or
// module addresses captured by API aliases.
func InitializeServices(db *gorm.DB, cfg config.Config) {
	global.OPTIMUS_DB = db
	service.ServiceGroupApp.SystemServiceGroup = newSystemServiceGroup(db, cfg)
	service.ServiceGroupApp.CloudServiceGroup = cloudInit.NewServiceGroup(db, cfg)
	service.ServiceGroupApp.ProjectServiceGroup = projectInit.NewServiceGroup(db, cfg)
}
