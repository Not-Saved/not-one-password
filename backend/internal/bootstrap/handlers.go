package bootstrap

import (
	"main/internal/adapters/handler"
)

type Handlers struct {
	*handler.UserHandler
}

func NewHandlers(s *Services) *Handlers {
	return &Handlers{
		UserHandler: handler.NewUserHandler(s.UserService),
	}
}
