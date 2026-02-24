package bootstrap

import (
	"main/internal/core/services"
)

type Services struct {
	*services.UserService
}

func NewServices(r *Repositories) *Services {
	return &Services{
		UserService: services.NewUserService(r.UserRepository, r.SessionRepository),
	}
}
