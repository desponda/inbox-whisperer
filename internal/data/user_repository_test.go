package data

import (
	"context"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/google/uuid"
)

func TestUserRepository_CRUD(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()

	userID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	user := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	// Create
	err := db.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Get
	got, err := db.GetUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}
	if got.Email != user.Email {
		t.Errorf("expected email %q, got %q", user.Email, got.Email)
	}

	// Update
	user.Email = "updated@example.com"
	err = db.UpdateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}
	got, err = db.GetUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUser after update failed: %v", err)
	}
	if got.Email != "updated@example.com" {
		t.Errorf("expected updated email, got %q", got.Email)
	}

	// List
	users, err := db.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	found := false
	for _, u := range users {
		if u.ID == userID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("user not found in ListUsers")
	}

	// Deactivate
	err = db.DeactivateUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("DeactivateUser failed: %v", err)
	}
	got, err = db.GetUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUser after deactivate failed: %v", err)
	}
	if !got.Deactivated {
		t.Errorf("expected user to be deactivated")
	}

	// Delete
	err = db.Delete(context.Background(), userID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = db.GetUser(context.Background(), userID)
	if err == nil {
		t.Errorf("expected error after delete, got nil")
	}
}
