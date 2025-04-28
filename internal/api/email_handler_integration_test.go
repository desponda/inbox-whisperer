package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/api/testutils"
	"github.com/desponda/inbox-whisperer/internal/auth/middleware"
	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/desponda/inbox-whisperer/internal/common"
	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/service/email"
	"github.com/desponda/inbox-whisperer/internal/service/provider"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// mockEmailService implements service.EmailService for testing
type mockEmailService struct {
	fetchMessagesErr       error
	fetchMessageContentErr error
	messages               []models.EmailMessage
}

func (m *mockEmailService) FetchMessages(ctx context.Context, token *oauth2.Token) ([]models.EmailMessage, error) {
	if m.fetchMessagesErr != nil {
		return nil, m.fetchMessagesErr
	}
	if m.messages != nil {
		return m.messages, nil
	}
	now := time.Now()
	return []models.EmailMessage{
		{
			EmailMessageID: "123",
			ThreadID:       "test-thread",
			Subject:        "Test Email",
			Sender:         "sender@example.com",
			Recipient:      "recipient@example.com",
			InternalDate:   now.Unix(),
			Date:           now.Format(time.RFC3339),
		},
	}, nil
}

func (m *mockEmailService) FetchMessageContent(ctx context.Context, token *oauth2.Token, messageID string) (*models.EmailMessage, error) {
	if m.fetchMessageContentErr != nil {
		return nil, m.fetchMessageContentErr
	}
	now := time.Now()
	return &models.EmailMessage{
		EmailMessageID: messageID,
		ThreadID:       "test-thread",
		Subject:        "Test Email",
		Sender:         "sender@example.com",
		Recipient:      "recipient@example.com",
		InternalDate:   now.Unix(),
		Date:           now.Format(time.RFC3339),
		Body:           "Test body",
		HTMLBody:       "<p>Test body</p>",
	}, nil
}

func (m *mockEmailService) FetchMessage(ctx context.Context, token *oauth2.Token, messageID string) (*models.EmailMessage, error) {
	return m.FetchMessageContent(ctx, token, messageID)
}

func (m *mockEmailService) FetchSummaries(ctx context.Context, token *oauth2.Token) ([]models.EmailSummary, error) {
	msgs, err := m.FetchMessages(ctx, token)
	if err != nil {
		return nil, err
	}
	var summaries []models.EmailSummary
	for _, msg := range msgs {
		summaries = append(summaries, models.EmailSummary{
			ID:           msg.EmailMessageID,
			ThreadID:     msg.ThreadID,
			Subject:      msg.Subject,
			Sender:       msg.Sender,
			Snippet:      msg.Snippet,
			InternalDate: msg.InternalDate,
			Date:         msg.Date,
			Provider:     msg.Provider,
		})
	}

	// Apply pagination logic as in production
	params := common.PaginationFromContext(ctx)
	filtered := make([]models.EmailSummary, 0, len(summaries))
	for _, summary := range summaries {
		if params.AfterTimestamp > 0 && summary.InternalDate < params.AfterTimestamp {
			continue
		}
		if params.AfterID != "" && summary.ID <= params.AfterID {
			continue
		}
		filtered = append(filtered, summary)
	}
	if params.Limit > 0 && len(filtered) > params.Limit {
		filtered = filtered[:params.Limit]
	}
	return filtered, nil
}

