package project

import (
	"errors"

	projectModel "optimus/server/model/project"
	systemModel "optimus/server/model/system"
	systemService "optimus/server/service/system"

	"gorm.io/gorm"
)

func AutoMigrateProject(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	return db.AutoMigrate(
		&projectModel.Project{},
		&projectModel.UserProject{},
	)
}

func DefaultProjectPermissions() []systemModel.SystemPermission {
	return []systemModel.SystemPermission{
		{Code: "project.menu", Name: "Projects", Type: "menu", Path: "/projects", Sort: 30},
		{Code: "project.view", Name: "View projects", Type: "api", Method: "GET", Path: "/projects", Sort: 700},
		{Code: "project.detail", Name: "View project detail", Type: "api", Method: "GET", Path: "/projects/:code", Sort: 701},
		{Code: "project.create", Name: "Create project", Type: "api", Method: "POST", Path: "/projects", Sort: 702},
		{Code: "project.update", Name: "Update project", Type: "api", Method: "PUT", Path: "/projects/:code", Sort: 703},
		{Code: "project.delete", Name: "Delete project", Type: "api", Method: "DELETE", Path: "/projects/:code", Sort: 704},
		{Code: "project.members.view", Name: "View project members", Type: "api", Method: "GET", Path: "/projects/:code/members", Sort: 710},
		{Code: "project.members.add", Name: "Add project member", Type: "api", Method: "POST", Path: "/projects/:code/members", Sort: 711},
		{Code: "project.members.update", Name: "Update project member role", Type: "api", Method: "PATCH", Path: "/projects/:code/members/:userId", Sort: 712},
		{Code: "project.members.remove", Name: "Remove project member", Type: "api", Method: "DELETE", Path: "/projects/:code/members/:userId", Sort: 713},
	}
}

func SeedProjectData(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}
	for _, perm := range DefaultProjectPermissions() {
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
	for _, perm := range DefaultProjectPermissions() {
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
