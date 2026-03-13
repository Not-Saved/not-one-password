package domain

import (
	"time"
)

type Vault struct {
	UserID    int32
	Vault     []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}
