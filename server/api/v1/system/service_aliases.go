package system

import "optimus/server/service"

// RefreshServiceAliases publishes startup-assembled service values before routes are registered.
// Calling it while requests are running is not supported.
func RefreshServiceAliases() {
	group := &service.ServiceGroupApp.SystemServiceGroup
	authService = group.AuthService
	userService = group.UserService
	roleService = group.RoleService
	ldapService = group.LDAPService
	auditService = group.AuditService
	statusService = group.Service
}
