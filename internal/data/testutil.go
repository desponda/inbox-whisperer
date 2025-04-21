package data

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/testcontainers/testcontainers-go"
)

// SetupTestDB starts a Postgres container and returns a DB and cleanup func
func SetupTestDB(t *testing.T) (*DB, func()) {
	ctx := context.Background()
	pgContainer, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// Create users table
	_, err = pool.Exec(ctx, `CREATE TABLE users (
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

	db := &DB{Pool: pool}
	cleanup := func() {
		pool.Close()
		_ = pgContainer.Terminate(ctx)
	}
	return db, cleanup
}
