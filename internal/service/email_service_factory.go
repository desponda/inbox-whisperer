package service

import (
	"context"
	"sort"
	"sync"
	"time"
	"fmt"
	"golang.org/x/oauth2"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/service/gmail"
)

type CtxKeyUserID struct{}
type CtxKeyLimit struct{}

var summaryCache sync.Map // per-user summary cache



type MultiProviderEmailService struct {
	Factory *EmailProviderFactory
}

func NewMultiProviderEmailService(factory *EmailProviderFactory) *MultiProviderEmailService {
	return &MultiProviderEmailService{Factory: factory}
}



func (s *MultiProviderEmailService) FetchMessages(ctx context.Context, token *oauth2.Token) ([]models.EmailMessage, error) {
	userIDVal := ctx.Value(CtxKeyUserID{})
if userIDVal == nil {
	return nil, fmt.Errorf("userID context key missing")
}
userID, ok := userIDVal.(string)
if !ok {
	return nil, fmt.Errorf("userID context key not a string")
}
	providers, err := s.Factory.ProvidersForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	limit := 10
if l, ok := ctx.Value(CtxKeyLimit{}).(int); ok && l > 0 {
	limit = l
}
params := gmail.FetchParams{Limit: limit} // limit extracted from context, default 10
	allSummaries := make([]models.EmailSummary, 0)
	for _, prov := range providers {
		summaries, err := prov.FetchSummaries(ctx, userID, params)
		if err != nil {
			continue // skip errored providers
		}
		for _, s := range summaries {
			allSummaries = append(allSummaries, models.EmailSummary{
				ID:           s.ID,
				ThreadID:     s.ThreadID,
				Subject:      s.Subject,
				Sender:       s.Sender,
				Snippet:      s.Snippet,
				InternalDate: s.InternalDate,
				Date:         "", // Gmail summary doesn't yet provide Date
				Provider:     s.Provider,
			})
		}
	}
	// Deduplicate by ID (favor newest InternalDate)
	deduped := make(map[string]models.EmailSummary)
	for _, s := range allSummaries {
		if prev, ok := deduped[s.ID]; !ok || s.InternalDate > prev.InternalDate {
			deduped[s.ID] = s
		}
	}
	// Convert to slice and sort by InternalDate desc
	result := make([]models.EmailSummary, 0, len(deduped))
	for _, v := range deduped {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].InternalDate > result[j].InternalDate })
	// Caching summary results for efficiency
	cacheKey := userID
	if l, ok := ctx.Value(CtxKeyLimit{}).(int); ok && l > 0 {
		cacheKey = fmt.Sprintf("%s:%d", userID, l)
	}
	type cacheEntry struct {
		Summaries []models.EmailMessage
		Expires  time.Time
	}
	if v, ok := summaryCache.Load(cacheKey); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.Expires) {
			return entry.Summaries, nil
		}
	}

	// Convert to []models.EmailMessage for now
final := make([]models.EmailMessage, len(result))
	for i, s := range result {
		final[i] = models.EmailMessage{
			EmailMessageID: s.ID,
			ThreadID:       s.ThreadID,
			Subject:        s.Subject,
			Sender:         s.Sender,
			Snippet:        s.Snippet,
			InternalDate:   s.InternalDate,
			Date:           s.Date,
			// ...other fields
		}
	}
	// Update cache for this user/limit
	summaryCache.Store(cacheKey, cacheEntry{
		Summaries: final,
		Expires:  time.Now().Add(60 * time.Second),
	})
	return final, nil
}

// FetchMessageContent fetches the full message from the right provider.
func (s *MultiProviderEmailService) FetchMessageContent(ctx context.Context, token *oauth2.Token, id string) (*models.EmailMessage, error) {
	userIDVal := ctx.Value(CtxKeyUserID{})
if userIDVal == nil {
	return nil, fmt.Errorf("userID context key missing")
}
userID, ok := userIDVal.(string)
if !ok {
	return nil, fmt.Errorf("userID context key not a string")
}
	providers, err := s.Factory.ProvidersForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, prov := range providers {
		msg, err := prov.FetchMessage(ctx, token, id)
		if err == nil {
			return &models.EmailMessage{
				EmailMessageID: msg.EmailMessageID,
				ThreadID: msg.ThreadID,
				Subject: msg.Subject,
				Sender: msg.Sender,
				Recipient: msg.Recipient,
				Snippet: msg.Snippet,
				Body: msg.Body,
				InternalDate: msg.InternalDate,
				Date: msg.Date,
				// ...other fields
			}, nil
		}
	}
	return nil, err
}
