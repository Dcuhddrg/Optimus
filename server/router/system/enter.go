package system

import (
	"github.com/gin-gonic/gin"
	api "optimus/server/api/v1"
)

type RouterGroup struct {
	BaseRouter
	AuthRouter
	UserRouter
	RoleRouter
	LDAPRouter
	AuditRouter
}

var (
	healthApi = &api.ApiGroupApp.SystemApiGroup.HealthApi
	baseApi   = &api.ApiGroupApp.SystemApiGroup.BaseApi
	authApi   = &api.ApiGroupApp.SystemApiGroup.AuthApi
	userApi   = &api.ApiGroupApp.SystemApiGroup.UserApi
	roleApi   = &api.ApiGroupApp.SystemApiGroup.RoleApi
	ldapApi   = &api.ApiGroupApp.SystemApiGroup.LDAPApi
	auditApi  = &api.ApiGroupApp.SystemApiGroup.AuditApi
)

func (r *RouterGroup) InitSystemRouter(publicGroup *gin.RouterGroup, privateGroup *gin.RouterGroup) {
	r.BaseRouter.InitBaseRouter(publicGroup)
	r.AuthRouter.InitAuthRouter(publicGroup, privateGroup)
	r.UserRouter.InitUserRouter(privateGroup)
	r.RoleRouter.InitRoleRouter(privateGroup)
	r.LDAPRouter.InitLDAPRouter(privateGroup)
	r.AuditRouter.InitAuditRouter(privateGroup)
}
