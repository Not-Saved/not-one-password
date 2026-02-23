package domain

import "time"

// User represents the core domain model for a user,
// independent of any infrastructure or persistence details.
type User struct {
	ID        int32
	Name      string
	Email     string
	CreatedAt time.Time
}
