package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/api/testutils"
	"github.com/desponda/inbox-whisperer/internal/auth/middleware"
	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/desponda/inbox-whisperer/internal/mocks"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/service/email"
	"github.com/desponda/inbox-whisperer/internal/service/provider"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// mockTokenRepo implements data.UserTokenRepository for testing
type mockTokenRepo struct {
	tokens map[string]*oauth2.Token
}

func newMockTokenRepo() *mockTokenRepo {
	return &mockTokenRepo{
		tokens: make(map[string]*oauth2.Token),
	}
}

func (m *mockTokenRepo) SaveUserToken(ctx context.Context, userID string, token *oauth2.Token) error {
	m.tokens[userID] = token
	return nil
}

func (m *mockTokenRepo) GetUserToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	if token, ok := m.tokens[userID]; ok {
		return token, nil
	}
	return nil, nil
}

// setupTestSessionManager creates a mock session manager for testing
func setupTestSessionManager() (session.Manager, session.Store) {
	store := testutils.NewMockStore()
	sessionManager := testutils.NewMockSessionManager(func(w http.ResponseWriter, r *http.Request) (session.Session, error) {
		sess := testutils.NewMockSession(
			"test-session",
			"11111111-1111-1111-1111-111111111111",
			"",
			map[string]interface{}{
				"oauth_token": &oauth2.Token{
					AccessToken:  "test-token",
					TokenType:    "Bearer",
					RefreshToken: "refresh-token",
					Expiry:       time.Now().Add(time.Hour),
				},
			},
			time.Now(),
			time.Now().Add(time.Hour),
		)
		store.Save(r.Context(), sess)
		return sess, nil
	})
	sessionManager.StoreFunc = func() session.Store {
		return store
	}
	return sessionManager, store
}

// setupTestOAuthConfig creates a test OAuth config
func setupTestOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/gmail.readonly",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost/auth",
			TokenURL: "http://localhost/token",
		},
	}
}

// setupTestRequest creates a request with session and OAuth middleware applied
func setupTestRequest(method, path string, sessionManager session.Manager) (*http.Request, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()

	// Create token repository and OAuth config
	tokenRepo := newMockTokenRepo()
	oauthConfig := setupTestOAuthConfig()

	// Save a test token
	tokenRepo.SaveUserToken(r.Context(), "11111111-1111-1111-1111-111111111111", &oauth2.Token{
		AccessToken:  "test-token",
		TokenType:    "Bearer",
		RefreshToken: "refresh-token",
		Expiry:       time.Now().Add(time.Hour),
	})

	// Apply session middleware
	sessionMiddleware := middleware.NewSessionMiddleware(sessionManager)
	oauthMiddleware := middleware.NewOAuthMiddleware(sessionManager, tokenRepo, oauthConfig)

	// Capture the request as modified by the middleware chain
	var reqWithCtx *http.Request
	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		reqWithCtx = req
	})
	handler := sessionMiddleware.Handler(oauthMiddleware.Handler(next))
	handler.ServeHTTP(w, r)

	return reqWithCtx, w
}

