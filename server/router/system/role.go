package system

import "github.com/gin-gonic/gin"

type RoleRouter struct{}

func (r *RoleRouter) InitRoleRouter(privateGroup *gin.RouterGroup) {
	roleRouter := privateGroup.Group("system/roles")
	permissionRouter := privateGroup.Group("system/permissions")
	roleRouter.GET("", roleApi.Roles)
	roleRouter.POST("", roleApi.CreateRole)
	roleRouter.PUT(":id", roleApi.UpdateRole)
	roleRouter.PATCH(":id/status", roleApi.UpdateRoleStatus)
	roleRouter.PUT(":id/permissions", roleApi.UpdateRolePermissions)
	permissionRouter.GET("tree", roleApi.PermissionTree)
}
