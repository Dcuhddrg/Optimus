// Package project seeds the project module, as GVA's source packages.
package project

import (
	"context"
	"errors"

	projectModel "optimus/server/model/project"
	systemModel "optimus/server/model/system"
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
	for _, perm := range projectService.DefaultProjectPermissions() {
		p := perm
		if err := db.Where("code = ?", p.Code).FirstOrCreate(&p).Error; err != nil {
			return err
		}
	}
	return seedAdminProjectPermissions(db)
}

func seedAdminProjectPermissions(db *gorm.DB) error {
	var adminRole systemModel.SystemRole
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	for _, perm := range projectService.DefaultProjectPermissions() {
		var p systemModel.SystemPermission
		if err := db.Where("code = ?", perm.Code).First(&p).Error; err != nil {
			return err
		}
		if err := systemService.GrantPermissions(db, adminRole.ID, []systemModel.SystemPermission{p}); err != nil {
			return err
		}
	}
	return nil
}
