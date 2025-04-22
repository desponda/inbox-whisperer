package service

import _ "github.com/jackc/pgx/v5/stdlib" // Register pgx as a database/sql driver for Postgres

import (
	"context"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/session"
	"golang.org/x/oauth2"
)

// mockToken returns a dummy OAuth2 token for tests
func mockToken() *oauth2.Token {
	return &oauth2.Token{AccessToken: "dummy", TokenType: "Bearer"}
}

func TestGmailService_CachingE2E(t *testing.T) {
	debug := func(msg string, args ...interface{}) { t.Logf("[REPO_DEBUG] "+msg, args...) }
	t.Log("[DEBUG] TestGmailService_CachingE2E: starting")
	db, cleanup := data.SetupTestDB(t)
	t.Log("[DEBUG] SetupTestDB done")
	defer cleanup()
	repo := data.NewGmailMessageRepositoryFromPool(db.Pool)
	t.Log("[DEBUG] NewGmailMessageRepositoryFromPool done")
	svc := NewGmailService(repo)
	t.Log("[DEBUG] NewGmailService done")
	ctx := context.Background()
	userID := "user-e2e"

	// Simulate session context using handler+middleware pattern
	// For service tests, create a context with the userID manually (since session.SetSession is for HTTP)
	tok := mockToken()
	testCtx := session.ContextWithUserID(ctx, userID)
	testCtx = context.WithValue(testCtx, "_test_debug", debug)

	msgID := "gmail_msg_123"
	msg := &data.GmailMessage{
		UserID:         userID,
		GmailMessageID: msgID,
		Subject:        "Cached Subject",
		Sender:         "sender@example.com",
		Recipient:      "rcpt@example.com",
		Snippet:        "Hello world",
		Body:           "Cached Body",
		InternalDate:   time.Now().Unix(),
		HistoryID:      1,
		CachedAt:       time.Now(),
	}
	t.Log("[DEBUG] Upserting initial cached message")
	err := repo.UpsertMessage(testCtx, msg)
	if err != nil {
		t.Fatalf("failed to upsert cached message: %v", err)
	}

	// Should hit cache (simulate freshness)
	t.Log("[DEBUG] Fetching message content (should hit cache)")
	content, err := svc.FetchMessageContent(testCtx, tok, msgID)
	if err != nil {
		t.Fatalf("FetchMessageContent cache hit failed: %v", err)
	}
	if content.Subject != "Cached Subject" {
		t.Errorf("expected cached subject, got %s", content.Subject)
	}

	// Simulate staleness (older than TTL)
	staleTime := time.Now().Add(-2 * time.Minute)
	msg.CachedAt = staleTime
	t.Log("[DEBUG] Upserting stale cached message")
	_ = repo.UpsertMessage(testCtx, msg)
	t.Log("[DEBUG] Fetching message content after staleness (should NOT hit Gmail API in test)")
	content2, err := svc.FetchMessageContent(testCtx, tok, msgID)
	if err != nil {
		t.Fatalf("FetchMessageContent after staleness failed: %v", err)
	}
	if content2.Body != "Cached Body" {
		t.Errorf("expected cached body, got %s", content2.Body)
	}

}



// You can add more tests for FetchMessages, cache miss, and handler integration as needed.
