package middleware

import "main/internal/core/services"

type Middleware struct {
	UserService *services.UserService
}

func NewMiddleware(UserService *services.UserService) *Middleware {
	return &Middleware{UserService: UserService}
}
