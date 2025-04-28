package middleware

import (
	"net/http"
	"path"
	"strings"

	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/rs/zerolog/log"
)

// SessionMiddleware provides session management for HTTP requests
type SessionMiddleware struct {
	manager session.Manager
	// List of paths that don't require authentication
	publicPaths []string
}

// NewSessionMiddleware creates a new session middleware
func NewSessionMiddleware(manager session.Manager, publicPaths ...string) *SessionMiddleware {
	// Normalize public paths
	normalized := make([]string, len(publicPaths))
	for i, p := range publicPaths {
		normalized[i] = path.Clean("/" + strings.TrimPrefix(p, "/"))
	}
	return &SessionMiddleware{
		manager:     manager,
		publicPaths: normalized,
	}
}

// Handler wraps an http.Handler with session management
func (m *SessionMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Clean the request path for consistent comparison
		cleanPath := path.Clean("/" + strings.TrimPrefix(r.URL.Path, "/"))

		// Start or retrieve session
		session, err := m.manager.Start(w, r)
		if err != nil {
			log.Error().
				Err(err).
				Str("path", cleanPath).
				Str("method", r.Method).
				Msg("Session middleware failed to start session")
			http.Error(w, "Failed to initialize session", http.StatusInternalServerError)
			return
		}

		// Check if path is public
		if m.isPublicPath(cleanPath) {
			log.Debug().
				Str("path", cleanPath).
				Str("method", r.Method).
				Msg("Allowing access to public path")
			next.ServeHTTP(w, r)
			return
		}

		// Check if user is authenticated
		if session.UserID() == "" {
			log.Info().
				Str("session_id", session.ID()).
				Str("path", cleanPath).
				Str("method", r.Method).
				Str("remote_addr", r.RemoteAddr).
				Msg("Access denied - no user ID in session")
			http.Error(w, "Please log in to access this resource", http.StatusUnauthorized)
			return
		}

		// User is authenticated
		log.Debug().
			Str("session_id", session.ID()).
			Str("user_id", session.UserID()).
			Str("path", cleanPath).
			Str("method", r.Method).
			Msg("Authenticated request")

		// Refresh session if needed
		if err := m.manager.Refresh(w, r); err != nil {
			log.Error().
				Err(err).
				Str("session_id", session.ID()).
				Str("user_id", session.UserID()).
				Str("path", cleanPath).
				Str("method", r.Method).
				Msg("Failed to refresh session")
			// Continue anyway, just log the error
		}

		next.ServeHTTP(w, r)
	})
}

// isPublicPath checks if the given path matches any of the public paths
func (m *SessionMiddleware) isPublicPath(cleanPath string) bool {
	for _, publicPath := range m.publicPaths {
		// Check for exact match
		if cleanPath == publicPath {
			return true
		}
		// Check if public path is a prefix (for API endpoints)
		if strings.HasSuffix(publicPath, "/*") {
			prefix := strings.TrimSuffix(publicPath, "/*")
			if strings.HasPrefix(cleanPath, prefix) {
				return true
			}
		}
	}
	return false
}
