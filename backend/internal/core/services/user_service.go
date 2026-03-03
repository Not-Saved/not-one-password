package services

import (
	"context"
	"fmt"
	"main/internal/core/domain"
	"main/internal/core/ports"
	"main/internal/utils"
)

type UserService struct {
	userRepository    ports.UserRepository
	sessionRepository ports.SessionRepository
}

func NewUserService(userRepo ports.UserRepository, sessionRepo ports.SessionRepository) *UserService {
	return &UserService{userRepository: userRepo, sessionRepository: sessionRepo}
}

func (s *UserService) GetUsers(ctx context.Context) ([]domain.User, error) {
	return s.userRepository.GetUsers(ctx)
}

func (s *UserService) GetUserByID(ctx context.Context, id int32) (*domain.User, error) {
	return s.userRepository.GetUserByID(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, name, email, password string) (*domain.User, error) {
	existingUser, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, fmt.Errorf("User already exists")
	}

	hashedPassword, err := utils.NewPassword(password)
	if err != nil {
		return nil, err
	}
	return s.userRepository.CreateUser(ctx, name, email, string(hashedPassword))
}
