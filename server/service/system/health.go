package system

import "optimus/server/global"

type Service struct{}

type HealthPayload struct {
	Status  string
	Healthy bool
}

type PingPayload struct {
	Message string
}

func (s *Service) Health() HealthPayload {
	if global.OPTIMUS_STARTUP_ERR != nil {
		return HealthPayload{Status: "unhealthy"}
	}
	return HealthPayload{Status: "ok", Healthy: true}
}

func (s *Service) Ping() PingPayload {
	return PingPayload{Message: "pong"}
}
