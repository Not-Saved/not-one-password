package bootstrap

import (
	"database/sql"
	"main/internal/adapters/repository"
	"main/internal/core/ports"

	"github.com/go-redis/redis/v8"
)

type Repositories struct {
	ports.UserRepository
	ports.SessionRepository
	ports.UserIntentRepository
}

func NewRepositories(db *sql.DB, rdb *redis.Client) *Repositories {
	return &Repositories{
		UserRepository:       repository.NewUserRepositoryPg(db),
		SessionRepository:    repository.NewSessionRepositoryRedis(rdb),
		UserIntentRepository: repository.NewUserIntentRepositoryRedis(rdb),
	}
}
