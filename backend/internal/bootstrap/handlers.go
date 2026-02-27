package bootstrap

import (
	"main/internal/adapters/handler"
)

type Handlers struct {
	*handler.UserHandler
	*handler.AuthHandler
}

func NewHandlers(s *Services) *Handlers {
	return &Handlers{
		UserHandler: handler.NewUserHandler(s.UserService),
		AuthHandler: handler.NewAuthHandler(s.UserService),
	}
}
