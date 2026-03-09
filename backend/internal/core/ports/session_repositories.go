package ports

import (
	"context"
	"main/internal/core/domain"
)

type SessionRepository interface {
	NewAccessToken(
		ctx context.Context,
		userID int32,
		deviceID string,
	) (*domain.AccessSessionLight, error)

	NewRefreshToken(
		ctx context.Context,
		userID int32,
		deviceID string,
	) (*domain.RefreshSessionLight, error)

	GetAccessSessionByToken(ctx context.Context, token string) (*domain.AccessSession, error)
	GetRefreshSessionByToken(ctx context.Context, token string) (*domain.RefreshSession, error)
}
