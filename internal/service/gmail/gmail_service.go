package gmail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/session"
	"github.com/desponda/inbox-whisperer/internal/notify"
	
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var ErrNotFound = errors.New("not found")

// extractUserIDFromContext gets the user ID from context
func extractUserIDFromContext(ctx context.Context) string {
	return session.GetUserID(ctx)
}

// GmailService fetches and caches Gmail messages for a user (DB-backed, 1-min TTL)
//go:generate mockgen -destination=internal/service/mocks/mock_gmail_api.go -package=mocks . GmailAPI

type GmailAPI interface {
	UsersMessagesGet(userID, msgID string) interface{
		Do(ctx context.Context) (*gmail.Message, error)
	}
}

type UsersMessagesGetCall interface {
	Do(ctx context.Context) (*gmail.Message, error)
}

type GmailService struct {
	Repo data.EmailMessageRepository
	GmailAPI GmailAPI // optional: for testing
}


func NewGmailService(repo data.EmailMessageRepository) *GmailService {
	return &GmailService{Repo: repo, GmailAPI: nil}
}

// NewGmailServiceWithAPI injects a mock GmailAPI (for tests)
func NewGmailServiceWithAPI(repo data.EmailMessageRepository, api GmailAPI) *GmailService {
	return &GmailService{Repo: repo, GmailAPI: api}
}

// getGmailClient creates a Gmail API client from an OAuth2 token
func getGmailClient(ctx context.Context, token *oauth2.Token) (*gmail.Service, error) {
	ts := oauth2.StaticTokenSource(token)
	return gmail.NewService(ctx, option.WithTokenSource(ts))
}

// MessageSummary is a minimal summary of a Gmail message
type MessageSummary struct {
	ID        string `json:"id"`
	ThreadID  string `json:"thread_id"`
	Snippet   string `json:"snippet"`
	From      string `json:"from"`
	Subject   string `json:"subject"`
	InternalDate int64 `json:"internal_date"`
	CursorAfterID string `json:"cursor_after_id,omitempty"`
	CursorAfterInternalDate int64 `json:"cursor_after_internal_date,omitempty"`
}

// MessageContent is the full content of a Gmail message


// FetchMessageContent fetches the full content of a Gmail message by ID, using cache if fresh.
func (s *GmailService) FetchMessageContent(ctx context.Context, token *oauth2.Token, id string) (*models.EmailMessage, error) {
	userID := extractUserIDFromContext(ctx)
	cached, err := s.Repo.GetMessageByID(ctx, userID, id)
	if err == nil && cached != nil && time.Since(cached.CachedAt) < time.Minute {
		return cached, nil
	}
	msg, err := s.fetchGmailMessage(ctx, token, id)
	if err != nil {
		return nil, err
	}
	dbMsg := &models.EmailMessage{
		UserID:         userID,
		EmailMessageID: msg.Id,
		ThreadID:       msg.ThreadId,
		Subject:        getHeader(msg.Payload.Headers, "Subject"),
		Sender:         getHeader(msg.Payload.Headers, "From"),
		Recipient:      getHeader(msg.Payload.Headers, "To"),
		Snippet:        msg.Snippet,
		Body:           extractPlainTextBody(msg.Payload),
		InternalDate:   msg.InternalDate,
		Date:           getHeader(msg.Payload.Headers, "Date"),
		HistoryID:      int64(msg.HistoryId),
		CachedAt:       time.Now(),
		RawJSON:        mustMarshalRawJSON(msg),
	}
	// Update cache asynchronously (log error if any)
	go func() {
		if err := s.Repo.UpsertMessage(ctx, dbMsg); err != nil {
			log.Printf("failed to upsert message: %v", err)
		}
	}()
	return dbMsg, nil
}

// mustMarshalRawJSON marshals a Gmail message to json.RawMessage (for RawJSON field), returns empty on error
func mustMarshalRawJSON(msg *gmail.Message) json.RawMessage {
	b, err := json.Marshal(msg)
	if err != nil {
		return json.RawMessage([]byte("{}"))
	}
	return b
}

// fetchGmailMessage fetches a Gmail message using the Gmail API or injected mock
func (s *GmailService) fetchGmailMessage(ctx context.Context, token *oauth2.Token, id string) (*gmail.Message, error) {
	// Prefer injected GmailAPI if present
	if s.GmailAPI == nil {
		return fetchGmailMessageClient(ctx, token, id)
	}
	call := s.GmailAPI.UsersMessagesGet("me", id)
	msg, err := call.Do(ctx)
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return msg, nil
}

// isNotFoundError returns true if the error represents a Gmail 404 not found
func isNotFoundError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "404")
}

