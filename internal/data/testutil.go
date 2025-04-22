package data

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/testcontainers/testcontainers-go"
)

// SetupTestDB starts a Postgres container and returns a DB and cleanup func
func SetupTestDB(t *testing.T) (*DB, func()) {
	log := func(msg string, args ...interface{}) {
		fmt.Printf("[SetupTestDB] "+msg+"\n", args...)
	}
	ctx := context.Background()
	log("Starting postgres test container...")
	pgContainer, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second)),
	)
	if err != nil {
		log("FAILED to start postgres container: %v", err)
		t.Fatalf("failed to start postgres container: %v", err)
	} else {
		log("Started postgres container successfully")
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	log("DSN: %s", dsn)
	if err != nil {
		log("FAILED to get connection string: %v", err)
		t.Fatalf("failed to get connection string: %v", err)
	}

	log("Connecting to Postgres test DB via pgxpool...")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log("FAILED to connect to test db: %v", err)
		t.Fatalf("failed to connect to test db: %v", err)
	} else {
		log("Connected to test DB via pgxpool")
	}

	// Create users table
	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		email TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL
	)`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	// Create user_tokens table (from migration)
	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS user_tokens (
		user_id TEXT PRIMARY KEY,
		token_json TEXT NOT NULL,
		updated_at TIMESTAMP NOT NULL DEFAULT now()
	)`)
	if err != nil {
		t.Fatalf("failed to create user_tokens table: %v", err)
	}

	// Create gmail_messages table (from migration)
	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS gmail_messages (
		id SERIAL PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		gmail_message_id VARCHAR(255) NOT NULL,
		thread_id VARCHAR(255),
		subject TEXT,
		sender TEXT,
		recipient TEXT,
		snippet TEXT,
		body TEXT,
		internal_date BIGINT,
		history_id BIGINT,
		cached_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		last_fetched_at TIMESTAMPTZ,
		category VARCHAR(64),
		categorization_confidence FLOAT,
		raw_json JSONB,
		UNIQUE(user_id, gmail_message_id)
	);
	CREATE INDEX IF NOT EXISTS idx_gmail_messages_user_msg ON gmail_messages(user_id, gmail_message_id);
	`)
	if err != nil {
		t.Fatalf("failed to create gmail_messages table: %v", err)
	}

	db := &DB{Pool: pool}
	cleanup := func() {
		pool.Close()
		_ = pgContainer.Terminate(ctx)
	}
	return db, cleanup
}
