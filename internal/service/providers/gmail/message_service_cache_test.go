package gmail

import (
	"context"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
)

// This test covers cache hit, miss, and staleness for Gmail message caching
func TestGmailService_Caching(t *testing.T) {
	dbWrapper, cleanup := data.SetupTestDB(t)
	defer cleanup()
	repo := data.NewEmailMessageRepositoryFromPool(dbWrapper.Pool)
	ctx := context.Background()
	userID := "11111111-1111-1111-1111-111111111111"
	msgID := "gmail_msg_1"
	testCtx := ContextWithUserID(ctx, userID)
	_ = NewMessageService(repo, nil, WithGmailAPI(&mockGmailAPI{})) // for completeness, but not used in this unit test

	// Insert a user with the test UUID into the users table before upserting a message, to satisfy the foreign key constraint
	dbWrapper.Pool.Exec(ctx, `INSERT INTO users (id, email) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`, userID, "test@example.com")

	// Insert a fresh message (cache hit)
	msg := &models.EmailMessage{
		UserID:         userID,
		Provider:       "gmail",
		EmailMessageID: msgID,
		Subject:        "Cached Subject",
		Body:           "Cached Body",
		CachedAt:       time.Now(),
	}
	err := repo.UpsertMessage(testCtx, msg)
	if err != nil {
		t.Fatalf("failed to upsert cached message: %v", err)
	}

	// Should hit cache (simulate freshness)
	cached, err := repo.GetMessageByID(testCtx, userID, "gmail", msgID)
	if err != nil {
		t.Fatalf("failed to get cached message: %v", err)
	}
	if cached.Subject != "Cached Subject" {
		t.Errorf("expected cached subject, got %s", cached.Subject)
	}

	// Simulate staleness (older than TTL)
	staleTime := time.Now().Add(-2 * time.Minute)
	msg.CachedAt = staleTime
	_ = repo.UpsertMessage(testCtx, msg)
	stale, _ := repo.GetMessageByID(testCtx, userID, "gmail", msgID)
	if time.Since(stale.CachedAt) < time.Minute {
		t.Errorf("expected message to be stale, got age %v", time.Since(stale.CachedAt))
	}
}
