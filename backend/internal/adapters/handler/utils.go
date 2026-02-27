package handler

import (
	"main/internal/core/domain"
	"main/internal/oapi"

	"github.com/oapi-codegen/runtime/types"
)

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
