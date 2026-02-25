package ports

import (
	"context"
	"main/internal/core/domain"
	"time"
)

type UserRepository interface {
	ListUsers(ctx context.Context) ([]domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateUser(ctx context.Context, name, email, passwordHash string) (*domain.User, error)
}

type SessionRepository interface {
	CreateSession(
		ctx context.Context,
		user *domain.User,
		tokenHash string,
		expiresAt time.Time,
		userAgent string,
		ipAddress string,
	) (*domain.Session, error)
	GetSessionByToken(ctx context.Context, token string) (*domain.Session, error)
}
