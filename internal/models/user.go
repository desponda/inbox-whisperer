package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
// Extend with more fields as needed
// ID is a UUID string

type User struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	Deactivated bool      `json:"deactivated"`
}

// UserIdentity represents an external auth provider identity
// (e.g. Google, GitHub, etc)
type UserIdentity struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	Provider       string    `json:"provider"`
	ProviderUserID string    `json:"provider_user_id"`
	Email          string    `json:"email"`
	CreatedAt      time.Time `json:"created_at"`
}
