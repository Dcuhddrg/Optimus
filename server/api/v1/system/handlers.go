package system

import (
	"net/http"

	systemService "optimus/server/service/system"

	"github.com/gin-gonic/gin"
)

type systemStatusService interface {
	Health() systemService.HealthPayload
	Ping() systemService.PingPayload
}

type AuthHandler struct {
	service *systemService.AuthService
}

type UserHandler struct {
	service *systemService.UserService
}

type RoleHandler struct {
	service *systemService.RoleService
}

type LDAPHandler struct {
	service *systemService.LDAPService
}

type AuditHandler struct {
	service *systemService.AuditService
}

type BaseHandler struct {
	service systemStatusService
}

type HealthHandler struct {
	service systemStatusService
}

type SystemHandlers struct {
	Auth   *AuthHandler
	User   *UserHandler
	Role   *RoleHandler
	LDAP   *LDAPHandler
	Audit  *AuditHandler
	Base   *BaseHandler
	Health *HealthHandler
}

type SystemHandlerDeps struct {
	Auth   *systemService.AuthService
	User   *systemService.UserService
	Role   *systemService.RoleService
	LDAP   *systemService.LDAPService
	Audit  *systemService.AuditService
	System systemStatusService
}

func NewSystemHandlers(deps SystemHandlerDeps) *SystemHandlers {
	systemDeps := deps.System
	if systemDeps == nil {
		systemDeps = &systemService.Service{}
	}

	return &SystemHandlers{
		Auth:   &AuthHandler{service: deps.Auth},
		User:   &UserHandler{service: deps.User},
		Role:   &RoleHandler{service: deps.Role},
		LDAP:   &LDAPHandler{service: deps.LDAP},
		Audit:  &AuditHandler{service: deps.Audit},
		Base:   &BaseHandler{service: systemDeps},
		Health: &HealthHandler{service: systemDeps},
	}
}

func writeServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"message": message})
}
