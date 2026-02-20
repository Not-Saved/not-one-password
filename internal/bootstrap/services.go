package bootstrap

import "main/internal/core/services"

type Services struct {
	User *services.UserService
}

func NewServices(r *Repositories) *Services {
	return &Services{
		User: services.NewUserService(r.User),
	}
}
