package handler

import (
	"context"
	"main/internal/adapters/middleware"
	"main/internal/core/services"
	"main/internal/oapi"
	"net/http"
	"time"
)

type AuthHandler struct {
	userService *services.UserService
}

func NewAuthHandler(service *services.UserService) *AuthHandler {
	return &AuthHandler{userService: service}
}

func (h *AuthHandler) LoginUser(ctx context.Context, request oapi.LoginUserRequestObject) (oapi.LoginUserResponseObject, error) {
	ip, _ := ctx.Value(middleware.IpContextKey).(string)
	userAgent, _ := ctx.Value(middleware.UserAgentContextKey).(string)

	user, session, err := h.userService.LoginUser(ctx, string(request.Body.Email), request.Body.Password, userAgent, ip)

	if err != nil {
		return oapi.LoginUser401JSONResponse{Code: 401, Message: "invalid email or password"}, nil
	}

	response := mapToAPIUser(*user)

	cookie := NewAuthCookie(session.Token, session.ExpiresAt)

	return oapi.LoginUser200JSONResponse{
		Headers: oapi.LoginUser200ResponseHeaders{SetCookie: cookie.String()},
		Body:    response,
	}, nil
}

func NewAuthCookie(token string, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     middleware.AuthCookieName,
		HttpOnly: true,
		Secure:   true,
		Expires:  expiresAt,
		Value:    token,
	}
}
