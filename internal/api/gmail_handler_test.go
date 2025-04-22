package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/desponda/inbox-whisperer/internal/session"
	"golang.org/x/oauth2"
)

// --- Mocks ---
type mockGmailService struct {
	FetchMessagesFunc func(ctx context.Context, token *oauth2.Token) ([]service.MessageSummary, error)
}

func (m *mockGmailService) FetchMessages(ctx context.Context, token *oauth2.Token) ([]service.MessageSummary, error) {
	return m.FetchMessagesFunc(ctx, token)
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
