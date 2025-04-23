package api

import (
	"context"
	"net/http"
	"github.com/desponda/inbox-whisperer/internal/session"
	"github.com/desponda/inbox-whisperer/internal/data"
)

// Context keys for userID and token
const (
	ContextUserIDKey = "userID"
	ContextTokenKey  = "userToken"
)

// AuthMiddleware ensures the user is authenticated and attaches userID to context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := session.GetUserID(r.Context())
		if userID == "" {
			
			http.Error(w, "not authenticated: no user session", http.StatusUnauthorized)
			return
		}
		
		ctx := context.WithValue(r.Context(), ContextUserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TokenMiddleware fetches the user's token and attaches it to context
func TokenMiddleware(userTokens data.UserTokenRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(ContextUserIDKey).(string)
			if !ok || userID == "" {
				
				http.Error(w, "not authenticated: no userID in context", http.StatusUnauthorized)
				return
			}
			
			tok, err := userTokens.GetUserToken(r.Context(), userID)
			if err != nil {
				
			}
			if tok == nil {
				
				http.Error(w, "not authenticated: no token found for user", http.StatusUnauthorized)
				return
			}
			
			ctx := context.WithValue(r.Context(), ContextTokenKey, tok)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
