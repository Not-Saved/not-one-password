package services

import (
	"context"
	"fmt"
	"main/internal/core/domain"
	"main/internal/core/ports"
	"main/internal/utils"
)

type UserService struct {
	userRepository       ports.UserRepository
	sessionRepository    ports.SessionRepository
	userIntentRepository ports.UserIntentRepository
}

func NewUserService(userRepo ports.UserRepository, userIntentRepo ports.UserIntentRepository, sessionRepo ports.SessionRepository) *UserService {
	return &UserService{userRepository: userRepo, userIntentRepository: userIntentRepo, sessionRepository: sessionRepo}
}

func (s *UserService) GetUsers(ctx context.Context) ([]domain.User, error) {
	return s.userRepository.GetUsers(ctx)
}

func (s *UserService) GetUserByPublicID(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.userRepository.GetUserByPublicID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("User doesn't exist")
	}

	return user, nil
}

func (s *UserService) CreateUser(ctx context.Context, name, email, password string) (*domain.RegistrationIntentToken, error) {
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

	registrationToken, err := s.userIntentRepository.CreateRegistrationIntent(ctx, domain.RegistrationIntentUser{
		Name:         name,
		PasswordHash: hashedPassword,
		Email:        email,
	})
	if err != nil {
		return nil, err
	}

	//TODO: email

	return registrationToken, nil
}

func (s *UserService) ConfirmUser(ctx context.Context, code string) (*domain.User, error) {
	registrationIntent, err := s.userIntentRepository.ConsumeRegistrationIntent(ctx, code)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepository.CreateUser(ctx, registrationIntent.Name, registrationIntent.Email, registrationIntent.PasswordHash)
	if err != nil {
		return nil, err
	}

	return user, nil
}
