package gorilla

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
)

// Serializer defines how session values are marshaled and unmarshaled
// for storage in the database.
type Serializer interface {
	Marshal(v map[string]interface{}) ([]byte, error)
	Unmarshal(data []byte) (map[string]interface{}, error)
}

// JSONSerializer implements Serializer using encoding/json.
type JSONSerializer struct{}

func (j *JSONSerializer) Marshal(v map[string]interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (j *JSONSerializer) Unmarshal(data []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// Store implements session.Store using a SQL database
type Store struct {
	db         *sql.DB
	tableName  string
	maxAge     time.Duration
	keyPrefix  string
	serializer Serializer
	Gorilla    sessions.Store
}

// StoreConfig configures the store
type StoreConfig struct {
	DB         *sql.DB
	TableName  string
	MaxAge     time.Duration
	KeyPrefix  string
	Serializer Serializer
}

// NewStore creates a new store
func NewStore(config StoreConfig) (*Store, error) {
	if config.DB == nil {
		return nil, fmt.Errorf("db is required")
	}
	if config.TableName == "" {
		config.TableName = "sessions"
	}
	if config.MaxAge == 0 {
		config.MaxAge = 24 * time.Hour
	}
	if config.Serializer == nil {
		config.Serializer = &JSONSerializer{}
	}

	store := &Store{
		db:         config.DB,
		tableName:  config.TableName,
		maxAge:     config.MaxAge,
		keyPrefix:  config.KeyPrefix,
		serializer: config.Serializer,
		Gorilla:    sessions.NewCookieStore([]byte("secret-key")),
	}

	if err := store.createTable(); err != nil {
		log.Error().
			Err(err).
			Str("table", config.TableName).
			Msg("Failed to create session table")
		return nil, err
	}

	log.Info().
		Str("table", config.TableName).
		Msg("Session table created or verified")
	return store, nil
}

// Create creates a new session
func (s *Store) Create(ctx context.Context) (session.Session, error) {
	sess := NewSession(generateID(), "", make(map[string]interface{}), time.Now(), time.Now().Add(s.maxAge))
	if err := s.Save(ctx, sess); err != nil {
		return nil, err
	}
	return sess, nil
}

// Get retrieves a session by ID
func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	log.Debug().
		Str("session_id", id).
		Msg("Retrieving session")

	var (
		userID    string
		values    []byte
		createdAt time.Time
		expiresAt time.Time
	)

	query := fmt.Sprintf("SELECT user_id, values, created_at, expires_at FROM %s WHERE id = $1 AND expires_at > $2", s.tableName)
	err := s.db.QueryRowContext(ctx, query, id, time.Now()).Scan(&userID, &values, &createdAt, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug().
				Str("session_id", id).
				Msg("Session not found or expired")
			return nil, fmt.Errorf("session not found")
		}
		return nil, err
	}

	var sessionValues map[string]interface{}
	if err := json.Unmarshal(values, &sessionValues); err != nil {
		log.Error().
			Err(err).
			Str("session_id", id).
			Msg("Failed to unmarshal session values")
		return nil, err
	}

	sess := NewSession(id, userID, sessionValues, createdAt, expiresAt)

	log.Debug().
		Str("session_id", id).
		Str("user_id", userID).
		Msg("Session retrieved successfully")
	return sess, nil
}

// Save persists a session
func (s *Store) Save(ctx context.Context, sess session.Session) error {
	// Cast to *Session if possible
	gsess, ok := sess.(*Session)
	if !ok {
		return fmt.Errorf("invalid session type for gorilla store")
	}
	values, err := json.Marshal(gsess.Values())
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, user_id, values, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE
		SET user_id = $2,
			values = $3,
			expires_at = $5
	`, s.tableName)

	_, err = s.db.ExecContext(ctx, query,
		gsess.ID(),
		gsess.UserID(),
		values,
		gsess.CreatedAt(),
		gsess.ExpiresAt(),
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", gsess.ID()).
			Msg("Failed to save session")
		return err
	}

	log.Debug().
		Str("session_id", gsess.ID()).
		Msg("Session saved successfully")
	return nil
}

// Delete removes a session
func (s *Store) Delete(ctx context.Context, id string) error {
	log.Debug().
		Str("session_id", id).
		Msg("Deleting session")

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", s.tableName)
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", id).
			Msg("Failed to delete session")
		return err
	}

	log.Debug().
		Str("session_id", id).
		Msg("Session deleted successfully")
	return nil
}

// Cleanup removes expired sessions
func (s *Store) Cleanup(ctx context.Context) error {
	log.Debug().Msg("Cleaning up expired sessions")

	query := fmt.Sprintf("DELETE FROM %s WHERE expires_at <= $1", s.tableName)
	result, err := s.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to cleanup expired sessions")
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	log.Info().
		Int64("count", rowsAffected).
		Msg("Expired sessions cleaned up")
	return nil
}

// createTable creates the sessions table if it doesn't exist
func (s *Store) createTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id VARCHAR(255) PRIMARY KEY,
			user_id VARCHAR(255),
			values JSONB,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL
		)
	`, s.tableName)

	_, err := s.db.Exec(query)
	return err
}

// generateID returns a new unique session ID
func generateID() string {
	return uuid.New().String()
}

// NewSession constructs a new *Session for the gorilla store
func NewSession(id, userID string, values map[string]interface{}, createdAt, expiresAt time.Time) *Session {
	// Convert values to map[interface{}]interface{}
	internalValues := make(map[interface{}]interface{}, len(values))
	for k, v := range values {
		internalValues[k] = v
	}
	return &Session{
		id:        id,
		userID:    userID,
		createdAt: createdAt,
		expiresAt: expiresAt,
		values:    internalValues,
	}
}
