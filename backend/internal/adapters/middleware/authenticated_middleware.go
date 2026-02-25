package middleware

import (
	"context"
	"main/internal/core/services"
	"net/http"
)

const (
	SessionContextKey contextKey = "session"
)

func AuthenticatedMiddleware(next http.Handler, userService *services.UserService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("session_token")

		var ctx context.Context
		if err == nil && token != nil {
			session, err := userService.GetSessionByToken(r.Context(), token.Value)
			if err == nil {
				ctx = context.WithValue(r.Context(), SessionContextKey, session)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
