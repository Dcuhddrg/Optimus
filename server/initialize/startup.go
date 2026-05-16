package initialize

import (
	"context"

	"optimus/server/global"
	cloudService "optimus/server/service/cloud"
	projectService "optimus/server/service/project"
	systemService "optimus/server/service/system"
)

func Startup() {
	db := Gorm()
	if db == nil {
		return
	}

	Redis()
	Casbin()

	if err := systemService.AutoMigrateSystem(db); err != nil {
		logDependencyInitFailure("system", "auto migrate", err)
		return
	}
	if err := systemService.SeedSystemData(db, global.OPTIMUS_CONFIG.System.AdminInitialPassword); err != nil {
		logDependencyInitFailure("system", "seed data", err)
	}

	if err := cloudService.AutoMigrateCloud(db); err != nil {
		logDependencyInitFailure("cloud", "auto migrate", err)
		return
	}
	if err := cloudService.SeedCloudData(db); err != nil {
		logDependencyInitFailure("cloud", "seed data", err)
	}

	if err := projectService.AutoMigrateProject(db); err != nil {
		logDependencyInitFailure("project", "auto migrate", err)
		return
	}
	if err := projectService.SeedProjectData(db); err != nil {
		logDependencyInitFailure("project", "seed data", err)
	}

	startAuditRetentionWorker(context.Background(), db, global.OPTIMUS_CONFIG.System.AuditLogRetentionDays)
}
