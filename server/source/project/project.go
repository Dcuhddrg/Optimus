// Package project seeds the project module, as GVA's source packages.
package project

import (
	"context"
	"errors"

	projectModel "optimus/server/model/project"
	projectService "optimus/server/service/project"
	systemService "optimus/server/service/system"

	"gorm.io/gorm"
)

const initOrderProject = 300

type initProject struct{}

func init() {
	systemService.RegisterInit(initOrderProject, &initProject{})
}

func (i *initProject) InitializerName() string { return "project" }

func (i *initProject) MigrateTable(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, systemService.ErrMissingDBContext
	}
	return ctx, projectService.AutoMigrateProject(db)
}

func (i *initProject) TableCreated(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return false
	}
	return db.Migrator().HasTable(&projectModel.Project{})
}

func (i *initProject) DataInserted(ctx context.Context) bool { return false }

func (i *initProject) InitializeData(ctx context.Context) (context.Context, error) {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		return ctx, systemService.ErrMissingDBContext
	}
	return ctx, SeedProjectData(db)
}

func SeedProjectData(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	// Grant the admin role only permissions introduced by this run, so a
	// revoked permission stays revoked across restarts.
	created, err := systemService.SeedPermissionCatalog(db, projectService.DefaultProjectPermissions())
	if err != nil {
		return err
	}
	return systemService.GrantToAdminRole(db, created)
}
