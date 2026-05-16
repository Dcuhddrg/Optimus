package router

import (
	"optimus/server/router/cloud"
	"optimus/server/router/project"
	"optimus/server/router/system"
)

type RouterGroup struct {
	System  system.RouterGroup
	Cloud   cloud.RouterGroup
	Project project.RouterGroup
}

var RouterGroupApp = new(RouterGroup)
