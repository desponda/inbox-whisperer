package models

import "time"

// UserToken represents a user's OAuth token (domain model)
type UserToken struct {
	UserID    string
	TokenJSON string
	UpdatedAt time.Time
}
