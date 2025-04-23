package gmail

import (
	"context"
	"errors"

	"github.com/desponda/inbox-whisperer/internal/models"
	"golang.org/x/oauth2"
)

type ctxKey string

const (
	ctxKeyUserID ctxKey = "user_id"
	ctxKeyLimit  ctxKey = "limit"
)

type FetchParams struct {
	AfterID           string
	AfterInternalDate int64
	Limit             int
}

type EmailProvider interface {
	FetchSummaries(ctx context.Context, userID string, params FetchParams) ([]models.EmailSummary, error)
	FetchMessage(ctx context.Context, userToken interface{}, messageID string) (*models.EmailMessage, error)
}

type GmailProvider struct {
	Service *GmailService
}

func NewGmailProvider(svc *GmailService) *GmailProvider {
	return &GmailProvider{Service: svc}
}

func (g *GmailProvider) FetchSummaries(ctx context.Context, userID string, params FetchParams) ([]models.EmailSummary, error) {
	ctx = context.WithValue(ctx, ctxKeyUserID, userID)
	ctx = context.WithValue(ctx, ctxKeyLimit, params.Limit)
	msgs, err := g.Service.FetchMessages(ctx, nil)
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
			Provider:     "gmail",
		})
	}
	return summaries, nil
}

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