// fetchGmailMessageClient fetches a Gmail message using the real Gmail client
func fetchGmailMessageClient(ctx context.Context, token *oauth2.Token, id string) (*gmail.Message, error) {
	client, err := getGmailClient(ctx, token)
	if err != nil {
		return nil, err
	}
	msg, err := client.Users.Messages.Get("me", id).Format("full").Do()
	if err != nil {
		if isNotFoundError(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return msg, nil
}




type CtxKeyAfterID struct{}
type CtxKeyAfterInternalDate struct{}

// FetchMessages returns only cached summaries (no full content/body) for a fast inbox load.
// It triggers a background sync with Gmail to fetch new/updated summaries.
// After sync, subsequent calls will see fresh data. Full content is fetched via FetchMessageContent.
func (s *GmailService) FetchMessages(ctx context.Context, token *oauth2.Token) ([]models.EmailMessage, error) {
	userID := extractUserIDFromContext(ctx)
	afterID, _ := ctx.Value(CtxKeyAfterID{}).(string)
	afterInternalDate, _ := ctx.Value(CtxKeyAfterInternalDate{}).(int64)
	pageSize := 10 // could be param

	// 1. Return cached summaries instantly
	msgs, err := s.fetchUserMessages(ctx, userID, pageSize, afterInternalDate, afterID)
	if err != nil {
		return nil, err
	}
	result := make([]models.EmailMessage, len(msgs))
	for i, m := range msgs {
		if m != nil {
			// Only summary fields; body is omitted
			result[i] = models.EmailMessage{
				EmailMessageID: m.EmailMessageID,
				ThreadID:       m.ThreadID,
				Subject:        m.Subject,
				Sender:         m.Sender,
				Snippet:        m.Snippet,
				InternalDate:   m.InternalDate,
				Date:           m.Date,
			}
		}
	}

	// 2. Trigger background sync for fresh Gmail data
	if token != nil {
		go func() {
			_ = s.syncLatestSummariesFromGmail(ctx, token, userID, pageSize)
			// Log error, but do not block user experience
		}()
	}

	return result, nil
}


// fetchUserMessages fetches messages for a user with cursor-based pagination (from DB only)
func (s *GmailService) fetchUserMessages(ctx context.Context, userID string, pageSize int, afterInternalDate int64, afterID string) ([]*models.EmailMessage, error) {
	if afterID != "" && afterInternalDate > 0 {
		return s.Repo.GetMessagesForUserCursor(ctx, userID, pageSize, afterInternalDate, afterID)
	}
	return s.Repo.GetMessagesForUserCursor(ctx, userID, pageSize, 0, "")
}

// syncLatestSummariesFromGmail fetches the latest message summaries from Gmail API and upserts them into the DB.
// This is run in the background after each inbox load for best UX.
func (s *GmailService) syncLatestSummariesFromGmail(ctx context.Context, token *oauth2.Token, userID string, pageSize int) error {
	// Use Gmail API's Users.Messages.List to get latest message IDs
	client, err := getGmailClient(ctx, token)
	if err != nil {
		return err
	}
	listCall := client.Users.Messages.List("me").MaxResults(int64(pageSize)).LabelIds("INBOX").Q("")
	resp, err := listCall.Do()
	if err != nil {
		return err
	}
	for _, m := range resp.Messages {
		// Fetch summary (minimal fields, not full body)
		msg, err := client.Users.Messages.Get("me", m.Id).Format("metadata").Do()
		if err != nil {
			continue // skip errored messages
		}
		dbMsg := &models.EmailMessage{
			UserID:         userID,
			EmailMessageID: msg.Id,
			ThreadID:       msg.ThreadId,
			Subject:        getHeader(msg.Payload.Headers, "Subject"),
			Sender:         getHeader(msg.Payload.Headers, "From"),
			Snippet:        msg.Snippet,
			InternalDate:   msg.InternalDate,
			Date:           getHeader(msg.Payload.Headers, "Date"),
			CachedAt:       time.Now(),
			RawJSON:        mustMarshalRawJSON(msg),
		}
		_ = s.Repo.UpsertMessage(ctx, dbMsg)
	}
	// Notify client (poll endpoint) after sync completes for instant refresh
	userID = extractUserIDFromContext(ctx)
	if userID != "" {
		notify.SetGmailSyncStatus(userID)
	}
	return nil
}



// getHeader returns the value for a given header name (case-insensitive)
func getHeader(headers []*gmail.MessagePartHeader, name string) string {
	for _, h := range headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}
	return ""
}

// extractPlainTextBody decodes the plain text body from a Gmail message payload
func extractPlainTextBody(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}
	// Single-part: try direct body
	if payload.Body != nil && payload.Body.Data != "" {
		if decoded, err := decodeGmailBody(payload.Body.Data); err == nil {
			return decoded
		}
	}
	// Multi-part: search for text/plain part
	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
			if decoded, err := decodeGmailBody(part.Body.Data); err == nil {
				return decoded
			}
		}
	}
	return ""
}


// decodeGmailBody decodes a base64url-encoded Gmail message body
func decodeGmailBody(data string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
