package data

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/desponda/inbox-whisperer/internal/models"
)

type EmailMessageRepository interface {
	UpsertMessage(ctx context.Context, msg *models.EmailMessage) error
	GetMessageByID(ctx context.Context, userID, emailMessageID string) (*models.EmailMessage, error)
	GetMessagesForUser(ctx context.Context, userID string, limit, offset int) ([]*models.EmailMessage, error)
	GetMessagesForUserCursor(ctx context.Context, userID string, limit int, afterInternalDate int64, afterMsgID string) ([]*models.EmailMessage, error)
	DeleteMessagesForUser(ctx context.Context, userID string) error
}

type emailMessageRepository struct {
	pool *pgxpool.Pool
}

// NewEmailMessageRepositoryFromPool creates a repository using a pgxpool.Pool (only implementation)
func NewEmailMessageRepositoryFromPool(pool *pgxpool.Pool) EmailMessageRepository {
	return &emailMessageRepository{pool: pool}
}

func (r *emailMessageRepository) UpsertMessage(ctx context.Context, msg *models.EmailMessage) error {
	query := `INSERT INTO email_messages
		(user_id, email_message_id, thread_id, subject, sender, recipient, snippet, body, internal_date, history_id, cached_at, last_fetched_at, category, categorization_confidence, raw_json)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		ON CONFLICT (user_id, email_message_id) DO UPDATE SET
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
		msg.EmailMessageID,
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
	return err
}

func (r *emailMessageRepository) GetMessageByID(ctx context.Context, userID, emailMessageID string) (*models.EmailMessage, error) {
	query := `SELECT id, user_id, email_message_id, thread_id, subject, sender, recipient, snippet, body, internal_date, history_id, cached_at, last_fetched_at, category, categorization_confidence, raw_json FROM email_messages WHERE user_id=$1 AND email_message_id=$2`
	row := r.pool.QueryRow(ctx, query, userID, emailMessageID)
	var msg models.EmailMessage
	err := row.Scan(&msg.ID, &msg.UserID, &msg.EmailMessageID, &msg.ThreadID, &msg.Subject, &msg.Sender, &msg.Recipient, &msg.Snippet, &msg.Body, &msg.InternalDate, &msg.HistoryID, &msg.CachedAt, &msg.LastFetchedAt, &msg.Category, &msg.CategorizationConfidence, &msg.RawJSON)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *emailMessageRepository) GetMessagesForUser(ctx context.Context, userID string, limit, offset int) ([]*models.EmailMessage, error) {
	query := `SELECT id, user_id, email_message_id, thread_id, subject, sender, recipient, snippet, body, internal_date, history_id, cached_at, last_fetched_at, category, categorization_confidence, raw_json FROM email_messages WHERE user_id=$1 ORDER BY internal_date DESC, email_message_id DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []*models.EmailMessage
	for rows.Next() {
		var msg models.EmailMessage
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.EmailMessageID, &msg.ThreadID, &msg.Subject, &msg.Sender, &msg.Recipient, &msg.Snippet, &msg.Body, &msg.InternalDate, &msg.HistoryID, &msg.CachedAt, &msg.LastFetchedAt, &msg.Category, &msg.CategorizationConfidence, &msg.RawJSON)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, &msg)
	}
	return msgs, nil
}

func (r *emailMessageRepository) GetMessagesForUserCursor(ctx context.Context, userID string, limit int, afterInternalDate int64, afterMsgID string) ([]*models.EmailMessage, error) {
	var (
		query string
		rows pgx.Rows
		err error
	)
	if afterInternalDate > 0 && afterMsgID != "" {
		query = `SELECT id, user_id, email_message_id, thread_id, subject, sender, recipient, snippet, body, internal_date, history_id, cached_at, last_fetched_at, category, categorization_confidence, raw_json FROM email_messages WHERE user_id=$1 AND (internal_date, email_message_id) < ($2, $3) ORDER BY internal_date DESC, email_message_id DESC LIMIT $4`
		rows, err = r.pool.Query(ctx, query, userID, afterInternalDate, afterMsgID, limit)
	} else {
		query = `SELECT id, user_id, email_message_id, thread_id, subject, sender, recipient, snippet, body, internal_date, history_id, cached_at, last_fetched_at, category, categorization_confidence, raw_json FROM email_messages WHERE user_id=$1 ORDER BY internal_date DESC, email_message_id DESC LIMIT $2`
		rows, err = r.pool.Query(ctx, query, userID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var msgs []*models.EmailMessage
	for rows.Next() {
		var msg models.EmailMessage
		err := rows.Scan(&msg.ID, &msg.UserID, &msg.EmailMessageID, &msg.ThreadID, &msg.Subject, &msg.Sender, &msg.Recipient, &msg.Snippet, &msg.Body, &msg.InternalDate, &msg.HistoryID, &msg.CachedAt, &msg.LastFetchedAt, &msg.Category, &msg.CategorizationConfidence, &msg.RawJSON)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, &msg)
	}
	return msgs, nil
}

func (r *emailMessageRepository) DeleteMessagesForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM email_messages WHERE user_id=$1`, userID)
	return err
}
