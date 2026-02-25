package repository

import (
	"context"
	"database/sql"
	"errors"
	"main/internal/core/domain"
	db "main/internal/db/sqlc"
)

type UserRepositoryPg struct {
	queries *db.Queries
}

func NewUserRepositoryPg(dbConn *sql.DB) *UserRepositoryPg {
	return &UserRepositoryPg{
		queries: db.New(dbConn),
	}
}

func (r *UserRepositoryPg) ListUsers(ctx context.Context) ([]domain.User, error) {
	dbUsers, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	users := make([]domain.User, 0, len(dbUsers))
	for _, u := range dbUsers {
		users = append(users, *toDomainUser(u))
	}
	return users, nil
}

func (r *UserRepositoryPg) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	dbUser, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	u := toDomainUser(dbUser)
	return u, nil
}

func (r *UserRepositoryPg) CreateUser(ctx context.Context, name, email, passwordHash string) (*domain.User, error) {
	dbUser, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return nil, err
	}
	return toDomainUser(dbUser), nil
}

func toDomainUser(u db.User) *domain.User {
	return &domain.User{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		CreatedAt:    u.CreatedAt,
		PasswordHash: u.PasswordHash,
	}
}
