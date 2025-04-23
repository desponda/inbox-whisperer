package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/desponda/inbox-whisperer/internal/service/gmail"
	"github.com/desponda/inbox-whisperer/internal/session"
	"golang.org/x/oauth2"
)

type EmailHandler struct {
	Service    service.EmailService
	UserTokens data.UserTokenRepository
}

func NewEmailHandler(svc service.EmailService, userTokens data.UserTokenRepository) *EmailHandler {
	return &EmailHandler{Service: svc, UserTokens: userTokens}
}

// FetchMessagesHandler supports cursor-based pagination via after_id and after_internal_date query params.
// Example: /api/email/fetch?after_id=abc123&after_internal_date=1713750000
func (h *EmailHandler) FetchMessagesHandler(w http.ResponseWriter, r *http.Request) {
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
	// Parse pagination params
	afterID := r.URL.Query().Get("after_id")
	afterInternalDate := int64(0)
	if v := r.URL.Query().Get("after_internal_date"); v != "" {
		fmt.Sscanf(v, "%d", &afterInternalDate)
	}
	ctx := r.Context()
	if afterID != "" && afterInternalDate > 0 {
		ctx = context.WithValue(ctx, "after_id", afterID)
		ctx = context.WithValue(ctx, "after_internal_date", afterInternalDate)
	}
	msgs, err := h.Service.FetchMessages(ctx, tok)
	if err != nil {
		if err.Error() == "not found" {
			http.Error(w, "messages not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch messages: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}

// GetMessageContentHandler fetches and returns the full content of a single email by ID
func (h *EmailHandler) GetMessageContentHandler(w http.ResponseWriter, r *http.Request) {
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

// RegisterEmailRoutes adds the Email API endpoints
// Uses *data.DB (with pgxpool.Pool) for production DB injection
func RegisterEmailRoutes(r chi.Router, userTokens data.UserTokenRepository, db *data.DB) {
	repo := data.NewEmailMessageRepositoryFromPool(db.Pool)
	factory := service.NewEmailProviderFactory()
	// Register GmailProvider
	gmailSvc := gmail.NewGmailService(repo)
	factory.RegisterProvider(service.ProviderGmail, func(cfg service.ProviderConfig) (service.EmailProvider, error) {
		return gmail.NewGmailProvider(gmailSvc), nil
	})
	// For demo, link Gmail for all users (in real app, link by user config)
	factory.LinkProvider("demo-user", service.ProviderConfig{UserID: "demo-user", Type: service.ProviderGmail})
	// Use MultiProviderEmailService
	var svc service.EmailService = service.NewMultiProviderEmailService(factory)
	h := NewEmailHandler(svc, userTokens)
	r.Get("/api/email/fetch", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := session.GetUserID(ctx)
		if userID == "" {
			http.Error(w, "not authenticated: no user session", http.StatusUnauthorized)
			return
		}
		ctx = context.WithValue(ctx, "user_id", userID)
		r = r.WithContext(ctx)
		h.FetchMessagesHandler(w, r)
	})
	r.Get("/api/email/messages/{id}", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := session.GetUserID(ctx)
		if userID == "" {
			http.Error(w, "not authenticated: no user session", http.StatusUnauthorized)
			return
		}
		ctx = context.WithValue(ctx, "user_id", userID)
		r = r.WithContext(ctx)
		h.GetMessageContentHandler(w, r)
	})
}
