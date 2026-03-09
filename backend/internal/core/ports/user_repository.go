package ports

import (
	"context"
	"main/internal/core/domain"
)

type UserRepository interface {
	GetUsers(ctx context.Context) ([]domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id int32) (*domain.User, error)
	CreateUser(ctx context.Context, name, email, passwordHash string) (*domain.User, error)
}
