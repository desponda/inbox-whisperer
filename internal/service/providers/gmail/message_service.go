package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/notify"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"encoding/base64"

	"golang.org/x/oauth2"
)

var (
	// EnableBackgroundSync controls whether to sync with Gmail in background
	EnableBackgroundSync = true
)

// MessageService handles fetching and caching Gmail messages
// It provides a fast, cached layer over the Gmail API with background syncing
type MessageService struct {
	repo        data.EmailMessageRepository
	api         GmailAPI
	oauthConfig *oauth2.Config
	syncEnabled bool
}

type MessageServiceOption func(*MessageService)

func WithGmailAPI(api GmailAPI) MessageServiceOption {
	return func(s *MessageService) { s.api = api }
}

// NewMessageService creates a new Gmail message service
func NewMessageService(repo data.EmailMessageRepository, oauthConfig *oauth2.Config, opts ...MessageServiceOption) *MessageService {
	svc := &MessageService{
		repo:        repo,
		oauthConfig: oauthConfig,
		syncEnabled: true,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// EnableSync controls whether background syncing is enabled
func (s *MessageService) EnableSync(enabled bool) {
	s.syncEnabled = enabled
}

// FetchMessages returns cached message summaries and triggers a background sync
// This provides fast inbox loading while ensuring data stays fresh
func (s *MessageService) FetchMessages(ctx context.Context, token *oauth2.Token) ([]models.EmailMessage, error) {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user ID from context")
		return nil, err
	}

	afterID, afterInternalDate := GetPaginationFromContext(ctx)

	// Extract limit from context if present
	pageSize := 10 // default
	if params, ok := ctx.Value(gmailParamsKey{}).(*FetchParams); ok && params.Limit > 0 {
		pageSize = params.Limit
	}

	// 1. Return cached summaries instantly
	cached, err := s.fetchFromCache(ctx, userID, pageSize, afterInternalDate, afterID)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Msg("Failed to fetch messages from cache")
		return nil, err
	}

	msgs := cached.([]*models.EmailMessage)
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

	// 2. Trigger background sync if enabled
	if token != nil && s.syncEnabled {
		go func() {
			if err := s.syncWithGmail(ctx, token, userID); err != nil {
				log.Error().
					Err(err).
					Str("user_id", userID).
					Msg("Failed to sync with Gmail")
			} else {
				log.Debug().
					Str("user_id", userID).
					Msg("Successfully synced with Gmail")
			}
		}()
	}

	return result, nil
}

// FetchMessageContent gets a full message by ID, using cache when possible
func (s *MessageService) FetchMessageContent(ctx context.Context, token *oauth2.Token, id string) (*models.EmailMessage, error) {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user ID from context")
		return nil, err
	}

	// Try cache first
	cached, err := s.fetchFromCache(ctx, userID, id)
	if err == nil && cached != nil {
		if msg, ok := cached.(*models.EmailMessage); ok {
			log.Debug().
				Str("user_id", userID).
				Str("message_id", id).
				Msg("Returning cached message")
			return msg, nil
		}
	}

	// Fetch from Gmail API
	gmailMsg, err := s.fetchFromAPI(ctx, token, id)
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID).
			Str("message_id", id).
			Msg("Failed to fetch from Gmail API")
		return nil, err
	}

	// Convert to our model
	msg := &models.EmailMessage{
		UserID:         userID,
		Provider:       "gmail",
		EmailMessageID: gmailMsg.Id,
		ThreadID:       gmailMsg.ThreadId,
		Subject:        getHeader(gmailMsg.Payload.Headers, "Subject"),
		Sender:         getHeader(gmailMsg.Payload.Headers, "From"),
		Recipient:      getHeader(gmailMsg.Payload.Headers, "To"),
		Snippet:        gmailMsg.Snippet,
		Body:           extractPlainTextBody(gmailMsg.Payload),
		HTMLBody:       extractHTMLBody(gmailMsg.Payload),
		InternalDate:   gmailMsg.InternalDate,
		Date:           getHeader(gmailMsg.Payload.Headers, "Date"),
		HistoryID:      int64(gmailMsg.HistoryId),
		CachedAt:       time.Now(),
		RawJSON:        mustMarshalJSON(gmailMsg),
	}

	// Update cache in background
	go func() {
		if err := s.repo.UpsertMessage(ctx, msg); err != nil {
			log.Error().
				Err(err).
				Str("user_id", userID).
				Str("message_id", id).
				Msg("Failed to cache message")
		}
	}()

	return msg, nil
}

