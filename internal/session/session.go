package session

import (
	"context"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

type ctxKey int

const (
	userIDKey ctxKey = iota
	tokenKey
)

// InMemoryStore is a demo in-memory session/token store
var store = struct {
	sync.RWMutex
	data map[string]*SessionData
}{data: make(map[string]*SessionData)}

type SessionData struct {
	UserID string
	Token  string // Store access token for demo; in prod, store full oauth2.Token
	Values map[string]string // Arbitrary key-value pairs (e.g., oauth_state)
}

// Middleware manages session cookie and attaches user/token to context
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		var sessionID string
		if err != nil || cookie.Value == "" {
			sessionID = uuid.NewString()
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				Secure:   false,
			})
		} else {
			sessionID = cookie.Value
		}
		store.RLock()
		data, ok := store.data[sessionID]
		store.RUnlock()
		ctx := r.Context()
		if ok {
			ctx = context.WithValue(ctx, userIDKey, data.UserID)
			ctx = context.WithValue(ctx, tokenKey, data.Token)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ContextWithUserID returns a new context with the user ID set (for test and service use)
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// SetSession stores user and token for a session
func SetSession(w http.ResponseWriter, r *http.Request, userID, token string) {
	cookie, err := r.Cookie("session_id")
	var sessionID string
	if err != nil || cookie.Value == "" {
		sessionID = uuid.NewString()
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
		})
	} else {
		sessionID = cookie.Value
	}
	store.Lock()
	store.data[sessionID] = &SessionData{UserID: userID, Token: token, Values: map[string]string{}}
	store.Unlock()
}

// GetToken retrieves the token from context
func GetToken(ctx context.Context) string {
	tok, _ := ctx.Value(tokenKey).(string)
	return tok
}

// SetSessionValue sets a custom key-value pair in the session (e.g., CSRF state)
func SetSessionValue(w http.ResponseWriter, r *http.Request, key, value string) {
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie.Value == "" {
		return // no session
	}
	store.Lock()
	if data, ok := store.data[cookie.Value]; ok {
		if data.Values == nil {
			data.Values = map[string]string{}
		}
		data.Values[key] = value
	}
	store.Unlock()
}

// GetSessionValue retrieves a custom key from the session
func GetSessionValue(r *http.Request, key string) string {
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie.Value == "" {
		return ""
	}
	store.RLock()
	defer store.RUnlock()
	if data, ok := store.data[cookie.Value]; ok && data.Values != nil {
		return data.Values[key]
	}
	return ""
}

// GetUserID gets the user ID from context
func GetUserID(ctx context.Context) string {
	uid, _ := ctx.Value(userIDKey).(string)
	return uid
}