func setupTestRouter(t *testing.T, mockSvc *mockEmailService, tokenValue *oauth2.Token) (*chi.Mux, func(req *http.Request, w *httptest.ResponseRecorder), func()) {
	db, cleanup := data.SetupTestDB(t)

	// Create a mock store and session manager
	store := testutils.NewMockStore()
	sessionManager := testutils.NewMockSessionManager(func(w http.ResponseWriter, r *http.Request) (session.Session, error) {
		mockSession := testutils.NewMockSession(
			"test-session",
			"11111111-1111-1111-1111-111111111111",
			"",
			map[string]interface{}{
				"oauth_token": tokenValue,
			},
			time.Now(),
			time.Now().Add(24*time.Hour),
		)
		store.Save(r.Context(), mockSession)
		return mockSession, nil
	})
	sessionManager.StoreFunc = func() session.Store { return store }

	r := chi.NewRouter()
	emailHandler := NewEmailHandler(email.Service(mockSvc), sessionManager)

	sessionMiddleware := middleware.NewSessionMiddleware(sessionManager)
	oauthMiddleware := middleware.NewOAuthMiddleware(sessionManager, db, &oauth2.Config{})

	r.Group(func(r chi.Router) {
		r.Use(sessionMiddleware.Handler)
		r.Use(oauthMiddleware.Handler)
		r.Get("/api/emails", emailHandler.ListEmails)
		r.Get("/api/emails/{id}", emailHandler.GetEmail)
		r.Put("/api/emails/{id}", emailHandler.UpdateEmail)
		r.Delete("/api/emails/{id}", emailHandler.DeleteEmail)
	})

	// Helper to inject session cookie into the request
	injectSessionCookie := func(req *http.Request, w *httptest.ResponseRecorder) {
		// Start session to set cookie
		sessionManager.Start(w, req)
		for _, c := range w.Result().Cookies() {
			req.AddCookie(c)
		}
	}

	return r, injectSessionCookie, cleanup
}

