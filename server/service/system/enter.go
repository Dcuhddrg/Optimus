package system

// ServiceGroup embeds the stateless system services.
type ServiceGroup struct {
	AuthService
	UserService
	RoleService
	LDAPService
	AuditService
	RBACService
	Service
}
