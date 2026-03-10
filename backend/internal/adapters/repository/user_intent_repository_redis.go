package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"main/internal/core/domain"
	"main/internal/utils"
	"time"

	"github.com/go-redis/redis/v8"
)

type UserIntentRepositoryRedis struct {
	rdb *redis.Client
}

func NewUserIntentRepositoryRedis(rdb *redis.Client) *UserIntentRepositoryRedis {
	return &UserIntentRepositoryRedis{rdb: rdb}
}

const REGISTRATION_INTENT_EXPIRATION = 15 * time.Minute

func registrationIntentKey(code string) string {
	return fmt.Sprintf("registration_intent:%s", code)
}

func (r *UserIntentRepositoryRedis) CreateRegistrationIntent(ctx context.Context, user domain.RegistrationIntentUser) (*domain.RegistrationIntentToken, error) {
	code, err := utils.GenerateRandomString(32)
	if err != nil {
		return nil, err
	}

	user.Code = code
	key := registrationIntentKey(code)

	data, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	err = r.rdb.Set(ctx, key, data, REGISTRATION_INTENT_EXPIRATION).Err()
	if err != nil {
		return nil, err
	}

	return &domain.RegistrationIntentToken{
		Code: code,
	}, nil
}

func (r *UserIntentRepositoryRedis) GetRegistrationIntent(ctx context.Context, code string) (*domain.RegistrationIntentUser, error) {
	key := registrationIntentKey(code)

	// Get the JSON value
	data, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		// Key doesn't exist (expired or invalid code)
		return nil, fmt.Errorf("registration intent not found or expired")
	} else if err != nil {
		return nil, err
	}

	// Unmarshal JSON into struct
	var user domain.RegistrationIntentUser
	err = json.Unmarshal([]byte(data), &user)
	if err != nil {
		return nil, fmt.Errorf("failed to parse registration intent: %w", err)
	}

	return &user, nil
}

func (r *UserIntentRepositoryRedis) DeleteRegistrationIntent(ctx context.Context, code string) error {
	key := registrationIntentKey(code)
	_, err := r.rdb.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to delete registration intent: %w", err)
	}
	return nil
}
