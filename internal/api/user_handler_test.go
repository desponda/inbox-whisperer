package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/api/testutils"
	"github.com/desponda/inbox-whisperer/internal/common"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock UserService ---
type mockUserService struct {
	GetUserFunc        func(ctx context.Context, id uuid.UUID) (*models.User, error)
	ListUsersFunc      func(ctx context.Context) ([]*models.User, error)
	UpdateUserFunc     func(ctx context.Context, user *models.User) error
	DeactivateUserFunc func(ctx context.Context, id uuid.UUID) error
}

func (m *mockUserService) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return m.GetUserFunc(ctx, id)
}
func (m *mockUserService) ListUsers(ctx context.Context) ([]*models.User, error) {
	return m.ListUsersFunc(ctx)
}
func (m *mockUserService) UpdateUser(ctx context.Context, user *models.User) error {
	return m.UpdateUserFunc(ctx, user)
}
func (m *mockUserService) DeactivateUser(ctx context.Context, id uuid.UUID) error {
	return m.DeactivateUserFunc(ctx, id)
}
func (m *mockUserService) CreateUser(ctx context.Context, user *models.User) error {
	return errors.New("forbidden")
}
func (m *mockUserService) DeleteUser(ctx context.Context, id uuid.UUID) error { return nil }

var (
	testUserID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testUser   = &models.User{ID: testUserID, Email: "test@example.com", CreatedAt: time.Now().UTC()}
)

