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

	"github.com/desponda/inbox-whisperer/internal/api/testutils"
	"github.com/desponda/inbox-whisperer/internal/auth/service/oauth"
	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/desponda/inbox-whisperer/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	goauth2 "google.golang.org/api/oauth2/v2"
)

// Test utilities and mocks
func setupTestSession(t *testing.T) (session.Store, session.Manager, func()) {
	store := testutils.NewMockStore()
	manager := testutils.NewMockSessionManager(func(w http.ResponseWriter, r *http.Request) (session.Session, error) {
		// Always return a session, simulating real session creation
		cookie, err := r.Cookie("session_id")
		if err == nil {
			sess, _ := store.Get(context.Background(), cookie.Value)
			if sess != nil {
				return sess, nil
			}
		}
		sess, _ := store.Create(context.Background())
		store.Save(context.Background(), sess)
		return sess, nil
	})
	manager.StoreFunc = func() session.Store { return store }

	cleanup := func() {
		store.Reset()
	}

	return store, manager, cleanup
}

type mockOAuthService struct {
	shouldFailExchange bool
	shouldFailSave     bool
}

func (m *mockOAuthService) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, *goauth2.Userinfo, error) {
	if m.shouldFailExchange || code != "good" {
		return nil, nil, oauth.ErrTokenExchange
	}
	return &oauth2.Token{AccessToken: "test"}, &goauth2.Userinfo{Id: "11111111-1111-1111-1111-111111111111", Email: "test@example.com"}, nil
}

func (m *mockOAuthService) SaveUserAndToken(ctx context.Context, userInfo *goauth2.Userinfo, token *oauth2.Token) error {
	if m.shouldFailSave {
		return fmt.Errorf("failed to save user and token")
	}
	return nil
}

func (m *mockOAuthService) GetDB() interface{} {
	return &mockPool{}
}

type mockDB struct{}

func (db *mockDB) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return &mockRow{
		userID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	}
}

type mockRow struct {
	userID uuid.UUID
}

func (r *mockRow) Scan(dest ...interface{}) error {
	if len(dest) > 0 {
		switch d := dest[0].(type) {
		case *uuid.UUID:
			*d = r.userID
			return nil
		}
	}
	return nil
}

type mockPool struct{}

func (p *mockPool) QueryRow(ctx context.Context, sql string, args ...interface{}) interface{} {
	return &mockRow{
		userID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
	}
}

type stubUserTokens struct{}

func (s *stubUserTokens) SaveUserToken(ctx context.Context, userID string, tok *oauth2.Token) error {
	return nil
}

func (s *stubUserTokens) GetUserToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "test"}, nil
}

// TestHandleLogin tests the login handler functionality
func TestHandleLogin(t *testing.T) {
	store, manager, cleanup := setupTestSession(t)
	defer cleanup()

	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	appConfig := &config.AppConfig{
		Google: config.GoogleConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  ts.URL + "/auth/callback",
		},
	}

	mockTokens := &stubUserTokens{}
	h := NewAuthHandler(appConfig, mockTokens, manager)
	h.oauthService = &mockOAuthService{}

	mux.HandleFunc("/auth/login", h.HandleLogin)

	tests := []struct {
		name         string
		setupSession bool
		expectedCode int
	}{
		{
			name:         "successful login initiation",
			setupSession: true,
			expectedCode: http.StatusTemporaryRedirect,
		},
		{
			name:         "session creation failure",
			setupSession: false,
			expectedCode: http.StatusTemporaryRedirect,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cleanup()

			var sess session.Session
			if tc.name == "session creation failure" {
				sess, err := store.Create(context.Background())
				require.NoError(t, err)
				err = store.Save(context.Background(), sess)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("GET", "/auth/login", nil)
			if tc.name == "session creation failure" && sess != nil {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: sess.ID()})
			}
			w := httptest.NewRecorder()
			h.HandleLogin(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, tc.expectedCode, resp.StatusCode)

			if tc.expectedCode == http.StatusTemporaryRedirect {
				location := resp.Header.Get("Location")
				require.Contains(t, location, "accounts.google.com")
				require.Contains(t, location, "state=")
			}
		})
	}
}

