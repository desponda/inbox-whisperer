package mocks

import (
	"context"
	"golang.org/x/oauth2"
)

type MockUserTokenRepository struct{}

func (m *MockUserTokenRepository) GetUserToken(ctx context.Context, userID string) (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: "mock-token"}, nil
}

func (m *MockUserTokenRepository) SaveUserToken(ctx context.Context, userID string, token *oauth2.Token) error {
	return nil
}
