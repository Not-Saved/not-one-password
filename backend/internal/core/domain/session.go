package domain

import "time"

type Session struct {
	ID        string
	User      SessionUser
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt time.Time
	UserAgent string
	IpAddress string
}

type SessionUser struct {
	ID    int32
	Name  string
	Email string
}

type SessionLight struct {
	Token     string
	ExpiresAt time.Time
}
