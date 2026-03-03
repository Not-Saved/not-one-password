package middleware

import (
	"context"
	"encoding/json"
	"main/internal/core/domain"
	"main/internal/oapi"
	"net/http"
	"strings"
)

const (
	SessionContextKey      contextKey = "session"
	RefreshTokenContextKey contextKey = "refresh_token"
	AccessTokenContextKey  contextKey = "access_token"
)

func (m *Middleware) AuthMiddleware(next oapi.StrictHandlerFunc, operationID string) oapi.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		switch operationID {
		case "GetCurrentUser":
			return m.hasAccessToken(next, ctx, w, r, request)
		case "RefreshToken":
			return m.hasRefreshToken(next, ctx, w, r, request)
		default:
			return next(ctx, w, r, request)
		}
	}
}

func (m *Middleware) hasRefreshToken(next oapi.StrictHandlerFunc, ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeError(w, "missing Authorization header")
		return nil, nil
	}

	// Expect format: "Bearer <token>"
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		writeError(w, "invalid Authorization header format")
		return nil, nil
	}

	tokenString := strings.TrimPrefix(authHeader, prefix)

	session, err := m.AuthService.GetRefreshSessionByToken(ctx, tokenString)
	if err != nil {
		writeError(w, "invalid or expired token")
		return nil, nil
	}

	ctx = context.WithValue(ctx, SessionContextKey, session)
	ctx = context.WithValue(ctx, RefreshTokenContextKey, tokenString)

	return next(ctx, w, r, request)
}

func (m *Middleware) hasAccessToken(next oapi.StrictHandlerFunc, ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeError(w, "missing Authorization header")
		return nil, nil
	}

	// Expect format: "Bearer <token>"
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		writeError(w, "invalid Authorization header format")
		return nil, nil
	}

	tokenString := strings.TrimPrefix(authHeader, prefix)

	session, err := m.AuthService.GetAccessSessionByToken(ctx, tokenString)
	if err != nil {
		writeError(w, "invalid or expired token")
		return nil, nil
	}

	ctx = context.WithValue(ctx, SessionContextKey, session)
	ctx = context.WithValue(ctx, AccessTokenContextKey, tokenString)

	return next(ctx, w, r, request)
}

func writeError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(401)
	json.NewEncoder(w).Encode(oapi.ErrorResponse{
		Code:    401,
		Message: message,
	})
}

func GetAccessSession(ctx context.Context) (*domain.AccessSession, bool) {
	session, ok := ctx.Value(SessionContextKey).(*domain.AccessSession)
	return session, ok
}

func GetRefreshSession(ctx context.Context) (*domain.RefreshSession, bool) {
	session, ok := ctx.Value(SessionContextKey).(*domain.RefreshSession)
	return session, ok
}

func GetAccessToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(AccessTokenContextKey).(string)
	return token, ok
}

func GetRefreshToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(RefreshTokenContextKey).(string)
	return token, ok
}
