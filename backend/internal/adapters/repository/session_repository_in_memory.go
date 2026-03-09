package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"main/internal/core/domain"
	"main/internal/utils"

	"github.com/google/uuid"
)

// SessionRepositoryInMemory is a thread-safe in-memory session store
type SessionRepositoryInMemory struct {
	mu sync.RWMutex

	// tokenHash -> session
	accessSessions  map[string]domain.AccessSession
	refreshSessions map[string]domain.RefreshSession

	// userID+deviceID -> tokenHash
	userDeviceAccess  map[string]string
	userDeviceRefresh map[string]string
}

func NewSessionResositoryInMemory() *SessionRepositoryInMemory {
	return &SessionRepositoryInMemory{
		accessSessions:    make(map[string]domain.AccessSession),
		refreshSessions:   make(map[string]domain.RefreshSession),
		userDeviceAccess:  make(map[string]string),
		userDeviceRefresh: make(map[string]string),
	}
}

// Helper for user-device key
func deviceKey(userID, deviceID string) string {
	return fmt.Sprintf("%s:%s", userID, deviceID)
}

// GET SESSIONS
func (r *SessionRepositoryInMemory) GetAccessSessionByToken(ctx context.Context, token string) (*domain.AccessSession, error) {
	tokenHash := utils.HashToken(token)

	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.accessSessions[tokenHash]
	if !ok || time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session not found or expired")
	}
	return &session, nil
}

func (r *SessionRepositoryInMemory) GetRefreshSessionByToken(ctx context.Context, token string) (*domain.RefreshSession, error) {
	tokenHash := utils.HashToken(token)

	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.refreshSessions[tokenHash]
	if !ok || time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("refresh session not found or expired")
	}
	return &session, nil
}

// REVOKE OLD TOKENS
func (r *SessionRepositoryInMemory) revokeOldAccessToken(ctx context.Context, userID, deviceID string) {
	key := deviceKey(userID, deviceID)

	r.mu.Lock()
	defer r.mu.Unlock()

	if oldHash, ok := r.userDeviceAccess[key]; ok {
		delete(r.accessSessions, oldHash)
	}
}

func (r *SessionRepositoryInMemory) revokeOldRefreshToken(ctx context.Context, userID, deviceID string) {
	key := deviceKey(userID, deviceID)

	r.mu.Lock()
	defer r.mu.Unlock()

	if oldHash, ok := r.userDeviceRefresh[key]; ok {
		delete(r.refreshSessions, oldHash)
	}
}

// CREATE NEW TOKENS
func (r *SessionRepositoryInMemory) NewAccessToken(ctx context.Context, userID, deviceID string) (*domain.AccessSessionLight, error) {
	r.revokeOldAccessToken(ctx, userID, deviceID)

	token, tokenHash, err := GenerateTokenForSession()
	if err != nil {
		return nil, err
	}

	session := domain.AccessSession{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: tokenHash,
		DeviceID:  deviceID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ACCESS_TOKEN_EXPIRATION),
	}

	key := deviceKey(userID, deviceID)

	r.mu.Lock()
	r.accessSessions[tokenHash] = session
	r.userDeviceAccess[key] = tokenHash
	r.mu.Unlock()

	return &domain.AccessSessionLight{
		Token:     token,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

func (r *SessionRepositoryInMemory) NewRefreshToken(ctx context.Context, userID, deviceID string) (*domain.RefreshSessionLight, error) {
	r.revokeOldRefreshToken(ctx, userID, deviceID)

	token, tokenHash, err := GenerateTokenForSession()
	if err != nil {
		return nil, err
	}

	session := domain.RefreshSession{
		ID:        uuid.NewString(),
		UserID:    userID,
		TokenHash: tokenHash,
		DeviceID:  deviceID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(REFRESH_TOKEN_EXPIRATION),
	}

	key := deviceKey(userID, deviceID)

	r.mu.Lock()
	r.refreshSessions[tokenHash] = session
	r.userDeviceRefresh[key] = tokenHash
	r.mu.Unlock()

	return &domain.RefreshSessionLight{
		Token:     token,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

// DELETE SESSIONS
func (r *SessionRepositoryInMemory) DeleteAccessSession(ctx context.Context, tokenHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.accessSessions, tokenHash)
	return nil
}

func (r *SessionRepositoryInMemory) DeleteRefreshSession(ctx context.Context, tokenHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.refreshSessions, tokenHash)
	return nil
}
