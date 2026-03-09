package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/core/domain"
	"main/internal/oapi"
	"net/http"
	"strings"
)

const (
	SessionContextKey       contextKey = "session"
	TokenResponseContextKey contextKey = "token_response"
	SessionCookieName       string     = "SESSION_ID"
)

func (m *Middleware) AuthMiddleware(next oapi.StrictHandlerFunc, operationID string) oapi.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		switch operationID {
		case "GetCurrentUser", "ListUsers", "LogoutUser":
			return m.hasAccessToken(next, ctx, w, r, request)
		case "RefreshToken":
			return m.hasRefreshToken(next, ctx, w, r, request)
		default:
			return next(ctx, w, r, request)
		}
	}
}

func (m *Middleware) hasRefreshToken(next oapi.StrictHandlerFunc, ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
	tokenBase64, err := getBase64TokenFromRequest(r)
	if err != nil {
		writeError(w, "no valid authentication method detected")
		return nil, nil
	}

	tokenResponse, err := domain.NewTokensFromBase64(tokenBase64)
	if err != nil {
		writeError(w, "invalid token format")
		return nil, nil
	}

	session, err := m.AuthService.GetRefreshSessionByToken(ctx, tokenResponse.RefreshToken)
	if err != nil {
		writeError(w, "invalid or expired token")
		return nil, nil
	}

	ctx = context.WithValue(ctx, SessionContextKey, session)
	ctx = context.WithValue(ctx, TokenResponseContextKey, tokenResponse)

	return next(ctx, w, r, request)
}

func (m *Middleware) hasAccessToken(next oapi.StrictHandlerFunc, ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
	tokenBase64, err := getBase64TokenFromRequest(r)
	if err != nil {
		writeError(w, "no valid authentication method detected")
		return nil, nil
	}

	tokenResponse, err := domain.NewTokensFromBase64(tokenBase64)
	if err != nil {
		writeError(w, "invalid token format")
		return nil, nil
	}

	session, err := m.AuthService.GetAccessSessionByToken(ctx, tokenResponse.AccessToken)
	if err != nil {
		writeError(w, "invalid or expired token")
		return nil, nil
	}

	ctx = context.WithValue(ctx, SessionContextKey, session)
	ctx = context.WithValue(ctx, TokenResponseContextKey, tokenResponse)

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

func getBase64TokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")

	cookie, err := r.Cookie(SessionCookieName)
	if authHeader == "" && err != nil {
		return "", fmt.Errorf("Unauthorized")
	}

	var tokenBase64 string

	if authHeader != "" {
		// Expect format: "Bearer <token>"
		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			return "", fmt.Errorf("Unauthorized")
		}
		tokenBase64 = strings.TrimPrefix(authHeader, prefix)

	}

	if cookie != nil {
		tokenBase64 = cookie.Value
	}

	return tokenBase64, nil
}

func GetAccessSession(ctx context.Context) (*domain.AccessSession, bool) {
	session, ok := ctx.Value(SessionContextKey).(*domain.AccessSession)
	return session, ok
}

func GetRefreshSession(ctx context.Context) (*domain.RefreshSession, bool) {
	session, ok := ctx.Value(SessionContextKey).(*domain.RefreshSession)
	return session, ok
}

func GetTokenResponse(ctx context.Context) (*domain.Tokens, bool) {
	token, ok := ctx.Value(TokenResponseContextKey).(*domain.Tokens)
	return token, ok
}
