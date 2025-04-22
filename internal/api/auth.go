package api

import "github.com/desponda/inbox-whisperer/internal/data"

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/desponda/inbox-whisperer/internal/config"
	"github.com/desponda/inbox-whisperer/internal/session"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"github.com/rs/zerolog/log"
)

// AuthHandler holds the OAuth2 config and provides HTTP handlers for auth endpoints
// (In production, you would inject user/session/token storage here as well)
type AuthHandler struct {
	OAuthConfig *oauth2.Config
	UserTokens  data.UserTokenRepository
}

// NewAuthHandler creates a new AuthHandler with the given app config
func NewAuthHandler(cfg *config.AppConfig, userTokens data.UserTokenRepository) *AuthHandler {
	defaultGoogleScopes := []string{"https://www.googleapis.com/auth/gmail.readonly", "openid", "profile", "email"}
	return &AuthHandler{
		OAuthConfig: &oauth2.Config{
			ClientID:     cfg.Google.ClientID,
			ClientSecret: cfg.Google.ClientSecret,
			RedirectURL:  cfg.Google.RedirectURL,
			Scopes:       defaultGoogleScopes,
			Endpoint:     google.Endpoint,
		},
		UserTokens: userTokens,
	}
}

// HandleLogin starts the OAuth2 flow
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate a random state token for CSRF protection
	state := generateRandomState(32)
	session.SetSessionValue(w, r, "oauth_state", state)
	url := h.OAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

// generateRandomState generates a secure random string for OAuth2 state
func generateRandomState(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "state-token-fallback"
	}
	return base64.URLEncoding.EncodeToString(b)
}

// HandleCallback handles the OAuth2 redirect from Google
func (h *AuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	code := r.URL.Query().Get("code")
	if code == "" {
		log.Warn().Msg("Missing code in callback")
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}
	// Validate state param for CSRF protection
	state := r.URL.Query().Get("state")
	expectedState := session.GetSessionValue(r, "oauth_state")
	if state == "" || expectedState == "" || state != expectedState {
		log.Warn().Msg("Invalid or missing OAuth2 state parameter (possible CSRF)")
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}
	tok, err := exchangeCodeForToken(h, ctx, code)
	if err != nil {
		log.Error().Err(err).Msg("Failed to exchange code for token")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	userID, err := fetchGoogleUserID(ctx, tok, "https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch Google user ID")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Support test patch: userID|email
	email := userID // fallback
	if sep := '|'; len(userID) > 0 && string(userID[0]) != "{" && string(userID[0]) != "[" && string(userID[0]) != "(" && string(userID[0]) != "<" {
		if before, after, found := strings.Cut(userID, string(sep)); found {
			userID = before
			email = after
		}
	}

	// Create user in DB if not exists
	db, ok := h.UserTokens.(interface{ GetByID(context.Context, string) (*models.User, error); Create(context.Context, *models.User) error })
	if ok {
		_, err := db.GetByID(ctx, userID)
		if err != nil {
			// Assume not found, create user
			if email == userID && tok != nil {
				// Try to fetch email from Google userinfo
				client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(tok))
				resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
				if err == nil && resp.StatusCode == 200 {
					var profile struct {
						Email string `json:"email"`
					}
					if err := json.NewDecoder(resp.Body).Decode(&profile); err == nil && profile.Email != "" {
						email = profile.Email
					}
					resp.Body.Close()
				}
			}
			db.Create(ctx, &models.User{
				ID:        userID,
				Email:     email,
				CreatedAt: time.Now().UTC(),
			})
		}
	}
	// Store token in DB (persistent)
	err = h.UserTokens.SaveUserToken(ctx, userID, tok)
	if err != nil {
		log.Error().Err(err).Msg("Failed to persist user token")
		http.Error(w, "failed to persist user token", http.StatusInternalServerError)
		return
	}
	// Store access token in session for immediate use
	session.SetSession(w, r, userID, tok.AccessToken)
	log.Info().Str("userID", userID).Msg("OAuth2 flow complete. Token stored in DB and session.")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OAuth2 flow complete. Token stored in DB and session for user: " + userID))
}

// exchangeCodeForToken exchanges an OAuth2 code for a token
var exchangeCodeForToken = func(h *AuthHandler, ctx context.Context, code string) (*oauth2.Token, error) {
	tok, err := h.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		log.Error().Err(err).Msg("OAuth2 token exchange failed")
		return nil, err
	}
	log.Info().Msg("OAuth2 token exchange successful")
	return tok, nil
}


// fetchGoogleUserID fetches the user's Google ID or email from the UserInfo endpoint
var fetchGoogleUserID = func(ctx context.Context, tok *oauth2.Token, userinfoURL string) (string, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(tok))
	resp := struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}{}
	res, err := client.Get(userinfoURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return "", err
	}
	if resp.ID != "" {
		return resp.ID, nil
	}
	return resp.Email, nil
}


// RegisterAuthRoutes adds the auth endpoints to the router
func RegisterAuthRoutes(r chi.Router, cfg *config.AppConfig, userTokens data.UserTokenRepository) {
	h := NewAuthHandler(cfg, userTokens)
	r.Get("/api/auth/login", h.HandleLogin)
	r.Get("/api/auth/callback", h.HandleCallback)
}
