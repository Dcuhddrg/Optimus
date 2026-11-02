package project

import (
	"gorm.io/gorm"
	"optimus/server/config"
	projectService "optimus/server/service/project"
	projectStore "optimus/server/service/project/store"
)

func NewServiceGroup(db *gorm.DB, _ config.Config) projectService.ServiceGroup {
	if db == nil {
		return projectService.ServiceGroup{}
	}
	store := projectStore.NewGormProjectStore(db)
	return projectService.ServiceGroup{ProjectService: projectService.NewProjectService(store)}
}