// TestHandleCallback tests both the authentication flow and CSRF protection
func TestHandleCallback(t *testing.T) {
	store, manager, cleanup := setupTestSession(t)
	defer cleanup()

	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	appConfig := &config.AppConfig{
		Google: config.GoogleConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  ts.URL + "/auth/callback",
		},
		Server: config.ServerConfig{
			FrontendURL: "http://localhost:5173",
		},
	}

	mockTokens := &stubUserTokens{}
	h := NewAuthHandler(appConfig, mockTokens, manager)
	h.oauthService = &mockOAuthService{}

	mux.HandleFunc("/auth/callback", h.HandleCallback)

	tests := []struct {
		name          string
		setupState    bool
		state         string
		code          string
		expectedCode  int
		expectedError string
	}{
		{
			name:          "missing state",
			setupState:    false,
			state:         "",
			code:          "good",
			expectedCode:  http.StatusBadRequest,
			expectedError: "missing state parameter",
		},
		{
			name:          "invalid state",
			setupState:    true,
			state:         "badstate",
			code:          "good",
			expectedCode:  http.StatusBadRequest,
			expectedError: "invalid state parameter",
		},
		{
			name:          "valid state but invalid code",
			setupState:    true,
			state:         "goodstate",
			code:          "bad",
			expectedCode:  http.StatusBadRequest,
			expectedError: "failed to exchange code for token",
		},
		{
			name:         "valid state and code",
			setupState:   true,
			state:        "goodstate",
			code:         "good",
			expectedCode: http.StatusFound,
		},
		{
			name:          "expired state",
			setupState:    true,
			state:         "goodstate",
			code:          "good",
			expectedCode:  http.StatusBadRequest,
			expectedError: "state parameter has expired",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cleanup()

			if tc.setupState {
				sess := testutils.NewMockSession(
					"test-session",
					"11111111-1111-1111-1111-111111111111",
					"",
					map[string]interface{}{
						"oauth_state":      "goodstate",
						"state_created_at": time.Now().Add(-6 * time.Minute), // Expired for the expired state test
					},
					time.Now(),
					time.Now().Add(24*time.Hour),
				)
				if tc.name != "expired state" {
					sess.SetValue("state_created_at", time.Now()) // Not expired for other tests
				}
				err := store.Save(context.Background(), sess)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("GET", fmt.Sprintf("/auth/callback?state=%s&code=%s", tc.state, tc.code), nil)
			if tc.setupState {
				req.AddCookie(&http.Cookie{Name: "session_id", Value: "test-session"})
			}
			w := httptest.NewRecorder()
			h.HandleCallback(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, tc.expectedCode, resp.StatusCode)

			if tc.expectedError != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				var errorResp struct {
					Error string `json:"error"`
				}
				err = json.Unmarshal(body, &errorResp)
				require.NoError(t, err)
				require.Contains(t, errorResp.Error, tc.expectedError)
			}

			if tc.expectedCode == http.StatusFound {
				location := resp.Header.Get("Location")
				require.Contains(t, location, appConfig.Server.FrontendURL)
			}
		})
	}
}

func TestRegisterAuthRoutes(t *testing.T) {
	r := chi.NewRouter()
	cfg := &config.AppConfig{
		Google: config.GoogleConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "http://localhost/callback",
		},
		Server: config.ServerConfig{
			FrontendURL: "http://localhost:3000",
		},
	}

	store := testutils.NewMockStore()
	sessionManager := testutils.NewMockSessionManager(func(w http.ResponseWriter, r *http.Request) (session.Session, error) {
		sess := testutils.NewMockSession(
			"test-session",
			"11111111-1111-1111-1111-111111111111",
			"",
			map[string]interface{}{
				"oauth_state":      "test-state",
				"state_created_at": time.Now(),
			},
			time.Now(),
			time.Now().Add(24*time.Hour),
		)
		store.Save(r.Context(), sess)
		return sess, nil
	})
	sessionManager.StoreFunc = func() session.Store {
		return store
	}

	// Create mock user tokens repository
	userTokens := &stubUserTokens{}

	// Register routes under /auth prefix
	r.Route("/auth", func(r chi.Router) {
		RegisterAuthRoutes(r, cfg, userTokens, sessionManager)
	})

	// Test that routes are registered
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/auth/login", nil))
	require.Equal(t, http.StatusTemporaryRedirect, w.Code)
}
