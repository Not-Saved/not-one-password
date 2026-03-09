package handler

import (
	"main/internal/core/domain"
	"main/internal/oapi"

	"github.com/oapi-codegen/runtime/types"
)

func mapToAPIUser(u domain.User) oapi.UserResponse {
	id := u.PublicID
	name := u.Name
	email := types.Email(u.Email)

	return oapi.UserResponse{
		Id:    id,
		Name:  &name,
		Email: email,
	}
}
