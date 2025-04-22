package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/desponda/inbox-whisperer/internal/session"
	"golang.org/x/oauth2"
	"context"
)


func TestHandleCallback_CSRFProtection(t *testing.T) {
	// Setup AuthHandler with mock UserTokens
	h := &AuthHandler{
		OAuthConfig: &oauth2.Config{},
		UserTokens: &stubUserTokens{},
	}
	// Patch exchangeCodeForToken to always succeed
	savedExchange := exchangeCodeForToken
	exchangeCodeForToken = func(h *AuthHandler, ctx context.Context, code string) (*oauth2.Token, error) {
		if code == "good" || code == "somecode" {
			return &oauth2.Token{AccessToken: "tok"}, nil
		}
		return nil, context.DeadlineExceeded
	}
	defer func() { exchangeCodeForToken = savedExchange }()


	// Simulate request with missing state
	r := httptest.NewRequest("GET", "/auth/callback?code=somecode", nil)
	w := httptest.NewRecorder()
	h.HandleCallback(w, r)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing state, got %d", resp.StatusCode)
	}

	// Simulate request with incorrect state
	r = httptest.NewRequest("GET", "/auth/callback?code=somecode&state=wrong", nil)
	w = httptest.NewRecorder()
	session.SetSessionValue(w, r, "oauth_state", "expected")
	h.HandleCallback(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid state, got %d", resp.StatusCode)
	}

	// Simulate request with correct state
	// We need to set the session value and then send a request with the same cookie
	mux := http.NewServeMux()
	mux.HandleFunc("/setstate", func(w http.ResponseWriter, r *http.Request) {
		// Only create the session
		session.SetSession(w, r, "testuser", "testtoken")
		w.Write([]byte("ok"))
	})
mux.HandleFunc("/setstatevalue", func(w http.ResponseWriter, r *http.Request) {
		session.SetSessionValue(w, r, "oauth_state", "goodstate")
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
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
	resp2, err := client.Get(ts.URL + "/setstate")
	if err != nil {
		t.Fatalf("setstate failed: %v", err)
	}
	defer resp2.Body.Close()
	// Set the state value in the session
	respVal, err := client.Get(ts.URL + "/setstatevalue")
	if err != nil {
		t.Fatalf("setstatevalue failed: %v", err)
	}
	defer respVal.Body.Close()
	// Now do callback with correct state and session cookie
	resp3, err := client.Get(ts.URL + "/auth/callback?code=good&state=goodstate")
	if err != nil {
		t.Fatalf("callback failed: %v", err)
	}
	if resp3.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for valid state, got %d", resp3.StatusCode)
	}
}

// stubUserTokens implements only SaveUserToken for test
// (other methods are no-ops)
type stubUserTokens struct{}
func (s *stubUserTokens) SaveUserToken(ctx context.Context, userID string, tok *oauth2.Token) error { return nil }
func (s *stubUserTokens) GetUserToken(ctx context.Context, userID string) (*oauth2.Token, error) { return &oauth2.Token{AccessToken: "tok"}, nil }
