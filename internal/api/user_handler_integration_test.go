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

	// Create user
	user := &models.User{
		ID:        uuid.NewString(),
		Email:     "integration@example.com",
		CreatedAt: time.Now().UTC(),
	}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/users/", bytes.NewReader(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST /users/ got status %d, want %d", w.Code, http.StatusCreated)
	}
	t.Logf("POST /users/ response: %d %s", w.Code, w.Body.String())

	// Get user
	req = httptest.NewRequest("GET", "/users/"+user.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	t.Logf("GET /users/{id} response: %d %s", w.Code, w.Body.String())
	if w.Code != http.StatusOK {
		t.Fatalf("GET /users/{id} got status %d", w.Code)
	}
	var got models.User
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Email != user.Email {
		t.Errorf("got email %q, want %q", got.Email, user.Email)
	}

	// List users
	req = httptest.NewRequest("GET", "/users/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET /users got status %d", w.Code)
	}
	var users []models.User
	if err := json.NewDecoder(w.Body).Decode(&users); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(users) == 0 {
		t.Errorf("expected at least 1 user, got %d", len(users))
	}

	// Update user
	user.Email = "updated@example.com"
	body, _ = json.Marshal(user)
	req = httptest.NewRequest("PUT", "/users/"+user.ID, bytes.NewReader(body))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("PUT /users/{id} got status %d", w.Code)
	}

	// Get updated user
	req = httptest.NewRequest("GET", "/users/"+user.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET after update got status %d", w.Code)
	}
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Email != "updated@example.com" {
		t.Errorf("got email %q, want %q", got.Email, "updated@example.com")
	}

	// Delete user
	req = httptest.NewRequest("DELETE", "/users/"+user.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("DELETE /users/{id} got status %d", w.Code)
	}

	// Get after delete
	req = httptest.NewRequest("GET", "/users/"+user.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Fatalf("expected not found after delete, got status %d", w.Code)
	}
}
