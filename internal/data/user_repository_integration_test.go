//go:build integration
// +build integration

package data

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupTestDB(t *testing.T) (*DB, func()) {
	ctx := context.Background()
	pgContainer, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
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

	db := &DB{Pool: pool}
	cleanup := func() {
		pool.Close()
		_ = pgContainer.Terminate(ctx)
	}
	return db, cleanup
}

func TestUserRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if os.Getenv("SKIP_DB_INTEGRATION") == "1" {
		t.Skip("skipping DB integration test by env")
	}
	db, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	user := &models.User{
		ID:        uuid.NewString(),
		Email:     "test@example.com",
		CreatedAt: time.Now().UTC(),
	}

	// Create
	err := db.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// GetByID
	got, err := db.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Email != user.Email || got.ID != user.ID {
		t.Errorf("got %+v, want %+v", got, user)
	}
}
