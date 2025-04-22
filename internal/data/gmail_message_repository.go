package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
)

type GmailMessage struct {
	ID              int64     // Local DB primary key
	UserID          string
	GmailMessageID  string
	ThreadID        string
	Subject         string
	Sender          string
	Recipient       string
	Snippet         string
	Body            string
	InternalDate    int64
	HistoryID       int64
	CachedAt        time.Time
	LastFetchedAt   sql.NullTime
	Category        sql.NullString
	CategorizationConfidence sql.NullFloat64
	RawJSON         json.RawMessage
}

type GmailMessageRepository interface {
	UpsertMessage(ctx context.Context, msg *GmailMessage) error
	GetMessageByID(ctx context.Context, userID, gmailMessageID string) (*GmailMessage, error)
	GetMessagesForUser(ctx context.Context, userID string, limit, offset int) ([]*GmailMessage, error) // legacy, wraps cursor version
	GetMessagesForUserCursor(ctx context.Context, userID string, limit int, afterInternalDate int64, afterMsgID string) ([]*GmailMessage, error)
	DeleteMessagesForUser(ctx context.Context, userID string) error
}


type gmailMessageRepository struct {
	pool *pgxpool.Pool
} 

// NewGmailMessageRepositoryFromPool creates a repository using a pgxpool.Pool (only implementation)
func NewGmailMessageRepositoryFromPool(pool *pgxpool.Pool) GmailMessageRepository {
	return &gmailMessageRepository{pool: pool}
}

func (r *gmailMessageRepository) UpsertMessage(ctx context.Context, msg *GmailMessage) error {
	// DEBUG
	if v, ok := ctx.Value("_test_debug").(func(string, ...interface{})); ok && v != nil {
		v("[UpsertMessage] userID=%s, gmailMessageID=%s, subject=%s", msg.UserID, msg.GmailMessageID, msg.Subject)
	}

	query := `INSERT INTO gmail_messages
		(user_id, gmail_message_id, thread_id, subject, sender, recipient, snippet, body, internal_date, history_id, cached_at, last_fetched_at, category, categorization_confidence, raw_json)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (user_id, gmail_message_id) DO UPDATE SET
		thread_id=EXCLUDED.thread_id,
		subject=EXCLUDED.subject,
		sender=EXCLUDED.sender,
		recipient=EXCLUDED.recipient,
		snippet=EXCLUDED.snippet,
		body=EXCLUDED.body,
		internal_date=EXCLUDED.internal_date,
		history_id=EXCLUDED.history_id,
		cached_at=EXCLUDED.cached_at,
		last_fetched_at=EXCLUDED.last_fetched_at,
		category=EXCLUDED.category,
		categorization_confidence=EXCLUDED.categorization_confidence,
		raw_json=EXCLUDED.raw_json`
	_, err := r.pool.Exec(ctx, query,
		msg.UserID,
		msg.GmailMessageID,
		msg.ThreadID,
		msg.Subject,
		msg.Sender,
		msg.Recipient,
		msg.Snippet,
		msg.Body,
		msg.InternalDate,
		msg.HistoryID,
		msg.CachedAt,
		msg.LastFetchedAt,
		msg.Category,
		msg.CategorizationConfidence,
		msg.RawJSON,
	)
	if v, ok := ctx.Value("_test_debug").(func(string, ...interface{})); ok && v != nil && err != nil {
		v("[UpsertMessage] ERROR: %v", err)
	}
	return err
}

