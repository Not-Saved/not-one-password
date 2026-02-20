package bootstrap

import "main/internal/adapters/handler"

type Handlers struct {
	User *handler.UserHandler
}

func NewHandlers(s *Services) *Handlers {
	return &Handlers{
		User: handler.NewUserHandler(s.User),
	}
}
