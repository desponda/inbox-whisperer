package gmail

import (
	"context"
	"errors"

	"github.com/desponda/inbox-whisperer/internal/models"
	"golang.org/x/oauth2"
)

// Use models.EmailSummary (moved to internal/models/email_summary.go)



// FetchParams can include pagination, filters, etc.
type FetchParams struct {
	AfterID           string
	AfterInternalDate int64
	Limit             int
}

// EmailProvider abstracts any email provider (Gmail, Outlook, etc.)
type EmailProvider interface {
	// FetchSummaries fetches summary info for a user's emails (minimal fields, paginated)
	FetchSummaries(ctx context.Context, userID string, params FetchParams) ([]models.EmailSummary, error)

	// FetchMessage fetches the full content of a specific email by ID
	// userToken should be *oauth2.Token for Gmail, string for others, etc.
	FetchMessage(ctx context.Context, userToken interface{}, messageID string) (*models.EmailMessage, error)
}

// GmailProvider implements EmailProvider for Gmail
// (wraps existing GmailService logic, refactored)
type GmailProvider struct {
	Service *GmailService
}

func NewGmailProvider(svc *GmailService) *GmailProvider {
	return &GmailProvider{Service: svc}
}

// FetchSummaries implements EmailProvider for Gmail
// FetchSummaries delegates to GmailService.FetchMessages, which handles caching and background sync.
// This ensures all caching and sync logic is centralized in the service.
func (g *GmailProvider) FetchSummaries(ctx context.Context, userID string, params FetchParams) ([]models.EmailSummary, error) {
	// Pass userID and limit via context
	ctx = context.WithValue(ctx, "user_id", userID)
	ctx = context.WithValue(ctx, "limit", params.Limit)
	msgs, err := g.Service.FetchMessages(ctx, nil) // token is nil for cached fetch
	if err != nil {
		return nil, err
	}
	// Map []EmailMessage to []EmailSummary
	summaries := make([]models.EmailSummary, 0, len(msgs))
	for _, m := range msgs {
		summaries = append(summaries, models.EmailSummary{
			ID:           m.EmailMessageID,
			ThreadID:     m.ThreadID,
			Snippet:      m.Snippet,
			Sender:       m.Sender,
			Subject:      m.Subject,
			InternalDate: m.InternalDate,
			Date:         m.Date,
			Provider:     "gmail",
		})
	}
	return summaries, nil
}

// FetchMessage implements EmailProvider for Gmail
// userToken should be *oauth2.Token, not string
func (g *GmailProvider) FetchMessage(ctx context.Context, userToken interface{}, messageID string) (*models.EmailMessage, error) {
	token, ok := userToken.(*oauth2.Token)
	if !ok {
		return nil, errors.New("invalid token type for GmailProvider.FetchMessage")
	}
	msg, err := g.Service.FetchMessageContent(ctx, token, messageID)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
