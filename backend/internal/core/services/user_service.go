package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"main/internal/core/domain"
	"main/internal/core/ports"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepository    ports.UserRepository
	sessionRepository ports.SessionRepository
}

func NewUserService(userRepo ports.UserRepository, sessionRepo ports.SessionRepository) *UserService {
	return &UserService{userRepository: userRepo, sessionRepository: sessionRepo}
}

func (s *UserService) ListUsers(ctx context.Context) ([]domain.User, error) {
	return s.userRepository.ListUsers(ctx)
}

func (s *UserService) LoginUser(ctx context.Context, email, password, userAgent, ip string) (*domain.User, *domain.SessionLight, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, fmt.Errorf("Invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, nil, fmt.Errorf("Invalid email or password")
	}

	token, err := generateToken()
	if err != nil {
		return nil, nil, err
	}
	hashedToken := hashToken(token)
	session, err := s.sessionRepository.CreateSession(ctx, user, hashedToken, time.Now().Add(24*time.Hour), userAgent, ip)
	sessionLight := &domain.SessionLight{
		Token:     token,
		ExpiresAt: session.ExpiresAt,
	}

	if err != nil {
		return nil, nil, err
	}

	return user, sessionLight, nil
}

func (s *UserService) CreateUser(ctx context.Context, name, email, password string) (*domain.User, error) {
	existingUser, err := s.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, fmt.Errorf("User already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return s.userRepository.CreateUser(ctx, name, email, string(hashedPassword))
}

func (s *UserService) GetSessionByToken(ctx context.Context, token string) (*domain.Session, error) {
	hashedToken := hashToken(token)

	session, err := s.sessionRepository.GetSessionByToken(ctx, hashedToken)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
