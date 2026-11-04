package initialize

import (
	"strings"

	"optimus/server/global"
	"optimus/server/middleware"
	"optimus/server/router"
	"optimus/server/service"

	"github.com/gin-gonic/gin"
)

const defaultRouterPrefix = "/api/v1"

func Routers() *gin.Engine {
	engine := gin.New()
	// 全局中间件
	engine.Use(middleware.Logger(), middleware.Recovery())
	// 公共路由
	publicGroup := engine.Group(routerPrefix())
	// 私有路由
	privateGroup := engine.Group(routerPrefix())
	privateGroup.Use(middleware.JWTAuth(), middleware.CasbinHandlerWithPrefix(
		&service.ServiceGroupApp.SystemServiceGroup.RBACService, privateGroup.BasePath(), defaultRouterPrefix,
	))
	cloudGroup := privateGroup.Group("")
	if global.OPTIMUS_DB != nil {
		cloudGroup.Use(middleware.ProjectScope(&service.ServiceGroupApp.ProjectServiceGroup.ProjectService))
	}

	routerGroup := router.RouterGroupApp
	systemRouter := &routerGroup.System
	systemRouter.InitBaseRouter(publicGroup)
	systemRouter.InitAuthRouter(publicGroup, privateGroup)
	systemRouter.InitUserRouter(privateGroup)
	systemRouter.InitRoleRouter(privateGroup)
	systemRouter.InitLDAPRouter(privateGroup)
	systemRouter.InitAuditRouter(privateGroup)

	cloudRouter := &routerGroup.Cloud
	cloudRouter.InitAccountRouter(cloudGroup)
	cloudRouter.InitAssetRouter(cloudGroup)

	projectRouter := &routerGroup.Project
	projectRouter.InitProjectRouter(privateGroup)
	projectRouter.InitMemberRouter(privateGroup)

	return engine
}

func routerPrefix() string {
	prefix := strings.TrimSpace(global.OPTIMUS_CONFIG.System.RouterPrefix)
	if prefix == "" {
		return defaultRouterPrefix
	}

	if strings.HasPrefix(prefix, "/") {
		return prefix
	}

	return "/" + prefix
}
