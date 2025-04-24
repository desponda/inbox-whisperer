package api

import (
	"context"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/session"
)

// Skipping TestExchangeCodeForToken as it requires refactor for dependency injection.
// The main logic is covered in fetchGoogleUserID below.

func TestHandleLogin(t *testing.T) {
	dummyCfg := &oauth2.Config{
		ClientID:     "dummy",
		ClientSecret: "dummy",
		RedirectURL:  "http://localhost/api/auth/callback",
		Scopes:       []string{"openid"},
		Endpoint:     oauth2.Endpoint{AuthURL: "http://localhost/auth", TokenURL: "http://localhost/token"},
	}
	h := &AuthHandler{
		OAuthConfig: dummyCfg,
		UserTokens:  nil,
	}
	r := httptest.NewRequest("GET", "/api/auth/login", nil)
	w := httptest.NewRecorder()

	h.HandleLogin(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected redirect (302), got %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if !strings.Contains(loc, "state=") {
		t.Errorf("expected redirect URL to contain state param, got %s", loc)
	}
}

// For error-injection, use a unique stub for SaveUserToken error case
// Implements data.UserTokenRepository
// (GetUserToken returns a dummy token)
type stubFailUserTokens struct{}

func (s *stubFailUserTokens) SaveUserToken(ctx context.Context, userID string, tok *oauth2.Token) error {
	return context.DeadlineExceeded
}
func (s *stubFailUserTokens) GetUserToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "tok"}, nil
}

func TestHandleCallback(t *testing.T) {

	tests := []struct {
		name       string
		query      string
		failSave   bool
		wantStatus int
		wantInBody string
	}{
		{
			name:       "missing code",
			query:      "?state=valid",
			wantStatus: http.StatusBadRequest,
			wantInBody: "missing code",
		},
		{
			name:       "missing state",
			query:      "?code=good",
			wantStatus: http.StatusBadRequest,
			wantInBody: "state parameter",
		},
		{
			name:       "invalid state",
			query:      "?code=good&state=invalid",
			wantStatus: http.StatusBadRequest,
			wantInBody: "state parameter",
		},
		{
			name:       "token exchange fails",
			query:      "?code=bad&state=valid",
			wantStatus: http.StatusInternalServerError,
			wantInBody: "token exchange",
		},
		{
			name:       "db save fails",
			query:      "?code=good&state=valid",
			failSave:   true,
			wantStatus: http.StatusInternalServerError,
			wantInBody: "persist user token",
		},
		// Success case would require more extensive session and dependency mocking
	}

	cfg := &oauth2.Config{
		ClientID:     "dummy",
		ClientSecret: "dummy",
		RedirectURL:  "http://localhost/api/auth/callback",
		Scopes:       []string{"openid"},
		Endpoint:     oauth2.Endpoint{AuthURL: "http://localhost/auth", TokenURL: "http://localhost/token"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var userTokens data.UserTokenRepository
			if tc.failSave {
				userTokens = &stubFailUserTokens{}
			} else {
				userTokens = &stubUserTokens{} // from auth_csrf_test.go
			}
			h := &AuthHandler{
				OAuthConfig: cfg,
				UserTokens:  userTokens,
			}

			if tc.name == "token exchange fails" || tc.name == "db save fails" {
				// Patch exchangeCodeForToken to simulate error or success as needed
				savedExchange := exchangeCodeForToken
				if tc.name == "token exchange fails" {
					exchangeCodeForToken = func(h *AuthHandler, ctx context.Context, code string) (*oauth2.Token, error) {
						return nil, context.DeadlineExceeded
					}
				} else {
					exchangeCodeForToken = func(h *AuthHandler, ctx context.Context, code string) (*oauth2.Token, error) {
						return &oauth2.Token{AccessToken: "tok"}, nil
					}
				}
				defer func() { exchangeCodeForToken = savedExchange }()
				// Use a real HTTP server and cookie jar to set up session state
				mux := http.NewServeMux()
				mux.HandleFunc("/setstate", func(w http.ResponseWriter, r *http.Request) {
					session.SetSession(w, r, "testuser", "testtoken")
					if _, err := w.Write([]byte("ok")); err != nil {
						t.Fatalf("failed to write response: %v", err)
					}
				})
				mux.HandleFunc("/setstatevalue", func(w http.ResponseWriter, r *http.Request) {
					session.SetSessionValue(w, r, "oauth_state", "valid")
					if _, err := w.Write([]byte("ok")); err != nil {
						t.Fatalf("failed to write response: %v", err)
					}
				})
				mux.HandleFunc("/api/auth/callback", func(w http.ResponseWriter, r *http.Request) {
					h.HandleCallback(w, r)
				})
				ts := httptest.NewServer(session.Middleware(mux))
				defer ts.Close()
				jar, err := session.NewTestCookieJar()
				if err != nil {
					t.Fatalf("failed to create cookie jar: %v", err)
				}
				client := &http.Client{Jar: jar}
				// Create the session
				_, err = client.Get(ts.URL + "/setstate")
				if err != nil {
					t.Fatalf("setstate failed: %v", err)
				}
				// Set the state value in the session
				_, err = client.Get(ts.URL + "/setstatevalue")
				if err != nil {
					t.Fatalf("setstatevalue failed: %v", err)
				}
				// Now do callback with correct state
				resp, err := client.Get(ts.URL + "/api/auth/callback" + tc.query)
				if err != nil {
					t.Fatalf("callback failed: %v", err)
				}
				if resp.StatusCode != tc.wantStatus {
					t.Errorf("expected status %d, got %d", tc.wantStatus, resp.StatusCode)
				}
				b, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				body := string(b)
				if tc.wantInBody != "" && !strings.Contains(body, tc.wantInBody) {
					t.Errorf("expected body to contain %q, got %q", tc.wantInBody, body)
				}
				return
			}

			r := httptest.NewRequest("GET", "/api/auth/callback"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandleCallback(w, r)
			resp := w.Result()
			if resp.StatusCode != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, resp.StatusCode)
			}
			body := w.Body.String()
			if tc.wantInBody != "" && !strings.Contains(body, tc.wantInBody) {
				t.Errorf("expected body to contain %q, got %q", tc.wantInBody, body)
			}
		})
	}
}

// SessionData mirrors the struct in session/session.go for test setup
// Only Values is needed for CSRF state
type SessionData struct {
	Values map[string]string
}

func TestFetchGoogleUserID(t *testing.T) {
	// Mock HTTP server for userinfo
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "userinfo") {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"id": "abc123", "email": "foo@bar.com"}`)); err != nil {
				t.Fatalf("failed to write response: %v", err)
			}
			return
		}
		w.WriteHeader(404)
	}))
	defer ts.Close()

	tok := &oauth2.Token{AccessToken: "tok"}
	id, err := fetchGoogleUserID(context.Background(), tok, ts.URL+"/userinfo")
	if err != nil || id != "abc123" {
		t.Errorf("expected id 'abc123', got %v, %v", id, err)
	}
}
