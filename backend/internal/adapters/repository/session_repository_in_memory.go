package repository

import (
	"context"
	"fmt"
	"main/internal/core/domain"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SessionRepositoryInMemory struct {
	mu               sync.RWMutex
	sessionsByToken  map[string]domain.Session
	sessionsByUserID map[string]domain.Session
}

func NewSessionRepositoryInMemory() *SessionRepositoryInMemory {
	return &SessionRepositoryInMemory{
		sessionsByToken:  make(map[string]domain.Session),
		sessionsByUserID: make(map[string]domain.Session),
	}
}

func (r *SessionRepositoryInMemory) CreateSession(
	ctx context.Context,
	user *domain.User,
	tokenHash string,
	expiresAt time.Time,
	userAgent,
	ipAddress string,
) (*domain.Session, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cleanupExpiredLocked()

	if _, exists := r.sessionsByToken[tokenHash]; exists {
		return nil, fmt.Errorf("session with this token already exists")
	}

	session := domain.Session{
		ID:        uuid.NewString(),
		UserID:    user.ID,
		UserName:  user.Name,
		UserEmail: user.Email,
		TokenHash: tokenHash,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
		UserAgent: userAgent,
		IpAddress: ipAddress,
	}

	r.sessionsByToken[tokenHash] = session
	r.sessionsByUserID[strconv.FormatInt(int64(session.UserID), 10)] = session

	return &session, nil
}

func (r *SessionRepositoryInMemory) GetSessionByToken(
	ctx context.Context,
	token string,
) (*domain.Session, error) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessionsByToken[token]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		delete(r.sessionsByToken, token)
		delete(r.sessionsByUserID, strconv.FormatInt(int64(session.UserID), 10))
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

func (r *SessionRepositoryInMemory) cleanupExpiredLocked() {
	now := time.Now()
	for token, session := range r.sessionsByToken {
		if now.After(session.ExpiresAt) {
			delete(r.sessionsByToken, token)
		}
	}
}
