package gmail

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/models"
	"golang.org/x/oauth2"
)

// mockGmailAPI implements GmailAPI for unit tests
// Only implements UsersMessagesGet for error/success simulation

import "google.golang.org/api/gmail/v1"

type mockGmailAPI struct {
	msg      *gmail.Message
	err      error
	listResp *gmail.ListMessagesResponse
	listErr  error
	msgMap   map[string]*gmail.Message
	getErr   error
}

func (m *mockGmailAPI) UsersMessagesGet(userID, msgID string) interface {
	Do(context.Context) (*gmail.Message, error)
} {
	if m.msgMap != nil {
		return &mockUsersMessagesGetCall{msg: m.msgMap[msgID], err: m.getErr}
	}
	return &mockUsersMessagesGetCall{msg: m.msg, err: m.err}
}

func (m *mockGmailAPI) UsersMessagesList(userID string) interface {
	Do() (*gmail.ListMessagesResponse, error)
} {
	return &mockUsersMessagesListCall{resp: m.listResp, err: m.listErr}
}

type mockUsersMessagesListCall struct {
	resp *gmail.ListMessagesResponse
	err  error
}

func (c *mockUsersMessagesListCall) Do() (*gmail.ListMessagesResponse, error) {
	return c.resp, c.err
}

type mockUsersMessagesGetCall struct {
	msg *gmail.Message
	err error
}

func (c *mockUsersMessagesGetCall) Do(ctx context.Context) (*gmail.Message, error) {
	return c.msg, c.err
}

func TestGmailService_FetchMessageContent_ErrorPaths(t *testing.T) {
	repo := &dummyRepo{}
	tok := &oauth2.Token{AccessToken: "dummy"}
	ctx := context.Background()
	testCases := []struct {
		name      string
		mockMsg   *gmail.Message
		mockErr   error
		inputID   string
		expectErr string // empty means expect nil error
		wantID    string // if expectErr is empty, check returned message ID
	}{
		{
			name:      "Gmail API not found error",
			mockMsg:   nil,
			mockErr:   errors.New("not found"),
			inputID:   "missing-id",
			expectErr: "not found",
		},
		{
			name:      "Gmail API timeout error",
			mockMsg:   nil,
			mockErr:   errors.New("timeout"),
			inputID:   "timeout-id",
			expectErr: "timeout",
		},
		{
			name:    "Gmail API success",
			mockMsg: &gmail.Message{Id: "abc", Snippet: "Test", Payload: &gmail.MessagePart{Headers: []*gmail.MessagePartHeader{{Name: "From", Value: "a@b.com"}, {Name: "To", Value: "c@d.com"}, {Name: "Subject", Value: "Test"}, {Name: "Date", Value: time.Now().Format(time.RFC1123Z)}}}},
			mockErr: nil,
			inputID: "abc",
			wantID:  "abc",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewGmailServiceWithAPI(repo, &mockGmailAPI{msg: tc.mockMsg, err: tc.mockErr})
			msg, err := svc.fetchGmailMessage(ctx, tok, tc.inputID)
			if tc.expectErr != "" {
				if err == nil || err.Error() != tc.expectErr {
					t.Errorf("%s: expected error '%s', got %v", tc.name, tc.expectErr, err)
				} else {
					t.Logf("%s: got expected error '%s'", tc.name, tc.expectErr)
				}
			} else {
				if err != nil {
					t.Errorf("%s: expected nil error, got %v", tc.name, err)
				}
				if msg == nil || msg.Id != tc.wantID {
					t.Errorf("%s: expected message with Id '%s', got %+v", tc.name, tc.wantID, msg)
				} else {
					t.Logf("%s: got expected message with Id '%s'", tc.name, tc.wantID)
				}
			}
		})
	}
}

