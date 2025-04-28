package gmail

import "errors"

var (
	// ErrNotFound indicates a Gmail message was not found
	ErrNotFound = errors.New("not found")

	// ErrUserIDNotFound indicates user ID was not found in context
	ErrUserIDNotFound = errors.New("user ID not found in context")
)
