package gmail

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/session"
	"golang.org/x/oauth2"
	gmailapi "google.golang.org/api/gmail/v1"
)

type fakeRepo struct{}

func (f *fakeRepo) UpsertMessage(ctx context.Context, msg *models.EmailMessage) error { return nil }
func (f *fakeRepo) GetMessageByID(ctx context.Context, userID, emailMessageID string) (*models.EmailMessage, error) {
	return nil, nil
}
func (f *fakeRepo) GetMessagesForUser(ctx context.Context, userID string, limit, offset int) ([]*models.EmailMessage, error) {
	return nil, nil
}
func (f *fakeRepo) GetMessagesForUserCursor(ctx context.Context, userID string, limit int, afterInternalDate int64, afterMsgID string) ([]*models.EmailMessage, error) {
	return []*models.EmailMessage{{
		EmailMessageID: "id1",
		ThreadID:       "thread1",
		Snippet:        "snippet1",
		Sender:         "sender1@example.com",
		Subject:        "subject1",
		InternalDate:   123456,
		Date:           "2025-04-24T00:00:00Z",
	}}, nil
}

func (f *fakeRepo) DeleteMessagesForUser(ctx context.Context, userID string) error { return nil }

func TestGmailProvider_FetchSummaries(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewGmailService(repo, &mockGmailAPI{})
	provider := NewGmailProvider(svc)
	ctx := context.Background()
	ctx = session.ContextWithUserID(ctx, "user1")
	params := FetchParams{Limit: 10}

	summaries, err := provider.FetchSummaries(ctx, "user1", params)
	if err != nil {
		t.Fatalf("FetchSummaries failed: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].ID != "id1" {
		t.Errorf("expected summary ID 'id1', got %q", summaries[0].ID)
	}

	// Error path: underlying service error
	errRepo := &fakeRepoWithError{}
	svcErr := NewGmailService(errRepo, &mockGmailAPI{})
	providerErr := NewGmailProvider(svcErr)
	_, err = providerErr.FetchSummaries(ctx, "user1", params)
	if err == nil {
		t.Errorf("expected error from FetchSummaries, got nil")
	}
}

type fakeRepoWithError struct{}

func (f *fakeRepoWithError) UpsertMessage(ctx context.Context, msg *models.EmailMessage) error {
	return nil
}
func (f *fakeRepoWithError) GetMessageByID(ctx context.Context, userID, emailMessageID string) (*models.EmailMessage, error) {
	return nil, nil
}
func (f *fakeRepoWithError) GetMessagesForUser(ctx context.Context, userID string, limit, offset int) ([]*models.EmailMessage, error) {
	return nil, errors.New("repo error")
}
func (f *fakeRepoWithError) GetMessagesForUserCursor(ctx context.Context, userID string, limit int, afterInternalDate int64, afterMsgID string) ([]*models.EmailMessage, error) {
	return nil, errors.New("repo error")
}
func (f *fakeRepoWithError) DeleteMessagesForUser(ctx context.Context, userID string) error {
	return nil
}
func (f *fakeRepoWithError) SaveUserToken(ctx context.Context, userID, provider string, token interface{}) error {
	return nil
}
func (f *fakeRepoWithError) GetUserToken(ctx context.Context, userID, provider string) (interface{}, error) {
	return nil, nil
}

// ---

type fakeRepoForFetch struct {
	msg *models.EmailMessage
	err error
}

func (f *fakeRepoForFetch) UpsertMessage(ctx context.Context, msg *models.EmailMessage) error {
	return nil
}

func (f *fakeRepoForFetch) GetMessagesForUser(ctx context.Context, userID string, limit, offset int) ([]*models.EmailMessage, error) {
	return nil, nil
}
func (f *fakeRepoForFetch) GetMessagesForUserCursor(ctx context.Context, userID string, limit int, afterInternalDate int64, afterMsgID string) ([]*models.EmailMessage, error) {
	return nil, nil
}
func (f *fakeRepoForFetch) DeleteMessagesForUser(ctx context.Context, userID string) error {
	return nil
}

func (f *fakeRepoForFetch) SaveUserToken(ctx context.Context, userID, provider string, token interface{}) error {
	return nil
}
func (f *fakeRepoForFetch) GetUserToken(ctx context.Context, userID, provider string) (interface{}, error) {
	return nil, nil
}

// Only this method is relevant for FetchMessageContent
func (f *fakeRepoForFetch) GetMessageByID(ctx context.Context, userID, emailMessageID string) (*models.EmailMessage, error) {
	return f.msg, f.err
}

func TestGmailProvider_FetchMessage(t *testing.T) {
	repo := &fakeRepoForFetch{}
	mockMsg := &gmailapi.Message{
		Id:       "id42",
		Snippet:  "Test snippet",
		ThreadId: "th123",
		Payload: &gmailapi.MessagePart{
			Headers: []*gmailapi.MessagePartHeader{
				{Name: "Subject", Value: "Test subject"},
				{Name: "From", Value: "sender@example.com"},
				{Name: "To", Value: "rcpt@example.com"},
				{Name: "Date", Value: "Mon, 24 Apr 2023 12:34:56 +0000"},
			},
		},
		InternalDate: 1234567890,
		HistoryId:    42,
	}
	mockAPI := &mockGmailAPI{msg: mockMsg, err: nil}
	svc := NewGmailService(repo, mockAPI)
	provider := &GmailProvider{Service: svc}
	ctx := context.Background()
	ctx = session.ContextWithUserID(ctx, "user1")
	tok := &oauth2.Token{AccessToken: "dummy"}

	// Simulate cache miss (repo returns nil)
	repo.msg = nil
	repo.err = nil
	msg, err := provider.FetchMessage(ctx, tok, "id42")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if msg == nil || msg.EmailMessageID != "id42" {
		t.Errorf("expected message id 'id42', got %+v", msg)
	}

	// Simulate cache hit (repo returns cached message)
	repo.msg = &models.EmailMessage{
		UserID:         "user1",
		EmailMessageID: "id42",
		ThreadID:       "th123",
		Subject:        "Test subject",
		Sender:         "sender@example.com",
		Recipient:      "rcpt@example.com",
		Snippet:        "Test snippet",
		Body:           "Hello world!",
		InternalDate:   1234567890,
		Date:           "Mon, 24 Apr 2023 12:34:56 +0000",
		HistoryID:      42,
		CachedAt:       time.Now(),
		RawJSON:        []byte(`{"id":"id42"}`),
	}
	repo.err = nil
	msg, err = provider.FetchMessage(ctx, tok, "id42")
	if err != nil {
		t.Fatalf("expected no error on cache hit, got %v", err)
	}
	if msg == nil || msg.EmailMessageID != "id42" {
		t.Errorf("expected cached message id 'id42', got %+v", msg)
	}

	// Test error path
	repo.err = nil
	repo.msg = nil
	mockAPI.err = errors.New("fetch error")
	msg, err = provider.FetchMessage(ctx, tok, "id42")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}
