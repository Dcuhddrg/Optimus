package project

import (
	projectApi "optimus/server/api/v1/project"
	"optimus/server/config"
	projectService "optimus/server/service/project"
	projectStore "optimus/server/service/project/store"

	"gorm.io/gorm"
)

type ProjectModule struct {
	Handlers *projectApi.ProjectHandlers
	Service  *projectService.ProjectService
}

func NewProjectModule(db *gorm.DB, _ config.Config) *ProjectModule {
	if db == nil {
		return &ProjectModule{Handlers: projectApi.NewProjectHandlers(projectApi.ProjectHandlerDeps{})}
	}
	store := projectStore.NewGormProjectStore(db)
	svc := projectService.NewProjectService(store)
	return &ProjectModule{
		Handlers: projectApi.NewProjectHandlers(projectApi.ProjectHandlerDeps{Service: svc}),
		Service:  svc,
	}
}
