package ports

import (
	"context"
	"main/internal/core/domain"
)

type UserRepository interface {
	ListUsers(ctx context.Context) ([]domain.User, error)
	GetUser(ctx context.Context, id int32) (domain.User, error)
	CreateUser(ctx context.Context, name, email string) (domain.User, error)
}
