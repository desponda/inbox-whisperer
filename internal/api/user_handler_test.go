package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/session"
)

type mockUserService struct {
	GetUserFunc    func(ctx context.Context, id string) (*models.User, error)
	CreateUserFunc func(ctx context.Context, user *models.User) error
	ListUsersFunc  func(ctx context.Context) ([]*models.User, error)
	UpdateUserFunc func(ctx context.Context, user *models.User) error
	DeleteUserFunc func(ctx context.Context, id string) error
	DeactivateUserFunc func(ctx context.Context, id string) error
}

func (m *mockUserService) DeactivateUser(ctx context.Context, id string) error {
	if m.DeactivateUserFunc != nil {
		return m.DeactivateUserFunc(ctx, id)
	}
	return nil
}

func (m *mockUserService) ListUsers(ctx context.Context) ([]*models.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx)
	}
	return nil, nil
}

func (m *mockUserService) UpdateUser(ctx context.Context, user *models.User) error {
	if m.UpdateUserFunc != nil {
		return m.UpdateUserFunc(ctx, user)
	}
	return nil
}

func (m *mockUserService) DeleteUser(ctx context.Context, id string) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}

func (m *mockUserService) GetUser(ctx context.Context, id string) (*models.User, error) {
	return m.GetUserFunc(ctx, id)
}
func (m *mockUserService) CreateUser(ctx context.Context, user *models.User) error {
	return m.CreateUserFunc(ctx, user)
}

func TestUserHandler_GetUser(t *testing.T) {
	// Note: chi will return 404 for /users/ (no id), so the handler is never called in that case.
	tests := []struct {
		name       string
		id         string
		service    *mockUserService
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			id:   "abc",
			service: &mockUserService{
				GetUserFunc: func(ctx context.Context, id string) (*models.User, error) {
					return &models.User{ID: id, Email: "test@example.com"}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   "{\"id\":\"abc\",\"email\":\"test@example.com\",\"created_at\":\"0001-01-01T00:00:00Z\",\"deactivated\":false}\n",
		},
		{
			name: "not found",
			id:   "notfound",
			service: &mockUserService{
				GetUserFunc: func(ctx context.Context, id string) (*models.User, error) {
					return nil, context.DeadlineExceeded
				},
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"context deadline exceeded"}` + "\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &UserHandler{Service: tc.service}
			r := chi.NewRouter()
			r.Get("/users/{id}", h.GetUser)
			url := "/users/"
			if tc.id != "" {
				url += tc.id
			}
			req := httptest.NewRequest("GET", url, nil)
			// Simulate authenticated session
			sessionW := httptest.NewRecorder()
			session.SetSession(sessionW, req, tc.id, "dummy-token")
			for _, c := range sessionW.Result().Cookies() {
				req.AddCookie(c)
			}
			w := httptest.NewRecorder()
			rWithSession := session.Middleware(r)
			rWithSession.ServeHTTP(w, req)
			if w.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, w.Code)
			}
			if w.Body.String() != tc.wantBody {
				t.Errorf("expected body %q, got %q", tc.wantBody, w.Body.String())
			}
		})
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	h := &UserHandler{Service: &mockUserService{}}
	r := chi.NewRouter()
	r.Get("/users", h.ListUsers)
	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	h := &UserHandler{Service: &mockUserService{}}
	r := chi.NewRouter()
	r.Put("/users/{id}", h.UpdateUser)
	body, _ := json.Marshal(map[string]interface{}{})
	req := httptest.NewRequest("PUT", "/users/a", bytes.NewReader(body))
	// Simulate authenticated session
	sessionW := httptest.NewRecorder()
	session.SetSession(sessionW, req, "a", "dummy-token")
	for _, c := range sessionW.Result().Cookies() {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	rWithSession := session.Middleware(r)
	rWithSession.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	var errResp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Errorf("could not parse error response: %v", err)
	}
	if errResp["error"] != "no updatable fields" {
		t.Errorf("unexpected error message: %q", errResp["error"])
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	called := false
	h := &UserHandler{Service: &mockUserService{
		DeactivateUserFunc: func(ctx context.Context, id string) error {
			called = true
			if id != "a" {
				t.Errorf("unexpected id: %s", id)
			}
			return nil
		},
	}}
	r := chi.NewRouter()
	r.Delete("/users/{id}", h.DeleteUser)
	req := httptest.NewRequest("DELETE", "/users/a", nil)
	// Simulate authenticated session
	sessionW := httptest.NewRecorder()
	session.SetSession(sessionW, req, "a", "dummy-token")
	for _, c := range sessionW.Result().Cookies() {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	rWithSession := session.Middleware(r)
	rWithSession.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
	if !called {
		t.Error("expected DeleteUserFunc to be called")
	}
}

func TestUserHandler_DeleteUserMissingId(t *testing.T) {
	h := &UserHandler{Service: &mockUserService{}}
	r := chi.NewRouter()
	r.Delete("/users/{id}", h.DeleteUser)
	req := httptest.NewRequest("DELETE", "/users/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
	if w.Body.String() != "404 page not found\n" {
		t.Errorf("unexpected body: %q", w.Body.String())
	}
}

func TestUserHandler_DeleteUserServiceError(t *testing.T) {
	h := &UserHandler{Service: &mockUserService{
		DeactivateUserFunc: func(ctx context.Context, id string) error {
			return context.DeadlineExceeded
		},
	}}
	r := chi.NewRouter()
	r.Delete("/users/{id}", h.DeleteUser)
	req := httptest.NewRequest("DELETE", "/users/a", nil)
	// Simulate authenticated session
	sessionW := httptest.NewRecorder()
	session.SetSession(sessionW, req, "a", "dummy-token")
	for _, c := range sessionW.Result().Cookies() {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	rWithSession := session.Middleware(r)
	rWithSession.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
	var errResp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &errResp); err != nil {
		t.Errorf("could not parse error response: %v", err)
	}
	if errResp["error"] != "context deadline exceeded" {
		t.Errorf("unexpected error message: %q", errResp["error"])
	}
}

func TestUserHandler_CreateUser(t *testing.T) {
	h := &UserHandler{Service: &mockUserService{}}
	r := chi.NewRouter()
	r.Post("/users", h.CreateUser)
	body, _ := json.Marshal(map[string]interface{}{"email": "test@example.com"})
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}
