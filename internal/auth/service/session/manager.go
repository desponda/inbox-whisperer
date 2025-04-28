package session

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/rs/zerolog/log"
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

	log.Info().
		Dur("cleanup_interval", DefaultCleanupInterval).
		Bool("secure_cookie", secureCookie).
		Msg("session manager created with cleanup worker")
	return m
}

// Close stops the session manager and its cleanup worker
func (m *Manager) Close() error {
	m.cancel()
	if m.cleanupWorker != nil {
		m.cleanupWorker.Stop()
	}
	log.Info().Msg("session manager closed")
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
		log.Error().
			Err(err).
			Str("cookie_name", SessionCookieName).
			Msg("failed to get session cookie")
		return nil, fmt.Errorf("failed to get session cookie: %w", err)
	}

	log.Debug().
		Bool("has_cookie", cookie != nil).
		Str("cookie_value", func() string {
			if cookie != nil {
				return cookie.Value
			}
			return ""
		}()).
		Err(err).
		Msg("cookie check result")

	// If cookie exists, try to get session
	if cookie != nil {
		log.Debug().
			Str("session_id", cookie.Value).
			Msg("found session cookie")
		session, err := m.store.Get(r.Context(), cookie.Value)
		log.Debug().
			Bool("session_found", session != nil).
			Str("session_id", cookie.Value).
			Err(err).
			Msg("session lookup result")

		if err == nil && session != nil {
			log.Debug().
				Str("session_id", session.ID()).
				Str("user_id", session.UserID()).
				Str("expires_at", session.ExpiresAt().String()).
				Msg("retrieved existing session")
			return session, nil
		}
		log.Error().
			Err(err).
			Str("session_id", cookie.Value).
			Msg("failed to get session from store")
	}

	// Create new session
	session, err := m.store.Create(r.Context())
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to create new session")
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	log.Info().
		Str("session_id", session.ID()).
		Str("expires_at", session.ExpiresAt().String()).
		Msg("created new session")

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
	log.Debug().
		Str("cookie_name", cookie.Name).
		Str("cookie_value", cookie.Value).
		Str("expires", cookie.Expires.Format(time.RFC3339)).
		Msg("set session cookie")

	return session, nil
}

// Destroy removes a session
func (m *Manager) Destroy(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			log.Debug().Msg("no session cookie to destroy")
			return nil
		}
		log.Error().
			Err(err).
			Str("cookie_name", SessionCookieName).
			Msg("failed to get session cookie for destruction")
		return fmt.Errorf("failed to get session cookie: %w", err)
	}

	log.Info().
		Str("session_id", cookie.Value).
		Msg("destroying session")

	// Delete session from store
	if err := m.store.Delete(r.Context(), cookie.Value); err != nil {
		log.Error().
			Err(err).
			Str("session_id", cookie.Value).
			Msg("failed to delete session from store")
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

	log.Info().
		Str("session_id", cookie.Value).
		Msg("session destroyed successfully")
	return nil
}

// Refresh extends the session lifetime
func (m *Manager) Refresh(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		log.Error().
			Err(err).
			Str("cookie_name", SessionCookieName).
			Msg("failed to get session cookie for refresh")
		return fmt.Errorf("failed to get session cookie: %w", err)
	}

	log.Debug().
		Str("session_id", cookie.Value).
		Msg("refreshing session")

	// Get session
	session, err := m.store.Get(r.Context(), cookie.Value)
	if err != nil {
		log.Error().
			Err(err).
			Str("session_id", cookie.Value).
			Msg("failed to get session for refresh")
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

	log.Debug().
		Str("session_id", session.ID()).
		Str("user_id", session.UserID()).
		Str("expires_at", session.ExpiresAt().String()).
		Msg("session refreshed successfully")
	return nil
}
