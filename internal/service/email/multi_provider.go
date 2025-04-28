package email

import (
	"context"
	"fmt"
	"sort"

	"github.com/desponda/inbox-whisperer/internal/common"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/service/provider"
	"golang.org/x/oauth2"
)

// MultiProviderService aggregates multiple email providers into a single service
type MultiProviderService struct {
	factory *provider.Factory
}

// FetchSummaries returns a paginated list of email summaries for the first linked provider (for compatibility)
func (s *MultiProviderService) FetchSummaries(ctx context.Context, token *oauth2.Token) ([]models.EmailSummary, error) {
	userID, ok := common.UserIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("userID missing from context")
	}
	userLinks := s.factory.UserLinks(userID.String())
	if len(userLinks) == 0 {
		return nil, fmt.Errorf("no providers linked for user")
	}

	// Aggregate summaries from all providers (no deduplication)
	allSummaries := []models.EmailSummary{}
	for _, link := range userLinks {
		prov, err := s.factory.ProviderForUser(ctx, userID.String(), link.Type)
		if err != nil {
			return nil, err
		}
		summaries, err := prov.FetchSummaries(ctx, userID.String())
		if err != nil {
			return nil, err
		}
		allSummaries = append(allSummaries, summaries...)
	}
	// Sort: newest InternalDate first, then by ID for stability
	sort.Slice(allSummaries, func(i, j int) bool {
		if allSummaries[i].InternalDate == allSummaries[j].InternalDate {
			return allSummaries[i].ID < allSummaries[j].ID
		}
		return allSummaries[i].InternalDate > allSummaries[j].InternalDate
	})

	// Apply pagination from context
	params := common.PaginationFromContext(ctx)
	filtered := make([]models.EmailSummary, 0, len(allSummaries))
	for _, summary := range allSummaries {
		if params.AfterTimestamp > 0 && summary.InternalDate < params.AfterTimestamp {
			continue
		}
		if params.AfterID != "" && summary.ID <= params.AfterID {
			continue
		}
		filtered = append(filtered, summary)
	}
	if params.Limit > 0 && len(filtered) > params.Limit {
		filtered = filtered[:params.Limit]
	}
	return filtered, nil
}

// FetchMessage retrieves the full content of a single email message for the first linked provider
func (s *MultiProviderService) FetchMessage(ctx context.Context, token *oauth2.Token, id string) (*models.EmailMessage, error) {
	userID, ok := common.UserIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("userID missing from context")
	}
	userLinks := s.factory.UserLinks(userID.String())
	if len(userLinks) == 0 {
		return nil, fmt.Errorf("no providers linked for user")
	}
	prov, err := s.factory.ProviderForUser(ctx, userID.String(), userLinks[0].Type)
	if err != nil {
		return nil, err
	}
	return prov.FetchMessage(ctx, token, id)
}

// Change NewMultiProviderService to be exported
func NewMultiProviderService(factory *provider.Factory) Service {
	return &MultiProviderService{factory: factory}
}
