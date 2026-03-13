package services

import (
	"context"
	"main/internal/core/domain"
	"main/internal/core/ports"
	"time"
)

type VaultService struct {
	vaultRepo ports.VaultRepository
}

func NewVaultService(
	vaultRepo ports.VaultRepository,
) *VaultService {
	return &VaultService{vaultRepo: vaultRepo}
}

func (s *VaultService) GetVaultByUserID(ctx context.Context, userID string) (*domain.Vault, error) {
	return s.vaultRepo.GetVaultByUserID(ctx, userID)
}

func (s *VaultService) GetVaultUpdatedAtByUserID(ctx context.Context, userID string) (*time.Time, error) {
	return s.vaultRepo.GetVaultUpdatedAtByUserID(ctx, userID)
}

func (s *VaultService) InsertVaultByUserID(ctx context.Context, userID string, vault []byte) (*domain.Vault, error) {
	return s.vaultRepo.InsertVaultByUserID(ctx, userID, vault)
}
