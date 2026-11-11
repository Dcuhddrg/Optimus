package system

// ServiceGroup embeds the stateless system services.
type ServiceGroup struct {
	AuthService
	AuthStateService
	UserService
	RoleService
	LDAPService
	AuditService
	CasbinService
	JwtService
	Service
}
