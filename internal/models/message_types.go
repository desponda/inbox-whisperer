package models

// MessageSummary is a minimal summary of a Gmail message
type MessageSummary struct {
	ID                      string `json:"id"`
	ThreadID                string `json:"thread_id"`
	Snippet                 string `json:"snippet"`
	From                    string `json:"from"`
	Subject                 string `json:"subject"`
	InternalDate            int64  `json:"internal_date"`
	CursorAfterID           string `json:"cursor_after_id,omitempty"`
	CursorAfterInternalDate int64  `json:"cursor_after_internal_date,omitempty"`
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
