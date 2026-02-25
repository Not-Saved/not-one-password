package domain

import "time"

type User struct {
	ID           int32
	Name         string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
