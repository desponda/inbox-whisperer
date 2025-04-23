package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/desponda/inbox-whisperer/internal/api"
	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/mocks"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/desponda/inbox-whisperer/internal/session"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func setupTestRouterWithEmail(userTokens data.UserTokenRepository, svc service.EmailService) http.Handler {
	r := chi.NewRouter()
	r.Use(api.AuthMiddleware)
	r.Use(api.TokenMiddleware(userTokens))
	h := api.NewEmailHandler(svc, userTokens)
	r.Get("/api/email/messages", h.FetchMessagesHandler)
	r.Get("/api/email/messages/{id}", h.GetMessageContentHandler)
	return session.Middleware(r)
}

func TestEmailAPI_Integration_Unauthenticated(t *testing.T) {
	r := setupTestRouterWithEmail(&mocks.MockUserTokenRepository{}, &mocks.MockEmailService{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/email/messages", nil)
	r.ServeHTTP(w, req)
	resp := w.Result()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestEmailAPI_Integration_MissingToken(t *testing.T) {
	userTokens := &mockNoTokenUserTokenRepository{}
	r := setupTestRouterWithEmail(userTokens, &mocks.MockEmailService{})
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/email/messages", nil)
	ctx := context.WithValue(req.Context(), api.ContextUserIDKey, "user1")
	req = req.WithContext(ctx)
	r.ServeHTTP(w, req)
	resp := w.Result()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

type mockNoTokenUserTokenRepository struct{}

func (m *mockNoTokenUserTokenRepository) GetUserToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	return nil, nil
}
func (m *mockNoTokenUserTokenRepository) SaveUserToken(ctx context.Context, userID string, token *oauth2.Token) error {
	return nil
}

func TestEmailAPI_Integration_Success(t *testing.T) {
	userTokens := &mocks.MockUserTokenRepository{
	GetUserTokenFunc: func(ctx context.Context, userID string) (*oauth2.Token, error) {
		if userID == "user1" {
			return &oauth2.Token{AccessToken: "mock-token"}, nil
		}
		return nil, nil
	},
}
emailSvc := &mocks.MockEmailService{
	FetchMessagesFunc: func(ctx context.Context, token *oauth2.Token) ([]models.EmailMessage, error) {
		return []models.EmailMessage{{ID: 123, Subject: "Test"}}, nil
	},
}
r := setupTestRouterWithEmail(userTokens, emailSvc)
	w := httptest.NewRecorder()

	// Simulate a valid session using the session package helper
	req := httptest.NewRequest("GET", "/api/email/messages", nil)
	w2 := httptest.NewRecorder()
	session.SetSession(w2, req, "user1", "mock-token")
	// Copy the Set-Cookie header to the test request
	for _, c := range w2.Result().Cookies() {
		req.AddCookie(c)
	}

	r.ServeHTTP(w, req)
	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
