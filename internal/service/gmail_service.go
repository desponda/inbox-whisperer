package service

import (
	"context"
	"strings"
	"encoding/base64"
	"time"
	"errors"
	"encoding/json"

	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/session"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

)

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
	Repo data.GmailMessageRepository
	GmailAPI GmailAPI // optional: for testing
}


func NewGmailService(repo data.GmailMessageRepository) *GmailService {
	return &GmailService{Repo: repo, GmailAPI: nil}
}

// NewGmailServiceWithAPI injects a mock GmailAPI (for tests)
func NewGmailServiceWithAPI(repo data.GmailMessageRepository, api GmailAPI) *GmailService {
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
type MessageContent struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	From    string `json:"from"`
	To      string `json:"to"`
	Date    string `json:"date"`
	Body    string `json:"body"`
}

// FetchMessageContent fetches the full content of a Gmail message by ID, using cache if fresh.
func (s *GmailService) FetchMessageContent(ctx context.Context, token *oauth2.Token, id string) (*MessageContent, error) {
	userID := extractUserIDFromContext(ctx)
	if content, ok := s.tryCachedMessageContent(ctx, userID, id); ok {
		return content, nil
	}
	msg, err := s.fetchGmailMessage(ctx, token, id)
	if err != nil {
		return nil, err
	}
	content := buildMessageContent(msg)
	// Update cache asynchronously (ignore error)
	go s.cacheGmailMessage(ctx, userID, msg)
	return content, nil
}

// tryCachedMessageContent returns cached MessageContent if available and fresh
func (s *GmailService) tryCachedMessageContent(ctx context.Context, userID, id string) (*MessageContent, bool) {
	cached, err := s.Repo.GetMessageByID(ctx, userID, id)
	if err != nil || cached == nil || time.Since(cached.CachedAt) >= time.Minute {
		return nil, false
	}
	date := extractDateFromCachedMessage(cached)
	return &MessageContent{
		ID:      cached.GmailMessageID,
		Subject: cached.Subject,
		From:    cached.Sender,
		To:      cached.Recipient,
		Date:    date,
		Body:    cached.Body,
	}, true
}

// extractDateFromCachedMessage attempts to extract the Date header from cached.RawJSON, else returns empty string
func extractDateFromCachedMessage(cached *data.GmailMessage) string {
	if cached == nil || len(cached.RawJSON) == 0 {
		return ""
	}
	var raw struct {
		Payload struct {
			Headers []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"headers"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(cached.RawJSON, &raw); err != nil {
		return ""
	}
	for _, h := range raw.Payload.Headers {
		if strings.EqualFold(h.Name, "Date") {
			return h.Value
		}
	}
	return ""
}

// fetchGmailMessage fetches a Gmail message using the Gmail API or injected mock
func (s *GmailService) fetchGmailMessage(ctx context.Context, token *oauth2.Token, id string) (*gmail.Message, error) {
	// Prefer injected GmailAPI if present
	if s.GmailAPI == nil {
		return fetchGmailMessageClient(ctx, token, id)
	}
	call := s.GmailAPI.UsersMessagesGet("me", id)
	c, ok := call.(interface{ Do(context.Context) (*gmail.Message, error) })
	if !ok {
		return nil, errors.New("mock or injected GmailAPI does not implement Do(ctx)")
	}
	msg, err := c.Do(ctx)
	if err != nil {
		if isNotFoundError(err) {
			return nil, errors.New("not found")
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
			return nil, errors.New("not found")
		}
		return nil, err
	}
	return msg, nil
}

// buildMessageContent constructs MessageContent from a Gmail message
func buildMessageContent(msg *gmail.Message) *MessageContent {
	from := getHeader(msg.Payload.Headers, "From")
	to := getHeader(msg.Payload.Headers, "To")
	subject := getHeader(msg.Payload.Headers, "Subject")
	date := getHeader(msg.Payload.Headers, "Date")
	body := extractPlainTextBody(msg.Payload)
	return &MessageContent{
		ID:      msg.Id,
		Subject: subject,
		From:    from,
		To:      to,
		Date:    date,
		Body:    body,
	}
}

// cacheGmailMessage updates the Gmail message cache in the DB
func (s *GmailService) cacheGmailMessage(ctx context.Context, userID string, msg *gmail.Message) {
	dbMsg := &data.GmailMessage{
		UserID:         userID,
		GmailMessageID: msg.Id,
		ThreadID:       msg.ThreadId,
		Subject:        getHeader(msg.Payload.Headers, "Subject"),
		Sender:         getHeader(msg.Payload.Headers, "From"),
		Recipient:      getHeader(msg.Payload.Headers, "To"),
		Snippet:        msg.Snippet,
		Body:           extractPlainTextBody(msg.Payload),
		InternalDate:   msg.InternalDate,
		HistoryID:      int64(msg.HistoryId),
		CachedAt:       time.Now(),
	}
	_ = s.Repo.UpsertMessage(ctx, dbMsg)
}

// FetchMessages fetches the latest 10 messages (with cursor-based pagination)
func (s *GmailService) FetchMessages(ctx context.Context, token *oauth2.Token) ([]MessageSummary, error) {
	userID := extractUserIDFromContext(ctx)
	afterID, _ := ctx.Value("after_id").(string)
	afterInternalDate, _ := ctx.Value("after_internal_date").(int64)
	pageSize := 10 // could be param
	msgs, err := s.fetchUserMessages(ctx, userID, pageSize, afterInternalDate, afterID)
	if err != nil {
		return nil, err
	}
	return buildMessageSummaries(msgs), nil
}

// fetchUserMessages fetches messages for a user with cursor-based pagination
func (s *GmailService) fetchUserMessages(ctx context.Context, userID string, pageSize int, afterInternalDate int64, afterID string) ([]*data.GmailMessage, error) {
	if afterID != "" && afterInternalDate > 0 {
		return s.Repo.GetMessagesForUserCursor(ctx, userID, pageSize, afterInternalDate, afterID)
	}
	return s.Repo.GetMessagesForUserCursor(ctx, userID, pageSize, 0, "")
}

// buildMessageSummaries builds MessageSummary slices from []*data.GmailMessage
func buildMessageSummaries(msgs []*data.GmailMessage) []MessageSummary {
	summaries := make([]MessageSummary, 0, len(msgs))
	for _, m := range msgs {
		summaries = append(summaries, MessageSummary{
			ID:      m.GmailMessageID,
			ThreadID: m.ThreadID,
			Snippet:  m.Snippet,
			From:     m.Sender,
			Subject:  m.Subject,
			InternalDate: m.InternalDate,
			CursorAfterID: m.GmailMessageID,
			CursorAfterInternalDate: m.InternalDate,
		})
	}
	return summaries
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
