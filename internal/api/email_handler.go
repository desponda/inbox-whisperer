package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/service"
	gmail "github.com/desponda/inbox-whisperer/internal/service/gmail"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
)

type EmailHandler struct {
	Service    service.EmailService
	UserTokens data.UserTokenRepository
}

func NewEmailHandler(svc service.EmailService, userTokens data.UserTokenRepository) *EmailHandler {
	return &EmailHandler{Service: svc, UserTokens: userTokens}
}

func (h *EmailHandler) FetchMessagesHandler(w http.ResponseWriter, r *http.Request) {
	tokVal := r.Context().Value(ContextTokenKey)
	
	tok, ok := tokVal.(*oauth2.Token)
	if !ok || tok == nil {
		http.Error(w, "not authenticated: no token in context", http.StatusUnauthorized)
		return
	}
	ctx := h.extractPagination(r)
	msgs, err := h.Service.FetchMessages(ctx, tok)
	if err != nil {
		
		if errors.Is(err, gmail.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msgs)
}

func (h *EmailHandler) GetMessageContentHandler(w http.ResponseWriter, r *http.Request) {
	id, err := ValidateIDParam(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tokVal := r.Context().Value(ContextTokenKey)
	
	tok, ok := tokVal.(*oauth2.Token)
	if !ok || tok == nil {
		http.Error(w, "not authenticated: no token in context", http.StatusUnauthorized)
		return
	}
	msg, err := h.Service.FetchMessageContent(r.Context(), tok, id)
	if err != nil {
		
		if err.Error() == "not found" {
			http.Error(w, "email not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

func (h *EmailHandler) extractPagination(r *http.Request) context.Context {
	ctx := r.Context()
	afterID := r.URL.Query().Get("after_id")
	afterInternalDate := int64(0)
	if v := r.URL.Query().Get("after_internal_date"); v != "" {
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			afterInternalDate = parsed
		}
	}
	if afterID != "" && afterInternalDate > 0 {
		ctx = context.WithValue(ctx, "after_id", afterID)
		ctx = context.WithValue(ctx, "after_internal_date", afterInternalDate)
	}
	return ctx
}

// RegisterEmailRoutes adds the Email API endpoints
// Uses *data.DB (with pgxpool.Pool) for production DB injection
func RegisterEmailRoutes(r chi.Router, userTokens data.UserTokenRepository, db *data.DB) {
	factory := service.NewEmailProviderFactory()
	h := NewEmailHandler(service.NewMultiProviderEmailService(factory), userTokens)
	r.Get("/api/email/messages", h.FetchMessagesHandler)
	r.Get("/api/email/messages/{id}", h.GetMessageContentHandler)
}
