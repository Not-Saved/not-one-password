package handler

import (
	"context"
	"main/internal/adapters/middleware"
	"main/internal/core/domain"
	"main/internal/core/services"
	"main/internal/oapi"
	"main/internal/utils"
	"net/http"
	"time"
)

type AuthHandler struct {
	userService *services.UserService
	authService *services.AuthService
}

func NewAuthHandler(userService *services.UserService, authService *services.AuthService) *AuthHandler {
	return &AuthHandler{userService: userService, authService: authService}
}

func (h *AuthHandler) IssueToken(ctx context.Context, request oapi.IssueTokenRequestObject) (oapi.IssueTokenResponseObject, error) {
	_, access, refresh, err := h.authService.CreateToken(ctx, string(request.Body.Email), request.Body.Password, request.Body.DeviceID)

	if err != nil {
		return oapi.IssueToken401JSONResponse{Code: 401, Message: "invalid email or password"}, nil
	}

	tokenResponse, setCookie, err := tokenReponseAndSetCookieFromSessions(access, refresh)
	if err != nil {
		return nil, err
	}

	return oapi.IssueToken200JSONResponse{
		Headers: oapi.IssueToken200ResponseHeaders{SetCookie: setCookie.String()},
		Body:    *tokenResponse,
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, request oapi.RefreshTokenRequestObject) (oapi.RefreshTokenResponseObject, error) {
	tokenResponse, ok := middleware.GetTokenResponse(ctx)
	if !ok || tokenResponse == nil {
		return &oapi.RefreshToken401JSONResponse{
			Code:    401,
			Message: "Unauthorized",
		}, nil
	}

	newAccessSession, newRefreshSession, err := h.authService.RefreshToken(ctx, tokenResponse.RefreshToken)
	if err != nil {
		return &oapi.RefreshToken401JSONResponse{
			Code:    401,
			Message: "Invalid or expired refresh token",
		}, nil
	}

	newTokenResponse, setCookie, err := tokenReponseAndSetCookieFromSessions(newAccessSession, newRefreshSession)
	if err != nil {
		return nil, err
	}

	return oapi.RefreshToken200JSONResponse{
		Headers: oapi.RefreshToken200ResponseHeaders{SetCookie: setCookie.String()},
		Body:    *newTokenResponse,
	}, nil
}

func (h *AuthHandler) LogoutUser(ctx context.Context, r oapi.LogoutUserRequestObject) (oapi.LogoutUserResponseObject, error) {
	session, ok := middleware.GetAccessSession(ctx)
	if !ok || session == nil {
		return oapi.LogoutUser401JSONResponse{
			Code:    401,
			Message: "Unauthorized",
		}, nil
	}

	err := h.authService.Logout(ctx, session.UserID, session.DeviceID)
	if err != nil {
		return oapi.LogoutUser500JSONResponse{
			InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{
				Code:    500,
				Message: "Internal server error",
			},
		}, nil
	}

	emptySetCookie := newEmptySetCookie()
	return oapi.LogoutUser204Response{
		Headers: oapi.LogoutUser204ResponseHeaders{
			SetCookie: emptySetCookie.String(),
		}}, nil
}

func tokenReponseAndSetCookieFromSessions(access *domain.AccessSessionLight, refresh *domain.RefreshSessionLight) (*oapi.TokenResponse, *http.Cookie, error) {
	tokenResponse := domain.Tokens{
		AccessToken:  access.Token,
		RefreshToken: refresh.Token,
		ExpiresIn:    utils.SecondsUntilTime(access.ExpiresAt),
	}
	tokenBase64, err := tokenResponse.ToBase64()
	if err != nil {
		return nil, nil, err
	}
	setCookie := newSetCookie(tokenBase64, refresh.ExpiresAt)
	return &oapi.TokenResponse{Token: tokenBase64}, &setCookie, nil
}

func newSetCookie(value string, expiresAt time.Time) http.Cookie {
	return http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    value,
		MaxAge:   utils.SecondsUntilTime(expiresAt),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
}

func newEmptySetCookie() http.Cookie {
	return http.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
}
