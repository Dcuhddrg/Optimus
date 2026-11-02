package system

import "optimus/server/service"

type ApiGroup struct {
	AuthApi
	UserApi
	RoleApi
	LDAPApi
	AuditApi
	BaseApi
	HealthApi
}

var systemServices = &service.ServiceGroupApp.SystemServiceGroup
