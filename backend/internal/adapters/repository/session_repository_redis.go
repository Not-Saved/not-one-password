package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"main/internal/core/domain"
	"main/internal/utils"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type SessionRepositoryRedis struct {
	rdb *redis.Client
}

func NewSessionRepositoryRedis(rdb *redis.Client) *SessionRepositoryRedis {
	return &SessionRepositoryRedis{rdb: rdb}
}

// Redis keys
func accessSessionKey(tokenHash string) string {
	return fmt.Sprintf("session:access:%s", tokenHash)
}

func refreshSessionKey(tokenHash string) string {
	return fmt.Sprintf("session:refresh:%s", tokenHash)
}

func userDeviceAccessKey(userID, deviceID string) string {
	return fmt.Sprintf("user:%s:device:%s:access", userID, deviceID)
}

func userDeviceRefreshKey(userID, deviceID string) string {
	return fmt.Sprintf("user:%s:device:%s:refresh", userID, deviceID)
}

//
// ADD SESSIONS
//

func (r *SessionRepositoryRedis) addAccessSession(ctx context.Context, session domain.AccessSession) error {
	key := accessSessionKey(session.TokenHash)

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return r.rdb.Set(ctx, key, data, ACCESS_TOKEN_EXPIRATION).Err()
}

func (r *SessionRepositoryRedis) addRefreshSession(ctx context.Context, session domain.RefreshSession) error {
	key := refreshSessionKey(session.TokenHash)

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	return r.rdb.Set(ctx, key, data, REFRESH_TOKEN_EXPIRATION).Err()
}

//
// GET SESSION
//

func (r *SessionRepositoryRedis) GetAccessSessionByToken(ctx context.Context, token string) (*domain.AccessSession, error) {
	tokenHash := utils.HashToken(token)

	data, err := r.rdb.Get(ctx, accessSessionKey(tokenHash)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, err
	}

	var session domain.AccessSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *SessionRepositoryRedis) GetRefreshSessionByToken(ctx context.Context, token string) (*domain.RefreshSession, error) {
	tokenHash := utils.HashToken(token)

	data, err := r.rdb.Get(ctx, refreshSessionKey(tokenHash)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("refresh session not found")
		}
		return nil, err
	}

	var session domain.RefreshSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// REVOKE OLD TOKENS
// Revoke old access token for user+device
func (r *SessionRepositoryRedis) revokeOldAccessToken(ctx context.Context, userID, deviceID string) error {
	if oldHash, err := r.rdb.Get(ctx, userDeviceAccessKey(userID, deviceID)).Result(); err == nil {
		return r.rdb.Del(ctx, accessSessionKey(oldHash)).Err()
	}
	return nil
}

// Revoke old refresh token for user+device
func (r *SessionRepositoryRedis) revokeOldRefreshToken(ctx context.Context, userID, deviceID string) error {
	if oldHash, err := r.rdb.Get(ctx, userDeviceRefreshKey(userID, deviceID)).Result(); err == nil {
		return r.rdb.Del(ctx, refreshSessionKey(oldHash)).Err()
	}
	return nil
}

//
// CREATE NEW TOKENS WITH REVOCATION
//

func (r *SessionRepositoryRedis) NewAccessToken(ctx context.Context, userID, deviceID string) (*domain.AccessSessionLight, error) {
	if err := r.revokeOldAccessToken(ctx, userID, deviceID); err != nil {
		return nil, err
	}

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

	if err := r.addAccessSession(ctx, session); err != nil {
		return nil, err
	}

	// Store pointer for user+device
	if err := r.rdb.Set(ctx, userDeviceAccessKey(userID, deviceID), session.TokenHash, ACCESS_TOKEN_EXPIRATION).Err(); err != nil {
		return nil, err
	}

	return &domain.AccessSessionLight{
		Token:     token,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

func (r *SessionRepositoryRedis) NewRefreshToken(ctx context.Context, userID, deviceID string) (*domain.RefreshSessionLight, error) {
	if err := r.revokeOldRefreshToken(ctx, userID, deviceID); err != nil {
		return nil, err
	}

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

	if err := r.addRefreshSession(ctx, session); err != nil {
		return nil, err
	}

	// Store pointer for user+device
	if err := r.rdb.Set(ctx, userDeviceRefreshKey(userID, deviceID), session.TokenHash, REFRESH_TOKEN_EXPIRATION).Err(); err != nil {
		return nil, err
	}

	return &domain.RefreshSessionLight{
		Token:     token,
		ExpiresAt: session.ExpiresAt,
	}, nil
}

//
// DELETE SESSIONS
//

func (r *SessionRepositoryRedis) DeleteAccessSession(ctx context.Context, userID, deviceID string) error {
	return r.revokeOldAccessToken(ctx, userID, deviceID)
}

func (r *SessionRepositoryRedis) DeleteRefreshSession(ctx context.Context, userID, deviceID string) error {
	return r.revokeOldRefreshToken(ctx, userID, deviceID)
}
