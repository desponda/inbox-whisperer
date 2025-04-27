package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/desponda/inbox-whisperer/internal/config"
	"github.com/desponda/inbox-whisperer/internal/session"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

func TestHandleCallback_CSRFProtection(t *testing.T) {
	// Mock the token exchange
	originalExchange := exchangeCodeForToken
	exchangeCodeForToken = func(h *AuthHandler, ctx context.Context, code string) (*oauth2.Token, error) {
		if code == "good" {
			return &oauth2.Token{AccessToken: "test-token"}, nil
		}
		return nil, fmt.Errorf("invalid code")
	}
	defer func() { exchangeCodeForToken = originalExchange }()

	// Setup test variables
	var (
		mux    = http.NewServeMux()
		ts     = httptest.NewServer(session.Middleware(mux))
		client *http.Client
		jar    http.CookieJar
		err    error
	)
	defer ts.Close()

	// Create the auth handler with test config
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
	h := NewAuthHandler(appConfig, &stubUserTokens{})
	
	// Register auth routes
	mux.HandleFunc("/auth/callback", h.HandleCallback)
	mux.HandleFunc("/setstate", func(w http.ResponseWriter, r *http.Request) {
		// Create a session ID
		sessionID := uuid.New().String()
		
		// Set the cookie first
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
		})

		// Add the cookie to the request for the session functions to use
		r.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: sessionID,
		})

		// Create the session
		session.SetSession(w, r, "testuser", "testtoken")
		
		// Set the state value in the session
		session.SetSessionValue(w, r, "oauth_state", "goodstate")
		
		// Verify the state was set
		state := session.GetSessionValue(r, "oauth_state")
		fmt.Printf("[DEBUG] Set state value in session %s: %q\n", sessionID, state)
		if state != "goodstate" {
			t.Fatalf("state not set correctly, got %q", state)
		}

		_, err = w.Write([]byte("ok"))
		if err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	})
	
	// Create a cookie jar for the test client
	jar, err = session.NewTestCookieJar()
	if err != nil {
		t.Fatalf("failed to create cookie jar: %v", err)
	}
	// Don't follow redirects in test
	client = &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Try with no state
	resp, err := client.Get(ts.URL + "/auth/callback?code=good")
	if err != nil {
		t.Fatalf("callback failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing state, got %d", resp.StatusCode)
	}

	// Try with wrong state
	resp, err = client.Get(ts.URL + "/auth/callback?code=good&state=wrong")
	if err != nil {
		t.Fatalf("callback failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid state, got %d", resp.StatusCode)
	}

	// Create the session and set the state value
	resp2, err := client.Get(ts.URL + "/setstate")
	if err != nil {
		t.Fatalf("setstate failed: %v", err)
	}
	resp2.Body.Close()

	// Try with valid state
	resp3, err := client.Get(ts.URL + "/auth/callback?code=good&state=goodstate")
	if err != nil {
		t.Fatalf("callback failed: %v", err)
	}
	resp3Body, err := io.ReadAll(resp3.Body)
	resp3.Body.Close()
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	// Check that we get redirected
	if resp3.StatusCode != http.StatusFound {
		t.Errorf("expected 302 redirect for valid state, got %d, body: %q", resp3.StatusCode, resp3Body)
	}

	// Check the redirect URL
	location := resp3.Header.Get("Location")
	if location == "" {
		t.Error("expected Location header in redirect response")
	}

	// Verify state is still in session
	state := session.GetSessionValue(resp3.Request, "oauth_state")
	fmt.Printf("[DEBUG] State in session after callback: %q\n", state)
}

// stubUserTokens implements only SaveUserToken for test
// (other methods are no-ops)
type stubUserTokens struct{}

func (s *stubUserTokens) SaveUserToken(ctx context.Context, userID string, tok *oauth2.Token) error {
	return nil
}
func (s *stubUserTokens) GetUserToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "tok"}, nil
}