func (r *gmailMessageRepository) GetMessageByID(ctx context.Context, userID, gmailMessageID string) (*GmailMessage, error) {
	// DEBUG
	if v, ok := ctx.Value("_test_debug").(func(string, ...interface{})); ok && v != nil {
		v("[GetMessageByID] userID=%s, gmailMessageID=%s", userID, gmailMessageID)
	}

	query := `SELECT id, user_id, gmail_message_id, thread_id, subject, sender, recipient, snippet, body, internal_date, history_id, cached_at, last_fetched_at, category, categorization_confidence, raw_json FROM gmail_messages WHERE user_id=$1 AND gmail_message_id=$2`
	row := r.pool.QueryRow(ctx, query, userID, gmailMessageID)
	var msg GmailMessage
	err := row.Scan(&msg.ID, &msg.UserID, &msg.GmailMessageID, &msg.ThreadID, &msg.Subject, &msg.Sender, &msg.Recipient, &msg.Snippet, &msg.Body, &msg.InternalDate, &msg.HistoryID, &msg.CachedAt, &msg.LastFetchedAt, &msg.Category, &msg.CategorizationConfidence, &msg.RawJSON)
	if err != nil {
		if v, ok := ctx.Value("_test_debug").(func(string, ...interface{})); ok && v != nil {
			v("[GetMessageByID] ERROR: %v", err)
		}
		return nil, err
	}
	return &msg, nil
}

// GetMessagesForUser provides offset-based pagination for legacy compatibility (not recommended for large inboxes)
func (r *gmailMessageRepository) GetMessagesForUser(ctx context.Context, userID string, limit, offset int) ([]*GmailMessage, error) {
	query := `SELECT id, user_id, gmail_message_id, thread_id, subject, sender, recipient, snippet, body, internal_date, history_id, cached_at, last_fetched_at, category, categorization_confidence, raw_json FROM gmail_messages WHERE user_id=$1 ORDER BY internal_date DESC, gmail_message_id DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []*GmailMessage
	for rows.Next() {
		var msg GmailMessage
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.GmailMessageID, &msg.ThreadID, &msg.Subject, &msg.Sender, &msg.Recipient, &msg.Snippet, &msg.Body, &msg.InternalDate, &msg.HistoryID, &msg.CachedAt, &msg.LastFetchedAt, &msg.Category, &msg.CategorizationConfidence, &msg.RawJSON)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, &msg)
	}
	return msgs, nil
}

// GetMessagesForUserCursor returns a page of messages for a user, starting after the given cursor (internal_date, gmail_message_id)
func (r *gmailMessageRepository) GetMessagesForUserCursor(ctx context.Context, userID string, limit int, afterInternalDate int64, afterMsgID string) ([]*GmailMessage, error) {
	var (
		query string
		rows pgx.Rows
		err error
	)
	if afterInternalDate > 0 && afterMsgID != "" {
		// Use tuple comparison for stable pagination
		query = `SELECT id, user_id, gmail_message_id, thread_id, subject, sender, recipient, snippet, body, internal_date, history_id, cached_at, last_fetched_at, category, categorization_confidence, raw_json FROM gmail_messages WHERE user_id=$1 AND (internal_date, gmail_message_id) < ($2, $3) ORDER BY internal_date DESC, gmail_message_id DESC LIMIT $4`
		rows, err = r.pool.Query(ctx, query, userID, afterInternalDate, afterMsgID, limit)
	} else {
		query = `SELECT id, user_id, gmail_message_id, thread_id, subject, sender, recipient, snippet, body, internal_date, history_id, cached_at, last_fetched_at, category, categorization_confidence, raw_json FROM gmail_messages WHERE user_id=$1 ORDER BY internal_date DESC, gmail_message_id DESC LIMIT $2`
		rows, err = r.pool.Query(ctx, query, userID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []*GmailMessage
	for rows.Next() {
		var msg GmailMessage
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.GmailMessageID, &msg.ThreadID, &msg.Subject, &msg.Sender, &msg.Recipient, &msg.Snippet, &msg.Body, &msg.InternalDate, &msg.HistoryID, &msg.CachedAt, &msg.LastFetchedAt, &msg.Category, &msg.CategorizationConfidence, &msg.RawJSON)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, &msg)
	}
	return msgs, nil
}

func (r *gmailMessageRepository) DeleteMessagesForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM gmail_messages WHERE user_id=$1`, userID)
	return err
}
