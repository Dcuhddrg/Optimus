package system

import "optimus/server/service"

type AuthApi struct{}
type UserApi struct{}
type RoleApi struct{}
type LDAPApi struct{}
type AuditApi struct{}
type BaseApi struct{}
type HealthApi struct{}

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
