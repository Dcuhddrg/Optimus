package initialize

import (
	"context"
	"errors"
	"fmt"

	"optimus/server/global"
	"optimus/server/service"
	cloudService "optimus/server/service/cloud"
	projectService "optimus/server/service/project"
	systemService "optimus/server/service/system"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InitializeData migrates and seeds every module. A migration failure stops the
// remaining steps; seed failures are collected so later modules still run. Any
// returned error means the service is degraded and must report unhealthy.
func InitializeData(db *gorm.DB) error {
	if db == nil {
		return errors.New("database is not initialized")
	}

	var errs []error
	fail := func(component, action string, err error) {
		logDependencyInitFailure(component, action, err)
		errs = append(errs, fmt.Errorf("%s %s: %w", component, action, err))
	}

	if err := systemService.AutoMigrateSystem(db); err != nil {
		fail("system", "auto migrate", err)
		return errors.Join(errs...)
	}
	if err := systemService.SeedSystemData(db, global.OPTIMUS_CONFIG.System.AdminInitialPassword); err != nil {
		fail("system", "seed data", err)
	}

	if err := cloudService.AutoMigrateCloud(db); err != nil {
		fail("cloud", "auto migrate", err)
		return errors.Join(errs...)
	}
	if err := cloudService.SeedCloudData(db); err != nil {
		fail("cloud", "seed data", err)
	}

	if err := projectService.AutoMigrateProject(db); err != nil {
		fail("project", "auto migrate", err)
		return errors.Join(errs...)
	}
	if err := projectService.SeedProjectData(db); err != nil {
		fail("project", "seed data", err)
	}
	if err := cloudService.BackfillCloudProjectScope(db); err != nil {
		fail("cloud", "backfill project scope", err)
	}

	startAuditRetentionWorker(context.Background(), &service.ServiceGroupApp.SystemServiceGroup.AuditService, global.OPTIMUS_CONFIG.System.AuditLogRetentionDays)
	return errors.Join(errs...)
}

func logDependencyInitFailure(component, action string, err error) {
	logger, ok := global.OPTIMUS_LOG.(*zap.Logger)
	if !ok || logger == nil {
		return
	}

	logger.WithOptions(zap.AddCallerSkip(1)).Error(component+" init failed", zap.String("action", action), zap.Error(err))
}
