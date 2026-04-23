package system

type Service struct{}

type HealthPayload struct {
	Status string
}

type PingPayload struct {
	Message string
}

func (s *Service) Health() HealthPayload {
	return HealthPayload{Status: "ok"}
}

func (s *Service) Ping() PingPayload {
	return PingPayload{Message: "pong"}
}
