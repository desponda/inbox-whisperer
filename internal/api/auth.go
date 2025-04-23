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
		
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	expectedState := session.GetSessionValue(r, "oauth_state")
	if state == "" || expectedState == "" || state != expectedState {
		
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}

	tok, err := exchangeCodeForToken(h, ctx, code)
	if err != nil {
		http.Error(w, "token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userID, email, err := getUserIDAndEmail(ctx, tok)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ensureUserExists(ctx, h.UserTokens, userID, email, tok)

	err = h.UserTokens.SaveUserToken(ctx, userID, tok)
	if err != nil {
		http.Error(w, "failed to persist user token", http.StatusInternalServerError)
		return
	}

	setSessionToken(w, r, userID, tok.AccessToken)
	http.Redirect(w, r, "/", http.StatusFound)
}

func getUserIDAndEmail(ctx context.Context, tok *oauth2.Token) (string, string, error) {
	userID, err := fetchGoogleUserID(ctx, tok, "https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", "", err
	}
	email := userID
	if sep := '|'; len(userID) > 0 && !strings.ContainsAny(string(userID[0]), "[{(<") {
		if before, after, found := strings.Cut(userID, string(sep)); found {
			userID = before
			email = after
		}
	}
	return userID, email, nil
}

func ensureUserExists(ctx context.Context, userTokens data.UserTokenRepository, userID, email string, tok *oauth2.Token) {
	db, ok := userTokens.(interface{ GetByID(context.Context, string) (*models.User, error); Create(context.Context, *models.User) error })
	if !ok {
		return
	}
	_, err := db.GetByID(ctx, userID)
	if err == nil {
		return
	}
	if email == userID && tok != nil {
		client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(tok))
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err == nil && resp.StatusCode == 200 {
			var profile struct { Email string `json:"email"` }
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

func setSessionToken(w http.ResponseWriter, r *http.Request, userID, token string) {
	
	session.SetSession(w, r, userID, token)
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OAuth2 flow complete. Token stored in DB and session for user: " + userID))
}

// exchangeCodeForToken exchanges an OAuth2 code for a token
var exchangeCodeForToken = func(h *AuthHandler, ctx context.Context, code string) (*oauth2.Token, error) {
	tok, err := h.OAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
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
