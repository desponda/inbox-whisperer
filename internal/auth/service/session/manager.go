package session

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/desponda/inbox-whisperer/internal/auth/session"
)

const (
	// SessionCookieName is the name of the cookie used to store the session ID
	SessionCookieName = "session_id"

	// DefaultSessionDuration is the default duration for new sessions
	DefaultSessionDuration = 24 * time.Hour
)

// Manager implements session.Manager
type Manager struct {
	store         session.Store
	cleanupWorker *CleanupWorker
	ctx           context.Context
	cancel        context.CancelFunc
	secureCookie  bool
}

// NewManager creates a new session manager (default Secure=true)
func NewManager(store session.Store) *Manager {
	return NewManagerWithSecure(store, true)
}

// NewManagerWithSecure creates a new session manager with configurable Secure flag
func NewManagerWithSecure(store session.Store, secureCookie bool) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		store:        store,
		ctx:          ctx,
		cancel:       cancel,
		secureCookie: secureCookie,
	}

	// Create and start cleanup worker
	m.cleanupWorker = NewCleanupWorker(store, DefaultCleanupInterval)
	m.cleanupWorker.Start(ctx)

	slog.Info("session manager created with cleanup worker", "cleanup_interval", DefaultCleanupInterval, "secure_cookie", secureCookie)
	return m
}

// Close stops the session manager and its cleanup worker
func (m *Manager) Close() error {
	m.cancel()
	if m.cleanupWorker != nil {
		m.cleanupWorker.Stop()
	}
	slog.Info("session manager closed")
	return nil
}

// Store returns the underlying session store
func (m *Manager) Store() session.Store {
	return m.store
}

// Start creates a new session or retrieves an existing one
func (m *Manager) Start(w http.ResponseWriter, r *http.Request) (session.Session, error) {
	// Try to get existing session from cookie
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil && err != http.ErrNoCookie {
		slog.Error("failed to get session cookie", "error", err, "cookie_name", SessionCookieName)
		return nil, fmt.Errorf("failed to get session cookie: %w", err)
	}

	slog.Debug("cookie check result",
		"has_cookie", cookie != nil,
		"cookie_value", func() string {
			if cookie != nil {
				return cookie.Value
			}
			return ""
		}(),
		"error", err)

	// If cookie exists, try to get session
	if cookie != nil {
		slog.Debug("found session cookie", "session_id", cookie.Value)
		session, err := m.store.Get(r.Context(), cookie.Value)
		slog.Debug("session lookup result",
			"session_found", session != nil,
			"session_id", cookie.Value,
			"error", err)

		if err == nil && session != nil {
			slog.Debug("retrieved existing session",
				"session_id", session.ID(),
				"user_id", session.UserID(),
				"expires_at", session.ExpiresAt())
			return session, nil
		}
		slog.Error("failed to get session from store", "error", err, "session_id", cookie.Value)
	}

	// Create new session
	session, err := m.store.Create(r.Context())
	if err != nil {
		slog.Error("failed to create new session", "error", err)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	slog.Info("created new session",
		"session_id", session.ID(),
		"expires_at", session.ExpiresAt())

	// Set cookie
	cookie = &http.Cookie{
		Name:     SessionCookieName,
		Value:    session.ID(),
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secureCookie,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt(),
	}
	http.SetCookie(w, cookie)
	slog.Debug("set session cookie",
		"cookie_name", cookie.Name,
		"cookie_value", cookie.Value,
		"expires", cookie.Expires)

	return session, nil
}

// Destroy removes a session
func (m *Manager) Destroy(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			slog.Debug("no session cookie to destroy")
			return nil
		}
		slog.Error("failed to get session cookie for destruction", "error", err, "cookie_name", SessionCookieName)
		return fmt.Errorf("failed to get session cookie: %w", err)
	}

	slog.Info("destroying session", "session_id", cookie.Value)

	// Delete session from store
	if err := m.store.Delete(r.Context(), cookie.Value); err != nil {
		slog.Error("failed to delete session from store", "error", err, "session_id", cookie.Value)
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Delete cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secureCookie,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	slog.Info("session destroyed successfully", "session_id", cookie.Value)
	return nil
}

// Refresh extends the session lifetime
func (m *Manager) Refresh(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		slog.Error("failed to get session cookie for refresh", "error", err, "cookie_name", SessionCookieName)
		return fmt.Errorf("failed to get session cookie: %w", err)
	}

	slog.Debug("refreshing session", "session_id", cookie.Value)

	// Get session
	session, err := m.store.Get(r.Context(), cookie.Value)
	if err != nil {
		slog.Error("failed to get session for refresh", "error", err, "session_id", cookie.Value)
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Update cookie expiry
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    session.ID(),
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secureCookie,
		SameSite: http.SameSiteStrictMode,
		Expires:  session.ExpiresAt(),
	})

	slog.Debug("session refreshed successfully",
		"session_id", session.ID(),
		"user_id", session.UserID(),
		"expires_at", session.ExpiresAt())
	return nil
}
