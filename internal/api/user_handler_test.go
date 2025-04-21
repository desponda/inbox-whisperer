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
)

type mockUserService struct {
	GetUserFunc    func(ctx context.Context, id string) (*models.User, error)
	CreateUserFunc func(ctx context.Context, user *models.User) error
	ListUsersFunc  func(ctx context.Context) ([]*models.User, error)
	UpdateUserFunc func(ctx context.Context, user *models.User) error
	DeleteUserFunc func(ctx context.Context, id string) error
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
			wantBody:   `{"id":"abc","email":"test@example.com","created_at":"0001-01-01T00:00:00Z"}` + "\n",
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
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, w.Code)
			}
			if w.Body.String() != tc.wantBody {
				t.Errorf("expected body %q, got %q", tc.wantBody, w.Body.String())
			}
		})
	}
}

func TestUserHandler_CreateUser(t *testing.T) {
	tests := []struct {
		name       string
		input      models.User
		service    *mockUserService
		wantStatus int
		wantBody   string
	}{
		{
			name:  "success",
			input: models.User{Email: "test@example.com"},
			service: &mockUserService{
				CreateUserFunc: func(ctx context.Context, user *models.User) error {
					return nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `{"id":"","email":"test@example.com","created_at":"0001-01-01T00:00:00Z"}` + "\n",
		},
		{
			name:  "missing email",
			input: models.User{},
			service: &mockUserService{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"missing email"}` + "\n",
		},
		{
			name:  "invalid body",
			input: models.User{}, // will send invalid JSON
			service: &mockUserService{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"invalid request body"}` + "\n",
		},
		{
			name:  "service error",
			input: models.User{Email: "fail@example.com"},
			service: &mockUserService{
				CreateUserFunc: func(ctx context.Context, user *models.User) error {
					return context.DeadlineExceeded
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"context deadline exceeded"}` + "\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &UserHandler{Service: tc.service}
			r := chi.NewRouter()
			r.Post("/users", h.CreateUser)
			var body []byte
			if tc.name == "invalid body" {
				body = []byte("notjson")
			} else {
				body, _ = json.Marshal(tc.input)
			}
			req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, w.Code)
			}
			if tc.name == "success" {
				var resp models.User
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("could not parse response: %v", err)
				}
				if resp.ID == "" {
					t.Errorf("expected non-empty id, got empty")
				}
				if resp.Email != tc.input.Email {
					t.Errorf("expected email %q, got %q", tc.input.Email, resp.Email)
				}
				if resp.CreatedAt.IsZero() {
					t.Errorf("expected non-zero created_at, got %v", resp.CreatedAt)
				}
			} else {
				if w.Body.String() != tc.wantBody {
					t.Errorf("expected body %q, got %q", tc.wantBody, w.Body.String())
				}
			}
		})
	}
}
