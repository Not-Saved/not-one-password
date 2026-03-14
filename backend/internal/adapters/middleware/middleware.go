package middleware

import (
	"main/internal/config"
	"main/internal/core/services"
)

type Middleware struct {
	Config      *config.Config
	AuthService *services.AuthService
}

func NewMiddleware(AuthService *services.AuthService, config *config.Config) *Middleware {
	return &Middleware{AuthService: AuthService, Config: config}
}
