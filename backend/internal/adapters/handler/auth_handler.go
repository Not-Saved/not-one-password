package handler

import (
	"context"
	"main/internal/adapters/middleware"
	"main/internal/core/services"
	"main/internal/oapi"
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

	return oapi.IssueToken200JSONResponse{
		Headers: oapi.IssueToken200ResponseHeaders{},
		Body: oapi.TokenResponse{
			AccessToken:  access.Token,
			RefreshToken: refresh.Token,
			ExpiresIn:    int(time.Until(access.ExpiresAt).Seconds()),
		},
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, request oapi.RefreshTokenRequestObject) (oapi.RefreshTokenResponseObject, error) {
	refreshToken, ok := middleware.GetRefreshToken(ctx)
	if !ok {
		return &oapi.RefreshToken500JSONResponse{
			InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{
				Code:    500,
				Message: "Internal Server Error",
			},
		}, nil
	}

	newAccessSession, newRefreshSession, err := h.authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		return &oapi.RefreshToken401JSONResponse{
			Code:    401,
			Message: "Invalid or expired refresh token",
		}, nil
	}

	return oapi.RefreshToken200JSONResponse{
		Headers: oapi.RefreshToken200ResponseHeaders{},
		Body: oapi.TokenResponse{
			AccessToken:  newAccessSession.Token,
			RefreshToken: newRefreshSession.Token,
			ExpiresIn:    int(time.Until(newAccessSession.ExpiresAt).Seconds()),
		},
	}, nil
}
