package data

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/models"
)

func TestEmailMessageRepository_CRUD(t *testing.T) {
	db, cleanup := SetupTestDB(t)
	defer cleanup()
	repo := NewEmailMessageRepositoryFromPool(db.Pool)
	ctx := context.Background()

	msg := &models.EmailMessage{
		UserID:                   "user-uuid-1",
		EmailMessageID:           "msg-1",
		ThreadID:                 "thread-1",
		Subject:                  "Test Subject",
		Sender:                   "sender@example.com",
		Recipient:                "rcpt@example.com",
		Snippet:                  "Snippet",
		Body:                     "Body",
		InternalDate:             time.Now().Unix(),
		HistoryID:                1,
		CachedAt:                 time.Now(),
		LastFetchedAt:            sql.NullTime{Time: time.Now(), Valid: true},
		Category:                 sql.NullString{String: "inbox", Valid: true},
		CategorizationConfidence: sql.NullFloat64{Float64: 0.99, Valid: true},
		RawJSON:                  []byte("{}"),
	}

	// Upsert (Create)
	err := repo.UpsertMessage(ctx, msg)
	if err != nil {
		t.Fatalf("UpsertMessage failed: %v", err)
	}

	// Get by ID
	got, err := repo.GetMessageByID(ctx, msg.UserID, msg.EmailMessageID)
	if err != nil {
		t.Fatalf("GetMessageByID failed: %v", err)
	}
	if got.Subject != msg.Subject {
		t.Errorf("expected subject %q, got %q", msg.Subject, got.Subject)
	}

	// Update
	msg.Subject = "Updated Subject"
	err = repo.UpsertMessage(ctx, msg)
	if err != nil {
		t.Fatalf("UpsertMessage (update) failed: %v", err)
	}
	got, err = repo.GetMessageByID(ctx, msg.UserID, msg.EmailMessageID)
	if err != nil {
		t.Fatalf("GetMessageByID after update failed: %v", err)
	}
	if got.Subject != "Updated Subject" {
		t.Errorf("expected updated subject, got %q", got.Subject)
	}

	// List for user
	msgs, err := repo.GetMessagesForUser(ctx, msg.UserID, 10, 0)
	if err != nil {
		t.Fatalf("GetMessagesForUser failed: %v", err)
	}
	if len(msgs) == 0 {
		t.Errorf("expected at least 1 message, got 0")
	}

	// List with cursor
	msgsCursor, err := repo.GetMessagesForUserCursor(ctx, msg.UserID, 10, 0, "")
	if err != nil {
		t.Fatalf("GetMessagesForUserCursor failed: %v", err)
	}
	if len(msgsCursor) == 0 {
		t.Errorf("expected at least 1 message with cursor, got 0")
	}

	// Delete for user
	err = repo.DeleteMessagesForUser(ctx, msg.UserID)
	if err != nil {
		t.Fatalf("DeleteMessagesForUser failed: %v", err)
	}
	msgs, err = repo.GetMessagesForUser(ctx, msg.UserID, 10, 0)
	if err != nil {
		t.Fatalf("GetMessagesForUser after delete failed: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages after delete, got %d", len(msgs))
	}
}
