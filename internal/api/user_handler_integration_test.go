//go:build integration
// +build integration

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/desponda/inbox-whisperer/internal/session"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func setupIntegrationServer(t *testing.T) (*chi.Mux, func()) {
	db, cleanup := data.SetupTestDB(t)
	svc := service.NewUserService(db)
	h := NewUserHandler(svc)
	r := chi.NewRouter()
	r.Route("/users", func(r chi.Router) {
		r.Get("/", h.ListUsers)
		r.Get("/{id}", h.GetUser)
		r.Post("/", h.CreateUser)
		r.Put("/{id}", h.UpdateUser)
		r.Delete("/{id}", h.DeleteUser)
	})
	return r, cleanup
}

func TestUserAPI_Integration(t *testing.T) {
	if testing.Short() || os.Getenv("SKIP_DB_INTEGRATION") == "1" {
		t.Skip("skipping integration test")
	}
	router, cleanup := setupIntegrationServer(t)
	defer cleanup()

	// List users (should be forbidden)
	req := httptest.NewRequest("GET", "/users/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("GET /users/ got status %d, want 403", w.Code)
	}

	// Create user (should be forbidden)
	user := &models.User{
		ID:        uuid.NewString(),
		Email:     "integration@example.com",
		CreatedAt: time.Now().UTC(),
	}
	body, _ := json.Marshal(user)
	req = httptest.NewRequest("POST", "/users/", bytes.NewReader(body))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("POST /users/ got status %d, want 403", w.Code)
	}

	// Update user (should return bad request for no updatable fields)
	req = httptest.NewRequest("PUT", "/users/"+user.ID, bytes.NewReader(body))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("PUT /users/{id} got status %d, want 400", w.Code)
	}

	// Delete user (should return 204 even if user doesn't exist)
	req = httptest.NewRequest("DELETE", "/users/"+user.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("DELETE /users/{id} got status %d, want 204", w.Code)
	}
}

func TestUserHandler_CreateAndGetUser_Integration(t *testing.T) {
	if testing.Short() || os.Getenv("SKIP_DB_INTEGRATION") == "1" {
		t.Skip("skipping integration test")
	}
	router, cleanup := setupIntegrationServer(t)
	defer cleanup()

	// Simulate authenticated session
	userID := uuid.NewString()
	token := "dummy-token"
	user := &models.User{
		ID:        userID,
		Email:     "integration@example.com",
		CreatedAt: time.Now().UTC(),
	}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/users/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	session.SetSession(w, req, userID, token)
	// Copy session_id cookie from response to request
	var sessionCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "session_id" {
			sessionCookie = c
		}
	}
	if sessionCookie != nil {
		req.AddCookie(sessionCookie)
	}
	routerWithSession := session.Middleware(router)
	routerWithSession.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("POST /users/ got status %d, want %d", w.Code, http.StatusForbidden)
	}
	t.Logf("POST /users/ response: %d %s", w.Code, w.Body.String())
// No further actions: direct user creation is forbidden, so no user should exist or be manipulated in this test.
}
