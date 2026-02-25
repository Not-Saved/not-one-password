package domain

import "time"

type Session struct {
	ID        string
	UserID    int32
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt time.Time
	UserAgent string
	IpAddress string
}
