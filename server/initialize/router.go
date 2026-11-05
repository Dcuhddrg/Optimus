package initialize

import (
	"optimus/server/docs"
	"optimus/server/global"
	"optimus/server/middleware"
	"optimus/server/router"
	"optimus/server/service"
	"optimus/server/utils"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Routers() *gin.Engine {
	engine := gin.New()
	// 全局中间件
	engine.Use(middleware.Logger(), middleware.Recovery())
	// Swagger 文档，与 GVA 一致在所有模式下注册
	docs.SwaggerInfo.BasePath = routerPrefix()
	engine.GET(routerPrefix()+"/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// 公共路由
	publicGroup := engine.Group(routerPrefix())
	// 私有路由
	privateGroup := engine.Group(routerPrefix())
	privateGroup.Use(middleware.JWTAuth(), middleware.CasbinHandler())
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
	return utils.RouterPrefix()
}
