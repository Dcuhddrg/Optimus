package system

import "context"

// Authorizer is service-owned to avoid importing HTTP middleware into Service.
type Authorizer interface {
	AuthorizeAPI(context.Context, string, string, []uint) (bool, error)
}

type StatusService interface {
	Health() HealthPayload
	Ping() PingPayload
}

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
