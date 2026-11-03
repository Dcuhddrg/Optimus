package project

import (
	"optimus/server/config"
	projectService "optimus/server/service/project"

	"gorm.io/gorm"
)

func NewServiceGroup(db *gorm.DB, _ config.Config) projectService.ServiceGroup {
	if db == nil {
		return projectService.ServiceGroup{}
	}
	return projectService.ServiceGroup{ProjectService: &projectService.ProjectService{}}
}