func TestGmailService_syncLatestSummariesFromGmail(t *testing.T) {
	repo := &fakeUpsertRepo{}
	mockAPI := &mockGmailAPI{
		listResp: &gmail.ListMessagesResponse{
			Messages: []*gmail.Message{{Id: "id1"}},
		},
		msgMap: map[string]*gmail.Message{
			"id1": {
				Id:       "id1",
				ThreadId: "th1",
				Snippet:  "snippet1",
				Payload: &gmail.MessagePart{
					Headers: []*gmail.MessagePartHeader{
						{Name: "Subject", Value: "subject1"},
						{Name: "From", Value: "sender1@example.com"},
						{Name: "Date", Value: "2025-04-24T00:00:00Z"},
					},
				},
				InternalDate: 123456,
				HistoryId:    42,
			},
		},
	}

	svc := NewGmailServiceWithAPI(repo, mockAPI)
	ctx := context.Background()
	userID := "user1"
	tok := &oauth2.Token{AccessToken: "dummy"}

	// Error path: list call fails
	mockAPI.listErr = errors.New("list error")
	err := svc.syncLatestSummariesFromGmail(ctx, tok, userID, 10)
	if err == nil {
		t.Errorf("expected error from list call, got nil")
	}

	// Error path: get call fails (should skip errored messages, but not return error)
	mockAPI.listErr = nil
	mockAPI.getErr = errors.New("get error")
	repo.upsertCount = 0
	err = svc.syncLatestSummariesFromGmail(ctx, tok, userID, 10)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if repo.upsertCount != 0 {
		t.Errorf("expected no upserts when get fails, got %d", repo.upsertCount)
	}

	// Success path
	mockAPI.getErr = nil
	repo.upsertCount = 0
	err = svc.syncLatestSummariesFromGmail(ctx, tok, userID, 10)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if repo.upsertCount == 0 {
		t.Errorf("expected upsert to be called, got 0")
	}
}

type fakeUpsertRepo struct {
	upsertCount int
}

func (f *fakeUpsertRepo) UpsertMessage(ctx context.Context, msg *models.EmailMessage) error {
	f.upsertCount++
	return nil
}
func (f *fakeUpsertRepo) GetMessage(ctx context.Context, userID, msgID string) (*models.EmailMessage, error) {
	return nil, nil
}
func (f *fakeUpsertRepo) GetMessageByID(ctx context.Context, userID, msgID string) (*models.EmailMessage, error) {
	return nil, nil
}
func (f *fakeUpsertRepo) ListMessages(ctx context.Context, userID string, pageSize int, afterInternalDate int64, afterID string) ([]*models.EmailMessage, error) {
	return nil, nil
}
func (f *fakeUpsertRepo) GetMessagesForUser(ctx context.Context, userID string, pageSize int, afterInternalDate int) ([]*models.EmailMessage, error) {
	return nil, nil
}
func (f *fakeUpsertRepo) GetMessagesForUserCursor(ctx context.Context, userID string, pageSize int, afterInternalDate int64, afterID string) ([]*models.EmailMessage, error) {
	return nil, nil
}
func (f *fakeUpsertRepo) DeleteMessagesForUser(ctx context.Context, userID string) error { return nil }

type dummyRepo struct{}

func (d *dummyRepo) UpsertMessage(ctx context.Context, msg *models.EmailMessage) error { return nil }
func (d *dummyRepo) GetMessage(ctx context.Context, userID, msgID string) (*models.EmailMessage, error) {
	return nil, errors.New("not found")
}
func (d *dummyRepo) GetMessageByID(ctx context.Context, userID, msgID string) (*models.EmailMessage, error) {
	return nil, errors.New("not found")
}
func (d *dummyRepo) ListMessages(ctx context.Context, userID string, pageSize int, afterInternalDate int64, afterID string) ([]*models.EmailMessage, error) {
	return nil, nil
}
func (d *dummyRepo) GetMessagesForUser(ctx context.Context, userID string, pageSize int, afterInternalDate int) ([]*models.EmailMessage, error) {
	return nil, nil
}
func (d *dummyRepo) GetMessagesForUserCursor(ctx context.Context, userID string, pageSize int, afterInternalDate int64, afterID string) ([]*models.EmailMessage, error) {
	return nil, nil
}
func (d *dummyRepo) DeleteMessagesForUser(ctx context.Context, userID string) error { return nil }
