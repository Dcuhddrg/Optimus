package system

import "github.com/gin-gonic/gin"

type UserRouter struct{}

func (r *UserRouter) InitUserRouter(privateGroup *gin.RouterGroup) {
	userRouter := privateGroup.Group("system/users")
	userRouter.GET("", userApi.Users)
	userRouter.PUT(":id", userApi.UpdateUser)
	userRouter.PATCH(":id/status", userApi.UpdateUserStatus)
	userRouter.PUT(":id/roles", userApi.UpdateUserRoles)
}
