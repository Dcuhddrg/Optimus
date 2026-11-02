package system

import "github.com/gin-gonic/gin"

type RoleRouter struct{}

func (r *RoleRouter) InitRoleRouter(privateGroup *gin.RouterGroup) {
	privateGroup.GET("/system/roles", roleApi.Roles)
	privateGroup.POST("/system/roles", roleApi.CreateRole)
	privateGroup.PUT("/system/roles/:id", roleApi.UpdateRole)
	privateGroup.PATCH("/system/roles/:id/status", roleApi.UpdateRoleStatus)
	privateGroup.PUT("/system/roles/:id/permissions", roleApi.UpdateRolePermissions)
	privateGroup.GET("/system/permissions/tree", roleApi.PermissionTree)
}
