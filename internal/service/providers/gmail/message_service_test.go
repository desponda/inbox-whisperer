package gmail

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/googleapi"
)

// mockGmailAPI implements GmailAPI for unit tests
type mockGmailAPI struct {
	msg      *gmail.Message
	err      error
	listResp *gmail.ListMessagesResponse
	listErr  error
	msgMap   map[string]*gmail.Message
	getErr   error
}

// Implements GmailAPI
func (m *mockGmailAPI) UsersMessagesGet(userID, msgID string) UsersMessagesGetCall {
	if m.err != nil {
		return &mockUsersMessagesGetCall{msg: nil, err: m.err}
	}
	if m.msgMap != nil {
		return &mockUsersMessagesGetCall{msg: m.msgMap[msgID], err: m.getErr}
	}
	return &mockUsersMessagesGetCall{msg: m.msg, err: nil}
}

type mockUsersMessagesGetCall struct {
	msg *gmail.Message
	err error
}

func (c *mockUsersMessagesGetCall) Do(...googleapi.CallOption) (*gmail.Message, error) {
	return c.msg, c.err
}

func (m *mockGmailAPI) UsersMessagesList(userID string) UsersMessagesListCall {
	return &mockUsersMessagesListCallWithPaging{allMessages: m.listResp, err: m.listErr}
}

type mockUsersMessagesListCallWithPaging struct {
	allMessages *gmail.ListMessagesResponse
	err         error
}

func (c *mockUsersMessagesListCallWithPaging) Do(...googleapi.CallOption) (*gmail.ListMessagesResponse, error) {
	if c.err != nil {
		return nil, c.err
	}
	if c.allMessages == nil {
		return &gmail.ListMessagesResponse{Messages: []*gmail.Message{}}, nil
	}
	return c.allMessages, nil
}

// mockToken returns a dummy OAuth2 token for tests
func mockToken() *oauth2.Token {
	return &oauth2.Token{AccessToken: "dummy", TokenType: "Bearer"}
}

type fakeUpsertRepo struct {
	upsertCount int
}

func (f *fakeUpsertRepo) UpsertMessage(ctx context.Context, msg *models.EmailMessage) error {
	f.upsertCount++
	return nil
}

func (f *fakeUpsertRepo) GetMessage(ctx context.Context, userID string, msgID string) (*models.EmailMessage, error) {
	return nil, nil
}

func (f *fakeUpsertRepo) GetMessageByID(ctx context.Context, userID, provider, msgID string) (*models.EmailMessage, error) {
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

func (f *fakeUpsertRepo) DeleteMessagesForUser(ctx context.Context, userID string) error {
	return nil
}

type dummyRepo struct{}

func (d *dummyRepo) UpsertMessage(ctx context.Context, msg *models.EmailMessage) error { return nil }
func (d *dummyRepo) GetMessage(ctx context.Context, userID, msgID string) (*models.EmailMessage, error) {
	return nil, nil
}
func (d *dummyRepo) GetMessageByID(ctx context.Context, userID, provider, msgID string) (*models.EmailMessage, error) {
	return nil, nil
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

func TestGmailService_FetchMessageContent_ErrorPaths(t *testing.T) {
	repo := &dummyRepo{}
	tok := &oauth2.Token{AccessToken: "dummy"}
	ctx := context.Background()
	ctx = ContextWithUserID(ctx, "11111111-1111-1111-1111-111111111111")
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
			mockAPI := &mockGmailAPI{msg: tc.mockMsg, err: tc.mockErr}
			svc := NewMessageService(repo, nil, WithGmailAPI(mockAPI))
			msg, err := svc.FetchMessageContent(ctx, tok, tc.inputID)
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
				if msg == nil || msg.EmailMessageID != tc.wantID {
					t.Errorf("%s: expected message with Id '%s', got %+v", tc.name, tc.wantID, msg)
				} else {
					t.Logf("%s: got expected message with Id '%s'", tc.name, tc.wantID)
				}
			}
		})
	}
}

