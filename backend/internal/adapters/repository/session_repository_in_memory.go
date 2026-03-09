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

const (
	ACCESS_TOKEN_EXPIRATION  = 15 * time.Minute
	REFRESH_TOKEN_EXPIRATION = 7 * 24 * time.Hour
)

type SessionRepositoryInMemory struct {
	mu                        sync.RWMutex
	accessSessionsByToken     map[string]domain.AccessSession
	refreshSessionsByToken    map[string]domain.RefreshSession
	accessSessionsByDeviceID  map[string]domain.AccessSession
	refreshSessionsByDeviceID map[string]domain.RefreshSession
}

func NewSessionRepositoryInMemory() *SessionRepositoryInMemory {
	return &SessionRepositoryInMemory{
		accessSessionsByToken:     make(map[string]domain.AccessSession),
		refreshSessionsByToken:    make(map[string]domain.RefreshSession),
		accessSessionsByDeviceID:  make(map[string]domain.AccessSession),
		refreshSessionsByDeviceID: make(map[string]domain.RefreshSession),
	}
}

func (r *SessionRepositoryInMemory) deleteFromAccessSessionsByToken(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, exist := r.accessSessionsByToken[token]
	if exist {
		delete(r.accessSessionsByDeviceID, session.DeviceID)
		delete(r.accessSessionsByToken, token)
	}
}

func (r *SessionRepositoryInMemory) deleteFromRefreshSessionsByToken(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, exist := r.refreshSessionsByToken[token]
	if exist {
		delete(r.refreshSessionsByDeviceID, session.DeviceID)
		delete(r.refreshSessionsByToken, token)
	}
}

func (r *SessionRepositoryInMemory) deleteFromAccessSessionsByDeviceID(deviceID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, exist := r.accessSessionsByDeviceID[deviceID]
	if exist {
		delete(r.accessSessionsByDeviceID, deviceID)
		delete(r.accessSessionsByToken, session.TokenHash)
	}
}

func (r *SessionRepositoryInMemory) deleteFromRefreshSessionsByDeviceID(deviceID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, exist := r.refreshSessionsByDeviceID[deviceID]
	if exist {
		delete(r.refreshSessionsByDeviceID, deviceID)
		delete(r.refreshSessionsByToken, session.TokenHash)
	}
}

func (r *SessionRepositoryInMemory) addToAccessSessions(token string, session domain.AccessSession) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.accessSessionsByToken[token] = session
	r.accessSessionsByDeviceID[session.DeviceID] = session
}

func (r *SessionRepositoryInMemory) addToRefreshSessions(token string, session domain.RefreshSession) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refreshSessionsByToken[token] = session
	r.refreshSessionsByDeviceID[session.DeviceID] = session
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
	userID int32,
	deviceID string,
) (*domain.AccessSessionLight, error) {
	token, err := utils.GenerateToken()
	if err != nil {
		return nil, err
	}

	tokenHash := utils.HashToken(token)
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

	r.deleteFromAccessSessionsByDeviceID(deviceID)
	r.addToAccessSessions(tokenHash, session)
	return &domain.AccessSessionLight{Token: token, ExpiresAt: session.ExpiresAt}, nil
}

func (r *SessionRepositoryInMemory) NewRefreshToken(
	ctx context.Context,
	userID int32,
	deviceID string,
) (*domain.RefreshSessionLight, error) {
	refreshToken, err := utils.GenerateToken()
	if err != nil {
		return nil, err
	}

	refreshTokenHash := utils.HashToken(refreshToken)
	if _, exists := r.getFromRefreshSessions(refreshTokenHash); exists {
		return nil, fmt.Errorf("refresh session with this token already exists")
	}

	refreshSession := domain.RefreshSession{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: refreshTokenHash,
		ExpiresAt: time.Now().Add(REFRESH_TOKEN_EXPIRATION),
		CreatedAt: time.Now(),
		DeviceID:  deviceID,
	}

	r.deleteFromRefreshSessionsByDeviceID(deviceID)
	r.addToRefreshSessions(refreshTokenHash, refreshSession)
	return &domain.RefreshSessionLight{Token: refreshToken, ExpiresAt: refreshSession.ExpiresAt}, nil
}

func (r *SessionRepositoryInMemory) LogContents() error {
	log.Println("Access Sessions:")
	for token, session := range r.accessSessionsByToken {
		log.Printf("Token: %s, Session: %+s\n", token, session.DeviceID)
	}

	log.Println("Refresh Sessions:")
	for token, session := range r.refreshSessionsByToken {
		log.Printf("Token: %s, Session: %+s\n", token, session.DeviceID)
	}

	return nil
}
