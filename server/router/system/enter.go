package system

import (
	systemApi "optimus/server/api/v1/system"

	"github.com/gin-gonic/gin"
)

type RouterGroup struct{}

func (r *RouterGroup) InitSystemRouter(publicGroup *gin.RouterGroup, privateGroup *gin.RouterGroup) {
	api := &systemApi.Api{}

	publicGroup.GET("/health", api.Health)
	publicGroup.GET("/ping", api.Ping)
	publicGroup.GET("/base/ping", api.Ping)
	publicGroup.POST("/system/auth/login", api.Login)

	privateGroup.GET("/system/auth/me", api.CurrentUser)
	privateGroup.POST("/system/auth/logout", api.Logout)
	privateGroup.GET("/system/users", api.Users)
	privateGroup.PUT("/system/users/:id", api.UpdateUser)
	privateGroup.PATCH("/system/users/:id/status", api.UpdateUserStatus)
	privateGroup.PUT("/system/users/:id/roles", api.UpdateUserRoles)
	privateGroup.GET("/system/roles", api.Roles)
	privateGroup.POST("/system/roles", api.CreateRole)
	privateGroup.PUT("/system/roles/:id", api.UpdateRole)
	privateGroup.PATCH("/system/roles/:id/status", api.UpdateRoleStatus)
	privateGroup.PUT("/system/roles/:id/permissions", api.UpdateRolePermissions)
	privateGroup.GET("/system/permissions/tree", api.PermissionTree)
	privateGroup.GET("/system/ldap/config", api.LDAPConfig)
	privateGroup.PUT("/system/ldap/config", api.UpdateLDAPConfig)
	privateGroup.POST("/system/ldap/test", api.TestLDAPConfig)
	privateGroup.GET("/system/audit-logs", api.AuditLogs)
}