func TestUserHandler_GetMe(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &mockUserService{
			GetUserFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return testUser, nil
			},
		}
		sessionManager := testutils.NewTestSessionManager("user")
		h := NewUserHandler(svc, sessionManager)
		r, w := testutils.SetupTestRequest("GET", "/api/users/me", sessionManager, "user")
		r = r.WithContext(common.ContextWithUserID(r.Context(), testUserID))
		h.GetMe(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.User
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, testUserID, resp.ID)
	})
	t.Run("not found", func(t *testing.T) {
		svc := &mockUserService{
			GetUserFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return nil, errors.New("not found")
			},
		}
		sessionManager := testutils.NewTestSessionManager("user")
		h := NewUserHandler(svc, sessionManager)
		r, w := testutils.SetupTestRequest("GET", "/api/users/me", sessionManager, "user")
		r = r.WithContext(common.ContextWithUserID(r.Context(), testUserID))
		h.GetMe(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestUserHandler_ListUsers(t *testing.T) {
	t.Run("admin success", func(t *testing.T) {
		svc := &mockUserService{
			ListUsersFunc: func(ctx context.Context) ([]*models.User, error) {
				return []*models.User{testUser}, nil
			},
		}
		sessionManager := testutils.NewTestSessionManager("admin")
		h := NewUserHandler(svc, sessionManager)
		r, w := testutils.SetupTestRequest("GET", "/api/users", sessionManager, "admin")
		r = r.WithContext(common.ContextWithRole(r.Context(), "admin"))
		h.ListUsers(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		var resp []*models.User
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Len(t, resp, 1)
		assert.Equal(t, testUserID, resp[0].ID)
	})
	t.Run("forbidden for non-admin", func(t *testing.T) {
		svc := &mockUserService{}
		sessionManager := testutils.NewTestSessionManager("user")
		h := NewUserHandler(svc, sessionManager)
		r, w := testutils.SetupTestRequest("GET", "/api/users", sessionManager, "user")
		r = r.WithContext(common.ContextWithRole(r.Context(), "user"))
		h.ListUsers(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestUserHandler_GetUser(t *testing.T) {
	t.Run("success as self", func(t *testing.T) {
		svc := &mockUserService{
			GetUserFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return testUser, nil
			},
		}
		sessionManager := testutils.NewTestSessionManager("user")
		h := NewUserHandler(svc, sessionManager)
		r, w := testutils.SetupTestRequest("GET", "/api/users/"+testUserID.String(), sessionManager, "user")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", testUserID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		r = r.WithContext(common.ContextWithUserID(r.Context(), testUserID))
		h.GetUser(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("forbidden for other user", func(t *testing.T) {
		svc := &mockUserService{}
		sessionManager := testutils.NewTestSessionManager("user")
		h := NewUserHandler(svc, sessionManager)
		otherID := uuid.New()
		r, w := testutils.SetupTestRequest("GET", "/api/users/"+otherID.String(), sessionManager, "user")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", otherID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		r = r.WithContext(common.ContextWithUserID(r.Context(), testUserID))
		h.GetUser(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("not found", func(t *testing.T) {
		svc := &mockUserService{
			GetUserFunc: func(ctx context.Context, id uuid.UUID) (*models.User, error) {
				return nil, errors.New("not found")
			},
		}
		sessionManager := testutils.NewTestSessionManager("admin")
		h := NewUserHandler(svc, sessionManager)
		r, w := testutils.SetupTestRequest("GET", "/api/users/"+testUserID.String(), sessionManager, "admin")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", testUserID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		r = r.WithContext(common.ContextWithUserID(r.Context(), uuid.New()))
		r = r.WithContext(common.ContextWithRole(r.Context(), "admin"))
		h.GetUser(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestUserHandler_UpdateUser(t *testing.T) {
	t.Run("success as self", func(t *testing.T) {
		svc := &mockUserService{
			UpdateUserFunc: func(ctx context.Context, user *models.User) error {
				return nil
			},
		}
		sessionManager := testutils.NewTestSessionManager("user")
		h := NewUserHandler(svc, sessionManager)
		body := io.NopCloser(bytes.NewReader([]byte(`{"email":"new@example.com"}`)))
		r, w := testutils.SetupTestRequest("PUT", "/api/users/"+testUserID.String(), sessionManager, "user")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", testUserID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		r = r.WithContext(common.ContextWithUserID(r.Context(), testUserID))
		r.Body = body
		h.UpdateUser(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("forbidden for other user", func(t *testing.T) {
		svc := &mockUserService{}
		sessionManager := testutils.NewTestSessionManager("user")
		h := NewUserHandler(svc, sessionManager)
		otherID := uuid.New()
		body := io.NopCloser(bytes.NewReader([]byte(`{"email":"new@example.com"}`)))
		r, w := testutils.SetupTestRequest("PUT", "/api/users/"+otherID.String(), sessionManager, "user")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", otherID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		r = r.WithContext(common.ContextWithUserID(r.Context(), testUserID))
		r.Body = body
		h.UpdateUser(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
	// Bad request: invalid body
	t.Run("bad request invalid body", func(t *testing.T) {
		svc := &mockUserService{}
		sessionManager := testutils.NewTestSessionManager("user")
		h := NewUserHandler(svc, sessionManager)
		body := io.NopCloser(bytes.NewReader([]byte("not-json")))
		r, w := testutils.SetupTestRequest("PUT", "/api/users/"+testUserID.String(), sessionManager, "user")
		r.Body = body
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", testUserID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		r = r.WithContext(common.ContextWithUserID(r.Context(), testUserID))
		h.UpdateUser(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserHandler_DeleteUser(t *testing.T) {
	t.Run("success as self", func(t *testing.T) {
		svc := &mockUserService{
			DeactivateUserFunc: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
		}
		sessionManager := testutils.NewTestSessionManager("user")
		h := NewUserHandler(svc, sessionManager)
		r, w := testutils.SetupTestRequest("DELETE", "/api/users/"+testUserID.String(), sessionManager, "user")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", testUserID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		r = r.WithContext(common.ContextWithUserID(r.Context(), testUserID))
		h.DeleteUser(w, r)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
	t.Run("forbidden for other user", func(t *testing.T) {
		svc := &mockUserService{}
		sessionManager := testutils.NewTestSessionManager("user")
		h := NewUserHandler(svc, sessionManager)
		otherID := uuid.New()
		r, w := testutils.SetupTestRequest("DELETE", "/api/users/"+otherID.String(), sessionManager, "user")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", otherID.String())
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		r = r.WithContext(common.ContextWithUserID(r.Context(), testUserID))
		h.DeleteUser(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestUserHandler_CreateUser(t *testing.T) {
	svc := &mockUserService{}
	sessionManager := testutils.NewTestSessionManager("admin")
	h := NewUserHandler(svc, sessionManager)
	// No need to declare body since handler always returns forbidden
	r, w := testutils.SetupTestRequest("POST", "/api/users", sessionManager, "admin")
	h.CreateUser(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}
