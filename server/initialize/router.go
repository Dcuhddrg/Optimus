package initialize

import (
	"optimus/server/model/common/response"
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
	publicGroup.Use(response.EnvelopeForRoutes(publicGroup.BasePath(), []response.Route{{Method: "POST", Path: "/system/auth/login"}}))
	// 私有路由
	privateGroup := engine.Group(routerPrefix())
	privateGroup.Use(response.EnvelopeForRoutes(privateGroup.BasePath(), []response.Route{
		{Method: "GET", Path: "/system/users"},
		{Method: "PUT", Path: "/system/users/:id"},
		{Method: "PATCH", Path: "/system/users/:id/status"},
		{Method: "PUT", Path: "/system/users/:id/roles"},
		{Method: "GET", Path: "/system/auth/me"},
		{Method: "POST", Path: "/system/auth/logout"},
		{Method: "GET", Path: "/system/audit-logs"},
		{Method: "GET", Path: "/system/audit-logs/export"},
	}), middleware.JWTAuth(), middleware.CasbinHandlerWithPrefix(
		service.ServiceGroupApp.SystemServiceGroup.RBAC, privateGroup.BasePath(), defaultRouterPrefix,
	))
	cloudGroup := privateGroup.Group("")
	if service.ServiceGroupApp.ProjectServiceGroup.ProjectService != nil {
		cloudGroup.Use(middleware.ProjectScope(service.ServiceGroupApp.ProjectServiceGroup.ProjectService))
	}

	routerGroup := router.RouterGroupApp
	routerGroup.System.InitSystemRouter(publicGroup, privateGroup)
	routerGroup.Cloud.InitCloudRouter(cloudGroup)
	routerGroup.Project.InitProjectRouter(privateGroup)

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