func TestGmailService_CachingE2E(t *testing.T) {
	t.Log("[DEBUG] TestGmailService_CachingE2E: starting")
	db, cleanup := data.SetupTestDB(t)
	t.Log("[DEBUG] SetupTestDB done")
	defer cleanup()
	repo := data.NewEmailMessageRepositoryFromPool(db.Pool)
	t.Log("[DEBUG] NewEmailMessageRepositoryFromPool done")

	// Setup mock API with test message
	testMsg := &gmail.Message{
		Id: "test_msg_1",
		Payload: &gmail.MessagePart{
			Headers: []*gmail.MessagePartHeader{
				{Name: "Subject", Value: "Test Subject"},
				{Name: "From", Value: "test@example.com"},
				{Name: "Date", Value: time.Now().Format(time.RFC1123Z)},
			},
		},
		Snippet:      "Test snippet",
		ThreadId:     "thread_1",
		InternalDate: time.Now().Unix(),
	}

	mockAPI := &mockGmailAPI{
		msg: testMsg,
		listResp: &gmail.ListMessagesResponse{
			Messages: []*gmail.Message{testMsg},
		},
	}

	svc := NewMessageService(repo, nil, WithGmailAPI(mockAPI))
	t.Log("[DEBUG] NewMessageService done")
	ctx := context.Background()
	userID := "11111111-1111-1111-1111-111111111111"

	// Simulate session context using handler+middleware pattern
	tok := mockToken()
	testCtx := ContextWithUserID(ctx, userID)
	testCtx = ContextWithDebug(testCtx, true)

	// Insert user before upserts
	db.Pool.Exec(ctx, `INSERT INTO users (id, email) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`, userID, "test@example.com")

	// Test message fetch
	msg, err := svc.FetchMessageContent(testCtx, tok, "test_msg_1")
	if err != nil {
		t.Fatalf("FetchMessageContent failed: %v", err)
	}
	if msg.Subject != "Test Subject" {
		t.Errorf("expected subject 'Test Subject', got %q", msg.Subject)
	}
	if msg.Provider != "gmail" {
		t.Errorf("expected provider 'gmail', got %q", msg.Provider)
	}
}

func TestExtractPlainTextBody(t *testing.T) {
	plain := "Hello, world!"
	encoded := base64.RawURLEncoding.EncodeToString([]byte(plain))
	payload := &gmail.MessagePart{
		MimeType: "text/plain",
		Body:     &gmail.MessagePartBody{Data: encoded},
	}
	if got := extractPlainTextBody(payload); got != plain {
		t.Errorf("expected plain body, got %q", got)
	}

	// Multipart: text/plain inside Parts
	encoded2 := base64.RawURLEncoding.EncodeToString([]byte("Second part"))
	multi := &gmail.MessagePart{
		MimeType: "multipart/alternative",
		Parts: []*gmail.MessagePart{
			{
				MimeType: "text/html",
				Body:     &gmail.MessagePartBody{Data: base64.RawURLEncoding.EncodeToString([]byte("<b>HTML</b>"))},
			},
			{
				MimeType: "text/plain",
				Body:     &gmail.MessagePartBody{Data: encoded2},
			},
		},
	}
	if got := extractPlainTextBody(multi); got != "Second part" {
		t.Errorf("expected 'Second part', got %q", got)
	}

	// Nil payload
	if got := extractPlainTextBody(nil); got != "" {
		t.Errorf("expected empty for nil payload, got %q", got)
	}
	// No valid plain text
	empty := &gmail.MessagePart{MimeType: "multipart/alternative", Parts: []*gmail.MessagePart{}}
	if got := extractPlainTextBody(empty); got != "" {
		t.Errorf("expected empty for no plain text, got %q", got)
	}
}

func TestGetHeader(t *testing.T) {
	headers := []*gmail.MessagePartHeader{
		{Name: "From", Value: "sender@example.com"},
		{Name: "To", Value: "rcpt@example.com"},
		{Name: "Subject", Value: "Test subject"},
		{Name: "Date", Value: "Wed, 01 Jan 2025 00:00:00 +0000"},
	}
	cases := []struct {
		name, key, want string
	}{
		{"Exact match", "From", "sender@example.com"},
		{"Case-insensitive", "subject", "Test subject"},
		{"Not found", "X-Header", ""},
	}
	for _, tc := range cases {
		if got := getHeader(headers, tc.key); got != tc.want {
			t.Errorf("%s: getHeader(%q) = %q, want %q", tc.name, tc.key, got, tc.want)
		}
	}
}

