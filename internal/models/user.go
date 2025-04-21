package models

import "time"

// User represents a user in the system
// Extend with more fields as needed
// ID is a UUID string

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
