package repository

import (
	"context"
	"database/sql"
	"errors"
	"main/internal/core/domain"
	db "main/internal/db/sqlc"
	"main/internal/utils"
	"time"
)

type VaultRepositoryPg struct {
	queries *db.Queries
}

func NewVaultRepositoryPg(dbConn *sql.DB) *VaultRepositoryPg {
	return &VaultRepositoryPg{
		queries: db.New(dbConn),
	}
}

func (r *VaultRepositoryPg) GetVaultByUserID(ctx context.Context, userID string) (*domain.Vault, error) {
	id, err := utils.Int32FromString(userID)
	if err != nil {
		return nil, err
	}

	dbVault, err := r.queries.GetVaultByUserID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	v := toDomainVault(dbVault)
	return v, nil
}

func (r *VaultRepositoryPg) GetVaultUpdatedAtByUserID(ctx context.Context, userID string) (*time.Time, error) {
	id, err := utils.Int32FromString(userID)
	if err != nil {
		return nil, err
	}

	dbVault, err := r.queries.GetVaultUpdatedAtByUserID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &dbVault.Time, nil
}

func (r *VaultRepositoryPg) InsertVaultByUserID(
	ctx context.Context,
	userID string,
	vault []byte,
) (*domain.Vault, error) {

	id, err := utils.Int32FromString(userID)
	if err != nil {
		return nil, err
	}

	newVault, err := r.queries.InsertVaultByUserID(ctx, db.InsertVaultByUserIDParams{
		UserID: id,
		Vault:  vault,
	})

	if err != nil {
		return nil, err
	}

	v := toDomainVault(newVault)
	return v, nil
}

func toDomainVault(v db.Vault) *domain.Vault {
	return &domain.Vault{
		UserID:    v.UserID,
		Vault:     v.Vault,
		CreatedAt: v.CreatedAt.Time,
		UpdatedAt: v.UpdatedAt.Time,
	}
}
