package handler

import (
	"context"
	"strconv"

	"main/internal/adapters/middleware"
	"main/internal/core/services"
	"main/internal/oapi"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{userService: service}
}

func (h *UserHandler) ListUsers(ctx context.Context, request oapi.ListUsersRequestObject) (oapi.ListUsersResponseObject, error) {
	users, err := h.userService.ListUsers(ctx)
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

	user, err := h.userService.CreateUser(ctx, request.Body.Name, string(request.Body.Email), request.Body.Password)
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

func (s *UserHandler) GetCurrentUser(ctx context.Context, request oapi.GetCurrentUserRequestObject) (oapi.GetCurrentUserResponseObject, error) {
	session, ok := middleware.GetSession(ctx)

	if !ok {
		return &oapi.GetCurrentUser500JSONResponse{
			InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{
				Code:    500,
				Message: "Internal Server Error",
			},
		}, nil
	}
	if session == nil {
		return oapi.GetCurrentUser401JSONResponse{
			Code:    401,
			Message: "Unauthorized",
		}, nil
	}
	return &oapi.GetCurrentUser200JSONResponse{
		Email:    session.UserEmail,
		Id:       strconv.FormatInt(int64(session.UserID), 10),
		Username: session.UserName,
	}, nil
}
