package middleware

import "main/internal/core/services"

type Middleware struct {
	AuthService *services.AuthService
}

func NewMiddleware(AuthService *services.AuthService) *Middleware {
	return &Middleware{AuthService: AuthService}
}
