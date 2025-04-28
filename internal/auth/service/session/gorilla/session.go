package gorilla

import (
	"time"

	"github.com/gorilla/sessions"
)

// Session implements interfaces.Session using Gorilla Sessions
type Session struct {
	id        string
	userID    string
	createdAt time.Time
	expiresAt time.Time
	values    map[interface{}]interface{}
	session   *sessions.Session
}

// ID returns the session ID
func (s *Session) ID() string {
	return s.id
}

// UserID returns the user ID
func (s *Session) UserID() string {
	return s.userID
}

// SetUserID sets the user ID
func (s *Session) SetUserID(userID string) {
	s.userID = userID
	s.values["user_id"] = userID
}

// CreatedAt returns when the session was created
func (s *Session) CreatedAt() time.Time {
	return s.createdAt
}

// ExpiresAt returns when the session expires
func (s *Session) ExpiresAt() time.Time {
	return s.expiresAt
}

// Values returns all session values as map[string]interface{}
func (s *Session) Values() map[string]interface{} {
	// Convert from Gorilla's map[interface{}]interface{} to our map[string]interface{}
	result := make(map[string]interface{}, len(s.values))
	for k, v := range s.values {
		if key, ok := k.(string); ok {
			result[key] = v
		}
	}
	return result
}

// SetValue sets a session value
func (s *Session) SetValue(key string, value interface{}) {
	s.values[key] = value
}

// GetValue gets a session value
func (s *Session) GetValue(key string) (interface{}, bool) {
	val, ok := s.values[key]
	return val, ok
}

// IsNew returns true if this is a new session
func (s *Session) IsNew() bool {
	return s.session.IsNew
}

// Options returns the session options
func (s *Session) Options() *sessions.Options {
	return s.session.Options
}

// SetOptions sets the session options
func (s *Session) SetOptions(options *sessions.Options) {
	s.session.Options = options
}

// gorillaValues returns the internal Gorilla session values
func (s *Session) gorillaValues() map[interface{}]interface{} {
	return s.values
}
