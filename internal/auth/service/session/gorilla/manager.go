package gorilla

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

// Manager implements session management using gorilla/sessions
type Manager struct {
	store       *Store
	sessionName string
}

// NewManager creates a new gorilla session manager
func NewManager(store *Store, sessionName string) *Manager {
	return &Manager{
		store:       store,
		sessionName: sessionName,
	}
}

// Start creates a new session or retrieves an existing one
func (m *Manager) Start(w http.ResponseWriter, r *http.Request) (session.Session, error) {
	// Try to get existing session from cookie
	gorillaSession, err := m.store.Gorilla.Get(r, m.sessionName)
	if err != nil {
		slog.Error("failed to get session from cookie", "error", err, "session_name", m.sessionName)
	}

	// If we have a valid session, wrap and return it
	if gorillaSession != nil && !gorillaSession.IsNew {
		userID, ok := gorillaSession.Values["user_id"].(string)
		if ok && userID != "" {
			slog.Debug("found existing session", "session_id", gorillaSession.ID, "user_id", userID)
			// Get full session from store
			session, err := m.store.Get(r.Context(), gorillaSession.ID)
			if err == nil {
				return session, nil
			}
			slog.Error("failed to get session from store", "error", err, "session_id", gorillaSession.ID)
		}
	}

	// Create new session
	sessionID := uuid.New().String()
	now := time.Now()
	slog.Info("creating new session", "session_id", sessionID)

	session := &Session{
		id:        sessionID,
		createdAt: now,
		expiresAt: now.Add(24 * time.Hour), // TODO: Make configurable
		values:    make(map[interface{}]interface{}),
		session:   sessions.NewSession(m.store.Gorilla, m.sessionName),
	}
	session.session.ID = sessionID
	session.session.IsNew = true

	// Save session to store
	if err := m.store.Save(r.Context(), session); err != nil {
		slog.Error("failed to save new session", "error", err, "session_id", sessionID)
		return nil, fmt.Errorf("failed to save new session: %w", err)
	}

	// Save session cookie
	session.session.Values = session.values
	if err := session.session.Save(r, w); err != nil {
		slog.Error("failed to save session cookie", "error", err, "session_id", sessionID)
		return nil, fmt.Errorf("failed to save session cookie: %w", err)
	}

	slog.Info("new session created successfully", "session_id", sessionID)
	return session, nil
}

// Destroy removes a session
func (m *Manager) Destroy(w http.ResponseWriter, r *http.Request) error {
	gorillaSession, err := m.store.Gorilla.Get(r, m.sessionName)
	if err != nil {
		slog.Error("failed to get session for destruction", "error", err, "session_name", m.sessionName)
		return fmt.Errorf("failed to get session: %w", err)
	}

	if !gorillaSession.IsNew {
		slog.Info("destroying session", "session_id", gorillaSession.ID)

		// Delete from store
		if err := m.store.Delete(r.Context(), gorillaSession.ID); err != nil {
			slog.Error("failed to delete session from store", "error", err, "session_id", gorillaSession.ID)
			return fmt.Errorf("failed to delete session: %w", err)
		}

		// Delete cookie
		gorillaSession.Options.MaxAge = -1
		if err := gorillaSession.Save(r, w); err != nil {
			slog.Error("failed to delete session cookie", "error", err, "session_id", gorillaSession.ID)
			return fmt.Errorf("failed to delete session cookie: %w", err)
		}

		slog.Info("session destroyed successfully", "session_id", gorillaSession.ID)
	}

	return nil
}

// Refresh extends the session lifetime
func (m *Manager) Refresh(w http.ResponseWriter, r *http.Request) error {
	gorillaSession, err := m.store.Gorilla.Get(r, m.sessionName)
	if err != nil {
		slog.Error("failed to get session for refresh", "error", err, "session_name", m.sessionName)
		return fmt.Errorf("failed to get session: %w", err)
	}

	if !gorillaSession.IsNew {
		slog.Debug("refreshing session", "session_id", gorillaSession.ID)

		// Get full session
		session, err := m.store.Get(r.Context(), gorillaSession.ID)
		if err != nil {
			slog.Error("failed to get session from store for refresh", "error", err, "session_id", gorillaSession.ID)
			return fmt.Errorf("failed to get session from store: %w", err)
		}

		// Update expiry
		s := session.(*Session)
		s.expiresAt = time.Now().Add(24 * time.Hour) // TODO: Make configurable

		// Save to store
		if err := m.store.Save(r.Context(), s); err != nil {
			slog.Error("failed to save refreshed session to store", "error", err, "session_id", gorillaSession.ID)
			return fmt.Errorf("failed to save refreshed session: %w", err)
		}

		// Update cookie
		if err := gorillaSession.Save(r, w); err != nil {
			slog.Error("failed to save refreshed session cookie", "error", err, "session_id", gorillaSession.ID)
			return fmt.Errorf("failed to save refreshed session cookie: %w", err)
		}

		slog.Debug("session refreshed successfully", "session_id", gorillaSession.ID)
	}

	return nil
}

// Store returns the underlying session store
func (m *Manager) Store() session.Store {
	return m.store
}
