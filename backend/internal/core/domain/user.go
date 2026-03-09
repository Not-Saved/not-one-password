package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	PublicID     uuid.UUID
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
