package ports

import (
	"context"
	"main/internal/core/domain"
	"time"
)

type UserRepository interface {
	ListUsers(ctx context.Context) ([]domain.User, error)
	GetUser(ctx context.Context, id int32) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateUser(ctx context.Context, name, email, passwordHash string) (domain.User, error)
}

type SessionRepository interface {
	CreateSession(
		ctx context.Context,
		userID int32,
		token string,
		expiresAt time.Time,
		userAgent string,
		ipAddress string,
	) (domain.Session, error)
	GetSessionByToken(ctx context.Context, token string) (domain.Session, error)
	ListActiveSessionsByUser(ctx context.Context, userID int32) ([]domain.Session, error)
	RevokeSessionByToken(ctx context.Context, token string) error
	RevokeAllSessionsByUser(ctx context.Context, userID int32) error
}
