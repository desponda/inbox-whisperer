package api

import "github.com/desponda/inbox-whisperer/internal/data"

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/desponda/inbox-whisperer/internal/session"
	"golang.org/x/oauth2"
)

type GmailHandler struct {
	Service      *service.GmailService
	UserTokens   data.UserTokenRepository
}

func NewGmailHandler(svc *service.GmailService, userTokens data.UserTokenRepository) *GmailHandler {
	return &GmailHandler{Service: svc, UserTokens: userTokens}
}

// FetchMessagesHandler uses the OAuth2 token from the session
func (h *GmailHandler) FetchMessagesHandler(w http.ResponseWriter, r *http.Request) {
	tokenStr := session.GetToken(r.Context())
	var tok *oauth2.Token
	var err error
	if tokenStr != "" {
		tok = &oauth2.Token{AccessToken: tokenStr}
	} else {
		userID := session.GetUserID(r.Context())
		if userID == "" {
			http.Error(w, "not authenticated: no user session", http.StatusUnauthorized)
			return
		}
		tok, err = h.UserTokens.GetUserToken(r.Context(), userID)
		if err != nil || tok == nil {
			http.Error(w, "not authenticated: no token found for user", http.StatusUnauthorized)
			return
		}
	}
	msgs, err := h.Service.FetchMessages(r.Context(), tok)
	if err != nil {
		http.Error(w, "failed to fetch messages: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}

// RegisterGmailRoutes adds the Gmail API endpoints
func RegisterGmailRoutes(r chi.Router, userTokens data.UserTokenRepository) {
	h := NewGmailHandler(service.NewGmailService(), userTokens)
	r.Get("/api/gmail/fetch", h.FetchMessagesHandler)
}

