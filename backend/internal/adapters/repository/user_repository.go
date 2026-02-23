package repository

import (
	"context"
	"database/sql"
	"main/internal/core/domain"
	db "main/internal/db/sqlc"
)

type UserRepository struct {
	queries *db.Queries
}

func NewUserRepository(dbConn *sql.DB) *UserRepository {
	return &UserRepository{
		queries: db.New(dbConn),
	}
}

func (r *UserRepository) ListUsers(ctx context.Context) ([]domain.User, error) {
	dbUsers, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	users := make([]domain.User, 0, len(dbUsers))
	for _, u := range dbUsers {
		users = append(users, toDomainUser(u))
	}
	return users, nil
}

func (r *UserRepository) GetUser(ctx context.Context, id int32) (domain.User, error) {
	dbUser, err := r.queries.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, err
	}
	return toDomainUser(dbUser), nil
}

func (r *UserRepository) CreateUser(ctx context.Context, name, email string) (domain.User, error) {
	dbUser, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Name:  name,
		Email: email,
	})
	if err != nil {
		return domain.User{}, err
	}
	return toDomainUser(dbUser), nil
}

func toDomainUser(u db.User) domain.User {
	return domain.User{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}
