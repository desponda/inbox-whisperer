package api

import (
	"github.com/desponda/inbox-whisperer/internal/data"
)

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/desponda/inbox-whisperer/internal/session"
	"golang.org/x/oauth2"
)

type GmailServiceInterface interface {
	FetchMessages(ctx context.Context, token *oauth2.Token) ([]service.MessageSummary, error)
	FetchMessageContent(ctx context.Context, token *oauth2.Token, id string) (*service.MessageContent, error)
}

type GmailHandler struct {
	Service    GmailServiceInterface
	UserTokens data.UserTokenRepository
}

func NewGmailHandler(svc GmailServiceInterface, userTokens data.UserTokenRepository) *GmailHandler {
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

// GetMessageContentHandler fetches and returns the full content of a single email by ID
func (h *GmailHandler) GetMessageContentHandler(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing email id", http.StatusBadRequest)
		return
	}
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
	msg, err := h.Service.FetchMessageContent(r.Context(), tok, id)
	if err != nil {
		if err.Error() == "not found" {
			http.Error(w, "email not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch email content: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

// RegisterGmailRoutes adds the Gmail API endpoints
// RegisterGmailRoutes adds the Gmail API endpoints
// Uses *data.DB (with pgxpool.Pool) for production DB injection
func RegisterGmailRoutes(r chi.Router, userTokens data.UserTokenRepository, db *data.DB) {
	repo := data.NewGmailMessageRepositoryFromPool(db.Pool)
	h := NewGmailHandler(service.NewGmailService(repo), userTokens)
	r.Get("/api/gmail/fetch", h.FetchMessagesHandler)
	r.Get("/api/gmail/messages/{id}", h.GetMessageContentHandler)
}

