package system

type SystemHandlers struct {
	Auth   *Api
	Base   *Api
	Health *Api
}

func NewSystemHandlers() *SystemHandlers {
	api := &Api{}

	return &SystemHandlers{
		Auth:   api,
		Base:   api,
		Health: api,
	}
}
