package domain

import (
	"time"
)

type AccessSession struct {
	ID        string
	UserID    int32
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt time.Time
	DeviceID  string
}

type RefreshSession struct {
	ID        string
	UserID    int32
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt time.Time
	DeviceID  string
}

type AccessSessionLight struct {
	Token     string
	ExpiresAt time.Time
}

type RefreshSessionLight struct {
	Token     string
	ExpiresAt time.Time
}
