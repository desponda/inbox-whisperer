package mocks

import (
	"context"
	"github.com/desponda/inbox-whisperer/internal/models"
	"golang.org/x/oauth2"
)

type MockEmailService struct {
	FetchMessagesFunc       func(ctx context.Context, token *oauth2.Token) ([]models.EmailMessage, error)
	FetchMessageContentFunc func(ctx context.Context, token *oauth2.Token, id string) (*models.EmailMessage, error)
}

func (m *MockEmailService) FetchMessages(ctx context.Context, token *oauth2.Token) ([]models.EmailMessage, error) {
	if m.FetchMessagesFunc != nil {
		return m.FetchMessagesFunc(ctx, token)
	}
	return nil, nil
}

func (m *MockEmailService) FetchMessageContent(ctx context.Context, token *oauth2.Token, id string) (*models.EmailMessage, error) {
	if m.FetchMessageContentFunc != nil {
		return m.FetchMessageContentFunc(ctx, token, id)
	}
	return nil, nil
}
