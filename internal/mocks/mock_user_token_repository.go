package mocks

import (
	"context"
	"golang.org/x/oauth2"
)

type MockUserTokenRepository struct {
	GetUserTokenFunc    func(ctx context.Context, userID string) (*oauth2.Token, error)
	SaveUserTokenFunc   func(ctx context.Context, userID string, token *oauth2.Token) error
}

func (m *MockUserTokenRepository) GetUserToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	if m.GetUserTokenFunc != nil {
		return m.GetUserTokenFunc(ctx, userID)
	}
	return &oauth2.Token{AccessToken: "mock-token"}, nil
}

func (m *MockUserTokenRepository) SaveUserToken(ctx context.Context, userID string, token *oauth2.Token) error {
	if m.SaveUserTokenFunc != nil {
		return m.SaveUserTokenFunc(ctx, userID, token)
	}
	return nil
}
