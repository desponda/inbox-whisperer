package service_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/desponda/inbox-whisperer/internal/common"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/service/email"
	"github.com/desponda/inbox-whisperer/internal/service/provider"
)

type dummyProvider struct {
	summaries []models.EmailSummary
}

func (d *dummyProvider) GetType() provider.Type {
	return provider.Gmail
}

func (d *dummyProvider) GetCapabilities() provider.Capabilities {
	return provider.Capabilities{
		SupportsSearch:    true,
		SupportsFolders:   true,
		SupportsLabels:    true,
		SupportsThreading: true,
	}
}

func (d *dummyProvider) FetchSummaries(ctx context.Context, userID string) ([]models.EmailSummary, error) {
	return d.summaries, nil
}

func (d *dummyProvider) FetchMessage(ctx context.Context, token interface{}, messageID string) (*models.EmailMessage, error) {
	return &models.EmailMessage{}, nil
}

func TestMultiProviderEmailService_FetchMessages_ProviderIDUniqueness(t *testing.T) {
	factory := provider.NewProviderFactory()
	userUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	userID := userUUID.String()
	p1 := &dummyProvider{summaries: []models.EmailSummary{
		{ID: "a", ThreadID: "t1", Subject: "S1", Sender: "f1", Snippet: "x", InternalDate: 200, Provider: "gmail"},
		{ID: "b", ThreadID: "t2", Subject: "S2", Sender: "f2", Snippet: "y", InternalDate: 100, Provider: "gmail"},
	}}
	p2 := &dummyProvider{summaries: []models.EmailSummary{
		{ID: "a", ThreadID: "t1", Subject: "S1-new", Sender: "f1", Snippet: "x2", InternalDate: 300, Provider: "outlook"},
		{ID: "c", ThreadID: "t3", Subject: "S3", Sender: "f3", Snippet: "z", InternalDate: 150, Provider: "outlook"},
	}}
	factory.RegisterProvider(provider.Gmail, func(cfg provider.Config) (provider.Provider, error) { return p1, nil })
	factory.RegisterProvider(provider.Outlook, func(cfg provider.Config) (provider.Provider, error) { return p2, nil })
	factory.LinkProvider(userID, provider.Config{UserID: userID, Type: provider.Gmail})
	factory.LinkProvider(userID, provider.Config{UserID: userID, Type: provider.Outlook})
	svc := email.NewMultiProviderService(factory)
	ctx := common.ContextWithUserID(context.Background(), userUUID)
	msgs, err := svc.FetchSummaries(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages (no deduplication), got %d", len(msgs))
	}
	// Ensure both 'a' from gmail and 'a' from outlook are present
	var foundGmailA, foundOutlookA bool
	for _, m := range msgs {
		if m.ID == "a" && m.Provider == "gmail" {
			foundGmailA = true
		}
		if m.ID == "a" && m.Provider == "outlook" {
			foundOutlookA = true
		}
	}
	if !foundGmailA || !foundOutlookA {
		t.Errorf("expected both gmail and outlook 'a' messages, got: %+v", msgs)
	}
}
