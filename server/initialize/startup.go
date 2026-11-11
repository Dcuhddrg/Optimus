package initialize

import (
	"context"
	"errors"

	"optimus/server/global"
	systemService "optimus/server/service/system"

	// Seed initializers register themselves in init(), as GVA's source packages.
	_ "optimus/server/source/cloud"
	_ "optimus/server/source/project"
	_ "optimus/server/source/system"

	"gorm.io/gorm"
)

// InitializeData migrates and seeds every module through the registered source
// initializers. Any returned error means the service is degraded and must
// report unhealthy (#51).
func InitializeData(db *gorm.DB) error {
	if db == nil {
		return errors.New("database is not initialized")
	}
	ctx := context.WithValue(context.Background(), "db", db)
	ctx = context.WithValue(ctx, "adminPassword", global.OPTIMUS_CONFIG.System.AdminInitialPassword)
	if err := systemService.InitDBServiceApp.InitDB(ctx); err != nil {
		return err
	}
	// Load the Casbin policy before serving. A failure is not fatal: requests
	// needing API permissions fail closed while a background retry runs.
	_ = systemService.CasbinServiceApp.FreshCasbin()
	return nil
}