func TestListEmails_Success(t *testing.T) {
	factory := provider.NewProviderFactory()
	mockProvider := &mocks.MockEmailService{
		FetchMessagesFunc: func(ctx context.Context, token *oauth2.Token) ([]models.EmailMessage, error) {
			require.Equal(t, "test-token", token.AccessToken)
			return []models.EmailMessage{{EmailMessageID: "123", Subject: "Test"}}, nil
		},
	}
	factory.RegisterProvider(provider.Gmail, func(cfg provider.Config) (provider.Provider, error) {
		return &dummyProvider{mockProvider}, nil
	})
	factory.LinkProvider("11111111-1111-1111-1111-111111111111", provider.Config{UserID: "11111111-1111-1111-1111-111111111111", Type: provider.Gmail})
	svc := email.NewMultiProviderService(factory)
	sessionManager, _ := setupTestSessionManager()
	h := NewEmailHandler(svc, sessionManager)

	r, w := setupTestRequest("GET", "/api/emails", sessionManager)
	h.ListEmails(w, r)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestListEmails_ServiceError(t *testing.T) {
	factory := provider.NewProviderFactory()
	mockProvider := &mocks.MockEmailService{
		FetchMessagesFunc: func(ctx context.Context, token *oauth2.Token) ([]models.EmailMessage, error) {
			return nil, provider.ErrNotFound
		},
	}
	factory.RegisterProvider(provider.Gmail, func(cfg provider.Config) (provider.Provider, error) {
		return &dummyProvider{mockProvider}, nil
	})
	factory.LinkProvider("11111111-1111-1111-1111-111111111111", provider.Config{UserID: "11111111-1111-1111-1111-111111111111", Type: provider.Gmail})
	svc := email.NewMultiProviderService(factory)
	sessionManager, _ := setupTestSessionManager()
	h := NewEmailHandler(svc, sessionManager)

	r, w := setupTestRequest("GET", "/api/emails", sessionManager)
	h.ListEmails(w, r)

	resp := w.Result()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestListEmails_ExpiredToken(t *testing.T) {
	factory := provider.NewProviderFactory()
	mockProvider := &mocks.MockEmailService{
		FetchMessagesFunc: func(ctx context.Context, token *oauth2.Token) ([]models.EmailMessage, error) {
			require.NotEqual(t, "expired-token", token.AccessToken)
			return []models.EmailMessage{{EmailMessageID: "123", Subject: "Test"}}, nil
		},
	}
	factory.RegisterProvider(provider.Gmail, func(cfg provider.Config) (provider.Provider, error) {
		return &dummyProvider{mockProvider}, nil
	})
	factory.LinkProvider("11111111-1111-1111-1111-111111111111", provider.Config{UserID: "11111111-1111-1111-1111-111111111111", Type: provider.Gmail})
	svc := email.NewMultiProviderService(factory)
	sessionManager, _ := setupTestSessionManager()
	h := NewEmailHandler(svc, sessionManager)

	// Create request with expired token
	r := httptest.NewRequest("GET", "/api/emails", nil)
	w := httptest.NewRecorder()

	tokenRepo := newMockTokenRepo()
	oauthConfig := setupTestOAuthConfig()

	// Save an expired token
	tokenRepo.SaveUserToken(r.Context(), "11111111-1111-1111-1111-111111111111", &oauth2.Token{
		AccessToken:  "expired-token",
		TokenType:    "Bearer",
		RefreshToken: "refresh-token",
		Expiry:       time.Now().Add(-time.Hour), // Expired
	})

	// Apply middleware
	sessionMiddleware := middleware.NewSessionMiddleware(sessionManager)
	oauthMiddleware := middleware.NewOAuthMiddleware(sessionManager, tokenRepo, oauthConfig)

	// Create handler chain
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ListEmails(w, r)
	})
	handler := sessionMiddleware.Handler(oauthMiddleware.Handler(next))
	handler.ServeHTTP(w, r)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEmail_MissingID(t *testing.T) {
	factory := provider.NewProviderFactory()
	svc := email.NewMultiProviderService(factory)
	sessionManager, _ := setupTestSessionManager()
	h := NewEmailHandler(svc, sessionManager)

	r, w := setupTestRequest("GET", "/api/emails/", sessionManager)
	h.GetEmail(w, r)

	resp := w.Result()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetEmail_Success(t *testing.T) {
	factory := provider.NewProviderFactory()
	mockProvider := &mocks.MockEmailService{
		FetchMessageContentFunc: func(ctx context.Context, token *oauth2.Token, id string) (*models.EmailMessage, error) {
			require.Equal(t, "test-token", token.AccessToken)
			return &models.EmailMessage{EmailMessageID: "1", Subject: "Test"}, nil
		},
	}
	factory.RegisterProvider(provider.Gmail, func(cfg provider.Config) (provider.Provider, error) {
		return &dummyProvider{mockProvider}, nil
	})
	factory.LinkProvider("11111111-1111-1111-1111-111111111111", provider.Config{UserID: "11111111-1111-1111-1111-111111111111", Type: provider.Gmail})
	svc := email.NewMultiProviderService(factory)
	sessionManager, _ := setupTestSessionManager()
	h := NewEmailHandler(svc, sessionManager)

	r, w := setupTestRequest("GET", "/api/emails/1", sessionManager)
	// Simulate chi URL param
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))

	h.GetEmail(w, r)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetEmail_ServiceError(t *testing.T) {
	factory := provider.NewProviderFactory()
	mockProvider := &mocks.MockEmailService{
		FetchMessageContentFunc: func(ctx context.Context, token *oauth2.Token, id string) (*models.EmailMessage, error) {
			return nil, provider.ErrNotFound
		},
	}
	factory.RegisterProvider(provider.Gmail, func(cfg provider.Config) (provider.Provider, error) {
		return &dummyProvider{mockProvider}, nil
	})
	factory.LinkProvider("11111111-1111-1111-1111-111111111111", provider.Config{UserID: "11111111-1111-1111-1111-111111111111", Type: provider.Gmail})
	svc := email.NewMultiProviderService(factory)
	sessionManager, _ := setupTestSessionManager()
	h := NewEmailHandler(svc, sessionManager)

	r, w := setupTestRequest("GET", "/api/emails/1", sessionManager)
	// Simulate chi URL param
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))

	h.GetEmail(w, r)

	resp := w.Result()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// dummyProvider adapts MockEmailService to EmailProvider interface
type dummyProvider struct {
	*mocks.MockEmailService
}

func (d *dummyProvider) FetchSummaries(ctx context.Context, userID string) ([]models.EmailSummary, error) {
	token, _ := ctx.Value(middleware.OAuthToken{}).(*oauth2.Token)
	msgs, err := d.FetchMessages(ctx, token)
	if err != nil {
		return nil, err
	}

	summaries := make([]models.EmailSummary, len(msgs))
	for i, msg := range msgs {
		summaries[i] = models.EmailSummary{
			ID:           msg.EmailMessageID,
			ThreadID:     msg.ThreadID,
			Subject:      msg.Subject,
			Sender:       msg.Sender,
			Snippet:      msg.Snippet,
			InternalDate: msg.InternalDate,
			Date:         msg.Date,
		}
	}
	return summaries, nil
}

func (d *dummyProvider) FetchMessage(ctx context.Context, userToken interface{}, messageID string) (*models.EmailMessage, error) {
	return d.FetchMessageContent(ctx, userToken.(*oauth2.Token), messageID)
}

func (d *dummyProvider) GetCapabilities() provider.Capabilities {
	return provider.Capabilities{}
}

func (d *dummyProvider) GetType() provider.Type {
	return provider.Gmail
}
