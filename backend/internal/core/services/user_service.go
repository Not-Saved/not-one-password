package services

import (
	"context"
	"main/internal/core/domain"
	"main/internal/core/ports"
)

type UserService struct {
	UserRepository ports.UserRepository
}

func NewUserService(userRepo ports.UserRepository) *UserService {
	return &UserService{UserRepository: userRepo}
}

func (s *UserService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.UserRepository.ListUsers(ctx)
}

func (s *UserService) GetUser(ctx context.Context, id int32) (domain.User, error) {
	return s.UserRepository.GetUser(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, name, email string) (domain.User, error) {
	return s.UserRepository.CreateUser(ctx, name, email)
}
