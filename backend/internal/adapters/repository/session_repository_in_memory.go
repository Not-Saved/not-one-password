package repository

import (
	"context"
	"fmt"
	"log"
	"main/internal/core/domain"
	"main/internal/utils"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SessionRepositoryInMemory struct {
	mu                     sync.RWMutex
	accessSessionsByToken  map[string]domain.AccessSession
	refreshSessionsByToken map[string]domain.RefreshSession
}

func NewSessionRepositoryInMemory() *SessionRepositoryInMemory {
	return &SessionRepositoryInMemory{
		accessSessionsByToken:  make(map[string]domain.AccessSession),
		refreshSessionsByToken: make(map[string]domain.RefreshSession),
	}
}

func (r *SessionRepositoryInMemory) deleteFromAccessSessionsByToken(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.accessSessionsByToken, token)
}

func (r *SessionRepositoryInMemory) deleteFromRefreshSessionsByToken(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.refreshSessionsByToken, token)
}

func (r *SessionRepositoryInMemory) addToAccessSessions(token string, session domain.AccessSession) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.accessSessionsByToken[token] = session
}

func (r *SessionRepositoryInMemory) addToRefreshSessions(token string, session domain.RefreshSession) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refreshSessionsByToken[token] = session
}

func (r *SessionRepositoryInMemory) getFromAccessSessions(token string) (domain.AccessSession, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, exists := r.accessSessionsByToken[token]
	return session, exists
}

func (r *SessionRepositoryInMemory) getFromRefreshSessions(token string) (domain.RefreshSession, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, exists := r.refreshSessionsByToken[token]
	return session, exists
}

func (r *SessionRepositoryInMemory) GetAccessSessionByToken(
	ctx context.Context,
	token string,
) (*domain.AccessSession, error) {
	hashedToken := utils.HashToken(token)

	session, exists := r.getFromAccessSessions(hashedToken)
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		r.deleteFromAccessSessionsByToken(hashedToken)
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

func (r *SessionRepositoryInMemory) GetRefreshSessionByToken(
	ctx context.Context,
	token string,
) (*domain.RefreshSession, error) {
	hashedToken := utils.HashToken(token)

	session, exists := r.getFromRefreshSessions(hashedToken)
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		r.deleteFromRefreshSessionsByToken(hashedToken)
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

func (r *SessionRepositoryInMemory) NewAccessToken(
	ctx context.Context,
	userID string,
	deviceID string,
) (*domain.AccessSessionLight, error) {
	token, tokenHash, err := GenerateTokenForSession()
	if err != nil {
		return nil, err
	}

	if _, exists := r.getFromAccessSessions(tokenHash); exists {
		return nil, fmt.Errorf("session with this token already exists")
	}

	session := domain.AccessSession{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(ACCESS_TOKEN_EXPIRATION),
		CreatedAt: time.Now(),
		DeviceID:  deviceID,
	}

	r.addToAccessSessions(tokenHash, session)
	return &domain.AccessSessionLight{Token: token, ExpiresAt: session.ExpiresAt}, nil
}

func (r *SessionRepositoryInMemory) NewRefreshToken(
	ctx context.Context,
	userID string,
	deviceID string,
) (*domain.RefreshSessionLight, error) {

	token, tokenHash, err := GenerateTokenForSession()
	if err != nil {
		return nil, err
	}

	if _, exists := r.getFromRefreshSessions(tokenHash); exists {
		return nil, fmt.Errorf("refresh session with this token already exists")
	}

	refreshSession := domain.RefreshSession{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(REFRESH_TOKEN_EXPIRATION),
		CreatedAt: time.Now(),
		DeviceID:  deviceID,
	}

	r.addToRefreshSessions(tokenHash, refreshSession)
	r.LogContents()
	return &domain.RefreshSessionLight{Token: token, ExpiresAt: refreshSession.ExpiresAt}, nil
}

func (r *SessionRepositoryInMemory) LogContents() error {
	log.Println("Access Sessions:")
	for token, session := range r.accessSessionsByToken {
		log.Printf("Token: %s, Session: %+s\n", token, session.UserID)
	}

	log.Println("Refresh Sessions:")
	for token, session := range r.refreshSessionsByToken {
		log.Printf("Token: %s, Session: %+s\n", token, session.UserID)
	}

	return nil
}
