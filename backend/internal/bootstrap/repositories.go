package bootstrap

import (
	"database/sql"
	"main/internal/adapters/repository"
	"main/internal/core/ports"
)

type Repositories struct {
	ports.UserRepository
	ports.SessionRepository
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		UserRepository:    repository.NewUserRepository(db),
		SessionRepository: repository.NewSessionRepository(db),
	}
}
