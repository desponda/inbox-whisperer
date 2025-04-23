package models

// EmailSummary is a provider-agnostic summary DTO for list endpoints
// (subject, sender, snippet, date, etc.)
type EmailSummary struct {
	ID           string
	ThreadID     string
	Subject      string
	Sender       string
	Snippet      string
	InternalDate int64
	Date         string
	Provider     string
}
