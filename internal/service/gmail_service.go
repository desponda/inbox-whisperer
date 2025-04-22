package service

import (
	"context"
	"fmt"
	"strings"
	"encoding/base64"
	"time"

	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/session"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// extractUserIDFromContext returns the user ID from context using the session package
func extractUserIDFromContext(ctx context.Context) string {
	return session.GetUserID(ctx)
}

// GmailService provides methods to fetch and cache Gmail messages for a user
// Caching is DB-backed and uses a 1-minute TTL for freshness
type GmailService struct {
	Repo data.GmailMessageRepository
}

func NewGmailService(repo data.GmailMessageRepository) *GmailService {
	return &GmailService{Repo: repo}
}

// getGmailClient returns a Gmail API client using the provided oauth2.Token
func getGmailClient(ctx context.Context, token *oauth2.Token) (*gmail.Service, error) {
	ts := oauth2.StaticTokenSource(token)
	return gmail.NewService(ctx, option.WithTokenSource(ts))
}

// MessageSummary contains minimal info about a Gmail message
// (expand as needed)
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

// MessageContent contains the full content of a Gmail message for display
// (matches OpenAPI EmailContent)
type MessageContent struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	From    string `json:"from"`
	To      string `json:"to"`
	Date    string `json:"date"`
	Body    string `json:"body"`
}

// FetchMessageContent fetches the full content of a Gmail message by ID
// This method uses helpers to keep orchestration clear and Gmail quirks isolated.
// FetchMessageContent returns the full content of a Gmail message, using cache if fresh.
func (s *GmailService) FetchMessageContent(ctx context.Context, token *oauth2.Token, id string) (*MessageContent, error) {
	userID := extractUserIDFromContext(ctx)
	cached, err := s.Repo.GetMessageByID(ctx, userID, id)
	if err == nil && cached != nil && time.Since(cached.CachedAt) < time.Minute {
		return &MessageContent{
			ID:      cached.GmailMessageID,
			Subject: cached.Subject,
			From:    cached.Sender,
			To:      cached.Recipient,
			Date:    "",
			Body:    cached.Body,
		}, nil
	}
	// Cache miss or stale; fetch from Gmail
	client, err := getGmailClient(ctx, token)
	if err != nil {
		return nil, err
	}
	msg, err := client.Users.Messages.Get("me", id).Format("full").Do()
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil, fmt.Errorf("not found")
		}
		return nil, err
	}
	from := getHeader(msg.Payload.Headers, "From")
	to := getHeader(msg.Payload.Headers, "To")
	subject := getHeader(msg.Payload.Headers, "Subject")
	date := getHeader(msg.Payload.Headers, "Date")
	body := extractPlainTextBody(msg.Payload)
	// Upsert cache
	dbMsg := &data.GmailMessage{
		UserID:         userID,
		GmailMessageID: msg.Id,
		ThreadID:       msg.ThreadId,
		Subject:        subject,
		Sender:         from,
		Recipient:      to,
		Snippet:        msg.Snippet,
		Body:           body,
		InternalDate:   msg.InternalDate,
		HistoryID:      int64(msg.HistoryId), // Convert uint64 to int64 for DB
		CachedAt:       time.Now(),
	}
	_ = s.Repo.UpsertMessage(ctx, dbMsg)
	return &MessageContent{
		ID:      msg.Id,
		Subject: subject,
		From:    from,
		To:      to,
		Date:    date,
		Body:    body,
	}, nil
}

// FetchMessages fetches the latest 10 messages using the user's OAuth2 token
// Uses helpers to keep orchestration clean.
// FetchMessages returns a list of message summaries, using cache if fresh.
// FetchMessages supports cursor-based pagination for snappy, scalable inbox browsing.
// If afterID and afterInternalDate are provided in context, paginates after that message.
func (s *GmailService) FetchMessages(ctx context.Context, token *oauth2.Token) ([]MessageSummary, error) {
	userID := extractUserIDFromContext(ctx)
	// Pagination params from context (set by handler)
	afterID, _ := ctx.Value("after_id").(string)
	afterInternalDate, _ := ctx.Value("after_internal_date").(int64)
	pageSize := 10 // could be param
	var msgs []*data.GmailMessage
	var err error
	if afterID != "" && afterInternalDate > 0 {
		msgs, err = s.Repo.GetMessagesForUserCursor(ctx, userID, pageSize, afterInternalDate, afterID)
	} else {
		msgs, err = s.Repo.GetMessagesForUserCursor(ctx, userID, pageSize, 0, "")
	}
	if err != nil {
		return nil, err
	}
	var summaries []MessageSummary
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
	return summaries, nil
}

// getHeader finds the value for a given header name from Gmail headers (case-insensitive)
func getHeader(headers []*gmail.MessagePartHeader, name string) string {
	for _, h := range headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}
	return ""
}

// extractPlainTextBody returns the decoded plain text body from a Gmail message payload
// Handles both single-part and multi-part messages. Gmail API bodies are base64url encoded.
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
// Gmail API uses base64url encoding, often without padding. base64.RawURLEncoding handles this correctly.
func decodeGmailBody(data string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