func TestEmailHandlerIntegration(t *testing.T) {
	validToken := &oauth2.Token{
		AccessToken: "test-token",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	t.Run("list emails", func(t *testing.T) {
		t.Run("success with default pagination", func(t *testing.T) {
			mockSvc := &mockEmailService{}
			r, injectSessionCookie, cleanup := setupTestRouter(t, mockSvc, validToken)
			defer cleanup()

			req := httptest.NewRequest(http.MethodGet, "/api/emails", nil)
			w := httptest.NewRecorder()

			injectSessionCookie(req, w)
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, "50", w.Header().Get("X-Page-Limit")) // Default limit

			var response []models.EmailSummary
			require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
			require.Len(t, response, 1)
			require.Equal(t, "123", response[0].ID)
		})

		t.Run("with custom pagination", func(t *testing.T) {
			mockSvc := &mockEmailService{
				messages: []models.EmailMessage{
					{EmailMessageID: "1", Subject: "Email 1", InternalDate: 1000},
					{EmailMessageID: "2", Subject: "Email 2", InternalDate: 2000},
					{EmailMessageID: "3", Subject: "Email 3", InternalDate: 3000},
				},
			}
			r, injectSessionCookie, cleanup := setupTestRouter(t, mockSvc, validToken)
			defer cleanup()

			req := httptest.NewRequest(http.MethodGet, "/api/emails?limit=2&after_id=1&after_timestamp=1000", nil)
			w := httptest.NewRecorder()

			injectSessionCookie(req, w)
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, "2", w.Header().Get("X-Page-Limit"))
			require.Equal(t, "3", w.Header().Get("X-Next-After-Id"))
			require.Equal(t, "3000", w.Header().Get("X-Next-After-Timestamp"))

			var response []models.EmailSummary
			require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
			require.Len(t, response, 2)
			require.Equal(t, "2", response[0].ID)
			require.Equal(t, "3", response[1].ID)
		})

		t.Run("with invalid pagination parameters", func(t *testing.T) {
			mockSvc := &mockEmailService{}
			r, injectSessionCookie, cleanup := setupTestRouter(t, mockSvc, validToken)
			defer cleanup()

			req := httptest.NewRequest(http.MethodGet, "/api/emails?limit=invalid&after_timestamp=notanumber", nil)
			w := httptest.NewRecorder()

			injectSessionCookie(req, w)
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)                // Should still work with default values
			require.Equal(t, "50", w.Header().Get("X-Page-Limit")) // Should use default limit
		})

		t.Run("service error", func(t *testing.T) {
			mockSvc := &mockEmailService{
				fetchMessagesErr: errors.New("service error"),
			}
			r, injectSessionCookie, cleanup := setupTestRouter(t, mockSvc, validToken)
			defer cleanup()

			req := httptest.NewRequest(http.MethodGet, "/api/emails", nil)
			w := httptest.NewRecorder()

			injectSessionCookie(req, w)
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusInternalServerError, w.Code)
		})

		t.Run("expired token", func(t *testing.T) {
			expiredToken := &oauth2.Token{
				AccessToken: "expired",
				Expiry:      time.Now().Add(-time.Hour),
			}
			mockSvc := &mockEmailService{}
			r, injectSessionCookie, cleanup := setupTestRouter(t, mockSvc, expiredToken)
			defer cleanup()

			req := httptest.NewRequest(http.MethodGet, "/api/emails", nil)
			w := httptest.NewRecorder()

			injectSessionCookie(req, w)
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusUnauthorized, w.Code)
		})
	})

	t.Run("get email", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			mockSvc := &mockEmailService{}
			r, injectSessionCookie, cleanup := setupTestRouter(t, mockSvc, validToken)
			defer cleanup()

			req := httptest.NewRequest(http.MethodGet, "/api/emails/test-id", nil)
			w := httptest.NewRecorder()

			injectSessionCookie(req, w)
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)

			var response struct {
				ID      string `json:"id"`
				Subject string `json:"subject"`
				From    string `json:"from"`
				To      string `json:"to"`
				Date    string `json:"date"`
				Body    string `json:"body"`
			}
			require.NoError(t, json.NewDecoder(w.Body).Decode(&response))
			require.Equal(t, "test-id", response.ID)
			require.Equal(t, "Test Email", response.Subject)
		})

		t.Run("not found", func(t *testing.T) {
			mockSvc := &mockEmailService{
				fetchMessageContentErr: provider.ErrNotFound,
			}
			r, injectSessionCookie, cleanup := setupTestRouter(t, mockSvc, validToken)
			defer cleanup()

			req := httptest.NewRequest(http.MethodGet, "/api/emails/not-found", nil)
			w := httptest.NewRecorder()

			injectSessionCookie(req, w)
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusNotFound, w.Code)
		})

		t.Run("service error", func(t *testing.T) {
			mockSvc := &mockEmailService{
				fetchMessageContentErr: errors.New("service error"),
			}
			r, injectSessionCookie, cleanup := setupTestRouter(t, mockSvc, validToken)
			defer cleanup()

			req := httptest.NewRequest(http.MethodGet, "/api/emails/test-id", nil)
			w := httptest.NewRecorder()

			injectSessionCookie(req, w)
			r.ServeHTTP(w, req)

			require.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})

	t.Run("update email", func(t *testing.T) {
		mockSvc := &mockEmailService{}
		r, injectSessionCookie, cleanup := setupTestRouter(t, mockSvc, validToken)
		defer cleanup()

		req := httptest.NewRequest(http.MethodPut, "/api/emails/test-id", nil)
		w := httptest.NewRecorder()

		injectSessionCookie(req, w)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotImplemented, w.Code)
	})

	t.Run("delete email", func(t *testing.T) {
		mockSvc := &mockEmailService{}
		r, injectSessionCookie, cleanup := setupTestRouter(t, mockSvc, validToken)
		defer cleanup()

		req := httptest.NewRequest(http.MethodDelete, "/api/emails/test-id", nil)
		w := httptest.NewRecorder()

		injectSessionCookie(req, w)
		r.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotImplemented, w.Code)
	})
}

// Ensure mockEmailService implements email.Service
var _ email.Service = (*mockEmailService)(nil)

func TestMain(m *testing.M) {
	// Start the shared test DB container
	_, _ = data.SetupTestDB(nil)
	code := m.Run()
	// Stop the shared test DB container
	data.StopTestDB()
	os.Exit(code)
}
