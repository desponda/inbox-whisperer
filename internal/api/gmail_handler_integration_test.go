package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/api/mocks"
	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/desponda/inbox-whisperer/internal/session"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
	"github.com/golang/mock/gomock"
)

// Use unique struct name to avoid redeclaration conflict with gmail_handler_test.go
// Integration test only

type integrationMockUserTokenRepo struct {
	GetUserTokenFunc  func(ctx context.Context, userID string) (*oauth2.Token, error)
	SaveUserTokenFunc func(ctx context.Context, userID string, token *oauth2.Token) error
}

func (m *integrationMockUserTokenRepo) GetUserToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	if m.GetUserTokenFunc != nil {
		return m.GetUserTokenFunc(ctx, userID)
	}
	return nil, nil
}

func (m *integrationMockUserTokenRepo) SaveUserToken(ctx context.Context, userID string, token *oauth2.Token) error {
	if m.SaveUserTokenFunc != nil {
		return m.SaveUserTokenFunc(ctx, userID, token)
	}
	return nil
}

// GmailHandler and NewGmailHandler are defined in gmail_handler.go in the same package, so we can use them directly.
func setupIntegration(t *testing.T) (*GmailHandler, *data.DB, func()) {
	db, cleanup := data.SetupTestDB(t)
	repo := data.NewGmailMessageRepositoryFromPool(db.Pool)
	svc := service.NewGmailService(repo)
	userTokens := &integrationMockUserTokenRepo{
		GetUserTokenFunc: func(ctx context.Context, userID string) (*oauth2.Token, error) {
			return &oauth2.Token{AccessToken: "integration-token"}, nil
		},
	}
	h := NewGmailHandler(svc, userTokens)
	return h, db, cleanup
}

func insertTestMessages(t *testing.T, repo data.GmailMessageRepository, userID string, count int) {
	now := time.Now().Unix()
	t.Logf("[DEBUG] Inserting %d test messages for user %s", count, userID)
	for i := 0; i < count; i++ {
		msg := &data.GmailMessage{
			UserID:         userID,
			GmailMessageID:  fmt.Sprintf("msg_%02d", i),
			ThreadID:       fmt.Sprintf("thread_%02d", i),
			Subject:        fmt.Sprintf("Subject %02d", i),
			Sender:         "sender@example.com",
			Recipient:      "rcpt@example.com",
			Snippet:        fmt.Sprintf("Snippet %02d", i),
			Body:           fmt.Sprintf("Body %02d", i),
			InternalDate:   now - int64(i*60),
			HistoryID:      int64(i),
			CachedAt:       time.Now(),
		}
		t.Logf("[DEBUG] Upserting message: %+v", msg)
		err := repo.UpsertMessage(context.Background(), msg)
		if err != nil {
			t.Fatalf("failed to upsert msg: %v", err)
		}
	}
}

func TestGmailHandlerIntegration_FetchMessages_Pagination(t *testing.T) {
	h, db, cleanup := setupIntegration(t)
	defer cleanup()
	repo := data.NewGmailMessageRepositoryFromPool(db.Pool)
	userID := "integration-user"
	insertTestMessages(t, repo, userID, 12)

	r := chi.NewRouter()
	r.Use(session.Middleware)
	r.Get("/api/gmail/fetch", h.FetchMessagesHandler)

	// Simulate session
	req := httptest.NewRequest("GET", "/api/gmail/fetch", nil)
	w := httptest.NewRecorder()
	session.SetSession(w, req, userID, "integration-token")
	cookie := w.Result().Cookies()[0]
	req.AddCookie(cookie)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)
	resp := w2.Result()
	defer resp.Body.Close()
	t.Logf("[DEBUG] First page response status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}
	var msgs []service.MessageSummary
	if err := json.NewDecoder(resp.Body).Decode(&msgs); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	t.Logf("[DEBUG] First page messages: %+v", msgs)
	if len(msgs) != 10 {
		t.Errorf("expected 10 messages on first page, got %d. Messages: %+v", len(msgs), msgs)
	}
	// Fetch second page using cursor
	last := msgs[len(msgs)-1]
	t.Logf("[DEBUG] Cursor for second page: after_id=%s, after_internal_date=%d", last.CursorAfterID, last.CursorAfterInternalDate)
	req2 := httptest.NewRequest("GET", "/api/gmail/fetch?after_id="+last.CursorAfterID+"&after_internal_date="+fmt.Sprintf("%d", last.CursorAfterInternalDate), nil)
	req2.AddCookie(cookie)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req2)
	resp2 := w3.Result()
	defer resp2.Body.Close()
	t.Logf("[DEBUG] Second page response status: %d", resp2.StatusCode)
	if resp2.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp2.Body)
		t.Fatalf("expected status 200, got %d. Body: %s", resp2.StatusCode, string(body))
	}
	var msgs2 []service.MessageSummary
	if err := json.NewDecoder(resp2.Body).Decode(&msgs2); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	t.Logf("[DEBUG] Second page messages: %+v", msgs2)
	if len(msgs2) != 2 {
		t.Errorf("expected 2 messages on second page, got %d. Messages: %+v", len(msgs2), msgs2)
	}
}

func TestGmailHandlerIntegration_GetMessageContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockGmailServiceInterface(ctrl)
	userID := "integration-user"
	msgID := "msg_00"
	mockSvc.EXPECT().FetchMessageContent(gomock.Any(), gomock.Any(), msgID).Return(&service.MessageContent{
		ID:      msgID,
		Subject: "Subject 00",
		Body:    "Body 00",
	}, nil)

	userTokens := &integrationMockUserTokenRepo{
		GetUserTokenFunc: func(ctx context.Context, userID string) (*oauth2.Token, error) {
			return &oauth2.Token{AccessToken: "integration-token"}, nil
		},
	}
	h := NewGmailHandler(mockSvc, userTokens)

	r := chi.NewRouter()
	r.Use(session.Middleware)
	r.Get("/api/gmail/messages/{id}", h.GetMessageContentHandler)

	req := httptest.NewRequest("GET", "/api/gmail/messages/"+msgID, nil)
	w := httptest.NewRecorder()
	session.SetSession(w, req, userID, "integration-token")
	cookie := w.Result().Cookies()[0]
	req.AddCookie(cookie)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req)
	resp := w2.Result()
	defer resp.Body.Close()
	t.Logf("[DEBUG] GetMessageContent response status: %d", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 200, got %d. Body: %s", resp.StatusCode, string(body))
	}
	var msg service.MessageContent
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	t.Logf("[DEBUG] Message content: %+v", msg)
	if msg.ID != msgID {
		t.Errorf("expected message ID %s, got %s. Message: %+v", msgID, msg.ID, msg)
	}
}
