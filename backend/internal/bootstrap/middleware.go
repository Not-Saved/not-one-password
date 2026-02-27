package bootstrap

import "main/internal/adapters/middleware"

type Middlewares struct {
	*middleware.Middleware
}

func NewMiddlewares(s *Services) *Middlewares {
	return &Middlewares{
		Middleware: middleware.NewMiddleware(s.UserService),
	}
}
