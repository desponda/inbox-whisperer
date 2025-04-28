package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/desponda/inbox-whisperer/internal/auth/middleware"
	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/desponda/inbox-whisperer/internal/common"
	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/service/email"
	"github.com/desponda/inbox-whisperer/internal/service/provider"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type EmailHandler struct {
	svc            email.Service
	sessionManager session.Manager
}

func NewEmailHandler(svc email.Service, sessionManager session.Manager) *EmailHandler {
	return &EmailHandler{
		svc:            svc,
		sessionManager: sessionManager,
	}
}

// ListEmails handles GET /api/emails
// Fetches a list of email messages and returns them as summaries
// Supports pagination via query parameters:
// - limit: number of items to return (default: 50)
// - after_id: return items after this message ID
// - after_timestamp: return items after this timestamp
func (h *EmailHandler) ListEmails(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	params := common.ParsePaginationFromQuery(
		r.URL.Query().Get("limit"),
		r.URL.Query().Get("after_id"),
		r.URL.Query().Get("after_timestamp"),
	)

	// Add pagination to context
	ctx := common.NewContextWithPagination(r.Context(), params)

	// Token is guaranteed to be valid by middleware
	token := ctx.Value(middleware.OAuthToken{}).(*oauth2.Token)
	summaries, err := h.svc.FetchSummaries(ctx, token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list emails")
		if err == email.ErrNotFound || err == provider.ErrNotFound {
			http.Error(w, "Emails not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch emails", http.StatusInternalServerError)
		return
	}

	// Convert EmailMessage to EmailSummary for response
	summariesForResponse := make([]models.EmailSummary, len(summaries))
	for i, summary := range summaries {
		summariesForResponse[i] = models.EmailSummary{
			ID:           summary.ID,
			ThreadID:     summary.ThreadID,
			Subject:      summary.Subject,
			Sender:       summary.Sender,
			Snippet:      summary.Snippet,
			InternalDate: summary.InternalDate,
			Date:         summary.Date,
			Provider:     summary.Provider,
		}
	}

	// Add pagination metadata to response headers
	if len(summaries) > 0 {
		lastSummary := summaries[len(summaries)-1]
		w.Header().Set("X-Next-After-Id", lastSummary.ID)
		w.Header().Set("X-Next-After-Timestamp", fmt.Sprintf("%d", lastSummary.InternalDate))
	}
	w.Header().Set("X-Page-Limit", fmt.Sprintf("%d", params.Limit))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summariesForResponse); err != nil {
		log.Error().Err(err).Msg("Failed to encode summaries")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// GetEmail handles GET /api/emails/{id}
// Fetches the full content of a single email message
func (h *EmailHandler) GetEmail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing email ID", http.StatusBadRequest)
		return
	}

	// Token is guaranteed to be valid by middleware
	token := r.Context().Value(middleware.OAuthToken{}).(*oauth2.Token)
	emailMsg, err := h.svc.FetchMessage(r.Context(), token, id)
	if err != nil {
		log.Error().Err(err).Str("email_id", id).Msg("Failed to get email")
		if err == email.ErrNotFound || err == provider.ErrNotFound {
			http.Error(w, "Email not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch email", http.StatusInternalServerError)
		return
	}

	// Convert EmailMessage to EmailContent for response
	content := struct {
		ID      string `json:"id"`
		Subject string `json:"subject"`
		From    string `json:"from"`
		To      string `json:"to"`
		Date    string `json:"date"`
		Body    string `json:"body"`
	}{
		ID:      emailMsg.EmailMessageID,
		Subject: emailMsg.Subject,
		From:    emailMsg.Sender,
		To:      emailMsg.Recipient,
		Date:    emailMsg.Date,
		Body:    emailMsg.Body,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(content); err != nil {
		log.Error().Err(err).Msg("Failed to encode content")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// DeleteEmail handles DELETE /api/emails/{id}
func (h *EmailHandler) DeleteEmail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing email ID", http.StatusBadRequest)
		return
	}

	// Not implemented yet
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// UpdateEmail handles PUT /api/emails/{id}
func (h *EmailHandler) UpdateEmail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing email ID", http.StatusBadRequest)
		return
	}

	// Not implemented yet
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// RegisterEmailRoutes adds the Email API endpoints
func RegisterEmailRoutes(r chi.Router, sessionManager session.Manager, db *data.DB, oauthConfig *oauth2.Config) {
	factory := provider.NewProviderFactory()
	h := NewEmailHandler(email.NewMultiProviderService(factory), sessionManager)

	// Create middleware chain
	sessionMiddleware := middleware.NewSessionMiddleware(sessionManager)
	oauthMiddleware := middleware.NewOAuthMiddleware(sessionManager, db, oauthConfig)

	// Apply middleware to routes
	r.Group(func(r chi.Router) {
		r.Use(sessionMiddleware.Handler)
		r.Use(oauthMiddleware.Handler)
		r.Get("/api/emails", h.ListEmails)
		r.Get("/api/emails/{id}", h.GetEmail)
		r.Put("/api/emails/{id}", h.UpdateEmail)
		r.Delete("/api/emails/{id}", h.DeleteEmail)
	})
}
