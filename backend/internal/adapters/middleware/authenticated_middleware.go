package middleware

import (
	"context"
	"encoding/json"
	"main/internal/core/domain"
	"main/internal/oapi"
	"net/http"
)

const (
	SessionContextKey contextKey = "session"
	AuthCookieName    string     = "session_token"
)

func (m *Middleware) AuthMiddleware(next oapi.StrictHandlerFunc, operationID string) oapi.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		switch operationID {
		case "GetCurrentUser":
			return m.isAuthenticated(next, ctx, w, r, request)
		default:
			return next(ctx, w, r, request)
		}

	}
}

func (m *Middleware) isAuthenticated(next oapi.StrictHandlerFunc, ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
	cookie, err := r.Cookie(AuthCookieName)

	if err != nil {
		writeError(w, "missing auth cookie")
		return nil, nil
	}

	session, err := m.UserService.GetSessionByToken(ctx, cookie.Value)
	if err != nil {
		writeError(w, err.Error())
		return nil, nil
	}

	ctx = context.WithValue(ctx, SessionContextKey, session)

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

func GetSession(ctx context.Context) (*domain.Session, bool) {
	session, ok := ctx.Value(SessionContextKey).(*domain.Session)
	return session, ok
}
