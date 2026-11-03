package system

import (
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
	healthApi = api.ApiGroupApp.SystemApiGroup.HealthApi
	baseApi   = api.ApiGroupApp.SystemApiGroup.BaseApi
	authApi   = api.ApiGroupApp.SystemApiGroup.AuthApi
	userApi   = api.ApiGroupApp.SystemApiGroup.UserApi
	roleApi   = api.ApiGroupApp.SystemApiGroup.RoleApi
	ldapApi   = api.ApiGroupApp.SystemApiGroup.LDAPApi
	auditApi  = api.ApiGroupApp.SystemApiGroup.AuditApi
)
