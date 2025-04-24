package data

import (
	"context"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/models"
)

func TestUserRepository_CRUD(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()
	repo := db // *DB implements UserRepository
	ctx := context.Background()

	user := &models.User{
		ID:        "11111111-1111-1111-1111-111111111111",
		Email:     "user1@example.com",
		CreatedAt: time.Now(),
	}

	// Create
	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// GetByID
	got, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Email != user.Email {
		t.Errorf("expected email %q, got %q", user.Email, got.Email)
	}

	// Update
	user.Email = "updated@example.com"
	err = repo.Update(ctx, user)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	got, err = repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID after update failed: %v", err)
	}
	if got.Email != "updated@example.com" {
		t.Errorf("expected updated email, got %q", got.Email)
	}

	// List
	users, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	found := false
	for _, u := range users {
		if u.ID == user.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("user not found in List")
	}

	// Delete
	err = repo.Delete(ctx, user.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = repo.GetByID(ctx, user.ID)
	if err == nil {
		t.Errorf("expected error after delete, got nil")
	}
}
