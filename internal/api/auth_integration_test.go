//go:build integration
// +build integration

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/desponda/inbox-whisperer/internal/config"
	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
)

type fakeGoogleServer struct {
	Server *httptest.Server
	UserID string
	Email  string
}

func newFakeGoogleServer() *fakeGoogleServer {
	f := &fakeGoogleServer{
		UserID: "test-google-id",
		Email:  "testuser@example.com",
	}
	f.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/userinfo" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"id":    f.UserID,
				"email": f.Email,
			})
			return
		}
		w.WriteHeader(404)
	}))
	return f
}

func TestAuth_FirstTimeLogin_CreatesUserAndToken(t *testing.T) {
	if testing.Short() || os.Getenv("SKIP_DB_INTEGRATION") == "1" {
		t.Skip("skipping integration test")
	}
	fakeGoogle := newFakeGoogleServer()
	defer fakeGoogle.Server.Close()

	db, cleanup := data.SetupTestDB(t)
	defer cleanup()

	cfg := &config.AppConfig{
		Google: config.GoogleConfig{
			ClientID:     "dummy",
			ClientSecret: "dummy",
			RedirectURL:  "http://localhost/api/auth/callback",
		},
	}
	h := NewAuthHandler(cfg, db)
	// Patch exchangeCodeForToken to mock token exchange
	origExchange := exchangeCodeForToken
	exchangeCodeForToken = func(_ *AuthHandler, ctx context.Context, code string) (*oauth2.Token, error) {
		if code == "good" {
			return &oauth2.Token{AccessToken: "tok"}, nil
		}
		return nil, context.DeadlineExceeded
	}
	defer func() { exchangeCodeForToken = origExchange }()

	// Patch fetchGoogleUserID to return both userID and email for test
	oldFetcher := fetchGoogleUserID
	fetchGoogleUserID = func(ctx context.Context, tok *oauth2.Token, userinfoURL string) (string, error) {
		return "test-google-id|testuser@example.com", nil
	}
	defer func() { fetchGoogleUserID = oldFetcher }()

	r := chi.NewRouter()
	r.Get("/api/auth/login", h.HandleLogin)
	r.Get("/api/auth/callback", h.HandleCallback)

	// (rest of test logic unchanged)

	// First OAuth login: should create a new user and token
	req := httptest.NewRequest("GET", "/api/auth/callback?code=good", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusFound && w.Code != http.StatusOK {
		t.Fatalf("expected redirect or 200, got %d", w.Code)
	}
	user, err := db.GetByID(context.Background(), fakeGoogle.UserID)
	if err != nil {
		t.Fatalf("user not created on first login: %v", err)
	}
	if user.Email != fakeGoogle.Email {
		t.Errorf("user email mismatch after first login: got %s, want %s", user.Email, fakeGoogle.Email)
	}
	tok, err := db.GetUserToken(context.Background(), fakeGoogle.UserID)
	if err != nil || tok == nil {
		t.Fatalf("token not saved on first login: %v", err)
	}

	// Second OAuth login: should NOT create a duplicate or overwrite user
	// (simulate user has updated their email elsewhere, ensure it does not get overwritten)
	updatedEmail := "shouldnotoverwrite@example.com"
	_, err = db.Pool.Exec(context.Background(), "UPDATE users SET email = $1 WHERE id = $2", updatedEmail, fakeGoogle.UserID)
	if err != nil {
		t.Fatalf("failed to update user email for overwrite test: %v", err)
	}
	// Simulate another OAuth callback
	req2 := httptest.NewRequest("GET", "/api/auth/callback?code=good", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusFound && w2.Code != http.StatusOK {
		t.Fatalf("expected redirect or 200, got %d", w2.Code)
	}
	// Ensure only one user exists
	rows, err := db.Pool.Query(context.Background(), "SELECT COUNT(*) FROM users WHERE id = $1", fakeGoogle.UserID)
	if err != nil {
		t.Fatalf("failed to query users: %v", err)
	}
	defer rows.Close()
	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			t.Fatalf("failed to scan count: %v", err)
		}
	}
	if count != 1 {
		t.Errorf("expected 1 user after second login, got %d", count)
	}
	// Ensure the user's email was NOT overwritten by the second OAuth login
	user, err = db.GetByID(context.Background(), fakeGoogle.UserID)
	if err != nil {
		t.Fatalf("user not found after second login: %v", err)
	}
	if user.Email != updatedEmail {
		t.Errorf("user email was overwritten on second login: got %s, want %s", user.Email, updatedEmail)
	}
}
