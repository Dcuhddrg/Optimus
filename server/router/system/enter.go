package system

import (
	"github.com/gin-gonic/gin"
	api "optimus/server/api/v1"
)

type RouterGroup struct{}

var (
	healthApi = &api.ApiGroupApp.SystemApiGroup.HealthApi
	baseApi   = &api.ApiGroupApp.SystemApiGroup.BaseApi
	authApi   = &api.ApiGroupApp.SystemApiGroup.AuthApi
	userApi   = &api.ApiGroupApp.SystemApiGroup.UserApi
	roleApi   = &api.ApiGroupApp.SystemApiGroup.RoleApi
	ldapApi   = &api.ApiGroupApp.SystemApiGroup.LDAPApi
	auditApi  = &api.ApiGroupApp.SystemApiGroup.AuditApi
)

func (r *RouterGroup) InitSystemRouter(publicGroup *gin.RouterGroup, privateGroup *gin.RouterGroup) {

	publicGroup.GET("/health", healthApi.Health)
	publicGroup.GET("/ping", baseApi.Ping)
	publicGroup.GET("/base/ping", baseApi.Ping)
	publicGroup.POST("/system/auth/login", authApi.Login)

	privateGroup.GET("/system/auth/me", authApi.CurrentUser)
	privateGroup.POST("/system/auth/logout", authApi.Logout)
	privateGroup.GET("/system/users", userApi.Users)
	privateGroup.PUT("/system/users/:id", userApi.UpdateUser)
	privateGroup.PATCH("/system/users/:id/status", userApi.UpdateUserStatus)
	privateGroup.PUT("/system/users/:id/roles", userApi.UpdateUserRoles)
	privateGroup.GET("/system/roles", roleApi.Roles)
	privateGroup.POST("/system/roles", roleApi.CreateRole)
	privateGroup.PUT("/system/roles/:id", roleApi.UpdateRole)
	privateGroup.PATCH("/system/roles/:id/status", roleApi.UpdateRoleStatus)
	privateGroup.PUT("/system/roles/:id/permissions", roleApi.UpdateRolePermissions)
	privateGroup.GET("/system/permissions/tree", roleApi.PermissionTree)
	privateGroup.GET("/system/ldap/config", ldapApi.LDAPConfig)
	privateGroup.PUT("/system/ldap/config", ldapApi.UpdateLDAPConfig)
	privateGroup.POST("/system/ldap/test", ldapApi.TestLDAPConfig)
	privateGroup.GET("/system/audit-logs", auditApi.QueryAuditLogs)
	privateGroup.GET("/system/audit-logs/export", auditApi.ExportAuditLogs)
}
