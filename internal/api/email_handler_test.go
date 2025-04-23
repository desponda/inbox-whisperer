package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/desponda/inbox-whisperer/internal/service/gmail"

	"github.com/desponda/inbox-whisperer/internal/mocks"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestFetchMessagesHandler_Unauthenticated(t *testing.T) {
	h := NewEmailHandler(&mocks.MockEmailService{}, &mocks.MockUserTokenRepository{})
	r := httptest.NewRequest("GET", "/api/email/fetch", nil)
	w := httptest.NewRecorder()

	h.FetchMessagesHandler(w, r)

	resp := w.Result()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetMessageContentHandler_MissingID(t *testing.T) {
	h := NewEmailHandler(&mocks.MockEmailService{}, &mocks.MockUserTokenRepository{})
	r := httptest.NewRequest("GET", "/api/email/messages/", nil)
	w := httptest.NewRecorder()

	h.GetMessageContentHandler(w, r)

	resp := w.Result()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestFetchMessagesHandler_Authenticated_Success(t *testing.T) {
	mockSvc := &mocks.MockEmailService{
		FetchMessagesFunc: func(ctx context.Context, token *oauth2.Token) ([]models.EmailMessage, error) {
			return []models.EmailMessage{{ID: 123, Subject: "Test"}}, nil
		},
	}
	h := NewEmailHandler(mockSvc, &mocks.MockUserTokenRepository{})

	r := httptest.NewRequest("GET", "/api/email/fetch", nil)
	ctx := context.WithValue(r.Context(), ContextUserIDKey, "user1")
ctx = context.WithValue(ctx, ContextTokenKey, &oauth2.Token{AccessToken: "test-token"})
r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	h.FetchMessagesHandler(w, r)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestFetchMessagesHandler_ServiceError(t *testing.T) {
	mockSvc := &mocks.MockEmailService{
		FetchMessagesFunc: func(ctx context.Context, token *oauth2.Token) ([]models.EmailMessage, error) {
			return nil, gmail.ErrNotFound
		},
	}
	h := NewEmailHandler(mockSvc, &mocks.MockUserTokenRepository{})

	r := httptest.NewRequest("GET", "/api/email/fetch", nil)
	ctx := context.WithValue(r.Context(), ContextUserIDKey, "user1")
ctx = context.WithValue(ctx, ContextTokenKey, &oauth2.Token{AccessToken: "test-token"})
r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	h.FetchMessagesHandler(w, r)

	resp := w.Result()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetMessageContentHandler_Authenticated_Success(t *testing.T) {
	mockSvc := &mocks.MockEmailService{
		FetchMessageContentFunc: func(ctx context.Context, token *oauth2.Token, id string) (*models.EmailMessage, error) {
			return &models.EmailMessage{ID: 1, Subject: "Test"}, nil
		},
	}
	h := NewEmailHandler(mockSvc, &mocks.MockUserTokenRepository{})

	r := httptest.NewRequest("GET", "/api/email/messages/1", nil)
	ctx := context.WithValue(r.Context(), ContextUserIDKey, "user1")
ctx = context.WithValue(ctx, ContextTokenKey, &oauth2.Token{AccessToken: "test-token"})
r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	// Simulate chi URL param
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))

	h.GetMessageContentHandler(w, r)

	resp := w.Result()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetMessageContentHandler_ServiceError(t *testing.T) {
	mockSvc := &mocks.MockEmailService{
		FetchMessageContentFunc: func(ctx context.Context, token *oauth2.Token, id string) (*models.EmailMessage, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	h := NewEmailHandler(mockSvc, &mocks.MockUserTokenRepository{})

	r := httptest.NewRequest("GET", "/api/email/messages/1", nil)
	ctx := context.WithValue(r.Context(), ContextUserIDKey, "user1")
ctx = context.WithValue(ctx, ContextTokenKey, &oauth2.Token{AccessToken: "test-token"})
r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("id", "1")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))

	h.GetMessageContentHandler(w, r)

	resp := w.Result()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
