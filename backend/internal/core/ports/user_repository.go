package ports

import (
	"context"
	"main/internal/core/domain"
)

type UserRepository interface {
	GetUsers(ctx context.Context) ([]domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByPublicID(ctx context.Context, id string) (*domain.User, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	CreateUser(ctx context.Context, name, email, passwordHash string) (*domain.User, error)
}
