package session

import (
	"context"
	"net/http"
	"time"
)

// Session represents a user session
type Session interface {
	// ID returns the session ID
	ID() string

	// UserID returns the user ID
	UserID() string

	// SetUserID sets the user ID
	SetUserID(userID string)

	// CreatedAt returns the creation time
	CreatedAt() time.Time

	// ExpiresAt returns the expiration time
	ExpiresAt() time.Time

	// Values returns the session values
	Values() map[string]interface{}

	// SetValue sets a session value
	SetValue(key string, value interface{})

	// GetValue gets a session value
	GetValue(key string) (interface{}, bool)
}

// Store manages session persistence
type Store interface {
	// Create creates a new session
	Create(ctx context.Context) (Session, error)

	// Get retrieves an existing session
	Get(ctx context.Context, id string) (Session, error)

	// Save persists session data
	Save(ctx context.Context, session Session) error

	// Delete removes a session
	Delete(ctx context.Context, id string) error

	// Cleanup removes expired sessions
	Cleanup(ctx context.Context) error
}

// StoreConfig configures a session store
type StoreConfig struct {
	// Type indicates the store type (e.g., "postgres", "redis")
	Type string

	// ConnectionString is the store-specific connection string
	ConnectionString string

	// SessionDuration is how long sessions should live
	SessionDuration time.Duration
}

// StoreFactory creates session stores
type StoreFactory interface {
	// CreateStore creates a new session store with the given configuration
	CreateStore(config *StoreConfig) (Store, error)
}

// Manager handles session lifecycle
type Manager interface {
	// Start creates a new session or retrieves an existing one
	Start(w http.ResponseWriter, r *http.Request) (Session, error)

	// Destroy removes a session
	Destroy(w http.ResponseWriter, r *http.Request) error

	// Refresh extends the session lifetime
	Refresh(w http.ResponseWriter, r *http.Request) error

	// Store returns the underlying session store
	Store() Store
}
