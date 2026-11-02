package system

// ServiceGroup holds initialized system services and shared capabilities.
type ServiceGroup struct {
	*AuthService
	*UserService
	*RoleService
	*LDAPService
	*AuditService
	RBAC   Authorizer
	Status StatusService
}
