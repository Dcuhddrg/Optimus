package project

import (
	projectApi "optimus/server/api/v1/project"
	"optimus/server/config"
	"optimus/server/service/project/infra/persistence"
	projectUsecase "optimus/server/service/project/usecase"

	"gorm.io/gorm"
)

type ProjectModule struct {
	Handlers *projectApi.ProjectHandlers
}

func NewProjectModule(db *gorm.DB, _ config.Config) *ProjectModule {
	if db == nil {
		return &ProjectModule{Handlers: projectApi.NewProjectHandlers(projectApi.ProjectHandlerDeps{})}
	}
	store := persistence.NewGormProjectStore(db)
	svc := projectUsecase.NewProjectService(store)
	return &ProjectModule{
		Handlers: projectApi.NewProjectHandlers(projectApi.ProjectHandlerDeps{Service: svc}),
	}
}
