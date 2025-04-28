package mocks

import (
	"context"
	"net/http"

	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"golang.org/x/oauth2"
)

// MockUserTokenRepository is a mock implementation of data.UserTokenRepository
type MockUserTokenRepository struct {
	GetUserTokenFunc  func(ctx context.Context, userID string) (*oauth2.Token, error)
	SaveUserTokenFunc func(ctx context.Context, userID string, token *oauth2.Token) error
	DestroyFunc       func(w http.ResponseWriter, r *http.Request) error
	StartFunc         func(w http.ResponseWriter, r *http.Request) (session.Session, error)
	RefreshFunc       func(w http.ResponseWriter, r *http.Request) error
	StoreFunc         func() session.Store
}

// GetUserToken implements data.UserTokenRepository
func (m *MockUserTokenRepository) GetUserToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	if m.GetUserTokenFunc != nil {
		return m.GetUserTokenFunc(ctx, userID)
	}
	return nil, nil
}

// SaveUserToken implements data.UserTokenRepository
func (m *MockUserTokenRepository) SaveUserToken(ctx context.Context, userID string, token *oauth2.Token) error {
	if m.SaveUserTokenFunc != nil {
		return m.SaveUserTokenFunc(ctx, userID, token)
	}
	return nil
}

func (m *MockUserTokenRepository) Destroy(w http.ResponseWriter, r *http.Request) error {
	if m.DestroyFunc != nil {
		return m.DestroyFunc(w, r)
	}
	return nil
}

func (m *MockUserTokenRepository) Start(w http.ResponseWriter, r *http.Request) (session.Session, error) {
	if m.StartFunc != nil {
		return m.StartFunc(w, r)
	}
	return nil, nil
}

func (m *MockUserTokenRepository) Refresh(w http.ResponseWriter, r *http.Request) error {
	if m.RefreshFunc != nil {
		return m.RefreshFunc(w, r)
	}
	return nil
}

func (m *MockUserTokenRepository) Store() session.Store {
	if m.StoreFunc != nil {
		return m.StoreFunc()
	}
	return nil
}
