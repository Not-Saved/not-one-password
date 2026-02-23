package bootstrap

import (
	"database/sql"
	"main/internal/adapters/repository"
)

type Repositories struct {
	User *repository.UserRepository
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		User: repository.NewUserRepository(db),
	}
}
