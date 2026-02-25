package handler

import (
	"context"
	"log"
	"net/http"

	"main/internal/core/domain"
	"main/internal/core/services"
	"main/internal/oapi"

	"github.com/oapi-codegen/runtime/types"
)

type UserHandler struct {
	UserService *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{UserService: service}
}

func (h *UserHandler) ListUsers(ctx context.Context, request oapi.ListUsersRequestObject) (oapi.ListUsersResponseObject, error) {
	users, err := h.UserService.ListUsers(ctx)
	if err != nil {
		return oapi.ListUsers500JSONResponse{
			InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	var response = make([]oapi.User, 0, len(users))

	for _, u := range users {
		response = append(response, mapToAPIUser(u))
	}

	return oapi.ListUsers200JSONResponse(response), nil
}

func (h *UserHandler) CreateUser(ctx context.Context, request oapi.CreateUserRequestObject) (oapi.CreateUserResponseObject, error) {
	if request.Body == nil {
		return oapi.CreateUser400JSONResponse{
			BadRequestJSONResponse: oapi.BadRequestJSONResponse{
				Code:    400,
				Message: "missing request body",
			},
		}, nil
	}

	user, err := h.UserService.CreateUser(ctx, request.Body.Name, string(request.Body.Email), request.Body.Password)
	if err != nil {
		return oapi.CreateUser400JSONResponse{
			BadRequestJSONResponse: oapi.BadRequestJSONResponse{
				Code:    400,
				Message: err.Error(),
			},
		}, nil
	}

	response := mapToAPIUser(*user)

	return oapi.CreateUser201JSONResponse(response), nil
}

func (h *UserHandler) LoginUser(ctx context.Context, request oapi.LoginUserRequestObject) (oapi.LoginUserResponseObject, error) {
	r := ctx.Value("httpRequest").(*http.Request)
	w := ctx.Value("httpResponseWriter").(http.ResponseWriter)

	userAgent := r.UserAgent()
	ip := r.RemoteAddr

	user, session, err := h.UserService.LoginUser(r.Context(), string(request.Body.Email), request.Body.Password, userAgent, ip)

	if err != nil {
		log.Printf("login failed: %v", err)
		return oapi.LoginUser401JSONResponse{Code: 401, Message: "invalid email or password"}, nil
	}

	response := mapToAPIUser(*user)

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.TokenHash,
		HttpOnly: true,
		Secure:   false,
		Expires:  session.ExpiresAt,
	})

	return oapi.LoginUser200JSONResponse(response), nil
}

func mapToAPIUser(u domain.User) oapi.User {
	id := u.ID
	name := u.Name
	email := types.Email(u.Email)
	createdAt := u.CreatedAt

	return oapi.User{
		ID:        &id,
		Name:      &name,
		Email:     &email,
		CreatedAt: &createdAt,
	}
}
