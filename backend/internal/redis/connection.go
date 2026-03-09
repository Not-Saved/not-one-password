package redis

import (
	"context"
	"fmt"
	"log"
	"main/internal/config"

	"github.com/go-redis/redis/v8"
)

func NewRedisConnection(cfg config.RedisConfig) *redis.Client {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: "", // no password if none set
		DB:       0,  // default DB
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	return rdb
}
