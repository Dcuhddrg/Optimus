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

type Api struct {
	handlers *SystemHandlers
}

var systemDefaultStatusService systemService.Service
var defaultCompatibilityHandlers = NewSystemHandlers(SystemHandlerDeps{})

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

func NewAPI(handlers *SystemHandlers) *Api {
	return &Api{handlers: handlers}
}

func (a *Api) currentHandlers() *SystemHandlers {
	if a != nil && a.handlers != nil {
		return a.handlers
	}
	return defaultCompatibilityHandlers
}

func (a *Api) Login(c *gin.Context) {
	a.currentHandlers().Auth.Login(c)
}

func (a *Api) CurrentUser(c *gin.Context) {
	a.currentHandlers().Auth.CurrentUser(c)
}

func (a *Api) Logout(c *gin.Context) {
	a.currentHandlers().Auth.Logout(c)
}

func (a *Api) Users(c *gin.Context) {
	a.currentHandlers().User.Users(c)
}

func (a *Api) UpdateUser(c *gin.Context) {
	a.currentHandlers().User.UpdateUser(c)
}

func (a *Api) UpdateUserStatus(c *gin.Context) {
	a.currentHandlers().User.UpdateUserStatus(c)
}

func (a *Api) UpdateUserRoles(c *gin.Context) {
	a.currentHandlers().User.UpdateUserRoles(c)
}

func (a *Api) Roles(c *gin.Context) {
	a.currentHandlers().Role.Roles(c)
}

func (a *Api) CreateRole(c *gin.Context) {
	a.currentHandlers().Role.CreateRole(c)
}

func (a *Api) UpdateRole(c *gin.Context) {
	a.currentHandlers().Role.UpdateRole(c)
}

func (a *Api) UpdateRoleStatus(c *gin.Context) {
	a.currentHandlers().Role.UpdateRoleStatus(c)
}

func (a *Api) UpdateRolePermissions(c *gin.Context) {
	a.currentHandlers().Role.UpdateRolePermissions(c)
}

func (a *Api) PermissionTree(c *gin.Context) {
	a.currentHandlers().Role.PermissionTree(c)
}

func (a *Api) LDAPConfig(c *gin.Context) {
	a.currentHandlers().LDAP.LDAPConfig(c)
}

func (a *Api) UpdateLDAPConfig(c *gin.Context) {
	a.currentHandlers().LDAP.UpdateLDAPConfig(c)
}

func (a *Api) TestLDAPConfig(c *gin.Context) {
	a.currentHandlers().LDAP.TestLDAPConfig(c)
}

func (a *Api) AuditLogs(c *gin.Context) {
	a.currentHandlers().Audit.AuditLogs(c)
}

func (a *Api) Ping(c *gin.Context) {
	a.currentHandlers().Base.Ping(c)
}

func (a *Api) Health(c *gin.Context) {
	a.currentHandlers().Health.Health(c)
}

func writeServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, gin.H{"message": message})
}
