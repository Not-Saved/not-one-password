package domain

import "time"

type Session struct {
	ID        string
	UserID    int32
	UserName  string
	UserEmail string
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt time.Time
	UserAgent string
	IpAddress string
}

type SessionLight struct {
	Token     string
	ExpiresAt time.Time
}
