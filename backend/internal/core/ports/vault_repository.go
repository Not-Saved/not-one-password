package ports

import (
	"context"
	"main/internal/core/domain"
)

type VaultRepository interface {
	GetVaultByUserID(ctx context.Context, userID string) (*domain.Vault, error)
	InsertVaultByUserID(ctx context.Context, userID string, vault []byte) (*domain.Vault, error)
}
