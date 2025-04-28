package email

import (
	"context"
	"errors"

	"github.com/desponda/inbox-whisperer/internal/models"
	"golang.org/x/oauth2"
)

// Common errors
var (
	ErrNotFound = errors.New("email not found")
)

// Service defines the interface for email operations
type Service interface {
	// FetchSummaries returns a paginated list of email summaries
	// Pagination parameters are extracted from the context
	FetchSummaries(ctx context.Context, token *oauth2.Token) ([]models.EmailSummary, error)

	// FetchMessage retrieves the full content of a single email message
	FetchMessage(ctx context.Context, token *oauth2.Token, id string) (*models.EmailMessage, error)
}
