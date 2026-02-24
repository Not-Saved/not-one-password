package domain

import "time"

// User represents the core domain model for a user,
// independent of any infrastructure or persistence details.
type User struct {
	ID           int32
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

// Session represents a user session in the domain layer.
type Session struct {
	ID        string
	UserID    int32
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt time.Time
	UserAgent string
	IpAddress string
}
