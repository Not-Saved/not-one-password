package bootstrap

import (
	"database/sql"
	"main/internal/adapters/notifier"
	"main/internal/adapters/repository"
	"main/internal/core/ports"
	"main/internal/smtp"

	"github.com/go-redis/redis/v8"
)

type Adapters struct {
	ports.UserRepository
	ports.SessionRepository
	ports.VaultRepository
	ports.UserIntentRepository
	ports.UserNotifier
}

func NewAdapters(db *sql.DB, rdb *redis.Client, smtp *smtp.SMTPClient) *Adapters {
	return &Adapters{
		UserRepository:       repository.NewUserRepositoryPg(db),
		SessionRepository:    repository.NewSessionRepositoryRedis(rdb),
		VaultRepository:      repository.NewVaultRepositoryPg(db),
		UserIntentRepository: repository.NewUserIntentRepositoryRedis(rdb),
		UserNotifier:         notifier.NewUserNotifierSMTP(smtp),
	}
}
