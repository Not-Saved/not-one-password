package bootstrap

import (
	"main/internal/adapters/middleware"
	"main/internal/config"
)

type Middlewares struct {
	*middleware.Middleware
}

func NewMiddlewares(s *Services, cfg *config.Config) *Middlewares {
	return &Middlewares{
		Middleware: middleware.NewMiddleware(s.AuthService, cfg),
	}
}
