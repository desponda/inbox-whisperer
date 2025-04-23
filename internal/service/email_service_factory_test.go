package service_test

import (
	"context"
	"testing"
	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/desponda/inbox-whisperer/internal/service/gmail"
	"github.com/desponda/inbox-whisperer/internal/models"
)

type dummyProvider struct {
	summaries []models.EmailSummary
}

func (d *dummyProvider) FetchSummaries(ctx context.Context, userID string, params gmail.FetchParams) ([]models.EmailSummary, error) {
	return d.summaries, nil
}
func (d *dummyProvider) FetchMessage(ctx context.Context, token interface{}, messageID string) (*models.EmailMessage, error) {
	return &models.EmailMessage{}, nil
}

func TestMultiProviderEmailService_FetchMessages_DeduplicationAndSorting(t *testing.T) {
	factory := service.NewEmailProviderFactory()
	p1 := &dummyProvider{summaries: []models.EmailSummary{
		{ID: "a", ThreadID: "t1", Subject: "S1", Sender: "f1", Snippet: "x", InternalDate: 200, Provider: "gmail"},
		{ID: "b", ThreadID: "t2", Subject: "S2", Sender: "f2", Snippet: "y", InternalDate: 100, Provider: "gmail"},
	}}
	p2 := &dummyProvider{summaries: []models.EmailSummary{
		{ID: "a", ThreadID: "t1", Subject: "S1-new", Sender: "f1", Snippet: "x2", InternalDate: 300, Provider: "outlook"},
		{ID: "c", ThreadID: "t3", Subject: "S3", Sender: "f3", Snippet: "z", InternalDate: 150, Provider: "outlook"},
	}}
	factory.RegisterProvider(service.ProviderGmail, func(cfg service.ProviderConfig) (service.EmailProvider, error) { return p1, nil })
	factory.RegisterProvider(service.ProviderOutlook, func(cfg service.ProviderConfig) (service.EmailProvider, error) { return p2, nil })
	factory.LinkProvider("user", service.ProviderConfig{UserID: "user", Type: service.ProviderGmail})
	factory.LinkProvider("user", service.ProviderConfig{UserID: "user", Type: service.ProviderOutlook})
	svc := service.NewMultiProviderEmailService(factory)
	ctx := context.WithValue(context.Background(), service.CtxKeyUserID{}, "user")
	msgs, err := svc.FetchMessages(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 deduped messages, got %d", len(msgs))
	}
	if msgs[0].EmailMessageID != "a" || msgs[0].InternalDate != 300 {
		t.Errorf("expected newest 'a' first, got %+v", msgs[0])
	}
	if msgs[1].EmailMessageID != "c" || msgs[2].EmailMessageID != "b" {
		t.Errorf("expected order c then b, got %+v, %+v", msgs[1], msgs[2])
	}
}
