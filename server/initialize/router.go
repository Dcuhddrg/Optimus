package initialize

import (
	"strings"

	"optimus/server/global"
	"optimus/server/middleware"
	"optimus/server/router"

	"github.com/gin-gonic/gin"
)

const defaultRouterPrefix = "/api/v1"

var newSystemModule = NewSystemModule

func Routers() *gin.Engine {
	systemModule := newSystemModule(global.OPTIMUS_DB, global.OPTIMUS_CONFIG)
	if systemModule == nil {
		systemModule = NewSystemModule(nil, global.OPTIMUS_CONFIG)
	}

	engine := gin.New()
	// 全局中间件
	engine.Use(middleware.Logger(), middleware.Recovery())
	// 公共路由
	publicGroup := engine.Group(routerPrefix())
	// 私有路由
	privateGroup := engine.Group(routerPrefix())
	privateGroup.Use(middleware.JWTAuth(), middleware.CasbinHandler(systemModule.Capabilities.RBAC))

	routerGroup := router.RouterGroupApp
	routerGroup.System.InitSystemRouter(publicGroup, privateGroup, systemModule.Handlers)

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