// Helper methods...
func (s *MessageService) fetchFromCache(ctx context.Context, userID string, args ...interface{}) (interface{}, error) {
	// For message lists
	if len(args) == 3 {
		pageSize := args[0].(int)
		afterInternalDate := args[1].(int64)
		afterID := args[2].(string)
		if afterID != "" && afterInternalDate > 0 {
			return s.repo.GetMessagesForUserCursor(ctx, userID, pageSize, afterInternalDate, afterID)
		}
		return s.repo.GetMessagesForUserCursor(ctx, userID, pageSize, 0, "")
	}

	// For single message
	if len(args) == 1 {
		messageID := args[0].(string)
		msg, err := s.repo.GetMessageByID(ctx, userID, "gmail", messageID)
		if err != nil || msg == nil || time.Since(msg.CachedAt) > time.Minute {
			return nil, err
		}
		return msg, nil
	}

	return nil, fmt.Errorf("invalid arguments to fetchFromCache")
}

func (s *MessageService) syncWithGmail(ctx context.Context, token *oauth2.Token, userID string) error {
	var listCall UsersMessagesListCall
	var getCall func(msgID string) UsersMessagesGetCall

	if s.api != nil {
		listCall = s.api.UsersMessagesList("me")
		getCall = func(msgID string) UsersMessagesGetCall {
			return s.api.UsersMessagesGet("me", msgID)
		}
	} else {
		client, err := s.getGmailClient(ctx, token)
		if err != nil {
			return err
		}
		listCall = client.Users.Messages.List("me")
		getCall = func(msgID string) UsersMessagesGetCall {
			return client.Users.Messages.Get("me", msgID)
		}
	}

	resp, err := listCall.Do()
	if err != nil {
		return err
	}
	for _, msg := range resp.Messages {
		if msg == nil {
			continue
		}
		get := getCall(msg.Id)
		msg, err := get.Do()
		if err != nil || msg == nil {
			continue
		} // skip errored messages
		dbMsg := &models.EmailMessage{
			UserID:         userID,
			Provider:       "gmail",
			EmailMessageID: msg.Id,
			ThreadID:       msg.ThreadId,
			Subject:        getHeader(msg.Payload.Headers, "Subject"),
			Sender:         getHeader(msg.Payload.Headers, "From"),
			Snippet:        msg.Snippet,
			InternalDate:   msg.InternalDate,
			Date:           getHeader(msg.Payload.Headers, "Date"),
			CachedAt:       time.Now(),
			RawJSON:        mustMarshalJSON(msg),
		}
		_ = s.repo.UpsertMessage(ctx, dbMsg)
	}
	// Notify client (poll endpoint) after sync completes for instant refresh
	if userID != "" {
		notify.SetGmailSyncStatus(userID)
	}
	return nil
}

func (s *MessageService) fetchFromAPI(ctx context.Context, token *oauth2.Token, id string) (*gmail.Message, error) {
	if s.api != nil {
		return s.api.UsersMessagesGet("me", id).Do()
	}
	client, err := s.getGmailClient(ctx, token)
	if err != nil {
		return nil, err
	}
	return client.Users.Messages.Get("me", id).Do()
}

// Utility functions...
func getHeader(headers []*gmail.MessagePartHeader, name string) string {
	if headers == nil {
		return ""
	}
	nameLower := strings.ToLower(name)
	for _, h := range headers {
		if strings.ToLower(h.Name) == nameLower {
			return h.Value
		}
	}
	return ""
}

func extractPlainTextBody(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}
	if payload.MimeType == "text/plain" && payload.Body != nil && payload.Body.Data != "" {
		data, err := base64.RawURLEncoding.DecodeString(payload.Body.Data)
		if err != nil {
			return ""
		}
		return string(data)
	}
	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
			data, err := base64.RawURLEncoding.DecodeString(part.Body.Data)
			if err != nil {
				return ""
			}
			return string(data)
		}
	}
	return ""
}

func extractHTMLBody(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}
	if payload.MimeType == "text/html" && payload.Body != nil && payload.Body.Data != "" {
		data, err := base64.RawURLEncoding.DecodeString(payload.Body.Data)
		if err != nil {
			return ""
		}
		return string(data)
	}
	for _, part := range payload.Parts {
		if part.MimeType == "text/html" && part.Body != nil && part.Body.Data != "" {
			data, err := base64.RawURLEncoding.DecodeString(part.Body.Data)
			if err != nil {
				return ""
			}
			return string(data)
		}
	}
	return ""
}

func mustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

func (s *MessageService) getGmailClient(ctx context.Context, token *oauth2.Token) (*gmail.Service, error) {
	client := s.oauthConfig.Client(ctx, token)
	return gmail.NewService(ctx, option.WithHTTPClient(client))
}
