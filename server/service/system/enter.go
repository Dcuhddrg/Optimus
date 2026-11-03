package system

// ServiceGroup holds initialized system services and shared capabilities.
type ServiceGroup struct {
	AuthService AuthOperations
	*UserService
	*RoleService
	*LDAPService
	AuditService AuditOperations
	RBAC         Authorizer
	Status       StatusService
}
