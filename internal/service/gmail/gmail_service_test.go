package gmail

import _ "github.com/jackc/pgx/v5/stdlib" // Register pgx as a database/sql driver for Postgres

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/desponda/inbox-whisperer/internal/models"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/data"

	"github.com/desponda/inbox-whisperer/internal/session"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	// "github.com/golang/mock/gomock"
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
	repo := data.NewEmailMessageRepositoryFromPool(db.Pool)
	t.Log("[DEBUG] NewEmailMessageRepositoryFromPool done")
	svc := NewGmailService(repo, &mockGmailAPI{})
	t.Log("[DEBUG] NewGmailService done")
	ctx := context.Background()
	userID := "user-e2e"

	// Simulate session context using handler+middleware pattern
	// For service tests, create a context with the userID manually (since session.SetSession is for HTTP)
	tok := mockToken()
	testCtx := session.ContextWithUserID(ctx, userID)
	type testCtxKey string
	testCtx = context.WithValue(testCtx, testCtxKey("_test_debug"), debug)

	msgID := "gmail_msg_123"
	msg := &models.EmailMessage{
		UserID:         userID,
		EmailMessageID: msgID,
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
	t.Logf("[DEBUG] First FetchMessageContent returned: %+v", content)
	if content.Subject != "Cached Subject" {
		t.Errorf("expected cached subject, got %s", content.Subject)
	}

	// Simulate staleness (older than TTL)
	staleTime := time.Now().Add(-2 * time.Minute)
	msg.CachedAt = staleTime
	t.Log("[DEBUG] Upserting stale cached message")
	err = repo.UpsertMessage(testCtx, msg)
	if err != nil {
		t.Fatalf("UpsertMessage for stale cache failed: %v", err)
	}
	t.Logf("[DEBUG] Cached message after staleness: %+v", msg)
	// t.Log("[DEBUG] Fetching message content after staleness (should NOT hit Gmail API in test)")
	// content2, err := svc.FetchMessageContent(testCtx, tok, msgID)
	// if err != nil {
	// 	t.Fatalf("FetchMessageContent after staleness failed: %v", err)
	// }
	// t.Logf("[DEBUG] Second FetchMessageContent returned: %+v", content2)
	// if content2.Body != "Cached Body" {
	// 	t.Errorf("expected cached body, got %s", content2.Body)
	// }
}

// mock for Get call

type mockGetCall struct {
	msg *gmail.Message
	err error
}

func (m *mockGetCall) Do(ctx context.Context) (*gmail.Message, error) {
	return m.msg, m.err
}

// mock for List call

type mockListCall struct {
	resp *gmail.ListMessagesResponse
	err  error
}

func (m *mockListCall) Do() (*gmail.ListMessagesResponse, error) {
	return m.resp, m.err
}

func TestExtractPlainTextBody(t *testing.T) {
	plain := "Hello, world!"
	encoded := base64.RawURLEncoding.EncodeToString([]byte(plain))
	payload := &gmail.MessagePart{
		MimeType: "text/plain",
		Body:     &gmail.MessagePartBody{Data: encoded},
	}
	if got := extractPlainTextBody(payload); got != plain {
		t.Errorf("expected plain body, got %q", got)
	}

	// Multipart: text/plain inside Parts
	encoded2 := base64.RawURLEncoding.EncodeToString([]byte("Second part"))
	multi := &gmail.MessagePart{
		MimeType: "multipart/alternative",
		Parts: []*gmail.MessagePart{
			{
				MimeType: "text/html",
				Body:     &gmail.MessagePartBody{Data: base64.RawURLEncoding.EncodeToString([]byte("<b>HTML</b>"))},
			},
			{
				MimeType: "text/plain",
				Body:     &gmail.MessagePartBody{Data: encoded2},
			},
		},
	}
	if got := extractPlainTextBody(multi); got != "Second part" {
		t.Errorf("expected 'Second part', got %q", got)
	}

	// Nil payload
	if got := extractPlainTextBody(nil); got != "" {
		t.Errorf("expected empty for nil payload, got %q", got)
	}
	// No valid plain text
	empty := &gmail.MessagePart{MimeType: "multipart/alternative", Parts: []*gmail.MessagePart{}}
	if got := extractPlainTextBody(empty); got != "" {
		t.Errorf("expected empty for no plain text, got %q", got)
	}
}

func TestGetHeader(t *testing.T) {
	headers := []*gmail.MessagePartHeader{
		{Name: "From", Value: "sender@example.com"},
		{Name: "To", Value: "rcpt@example.com"},
		{Name: "Subject", Value: "Test subject"},
		{Name: "Date", Value: "Wed, 01 Jan 2025 00:00:00 +0000"},
	}
	cases := []struct {
		name, key, want string
	}{
		{"Exact match", "From", "sender@example.com"},
		{"Case-insensitive", "subject", "Test subject"},
		{"Not found", "X-Header", ""},
	}
	for _, tc := range cases {
		if got := getHeader(headers, tc.key); got != tc.want {
			t.Errorf("%s: getHeader(%q) = %q, want %q", tc.name, tc.key, got, tc.want)
		}
	}
}

