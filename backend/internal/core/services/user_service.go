package services

import (
	"context"
	"main/internal/core/domain"
	"main/internal/core/ports"
)

type UserService struct {
	repo ports.UserRepository
}

func NewUserService(repo ports.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.repo.ListUsers(ctx)
}

func (s *UserService) GetUser(ctx context.Context, id int32) (domain.User, error) {
	return s.repo.GetUser(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, name, email string) (domain.User, error) {
	return s.repo.CreateUser(ctx, name, email)
}
