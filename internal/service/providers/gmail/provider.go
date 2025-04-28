package gmail

import (
	"context"
	"errors"

	"github.com/desponda/inbox-whisperer/internal/common"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/service/provider"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

var (
	ErrInvalidToken = errors.New("invalid or missing OAuth token")
)

// FetchParams represents Gmail-specific pagination parameters
type FetchParams struct {
	AfterID           string
	AfterInternalDate int64
	Limit             int
}

// Provider-specific context key for Gmail parameters
type gmailParamsKey struct{}

// GmailProviderAPI defines the Gmail-specific operations
// This is an internal interface specific to Gmail implementation
type GmailProviderAPI interface {
	// FetchSummaries retrieves email summaries from Gmail
	// The implementation handles Gmail-specific pagination and filtering
	FetchSummaries(ctx context.Context, userID string) ([]models.EmailSummary, error)

	// FetchMessage retrieves a single email from Gmail
	// Requires a valid Gmail OAuth token
	FetchMessage(ctx context.Context, userToken interface{}, messageID string) (*models.EmailMessage, error)
}

// Provider implements the provider.Provider interface for Gmail
type Provider struct {
	service *MessageService // Handles Gmail message operations
}

// Ensure Provider implements provider.Provider at compile time
var _ provider.Provider = (*Provider)(nil)

// NewProvider creates a new Gmail provider instance
func NewProvider(service *MessageService) *Provider {
	return &Provider{service: service}
}

// GetType returns Gmail as the provider type
func (g *Provider) GetType() provider.Type {
	return provider.Gmail
}

// GetCapabilities returns Gmail's supported features
func (g *Provider) GetCapabilities() provider.Capabilities {
	return provider.Capabilities{
		SupportsSearch:    true,
		SupportsFolders:   true,
		SupportsLabels:    true,
		SupportsThreading: true,
	}
}

// FetchSummaries fetches email summaries from Gmail
func (g *Provider) FetchSummaries(ctx context.Context, userID string) ([]models.EmailSummary, error) {
	uuidVal, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid userID format")
	}
	// Add user ID and provider to context
	ctx = common.ContextWithUserID(ctx, uuidVal)
	ctx = common.ContextWithProvider(ctx, "gmail")

	// Get pagination parameters from context
	params := common.PaginationFromContext(ctx)

	// Convert to Gmail-specific params and store in context
	ctx = context.WithValue(ctx, gmailParamsKey{}, &FetchParams{
		AfterID:           params.AfterID,
		AfterInternalDate: params.AfterTimestamp,
		Limit:             params.Limit,
	})

	token, err := g.getTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}
	msgs, err := g.service.FetchMessages(ctx, token)
	if err != nil {
		return nil, err
	}

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
			Provider:     "gmail", // Always set for Gmail provider
		})
	}
	return summaries, nil
}

// FetchMessage fetches a full email message from Gmail
func (g *Provider) FetchMessage(ctx context.Context, userToken interface{}, messageID string) (*models.EmailMessage, error) {
	token, ok := userToken.(*oauth2.Token)
	if !ok || token == nil {
		return nil, ErrInvalidToken
	}
	return g.service.FetchMessageContent(ctx, token, messageID)
}

// Helper method to get OAuth token from context
func (g *Provider) getTokenFromContext(ctx context.Context) (*oauth2.Token, error) {
	tokenVal := ctx.Value(common.OAuthTokenKey{})
	if tokenVal == nil {
		return nil, ErrInvalidToken
	}

	token, ok := tokenVal.(*oauth2.Token)
	if !ok || token == nil {
		return nil, ErrInvalidToken
	}

	return token, nil
}
