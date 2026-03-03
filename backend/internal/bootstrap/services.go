package bootstrap

import (
	"main/internal/core/services"
)

type Services struct {
	*services.UserService
	*services.AuthService
}

func NewServices(r *Repositories) *Services {
	return &Services{
		UserService: services.NewUserService(r.UserRepository, r.SessionRepository),
		AuthService: services.NewAuthService(r.UserRepository, r.SessionRepository),
	}
}
