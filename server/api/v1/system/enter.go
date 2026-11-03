package system

import "optimus/server/service"

type ApiGroup struct {
	AuthApi
	UserApi
	RoleApi
	LDAPApi
	AuditApi
	BaseApi
	HealthApi
}

// Each alias points to its service slot so startup can publish the service.
var (
	authService = &service.ServiceGroupApp.SystemServiceGroup.AuthService
	userService = &service.ServiceGroupApp.SystemServiceGroup.UserService
	roleService = &service.ServiceGroupApp.SystemServiceGroup.RoleService
	ldapService = &service.ServiceGroupApp.SystemServiceGroup.LDAPService
	auditService = &service.ServiceGroupApp.SystemServiceGroup.AuditService
	statusService = &service.ServiceGroupApp.SystemServiceGroup.Status
)
