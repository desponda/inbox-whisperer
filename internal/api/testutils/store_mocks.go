package testutils

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/desponda/inbox-whisperer/internal/auth/session"
)

// MockStore implements session.Store for testing
type MockStore struct {
	mu            sync.Mutex
	sessions      map[string]session.Session
	cleanupCalled int
	cleanupError  error
	nextID        int
}

func NewMockStore() *MockStore {
	return &MockStore{
		sessions: make(map[string]session.Session),
		nextID:   1,
	}
}

func (m *MockStore) Create(ctx context.Context) (session.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionID := fmt.Sprintf("test-session-%d", m.nextID)
	m.nextID++

	sess := NewMockSession(
		sessionID,
		"",
		"",
		make(map[string]interface{}),
		time.Now(),
		time.Now().Add(24*time.Hour),
	)
	m.sessions[sess.ID()] = sess
	return sess, nil
}

func (m *MockStore) Get(ctx context.Context, id string) (session.Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if sess, ok := m.sessions[id]; ok {
		return sess, nil
	}
	return nil, nil
}

func (m *MockStore) Save(ctx context.Context, sess session.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[sess.ID()] = sess
	return nil
}

func (m *MockStore) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, id)
	return nil
}

func (m *MockStore) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cleanupCalled++
	if m.cleanupError != nil {
		return m.cleanupError
	}

	now := time.Now()
	for id, sess := range m.sessions {
		if sess.ExpiresAt().Before(now) {
			delete(m.sessions, id)
		}
	}
	return nil
}

func (m *MockStore) GetCleanupCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cleanupCalled
}

func (m *MockStore) SetCleanupError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupError = err
}

// Reset clears all sessions and resets the cleanup counter
func (m *MockStore) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions = make(map[string]session.Session)
	m.cleanupCalled = 0
	m.cleanupError = nil
}
