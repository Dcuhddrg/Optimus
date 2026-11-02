package system

import "github.com/gin-gonic/gin"

type UserRouter struct{}

func (r *UserRouter) InitUserRouter(privateGroup *gin.RouterGroup) {
	privateGroup.GET("/system/users", userApi.Users)
	privateGroup.PUT("/system/users/:id", userApi.UpdateUser)
	privateGroup.PATCH("/system/users/:id/status", userApi.UpdateUserStatus)
	privateGroup.PUT("/system/users/:id/roles", userApi.UpdateUserRoles)
}
