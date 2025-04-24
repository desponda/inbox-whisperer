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
)

func TestUserRepository_Integration(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if os.Getenv("SKIP_DB_INTEGRATION") == "1" {
		t.Skip("skipping DB integration test by env")
	}

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
