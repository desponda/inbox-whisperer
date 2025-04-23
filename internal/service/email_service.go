package service

import "github.com/desponda/inbox-whisperer/internal/models"

import (
	"context"
	"golang.org/x/oauth2"
)

// EmailService defines generic email operations.
type EmailService interface {
	FetchMessages(ctx context.Context, token *oauth2.Token) ([]models.EmailMessage, error)
	FetchMessageContent(ctx context.Context, token *oauth2.Token, id string) (*models.EmailMessage, error)
}
