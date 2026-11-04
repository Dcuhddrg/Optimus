package system

import "github.com/gin-gonic/gin"

type AuthRouter struct{}

func (r *AuthRouter) InitAuthRouter(publicGroup, privateGroup *gin.RouterGroup) {
	publicAuthRouter := publicGroup.Group("system/auth")
	privateAuthRouter := privateGroup.Group("system/auth")
	publicAuthRouter.POST("login", authApi.Login)
	privateAuthRouter.GET("me", authApi.CurrentUser)
	privateAuthRouter.POST("logout", authApi.Logout)
}
