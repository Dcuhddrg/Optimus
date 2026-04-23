package service

import "optimus/server/service/system"

type ServiceGroup struct {
	System system.Service
}

var ServiceGroupApp = new(ServiceGroup)
