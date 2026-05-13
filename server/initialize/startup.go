package initialize

import (
	"context"

	"optimus/server/global"
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

	startAuditRetentionWorker(context.Background(), db, global.OPTIMUS_CONFIG.System.AuditLogRetentionDays)
}
