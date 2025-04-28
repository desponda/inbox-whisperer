package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/desponda/inbox-whisperer/internal/auth/service/oauth"
	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/desponda/inbox-whisperer/internal/config"
	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	goauth2 "google.golang.org/api/oauth2/v2"
)

const (
	minStateLength    = 32
	maxStateLength    = 64
	maxCodeLength     = 1024 // OAuth2 code length limit
	defaultTimeout    = 10 * time.Second
	googleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	googleJWKSURL     = "https://www.googleapis.com/oauth2/v3/certs"
)

var (
	ErrInvalidState   = errors.New("invalid or expired session state")
	ErrInvalidCode    = errors.New("invalid or missing authorization code")
	ErrTokenExchange  = errors.New("failed to exchange authorization code for token")
	ErrUserInfo       = errors.New("failed to retrieve user information")
	ErrInvalidEmail   = errors.New("invalid email format")
	ErrInvalidIDToken = errors.New("invalid ID token")
)

// OAuthService defines the interface for OAuth operations
type OAuthService interface {
	ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, *goauth2.Userinfo, error)
	SaveUserAndToken(ctx context.Context, userInfo *goauth2.Userinfo, token *oauth2.Token) error
	GetDB() interface{}
}

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	oauthConfig    *oauth2.Config
	oauthService   OAuthService
	frontendURL    string
	sessionManager session.Manager
}

