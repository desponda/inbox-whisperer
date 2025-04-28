package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/desponda/inbox-whisperer/internal/auth/models"
	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store implements session.Store using PostgreSQL
type Store struct {
	pool            *pgxpool.Pool
	tableName       string
	sessionDuration time.Duration
}

// NewStore creates a new postgres-backed session store
func NewStore(pool *pgxpool.Pool, tableName string, sessionDuration time.Duration) (*Store, error) {
	if pool == nil {
		return nil, fmt.Errorf("pgx pool cannot be nil")
	}

	store := &Store{
		pool:            pool,
		tableName:       tableName,
		sessionDuration: sessionDuration,
	}

	if err := store.createTableIfNotExists(); err != nil {
		return nil, fmt.Errorf("failed to create sessions table: %w", err)
	}

	return store, nil
}

// Get retrieves a session by ID
func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	var (
		userID    string
		values    []byte
		createdAt time.Time
		expiresAt time.Time
	)

	query := fmt.Sprintf(`
		SELECT user_id, values, created_at, expires_at
		FROM %s
		WHERE id = $1 AND expires_at > NOW()
	`, s.tableName)

	err := s.pool.QueryRow(ctx, query, id).Scan(&userID, &values, &createdAt, &expiresAt)
	if err != nil {
		return nil, fmt.Errorf("session not found or expired")
	}

	// Unmarshal session values
	var sessionValues map[string]interface{}
	if err := json.Unmarshal(values, &sessionValues); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session values: %w", err)
	}

	// Create session with duration from now until expiresAt
	duration := expiresAt.Sub(time.Now())
	sess := models.NewSession(id, duration)
	sess.SetUserID(userID)
	for k, v := range sessionValues {
		sess.SetValue(k, v)
	}

	return sess, nil
}

// Save persists the session
func (s *Store) Save(ctx context.Context, sess session.Session) error {
	// Marshal session values
	values, err := json.Marshal(sess.Values())
	if err != nil {
		return fmt.Errorf("failed to marshal session values: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, user_id, values, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE
		SET user_id = $2,
			values = $3,
			expires_at = $5
	`, s.tableName)

	_, err = s.pool.Exec(ctx, query,
		sess.ID(),
		sess.UserID(),
		values,
		sess.CreatedAt(),
		sess.ExpiresAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// Delete removes a session
func (s *Store) Delete(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, s.tableName)
	_, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// Cleanup removes expired sessions
func (s *Store) Cleanup(ctx context.Context) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE expires_at <= NOW()`, s.tableName)
	_, err := s.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}

// Create creates a new session
func (s *Store) Create(ctx context.Context) (session.Session, error) {
	id := uuid.New().String()
	duration := s.sessionDuration
	sess := models.NewSession(id, duration)

	if err := s.Save(ctx, sess); err != nil {
		return nil, fmt.Errorf("failed to save new session: %w", err)
	}

	return sess, nil
}

func (s *Store) createTableIfNotExists() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36),
			values JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`, s.tableName)

	_, err := s.pool.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create session table: %w", err)
	}

	// Create index on expires_at for efficient cleanup
	indexQuery := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %s_expires_at_idx ON %s (expires_at)
	`, s.tableName, s.tableName)

	_, err = s.pool.Exec(context.Background(), indexQuery)
	if err != nil {
		return fmt.Errorf("failed to create expires_at index: %w", err)
	}

	return nil
}
