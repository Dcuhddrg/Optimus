package system

import (
	systemApi "optimus/server/api/v1/system"

	"github.com/gin-gonic/gin"
)

type RouterGroup struct{}

func (r *RouterGroup) InitSystemRouter(publicGroup *gin.RouterGroup, privateGroup *gin.RouterGroup, handlers *systemApi.SystemHandlers) {
	if handlers == nil {
		handlers = systemApi.NewSystemHandlers(systemApi.SystemHandlerDeps{})
	}

	publicGroup.GET("/health", handlers.Health.Health)
	publicGroup.GET("/ping", handlers.Base.Ping)
	publicGroup.GET("/base/ping", handlers.Base.Ping)
	publicGroup.POST("/system/auth/login", handlers.Auth.Login)

	privateGroup.GET("/system/auth/me", handlers.Auth.CurrentUser)
	privateGroup.POST("/system/auth/logout", handlers.Auth.Logout)
	privateGroup.GET("/system/users", handlers.User.Users)
	privateGroup.PUT("/system/users/:id", handlers.User.UpdateUser)
	privateGroup.PATCH("/system/users/:id/status", handlers.User.UpdateUserStatus)
	privateGroup.PUT("/system/users/:id/roles", handlers.User.UpdateUserRoles)
	privateGroup.GET("/system/roles", handlers.Role.Roles)
	privateGroup.POST("/system/roles", handlers.Role.CreateRole)
	privateGroup.PUT("/system/roles/:id", handlers.Role.UpdateRole)
	privateGroup.PATCH("/system/roles/:id/status", handlers.Role.UpdateRoleStatus)
	privateGroup.PUT("/system/roles/:id/permissions", handlers.Role.UpdateRolePermissions)
	privateGroup.GET("/system/permissions/tree", handlers.Role.PermissionTree)
	privateGroup.GET("/system/ldap/config", handlers.LDAP.LDAPConfig)
	privateGroup.PUT("/system/ldap/config", handlers.LDAP.UpdateLDAPConfig)
	privateGroup.POST("/system/ldap/test", handlers.LDAP.TestLDAPConfig)
	privateGroup.GET("/system/audit-logs", handlers.Audit.AuditLogs)
}
