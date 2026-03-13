package ports

import (
	"context"
	"main/internal/core/domain"
	"time"
)

type VaultRepository interface {
	GetVaultByUserID(ctx context.Context, userID string) (*domain.Vault, error)
	GetVaultUpdatedAtByUserID(ctx context.Context, userID string) (*time.Time, error)
	InsertVaultByUserID(ctx context.Context, userID string, vault []byte) (*domain.Vault, error)
}
