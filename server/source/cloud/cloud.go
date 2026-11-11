// Package cloud seeds the cloud module, as GVA's source packages.
package cloud

import (
	"context"
	"errors"
	"fmt"

	cloudModel "optimus/server/model/cloud"
	projectModel "optimus/server/model/project"
	systemModel "optimus/server/model/system"
	cloudService "optimus/server/service/cloud"
	projectService "optimus/server/service/project"
	systemService "optimus/server/service/system"

	"gorm.io/gorm"
)

const (
	initOrderCloud             = 200
	initOrderCloudProjectScope = 400 // after project tables and data
	defaultProjectCode         = "default"
)

type initCloud struct{}

// initCloudProjectScope moves legacy cloud accounts into the default project.
type initCloudProjectScope struct{}

func init() {
	systemService.RegisterInit(initOrderCloud, &initCloud{})
	systemService.RegisterInit(initOrderCloudProjectScope, &initCloudProjectScope{})
}

func (i *initCloud) InitializerName() string { return "cloud" }

func (i *initCloud) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, systemService.ErrMissingDBContext
	}
	return ctx, cloudService.AutoMigrateCloud(db)
}

func (i *initCloud) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return db.Migrator().HasTable(&cloudModel.CloudAccount{})
}

func (i *initCloud) DataInserted(ctx context.Context) bool { return false }

func (i *initCloud) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, systemService.ErrMissingDBContext
	}
	return ctx, SeedCloudData(db)
}

func (i *initCloudProjectScope) InitializerName() string { return "cloud project scope" }

func (i *initCloudProjectScope) MigrateTable(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (i *initCloudProjectScope) TableCreated(ctx context.Context) bool { return true }

func (i *initCloudProjectScope) DataInserted(ctx context.Context) bool { return false }

func (i *initCloudProjectScope) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, systemService.ErrMissingDBContext
	}
	return ctx, BackfillCloudProjectScope(db)
}

func SeedCloudData(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	// Grant the admin role only permissions introduced by this run, so a
	// revoked permission stays revoked across restarts.
	created, err := systemService.SeedPermissionCatalog(db, cloudService.DefaultCloudPermissions())
	if err != nil {
		return err
	}
	return systemService.GrantToAdminRole(db, created)
}

// BackfillCloudProjectScope migrates legacy cloud accounts into the default project.
func BackfillCloudProjectScope(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	defaultProject, err := ensureDefaultProject(db)
	if err != nil {
		return err
	}
	return db.Model(&cloudModel.CloudAccount{}).
		Where("project_id = 0").
		Update("project_id", defaultProject.ID).Error
}

func ensureDefaultProject(db *gorm.DB) (projectModel.Project, error) {
	var project projectModel.Project
	err := db.Where("code = ?", defaultProjectCode).First(&project).Error
	if err == nil {
		return project, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return projectModel.Project{}, err
	}

	var admin systemModel.SystemUser
	if err := db.Where("username = ?", "admin").First(&admin).Error; err != nil {
		return projectModel.Project{}, fmt.Errorf("cannot find admin user to own default project: %w", err)
	}

	project = projectModel.Project{
		Name:        "Default",
		Code:        defaultProjectCode,
		Description: "Default project for legacy cloud accounts",
		OwnerID:     admin.ID,
		Status:      "active",
	}
	if err := db.Create(&project).Error; err != nil {
		return projectModel.Project{}, err
	}

	member := projectModel.UserProject{
		UserID:      admin.ID,
		ProjectID:   project.ID,
		ProjectRole: projectService.RoleOwner,
	}
	if err := db.Create(&member).Error; err != nil {
		return projectModel.Project{}, err
	}
	return project, nil
}
