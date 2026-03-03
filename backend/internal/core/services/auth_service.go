package services

import (
	"context"
	"fmt"
	"main/internal/core/domain"
	"main/internal/core/ports"
	"main/internal/utils"
)

type AuthService struct {
	userRepository    ports.UserRepository
	sessionRepository ports.SessionRepository
}

func NewAuthService(userRepo ports.UserRepository, sessionRepo ports.SessionRepository) *AuthService {
	return &AuthService{userRepository: userRepo, sessionRepository: sessionRepo}
}

func (s *AuthService) CreateToken(ctx context.Context, email, password, deviceID string) (*domain.User, *domain.AccessSessionLight, *domain.RefreshSessionLight, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, nil, nil, err
	}
	if user == nil {
		return nil, nil, nil, fmt.Errorf("Invalid email or password")
	}

	err = utils.CheckPassword(user.PasswordHash, password)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Invalid email or password")
	}

	accessSession, refreshSession, err := s.sessionRepository.CreateAccessAndRefreshSessions(ctx, user, deviceID)

	if err != nil {
		return nil, nil, nil, err
	}

	return user, accessSession, refreshSession, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken, deviceID string) (*domain.AccessSessionLight, *domain.RefreshSessionLight, error) {
	accessSession, refreshSession, err := s.sessionRepository.RefreshToken(ctx, refreshToken, deviceID)

	if err != nil {
		return nil, nil, err
	}

	return accessSession, refreshSession, nil
}

func (s *AuthService) GetAccessSessionByToken(ctx context.Context, token string) (*domain.AccessSession, error) {
	accessSession, err := s.sessionRepository.GetAccessSessionByToken(ctx, token)

	if err != nil {
		return nil, err
	}

	return accessSession, nil
}

func (s *AuthService) GetRefreshSessionByToken(ctx context.Context, token string) (*domain.RefreshSession, error) {
	refreshSession, err := s.sessionRepository.GetRefreshSessionByToken(ctx, token)

	if err != nil {
		return nil, err
	}

	return refreshSession, nil
}