// AuthHandlerDeps groups dependencies for NewAuthHandler
type AuthHandlerDeps struct {
	Config           *config.AppConfig
	UserTokens       data.UserTokenRepository
	UserRepo         data.UserRepository
	UserIdentityRepo data.UserIdentityRepository
	SessionManager   session.Manager
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(deps AuthHandlerDeps) *AuthHandler {
	cfg := deps.Config
	config := &oauth2.Config{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  cfg.Google.RedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/gmail.readonly", "openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}

	return &AuthHandler{
		oauthConfig:    config,
		oauthService:   oauth.NewService(config, deps.UserTokens, deps.UserRepo, deps.UserIdentityRepo),
		frontendURL:    cfg.Server.FrontendURL,
		sessionManager: deps.SessionManager,
	}
}

// HandleLogin starts the OAuth2 flow
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateRandomState(minStateLength)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate state token")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	session, err := h.sessionManager.Start(w, r)
	if err != nil || session == nil {
		log.Error().Err(err).Msg("Failed to start session")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	session.SetValue("oauth_state", state)
	session.SetValue("state_created_at", time.Now().UTC().Format(time.RFC3339))
	if err := h.sessionManager.Store().Save(r.Context(), session); err != nil {
		log.Error().Err(err).Msg("Failed to save session")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	url := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleCallback handles the OAuth2 redirect from Google
func (h *AuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Log all query params
	log.Debug().
		Str("path", r.URL.Path).
		Str("state", r.URL.Query().Get("state")).
		Str("code", r.URL.Query().Get("code")).
		Msg("OAuth callback received")

	// 1. Validate state parameter (CSRF protection)
	state := r.URL.Query().Get("state")
	if state == "" {
		log.Warn().Msg("Missing state parameter")
		h.handleAuthError(w, fmt.Errorf("missing state parameter"), http.StatusBadRequest)
		return
	}

	// 2. Get and validate session
	session, err := h.sessionManager.Start(w, r)
	if err != nil || session == nil {
		log.Error().Err(err).Msg("Failed to get session")
		h.handleAuthError(w, fmt.Errorf("invalid session"), http.StatusInternalServerError)
		return
	}

	// Log session values for debugging
	log.Debug().
		Interface("session_values", session.Values()).
		Msg("Session values at callback")

	// 3. Validate state against session
	if err := h.validateState(w, r, session); err != nil {
		log.Warn().Err(err).Msg("State validation failed")
		h.handleAuthError(w, err, http.StatusBadRequest)
		return
	}

	// 4. Validate authorization code
	code := r.URL.Query().Get("code")
	if err := h.validateAuthCode(code); err != nil {
		log.Warn().Err(err).Msg("Invalid authorization code")
		h.handleAuthError(w, fmt.Errorf("failed to exchange code for token"), http.StatusBadRequest)
		return
	}

	// 5. Exchange code for token and get user info
	tok, userInfo, err := h.oauthService.ExchangeCodeForToken(ctx, code)
	if err != nil {
		log.Error().
			Err(err).
			Str("code", code).
			Msg("Token exchange failed")
		if errors.Is(err, oauth.ErrTokenExchange) {
			h.handleAuthError(w, fmt.Errorf("failed to exchange code for token"), http.StatusBadRequest)
			return
		}
		h.handleAuthError(w, err, http.StatusInternalServerError)
		return
	}

	// Log token and user info for debugging (be careful with sensitive info in production)
	log.Debug().
		Interface("token", tok).
		Interface("userInfo", userInfo).
		Msg("Token and user info received from Google")

	// 6. Save user and token
	if err := h.oauthService.SaveUserAndToken(ctx, userInfo, tok); err != nil {
		log.Error().Err(err).Msg("Failed to save user/token")
		h.handleAuthError(w, err, http.StatusInternalServerError)
		return
	}

	// Look up our UUID for this Google user
	var userID uuid.UUID
	db := h.oauthService.GetDB()
	if db == nil {
		h.handleAuthError(w, errors.New("internal server error: cannot get DB"), http.StatusInternalServerError)
		return
	}

	row := db.(interface {
		QueryRow(context.Context, string, ...interface{}) interface{}
	}).QueryRow(ctx, `SELECT user_id FROM user_identities WHERE provider = $1 AND provider_user_id = $2`, "google", userInfo.Id)
	if row == nil {
		h.handleAuthError(w, errors.New("internal server error: cannot find user identity after save"), http.StatusInternalServerError)
		return
	}

	if err := row.(interface{ Scan(...interface{}) error }).Scan(&userID); err != nil {
		h.handleAuthError(w, errors.New("internal server error: cannot find user identity after save"), http.StatusInternalServerError)
		return
	}

	// 7. Update session
	session.SetUserID(userID.String())
	if err := h.sessionManager.Store().Save(ctx, session); err != nil {
		log.Error().Err(err).Msg("Failed to update session")
		h.handleAuthError(w, err, http.StatusInternalServerError)
		return
	}

	// 8. Redirect to frontend
	http.Redirect(w, r, h.frontendURL, http.StatusFound)
}

// Helper functions

func generateRandomState(length int) (string, error) {
	if length < minStateLength || length > maxStateLength {
		return "", fmt.Errorf("state length must be between %d and %d", minStateLength, maxStateLength)
	}

	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (h *AuthHandler) validateAuthCode(code string) error {
	if code == "" || len(code) > maxCodeLength {
		return ErrInvalidCode
	}
	// Removed character validation: Google OAuth codes can contain / and other chars
	return nil
}

func (h *AuthHandler) validateState(w http.ResponseWriter, r *http.Request, session session.Session) error {
	state := r.URL.Query().Get("state")
	expectedState, ok := session.GetValue("oauth_state")
	if !ok || expectedState.(string) != state {
		if err := h.sessionManager.Destroy(w, r); err != nil {
			log.Error().Err(err).Msg("Failed to destroy invalid session")
		}
		return fmt.Errorf("invalid state parameter")
	}

	createdAtRaw, ok := session.GetValue("state_created_at")
	if !ok {
		return fmt.Errorf("invalid state parameter")
	}

	var createdAt time.Time
	switch v := createdAtRaw.(type) {
	case time.Time:
		createdAt = v
	case string:
		var err error
		createdAt, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return fmt.Errorf("invalid state parameter: %w", err)
		}
	default:
		return fmt.Errorf("invalid state parameter type")
	}

	if time.Since(createdAt) > 5*time.Minute {
		if err := h.sessionManager.Destroy(w, r); err != nil {
			log.Error().Err(err).Msg("Failed to destroy expired session")
		}
		return fmt.Errorf("state parameter has expired")
	}

	return nil
}

func (h *AuthHandler) handleAuthError(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errMsg := "An error occurred during authentication"
	if err != nil {
		errMsg = err.Error()
	}

	response := map[string]string{
		"error": errMsg,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("Failed to encode error response")
	}
}

// RegisterAuthRoutes adds the auth endpoints to the router
func RegisterAuthRoutes(r chi.Router, deps AuthHandlerDeps) {
	h := NewAuthHandler(deps)
	r.Get("/login", h.HandleLogin)
	r.Get("/callback", h.HandleCallback)
}
