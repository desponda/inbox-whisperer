package models

import (
	"sync"
	"time"
)

// Session represents a user session
type Session struct {
	mu sync.RWMutex

	id        string
	userID    string
	createdAt time.Time
	expiresAt time.Time
	values    map[string]interface{}
}

// NewSession creates a new session
func NewSession(id string, duration time.Duration) *Session {
	now := time.Now()
	return &Session{
		id:        id,
		createdAt: now,
		expiresAt: now.Add(duration),
		values:    make(map[string]interface{}),
	}
}

// ID returns the session ID
func (s *Session) ID() string {
	return s.id
}

// UserID returns the user ID
func (s *Session) UserID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.userID
}

// SetUserID sets the user ID
func (s *Session) SetUserID(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userID = userID
}

// CreatedAt returns the creation time
func (s *Session) CreatedAt() time.Time {
	return s.createdAt
}

// ExpiresAt returns the expiration time
func (s *Session) ExpiresAt() time.Time {
	return s.expiresAt
}

// Values returns the session values
func (s *Session) Values() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]interface{})
	for k, v := range s.values {
		result[k] = v
	}
	return result
}

// SetValue sets a session value
func (s *Session) SetValue(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = value
}

// GetValue gets a session value
func (s *Session) GetValue(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.values[key]
	return value, ok
}

// Get retrieves a value from the session
func (s *Session) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.values[key]
	if !ok {
		return "", false
	}
	strVal, ok := val.(string)
	return strVal, ok
}

// Set stores a value in the session
func (s *Session) Set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = value
}

// Delete removes a value from the session
func (s *Session) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, key)
}

// Clear removes all values from the session
func (s *Session) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = make(map[string]interface{})
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.expiresAt)
}

// SetExpiresAt sets the expiration time for the session
func (s *Session) SetExpiresAt(t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.expiresAt = t
}
