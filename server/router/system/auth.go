package system

import "github.com/gin-gonic/gin"

type AuthRouter struct{}

func (r *AuthRouter) InitAuthRouter(publicGroup, privateGroup *gin.RouterGroup) {
	publicGroup.POST("/system/auth/login", authApi.Login)
	privateGroup.GET("/system/auth/me", authApi.CurrentUser)
	privateGroup.POST("/system/auth/logout", authApi.Logout)
}
