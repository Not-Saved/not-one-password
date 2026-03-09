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
}

func NewRepositories(db *sql.DB, redis *redis.Client) *Repositories {
	return &Repositories{
		UserRepository:    repository.NewUserRepositoryPg(db),
		SessionRepository: repository.NewSessionRepositoryRedis(redis),
	}
}
