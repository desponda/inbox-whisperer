package models

import (
	"time"
	"database/sql"
	"encoding/json"
)

// EmailMessage represents a cached email message (provider-agnostic, e.g. Gmail, Outlook, etc.)
type EmailMessage struct {
	ID              int64          // Local DB primary key
	UserID          string
	EmailMessageID  string         // Unified message ID (was GmailMessageID)
	ThreadID        string
	Subject         string
	Sender          string
	Recipient       string
	Snippet         string
	Body            string
	InternalDate    int64
	Date            string         // RFC 2822/3339 date string (provider-agnostic)
	HistoryID       int64
	CachedAt        time.Time
	LastFetchedAt   sql.NullTime
	Category        sql.NullString
	CategorizationConfidence sql.NullFloat64
	RawJSON         json.RawMessage
}
