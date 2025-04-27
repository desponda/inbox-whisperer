package session

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"encoding/json"
)

// ClearSession expires the session_id cookie and removes the session from the store
func ClearSession(w http.ResponseWriter, r *http.Request) {
	// Get and clear session from store
	cookie, err := r.Cookie("session_id")
	sessionID := ""
	if err == nil && cookie.Value != "" {
		sessionID = cookie.Value
	}
	if sessionID != "" {
		store.Lock()
		delete(store.data, sessionID)
		store.Unlock()
	}

	// Clear cookie for both / and /api paths to ensure it's fully removed
	for _, path := range []string{"/", "/api"} {
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     path,
			HttpOnly: true,
			Secure:   true, // Always use Secure for session cookies
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
		})
	}
}

type ctxKey int

const (
	userIDKey ctxKey = iota
	tokenKey
	sessionIDKey
)

// InMemoryStore is a demo in-memory session/token store
var store = struct {
	sync.RWMutex
	data map[string]*SessionData
}{data: make(map[string]*SessionData)}

type SessionData struct {
	UserID string
	Token  string            // Store access token for demo; in prod, store full oauth2.Token
	Values map[string]string // Arbitrary key-value pairs (e.g., oauth_state)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		var sessionID string
		if err != nil || cookie.Value == "" {
			// Create new session and set cookie
			sessionID = uuid.NewString()
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				Secure:   false,
			})
			store.Lock()
			if _, exists := store.data[sessionID]; !exists {
				store.data[sessionID] = &SessionData{Values: map[string]string{}}
			}
			store.Unlock()
		} else {
			sessionID = cookie.Value
		}

		// Retrieve session data if present
		store.RLock()
		data, ok := store.data[sessionID]
		store.RUnlock()

		ctx := r.Context()
		ctx = context.WithValue(ctx, sessionIDKey, sessionID)
		if ok {
			ctx = context.WithValue(ctx, userIDKey, data.UserID)
			ctx = context.WithValue(ctx, tokenKey, data.Token)
		}
		// Pass updated context to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ContextWithUserID returns a new context with the user ID set. Used for testing and service logic.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func SetSession(w http.ResponseWriter, r *http.Request, userID, token string) {

	// Log all cookies received
	for _, c := range r.Cookies() {
		log.Debug().Str("cookie_name", c.Name).Str("cookie_value", c.Value).Msg("cookie received in SetSession")
	}
	cookie, err := r.Cookie("session_id")
	var sessionID string
	if err != nil || cookie.Value == "" {
		sessionID = uuid.NewString()
		cookieObj := &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
		}
		http.SetCookie(w, cookieObj)
		log.Debug().Str("session_id", sessionID).
			Str("Path", cookieObj.Path).
			Bool("HttpOnly", cookieObj.HttpOnly).
			Bool("Secure", cookieObj.Secure).
			Msg("session cookie created in SetSession with attributes")
	} else {
		sessionID = cookie.Value
		log.Debug().Str("session_id", sessionID).Msg("session cookie received in SetSession")
	}
	store.Lock()
	store.data[sessionID] = &SessionData{UserID: userID, Token: token, Values: map[string]string{}}
	storeDump, _ := json.Marshal(store.data)
	log.Debug().Str("session_id", sessionID).Str("user_id", userID).RawJSON("session_store", storeDump).Msg("session set in SetSession")
	store.Unlock()
}

func GetToken(ctx context.Context) string {
	tok, _ := ctx.Value(tokenKey).(string)
	return tok
}

// SetSessionValue sets a custom key-value pair in the session (e.g., CSRF state)
func SetSessionValue(w http.ResponseWriter, r *http.Request, key, value string) {

	cookie, err := r.Cookie("session_id")
	sessionID := ""
	if err == nil && cookie.Value != "" {
		sessionID = cookie.Value
	} else {
		// Try context if cookie not found
		if sid, ok := r.Context().Value(sessionIDKey).(string); ok && sid != "" {
			sessionID = sid

		} else {

			return
		}
	}

	store.Lock()
	data, ok := store.data[sessionID]
	if !ok {
		store.Unlock()
		return
	}
	if data.Values == nil {
		data.Values = make(map[string]string)
	}
	data.Values[key] = value
	store.Unlock()
}

// GetSessionValue retrieves a custom key from the session
func GetSessionValue(r *http.Request, key string) string {

	cookie, err := r.Cookie("session_id")
	sessionID := ""
	if err == nil && cookie.Value != "" {
		sessionID = cookie.Value
	} else {
		// Try context if cookie not found
		if sid, ok := r.Context().Value(sessionIDKey).(string); ok && sid != "" {
			sessionID = sid

		} else {

			return ""
		}
	}

	store.RLock()
	defer store.RUnlock()
	if data, ok := store.data[sessionID]; ok && data.Values != nil {
		return data.Values[key]
	}
	return ""
}

// GetUserID gets the user ID from context
func GetUserID(ctx context.Context) string {
	uid, _ := ctx.Value(userIDKey).(string)
	return uid
}
