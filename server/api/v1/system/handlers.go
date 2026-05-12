package system

import (
	"net/http"

	systemService "optimus/server/service/system"
	systemUsecase "optimus/server/service/system/usecase"

	"github.com/gin-gonic/gin"
)

type systemStatusService interface {
	Health() systemService.HealthPayload
	Ping() systemService.PingPayload
}

type AuthHandler struct {
	service *systemUsecase.AuthService
}

type UserHandler struct {
	service *systemUsecase.UserService
}

type RoleHandler struct {
	service *systemUsecase.RoleService
}

type LDAPHandler struct {
	service *systemUsecase.LDAPService
}

type AuditHandler struct {
	service *systemUsecase.AuditService
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
	Auth   *systemUsecase.AuthService
	User   *systemUsecase.UserService
	Role   *systemUsecase.RoleService
	LDAP   *systemUsecase.LDAPService
	Audit  *systemUsecase.AuditService
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
