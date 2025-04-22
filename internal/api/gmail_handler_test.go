package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/desponda/inbox-whisperer/internal/session"
	"golang.org/x/oauth2"
)

// --- Mocks ---
type mockGmailService struct {
	FetchMessagesFunc      func(ctx context.Context, token *oauth2.Token) ([]service.MessageSummary, error)
	FetchMessageContentFunc func(ctx context.Context, token *oauth2.Token, id string) (*service.MessageContent, error)
}

func (m *mockGmailService) FetchMessages(ctx context.Context, token *oauth2.Token) ([]service.MessageSummary, error) {
	return m.FetchMessagesFunc(ctx, token)
}

func (m *mockGmailService) FetchMessageContent(ctx context.Context, token *oauth2.Token, id string) (*service.MessageContent, error) {
	return m.FetchMessageContentFunc(ctx, token, id)
}

type mockUserTokenRepo struct {
	GetUserTokenFunc  func(ctx context.Context, userID string) (*oauth2.Token, error)
	SaveUserTokenFunc func(ctx context.Context, userID string, token *oauth2.Token) error
}

func (m *mockUserTokenRepo) GetUserToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	if m.GetUserTokenFunc != nil {
		return m.GetUserTokenFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockUserTokenRepo) SaveUserToken(ctx context.Context, userID string, token *oauth2.Token) error {
	if m.SaveUserTokenFunc != nil {
		return m.SaveUserTokenFunc(ctx, userID, token)
	}
	return nil
}

// --- Test ---
func TestFetchMessagesHandler_TokenFromSession(t *testing.T) {
	// Arrange
	mockService := &mockGmailService{
		FetchMessagesFunc: func(ctx context.Context, token *oauth2.Token) ([]service.MessageSummary, error) {
			if token.AccessToken != "test-access-token" {
				t.Fatalf("expected access token 'test-access-token', got %v", token.AccessToken)
			}
			return []service.MessageSummary{{ID: "1", Snippet: "msg1"}, {ID: "2", Snippet: "msg2"}}, nil
		},
	}
	mockRepo := &mockUserTokenRepo{}
	h := &GmailHandler{
		Service:    mockService,
		UserTokens: mockRepo,
	}

	// Simulate session with token in context using session.SetSession and session.Middleware
	req := httptest.NewRequest("GET", "/api/gmail/fetch", nil)
	w := httptest.NewRecorder()
	session.SetSession(w, req, "user123", "test-access-token")
	// Copy the Set-Cookie header to the request
	cookie := w.Result().Cookies()[0]
	req.AddCookie(cookie)
	w2 := httptest.NewRecorder()
	handler := session.Middleware(http.HandlerFunc(h.FetchMessagesHandler))

	// Act
	handler.ServeHTTP(w2, req)
	resp := w2.Result()
	defer resp.Body.Close()

	// Assert
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	var msgs []service.MessageSummary
	if err := json.NewDecoder(resp.Body).Decode(&msgs); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(msgs) != 2 || msgs[0].Snippet != "msg1" || msgs[1].Snippet != "msg2" {
		t.Errorf("unexpected messages: %v", msgs)
	}
}

// More tests can be added for token-from-repo, error cases, etc.

func TestGetMessageContentHandler_TokenFromSession(t *testing.T) {
	mockService := &mockGmailService{
		FetchMessagesFunc: func(ctx context.Context, token *oauth2.Token) ([]service.MessageSummary, error) {
			return nil, nil
		},
		FetchMessageContentFunc: func(ctx context.Context, token *oauth2.Token, id string) (*service.MessageContent, error) {
			if token.AccessToken != "test-access-token" {
				t.Fatalf("expected access token 'test-access-token', got %v", token.AccessToken)
			}
			if id != "abc123" {
				t.Fatalf("expected id 'abc123', got %v", id)
			}
			return &service.MessageContent{
				ID:      "abc123",
				Subject: "Test Subject",
				From:    "from@example.com",
				To:      "to@example.com",
				Date:    "2025-04-22T00:00:00Z",
				Body:    "Hello, world!",
			}, nil
		},
	}
	mockRepo := &mockUserTokenRepo{}
	h := &GmailHandler{
		Service:    mockService,
		UserTokens: mockRepo,
	}

	// Simulate session with token in context using session.SetSession and session.Middleware
	req := httptest.NewRequest("GET", "/api/gmail/messages/abc123", nil)
	w := httptest.NewRecorder()
	session.SetSession(w, req, "user123", "test-access-token")
	cookie := w.Result().Cookies()[0]
	req.AddCookie(cookie)

	// Use chi router to properly set URL param
	r := chi.NewRouter()
	r.Use(session.Middleware)
	r.Get("/api/gmail/messages/{id}", h.GetMessageContentHandler)

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)
	resp := w2.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	var msg service.MessageContent
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if msg.ID != "abc123" || msg.Subject != "Test Subject" {
		t.Errorf("unexpected message content: %+v", msg)
	}
}
