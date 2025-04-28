package postgres

import (
	"context"
	"fmt"

	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Factory creates PostgreSQL session stores
type Factory struct{}

// NewFactory creates a new PostgreSQL store factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateStore creates a new PostgreSQL session store
func (f *Factory) CreateStore(config *session.StoreConfig) (session.Store, error) {
	if config.Type != "postgres" {
		return nil, fmt.Errorf("unsupported store type: %s", config.Type)
	}

	db, err := pgxpool.New(context.Background(), config.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	store, err := NewStore(db, "sessions", config.SessionDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	return store, nil
}