func TestGmailService_FetchMessages_Pagination(t *testing.T) {
	db, cleanup := data.SetupTestDB(t)
	defer cleanup()
	repo := data.NewEmailMessageRepositoryFromPool(db.Pool)
	mockAPI := &mockGmailAPI{}
	svc := NewGmailService(repo, mockAPI)
	ctx := context.Background()
	userID := "user-pagination"
	tok := mockToken()
	testCtx := session.ContextWithUserID(ctx, userID)

	// Insert 15 messages with descending InternalDate (newest first)
	var now = time.Now().Unix()
	for i := 0; i < 15; i++ {
		msg := &models.EmailMessage{
			UserID:         userID,
			EmailMessageID: fmt.Sprintf("msg_%02d", i),
			ThreadID:       fmt.Sprintf("thread_%02d", i),
			Subject:        fmt.Sprintf("Subject %02d", i),
			Sender:         "sender@example.com",
			Recipient:      "rcpt@example.com",
			Snippet:        fmt.Sprintf("Snippet %02d", i),
			Body:           fmt.Sprintf("Body %02d", i),
			InternalDate:   now - int64(i*60), // 1 min apart
			HistoryID:      int64(i),
			CachedAt:       time.Now(),
		}
		err := repo.UpsertMessage(testCtx, msg)
		if err != nil {
			t.Fatalf("failed to upsert msg %d: %v", i, err)
		}
	}

	// Prepare all messages for paging
	allMsgs := []*models.EmailMessage{}
	for i := 0; i < 15; i++ {
		msg, _ := repo.GetMessageByID(testCtx, userID, fmt.Sprintf("msg_%02d", i))
		allMsgs = append(allMsgs, msg)
	}
	// First page: 10 messages
	mockAPI.listResp = &gmail.ListMessagesResponse{Messages: []*gmail.Message{}}
	for i := 0; i < 10; i++ {
		m := allMsgs[i]
		mockAPI.listResp.Messages = append(mockAPI.listResp.Messages, &gmail.Message{
			Id: m.EmailMessageID,
			ThreadId: m.ThreadID,
			Snippet: m.Snippet,
			Payload: &gmail.MessagePart{Headers: []*gmail.MessagePartHeader{{Name: "Subject", Value: m.Subject}, {Name: "From", Value: m.Sender}, {Name: "Date", Value: time.Now().Format(time.RFC3339)}}},
			InternalDate: m.InternalDate,
			HistoryId: uint64(m.HistoryID),
		})
	}
	msgs, err := svc.FetchMessages(testCtx, tok)
	if err != nil {
		t.Fatalf("FetchMessages page 1 failed: %v", err)
	}
	if len(msgs) != 10 {
		t.Errorf("expected 10 messages, got %d", len(msgs))
	}
	last := msgs[len(msgs)-1]

	// Second page: next 5 messages
	mockAPI.listResp = &gmail.ListMessagesResponse{Messages: []*gmail.Message{}}
	for i := 10; i < 15; i++ {
		m := allMsgs[i]
		mockAPI.listResp.Messages = append(mockAPI.listResp.Messages, &gmail.Message{
			Id: m.EmailMessageID,
			ThreadId: m.ThreadID,
			Snippet: m.Snippet,
			Payload: &gmail.MessagePart{Headers: []*gmail.MessagePartHeader{{Name: "Subject", Value: m.Subject}, {Name: "From", Value: m.Sender}, {Name: "Date", Value: time.Now().Format(time.RFC3339)}}},
			InternalDate: m.InternalDate,
			HistoryId: uint64(m.HistoryID),
		})
	}
	ctx2 := context.WithValue(testCtx, CtxKeyAfterID{}, last.EmailMessageID)
	ctx2 = context.WithValue(ctx2, CtxKeyAfterInternalDate{}, last.InternalDate)
	err = svc.syncLatestSummariesFromGmail(ctx2, tok, "user1")
	if err != nil {
		t.Fatalf("syncLatestSummariesFromGmail page 2 failed: %v", err)
	}

	// Third page: no messages
	mockAPI.listResp = &gmail.ListMessagesResponse{Messages: []*gmail.Message{}}
	last2 := allMsgs[14]
	ctx3 := context.WithValue(testCtx, CtxKeyAfterID{}, last2.EmailMessageID)
	ctx3 = context.WithValue(ctx3, CtxKeyAfterInternalDate{}, last2.InternalDate)
	msgs3, err := svc.FetchMessages(ctx3, tok)
	if err != nil {
		t.Fatalf("FetchMessages page 3 failed: %v", err)
	}
	if len(msgs3) != 0 {
		t.Errorf("expected 0 messages on page 3, got %d", len(msgs3))
	}
}
