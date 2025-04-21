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

// InMemoryStore is a demo store for session <-> token
var store = struct {
	sync.RWMutex
	data map[string]*SessionData
}{data: make(map[string]*SessionData)}

type SessionData struct {
	UserID string
	Token  string // Store access token for demo; in prod, store full oauth2.Token
}

// Middleware sets/loads session cookie and attaches user/token to context
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

// SetSession stores user/token for a session
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
	store.data[sessionID] = &SessionData{UserID: userID, Token: token}
	store.Unlock()
}

// GetToken gets the token from context
func GetToken(ctx context.Context) string {
	tok, _ := ctx.Value(tokenKey).(string)
	return tok
}

// GetUserID gets the user ID from context
func GetUserID(ctx context.Context) string {
	uid, _ := ctx.Value(userIDKey).(string)
	return uid
}
