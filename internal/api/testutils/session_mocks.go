package testutils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/desponda/inbox-whisperer/internal/common"
)

// MockSession implements session.Session
type MockSession struct {
	id        string
	userID    string
	role      string
	values    map[string]interface{}
	createdAt time.Time
	expiresAt time.Time
}

func NewMockSession(id, userID, role string, values map[string]interface{}, createdAt, expiresAt time.Time) *MockSession {
	if values == nil {
		values = make(map[string]interface{})
	}
	return &MockSession{
		id:        id,
		userID:    userID,
		role:      role,
		values:    values,
		createdAt: createdAt,
		expiresAt: expiresAt,
	}
}

func (s *MockSession) ID() string                            { return s.id }
func (s *MockSession) UserID() string                        { return s.userID }
func (s *MockSession) SetUserID(id string)                   { s.userID = id }
func (s *MockSession) CreatedAt() time.Time                  { return s.createdAt }
func (s *MockSession) ExpiresAt() time.Time                  { return s.expiresAt }
func (s *MockSession) Values() map[string]interface{}        { return s.values }
func (s *MockSession) SetValue(k string, v interface{})      { s.values[k] = v }
func (s *MockSession) GetValue(k string) (interface{}, bool) { v, ok := s.values[k]; return v, ok }
func (s *MockSession) Role() string                          { return s.role }
func (s *MockSession) SetRole(role string)                   { s.role = role }

// MockSessionManager implements session.Manager
type MockSessionManager struct {
	StartFunc   func(w http.ResponseWriter, r *http.Request) (session.Session, error)
	DestroyFunc func(w http.ResponseWriter, r *http.Request) error
	RefreshFunc func(w http.ResponseWriter, r *http.Request) error
	StoreFunc   func() session.Store
}

func NewMockSessionManager(startFunc func(w http.ResponseWriter, r *http.Request) (session.Session, error)) *MockSessionManager {
	return &MockSessionManager{
		StartFunc: startFunc,
	}
}

func (m *MockSessionManager) Start(w http.ResponseWriter, r *http.Request) (session.Session, error) {
	if m.StoreFunc != nil {
		if cookie, err := r.Cookie("session_id"); err == nil {
			store := m.StoreFunc()
			if store != nil {
				sess, err := store.Get(r.Context(), cookie.Value)
				if err == nil && sess != nil {
					return sess, nil
				}
			}
		}
	}
	if m.StartFunc != nil {
		return m.StartFunc(w, r)
	}
	return nil, nil
}

func (m *MockSessionManager) Destroy(w http.ResponseWriter, r *http.Request) error {
	if m.DestroyFunc != nil {
		return m.DestroyFunc(w, r)
	}
	return nil
}

func (m *MockSessionManager) Refresh(w http.ResponseWriter, r *http.Request) error {
	if m.RefreshFunc != nil {
		return m.RefreshFunc(w, r)
	}
	return nil
}

func (m *MockSessionManager) Store() session.Store {
	if m.StoreFunc != nil {
		return m.StoreFunc()
	}
	return nil
}

// SetupTestRequest creates a new test request and response recorder with an optional session
func SetupTestRequest(method, path string, sm session.Manager, role ...string) (*http.Request, *httptest.ResponseRecorder) {
	r := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	if sm != nil {
		sess, _ := sm.Start(w, r)
		if sess != nil {
			sess.SetUserID("user1")
			if len(role) > 0 {
				sess.(interface{ SetRole(string) }).SetRole(role[0])
				r = r.WithContext(context.WithValue(r.Context(), common.RoleKey{}, role[0]))
			}
			r = r.WithContext(context.WithValue(r.Context(), common.UserIDKey{}, "user1"))
		}
	}
	return r, w
}

// NewTestSessionManager creates a new session manager for testing with a default user session
func NewTestSessionManager(role ...string) session.Manager {
	now := time.Now()
	userRole := ""
	if len(role) > 0 {
		userRole = role[0]
	}
	mockSession := NewMockSession(
		"test-session",
		"user1",
		userRole,
		make(map[string]interface{}),
		now,
		now.Add(24*time.Hour),
	)

	return NewMockSessionManager(func(w http.ResponseWriter, r *http.Request) (session.Session, error) {
		return mockSession, nil
	})
}
