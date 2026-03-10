package handler

import (
	"context"

	"main/internal/adapters/middleware"
	"main/internal/core/services"
	"main/internal/oapi"

	"github.com/oapi-codegen/runtime/types"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{userService: service}
}

func (h *UserHandler) ListUsers(ctx context.Context, request oapi.ListUsersRequestObject) (oapi.ListUsersResponseObject, error) {
	users, err := h.userService.GetUsers(ctx)
	if err != nil {
		return oapi.ListUsers500JSONResponse{
			InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	var response = make([]oapi.UserResponse, 0, len(users))

	for _, u := range users {
		response = append(response, mapToAPIUser(u))
	}

	return oapi.ListUsers200JSONResponse(response), nil
}

func (h *UserHandler) CreateUser(ctx context.Context, request oapi.CreateUserRequestObject) (oapi.CreateUserResponseObject, error) {
	_, err := h.userService.CreateUser(ctx, request.Body.Name, string(request.Body.Email), request.Body.Password)
	if err != nil {
		return oapi.CreateUser400JSONResponse{
			BadRequestJSONResponse: oapi.BadRequestJSONResponse{
				Code:    400,
				Message: err.Error(),
			},
		}, nil
	}

	return oapi.CreateUser201Response{}, nil
}

func (s *UserHandler) GetCurrentUser(ctx context.Context, request oapi.GetCurrentUserRequestObject) (oapi.GetCurrentUserResponseObject, error) {
	session, ok := middleware.GetAccessSession(ctx)

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
	user, err := s.userService.GetUserByPublicID(ctx, session.UserID)
	if err != nil {
		return &oapi.GetCurrentUser500JSONResponse{
			InternalServerErrorJSONResponse: oapi.InternalServerErrorJSONResponse{
				Code:    500,
				Message: "Internal Server Error",
			},
		}, nil
	}

	response := mapToAPIUser(*user)
	return oapi.GetCurrentUser200JSONResponse(response), nil
}

func (s *UserHandler) ConfirmUser(ctx context.Context, r oapi.ConfirmUserRequestObject) (oapi.ConfirmUserResponseObject, error) {
	user, err := s.userService.ConfirmUser(ctx, r.Params.Code)
	if err != nil {
		return oapi.ConfirmUser400JSONResponse{
			BadRequestJSONResponse: oapi.BadRequestJSONResponse{
				Code:    400,
				Message: err.Error(),
			},
		}, nil
	}

	return oapi.ConfirmUser200JSONResponse{
		Email: types.Email(user.Email),
		Id:    user.PublicID,
		Name:  &user.Email,
	}, nil
}