func TestGmailService_Pagination(t *testing.T) {
	repo := &fakeUpsertRepo{}

	// Create test messages
	var messages []*gmail.Message
	for i := 0; i < 15; i++ {
		msg := &gmail.Message{
			Id:           fmt.Sprintf("msg_%02d", i),
			ThreadId:     fmt.Sprintf("thread_%02d", i),
			Snippet:      fmt.Sprintf("Snippet %02d", i),
			InternalDate: time.Now().Add(-time.Duration(i) * time.Hour).Unix(),
			Payload: &gmail.MessagePart{
				Headers: []*gmail.MessagePartHeader{
					{Name: "Subject", Value: fmt.Sprintf("Subject %02d", i)},
					{Name: "From", Value: "sender@example.com"},
					{Name: "Date", Value: time.Now().Format(time.RFC1123Z)},
				},
			},
		}
		messages = append(messages, msg)
	}

	// Setup mock API with paginated responses
	mockAPI := &mockGmailAPI{
		listResp: &gmail.ListMessagesResponse{
			Messages: messages[:10], // First page: 10 messages
		},
		msg: messages[0], // Set first message as default
	}

	svc := NewMessageService(repo, nil, WithGmailAPI(mockAPI))
	ctx := context.Background()
	userID := "11111111-1111-1111-1111-111111111111"
	tok := mockToken()
	testCtx := ContextWithUserID(ctx, userID)

	// Insert user before upserts
	// No need for real DB in this test since we're using fakeUpsertRepo
	err := svc.syncWithGmail(testCtx, tok, userID)
	if err != nil {
		t.Fatalf("syncWithGmail page 1 failed: %v", err)
	}
	if repo.upsertCount != 10 {
		t.Errorf("expected 10 upserts on page 1, got %d", repo.upsertCount)
	}

	// Second page
	repo.upsertCount = 0
	mockAPI.listResp = &gmail.ListMessagesResponse{
		Messages: messages[10:], // Next 5 messages
	}
	mockAPI.msg = messages[10] // Update default message
	ctx2 := ContextWithPagination(testCtx, messages[9].Id, messages[9].InternalDate)
	err = svc.syncWithGmail(ctx2, tok, userID)
	if err != nil {
		t.Fatalf("syncWithGmail page 2 failed: %v", err)
	}
	if repo.upsertCount != 5 {
		t.Errorf("expected 5 upserts on page 2, got %d", repo.upsertCount)
	}

	// Third page (empty)
	repo.upsertCount = 0
	mockAPI.listResp = &gmail.ListMessagesResponse{
		Messages: nil,
	}
	ctx3 := ContextWithPagination(testCtx, messages[14].Id, messages[14].InternalDate)
	err = svc.syncWithGmail(ctx3, tok, userID)
	if err != nil {
		t.Fatalf("syncWithGmail page 3 failed: %v", err)
	}
	if repo.upsertCount != 0 {
		t.Errorf("expected 0 upserts on page 3, got %d", repo.upsertCount)
	}
}

func TestGmailService_SyncWithGmail(t *testing.T) {
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

	svc := NewMessageService(repo, nil, WithGmailAPI(mockAPI))
	ctx := context.Background()

	tok := &oauth2.Token{AccessToken: "dummy"}

	// Error path: list call fails
	mockAPI.listErr = errors.New("list error")
	err := svc.syncWithGmail(ctx, tok, "11111111-1111-1111-1111-111111111111")
	if err == nil {
		t.Errorf("expected error from list call, got nil")
	}

	// Error path: get call fails (should skip errored messages, but not return error)
	mockAPI.listErr = nil
	mockAPI.getErr = errors.New("get error")
	repo.upsertCount = 0
	err = svc.syncWithGmail(ctx, tok, "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if repo.upsertCount != 0 {
		t.Errorf("expected no upserts when get fails, got %d", repo.upsertCount)
	}

	// Success path
	mockAPI.getErr = nil
	repo.upsertCount = 0
	err = svc.syncWithGmail(ctx, tok, "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if repo.upsertCount == 0 {
		t.Errorf("expected upsert to be called, got 0")
	}
}

func TestGmailService_SyncWithGmail_WithUserInsert(t *testing.T) {
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

	svc := NewMessageService(repo, nil, WithGmailAPI(mockAPI))
	ctx := context.Background()

	tok := &oauth2.Token{AccessToken: "dummy"}

	// Error path: list call fails
	mockAPI.listErr = errors.New("list error")
	err := svc.syncWithGmail(ctx, tok, "11111111-1111-1111-1111-111111111111")
	if err == nil {
		t.Errorf("expected error from list call, got nil")
	}

	// Error path: get call fails (should skip errored messages, but not return error)
	mockAPI.listErr = nil
	mockAPI.getErr = errors.New("get error")
	repo.upsertCount = 0
	err = svc.syncWithGmail(ctx, tok, "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if repo.upsertCount != 0 {
		t.Errorf("expected no upserts when get fails, got %d", repo.upsertCount)
	}

	// Success path
	mockAPI.getErr = nil
	repo.upsertCount = 0
	err = svc.syncWithGmail(ctx, tok, "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if repo.upsertCount == 0 {
		t.Errorf("expected upsert to be called, got 0")
	}
}
