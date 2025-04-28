package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/rs/zerolog/log"
)

const (
	// CSRFTokenLength is the length of the CSRF token in bytes
	CSRFTokenLength = 32

	// CSRFTokenKey is the key used to store the CSRF token in the session
	CSRFTokenKey = "csrf_token"

	// CSRFHeaderName is the name of the header that should contain the CSRF token
	CSRFHeaderName = "X-CSRF-Token"

	// CSRFFormField is the name of the form field that can contain the CSRF token
	CSRFFormField = "csrf_token"
)

// CSRF middleware provides CSRF protection
type CSRF struct {
	manager session.Manager
}

// NewCSRF creates a new CSRF middleware
func NewCSRF(manager session.Manager) *CSRF {
	return &CSRF{manager: manager}
}

// Handler wraps an http.Handler with CSRF protection
func (m *CSRF) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for safe methods
		if isSafeMethod(r.Method) {
			// For GET requests, ensure a CSRF token exists
			session, err := m.manager.Start(w, r)
			if err != nil {
				log.Error().
					Err(err).
					Str("path", r.URL.Path).
					Msg("Failed to start session")
				http.Error(w, "Session error", http.StatusInternalServerError)
				return
			}

			// Generate a new token if one doesn't exist
			if _, ok := session.GetValue(CSRFTokenKey); !ok {
				token, err := generateCSRFToken()
				if err != nil {
					log.Error().
						Err(err).
						Str("session_id", session.ID()).
						Msg("Failed to generate CSRF token")
					http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
					return
				}
				log.Debug().
					Str("session_id", session.ID()).
					Msg("Generated new CSRF token")
				session.SetValue(CSRFTokenKey, token)
				if err := m.manager.Store().Save(r.Context(), session); err != nil {
					log.Error().
						Err(err).
						Str("session_id", session.ID()).
						Msg("Failed to save session with CSRF token")
					http.Error(w, "Failed to save session", http.StatusInternalServerError)
					return
				}
				log.Debug().
					Str("session_id", session.ID()).
					Msg("Saved CSRF token to session")
			} else {
				log.Debug().
					Str("session_id", session.ID()).
					Msg("Reusing existing CSRF token")
			}

			next.ServeHTTP(w, r)
			return
		}

		// For unsafe methods, validate the CSRF token
		session, err := m.manager.Start(w, r)
		if err != nil {
			log.Error().
				Err(err).
				Str("path", r.URL.Path).
				Msg("Failed to start session for CSRF validation")
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		// Get the expected token from the session
		expectedToken, ok := session.GetValue(CSRFTokenKey)
		if !ok {
			log.Warn().
				Str("session_id", session.ID()).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("No CSRF token found in session")
			http.Error(w, "CSRF token not found", http.StatusBadRequest)
			return
		}
		log.Debug().
			Str("session_id", session.ID()).
			Str("method", r.Method).
			Msg("Found CSRF token in session")

		// Get the actual token from the request
		actualToken := getCSRFToken(r)
		if actualToken == "" {
			log.Warn().
				Str("session_id", session.ID()).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("No CSRF token provided in request")
			http.Error(w, "CSRF token missing", http.StatusBadRequest)
			return
		}
		log.Debug().
			Str("session_id", session.ID()).
			Str("method", r.Method).
			Msg("Found CSRF token in request")

		// Compare tokens
		if !compareTokens(expectedToken.(string), actualToken) {
			log.Warn().
				Str("session_id", session.ID()).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("CSRF token mismatch")
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}
		log.Debug().
			Str("session_id", session.ID()).
			Str("method", r.Method).
			Msg("CSRF token validated successfully")

		// Token is valid, proceed
		next.ServeHTTP(w, r)
	})
}

// generateCSRFToken generates a new random CSRF token
func generateCSRFToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// getCSRFToken gets the CSRF token from the request
func getCSRFToken(r *http.Request) string {
	// Try the header first
	token := r.Header.Get(CSRFHeaderName)
	if token != "" {
		return token
	}

	// Then try form values
	return r.FormValue(CSRFFormField)
}

// compareTokens compares two CSRF tokens in a timing-safe manner
func compareTokens(a, b string) bool {
	return len(a) == len(b) && strings.Compare(a, b) == 0
}

// isSafeMethod returns true if the HTTP method is safe (GET, HEAD, OPTIONS)
func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}
