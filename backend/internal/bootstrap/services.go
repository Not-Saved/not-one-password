package bootstrap

import (
	"main/internal/core/services"
)

type Services struct {
	*services.UserService
	*services.AuthService
	*services.VaultService
}

func NewServices(r *Adapters) *Services {
	return &Services{
		UserService:  services.NewUserService(r.UserRepository, r.UserIntentRepository, r.UserNotifier, r.SessionRepository),
		AuthService:  services.NewAuthService(r.UserRepository, r.SessionRepository),
		VaultService: services.NewVaultService(r.VaultRepository),
	}
}
