package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/desponda/inbox-whisperer/internal/common"
	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

// OAuthToken is the context key for the OAuth token
type OAuthToken struct{}

// AuthError represents authentication-related errors
type AuthError struct {
	Message string
	Code    int
}

func (e AuthError) Error() string {
	return e.Message
}

// OAuthMiddleware provides OAuth token validation and context setup
type OAuthMiddleware struct {
	manager    session.Manager
	tokenStore data.UserTokenRepository
	config     *oauth2.Config // Added OAuth config for token refresh
}

// NewOAuthMiddleware creates a new OAuth middleware
func NewOAuthMiddleware(manager session.Manager, tokenStore data.UserTokenRepository, config *oauth2.Config) *OAuthMiddleware {
	return &OAuthMiddleware{
		manager:    manager,
		tokenStore: tokenStore,
		config:     config,
	}
}

// refreshToken attempts to refresh an expired OAuth token
func (m *OAuthMiddleware) refreshToken(ctx context.Context, token *oauth2.Token, userID string) (*oauth2.Token, error) {
	if token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	newToken, err := m.config.TokenSource(ctx, token).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Save the new token
	if err := m.tokenStore.SaveUserToken(ctx, userID, newToken); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to save refreshed token")
		// Continue anyway as we got a valid token
	}

	return newToken, nil
}

// getToken retrieves and validates the OAuth token
func (m *OAuthMiddleware) getToken(w http.ResponseWriter, r *http.Request, userID string) (*oauth2.Token, error) {
	// First try to get token from session
	session, err := m.manager.Start(w, r)
	if err != nil {
		return nil, AuthError{Message: "Failed to access session", Code: http.StatusInternalServerError}
	}

	values := session.Values()
	token, ok := values["oauth_token"].(*oauth2.Token)

	if !ok || token == nil {
		token, err = m.tokenStore.GetUserToken(r.Context(), userID)
		if err != nil {
			return nil, AuthError{Message: "Failed to retrieve OAuth token", Code: http.StatusUnauthorized}
		}
		if token == nil {
			return nil, AuthError{Message: "No OAuth token found", Code: http.StatusUnauthorized}
		}
	}

	// Check if token is expired and needs refresh
	if token.Expiry.Before(time.Now()) {
		newToken, err := m.refreshToken(r.Context(), token, userID)
		if err != nil {
			return nil, AuthError{Message: "OAuth token expired and refresh failed", Code: http.StatusUnauthorized}
		}
		token = newToken

		// Update session with new token
		session.SetValue("oauth_token", token)
		if err := m.manager.Refresh(w, r); err != nil {
			log.Error().Err(err).Msg("Failed to update session with refreshed token")
			// Continue anyway as we got a valid token
		}
	}

	return token, nil
}

// Handler wraps an http.Handler with OAuth token validation and context setup
func (m *OAuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := m.manager.Start(w, r)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get session")
			http.Error(w, "Session initialization failed", http.StatusInternalServerError)
			return
		}

		userID := session.UserID()
		if userID == "" {
			http.Error(w, "No active session found", http.StatusUnauthorized)
			return
		}

		// Parse userID as uuid.UUID
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			http.Error(w, "Invalid user ID in session", http.StatusUnauthorized)
			return
		}

		token, err := m.getToken(w, r, userID)
		if err != nil {
			if authErr, ok := err.(AuthError); ok {
				http.Error(w, authErr.Message, authErr.Code)
			} else {
				http.Error(w, "Authentication failed", http.StatusUnauthorized)
			}
			return
		}

		// Add userID and token to context as uuid.UUID
		ctx := context.WithValue(r.Context(), common.UserIDKey{}, userUUID)
		ctx = context.WithValue(ctx, OAuthToken{}, token)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
