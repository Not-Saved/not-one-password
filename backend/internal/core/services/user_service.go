package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"main/internal/core/domain"
	"main/internal/core/ports"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	UserRepository    ports.UserRepository
	SessionRepository ports.SessionRepository
}

func NewUserService(userRepo ports.UserRepository, sessionRepo ports.SessionRepository) *UserService {
	return &UserService{UserRepository: userRepo, SessionRepository: sessionRepo}
}

func (s *UserService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.UserRepository.ListUsers(ctx)
}

func (s *UserService) GetUser(ctx context.Context, id int32) (domain.User, error) {
	return s.UserRepository.GetUser(ctx, id)
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.UserRepository.GetUserByEmail(ctx, email)
}

func (s *UserService) LoginUser(ctx context.Context, email, password, userAgent, ip string) (*domain.User, *domain.Session, error) {
	user, err := s.UserRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, fmt.Errorf("Invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, nil, err
	}

	token, err := generateToken()
	if err != nil {
		return nil, nil, err
	}
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	session, err := s.SessionRepository.CreateSession(ctx, user.ID, string(hashedToken), time.Now().Add(24*time.Hour), userAgent, ip)

	if err != nil {
		return nil, nil, err
	}

	return user, &session, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (s *UserService) CreateUser(ctx context.Context, name, email, password string) (domain.User, error) {
	existingUser, err := s.UserRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	if existingUser != nil {
		return domain.User{}, fmt.Errorf("User already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	return s.UserRepository.CreateUser(ctx, name, email, string(hashedPassword))
}
