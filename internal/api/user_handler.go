package api

import (
	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/desponda/inbox-whisperer/internal/session"
	"github.com/go-chi/chi/v5"
	"net/http"
	"github.com/rs/zerolog/log"
)

// GetMe handles GET /api/users/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Extract session_id cookie for logging
	var sessionID string
	if cookie, err := r.Cookie("session_id"); err == nil {
		sessionID = cookie.Value
	}
	userID := session.GetUserID(r.Context())
	if userID == "" {
		log.Debug().Str("session_id", sessionID).Msg("GetMe: not authenticated, no userID in session")
		session.ClearSession(w, r)
		RespondError(w, http.StatusUnauthorized, "not authenticated: no userID in session")
		return
	}
	user, err := h.Service.GetUser(r.Context(), userID)
	if err != nil {
		log.Debug().Str("session_id", sessionID).Str("user_id", userID).Msg("GetMe: user not found")
		RespondError(w, http.StatusNotFound, "user not found")
		return
	}
	log.Debug().Str("session_id", sessionID).Str("user_id", userID).Msg("GetMe: user authenticated and found")
	RespondJSON(w, http.StatusOK, user)
}

type UserHandler struct {
	Service service.UserServiceInterface
}

// ListUsers handles GET /users
// Only admin should be able to list all users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	RespondError(w, http.StatusForbidden, "forbidden")
}

// UpdateUser handles PUT /users/{id}
// UpdateUser only allows updating safe fields (none in current model)
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := ValidateIDParam(r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	userIDVal := r.Context().Value(ContextUserIDKey)

	userID, ok := userIDVal.(string)
	if !ok || userID == "" {

		RespondError(w, http.StatusUnauthorized, "not authenticated: no userID in context")
		return
	}
	if id != userID {
		RespondError(w, http.StatusForbidden, "forbidden")
		return
	}
	var req struct {
		// Add safe fields here if/when model expands
	}
	if err := DecodeJSON(r, &req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// No updatable fields; respond with error
	RespondError(w, http.StatusBadRequest, "no updatable fields")
}

// DeleteUser handles DELETE /users/{id}
// DeleteUser performs a soft delete (deactivate)
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := ValidateIDParam(r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	userIDVal := r.Context().Value(ContextUserIDKey)

	userID, ok := userIDVal.(string)
	if !ok || userID == "" {

		RespondError(w, http.StatusUnauthorized, "not authenticated: no userID in context")
		return
	}
	if id != userID {
		RespondError(w, http.StatusForbidden, "forbidden")
		return
	}
	err = h.Service.DeactivateUser(r.Context(), id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func NewUserHandler(svc service.UserServiceInterface) *UserHandler {
	return &UserHandler{Service: svc}
}

// RequireSameUser is middleware that ensures the session user matches the {id} param
func RequireSameUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		sessionUserID := session.GetUserID(r.Context())
		if id == "" || sessionUserID == "" || id != sessionUserID {
			RespondError(w, http.StatusForbidden, "forbidden")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := ValidateIDParam(r)
	if err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}
	userIDVal := r.Context().Value(ContextUserIDKey)

	userID, ok := userIDVal.(string)
	if !ok || userID == "" {

		RespondError(w, http.StatusUnauthorized, "not authenticated: no userID in context")
		return
	}
	if id != userID {
		RespondError(w, http.StatusForbidden, "forbidden")
		return
	}
	user, err := h.Service.GetUser(r.Context(), id)
	if err != nil {
		RespondError(w, http.StatusNotFound, err.Error())
		return
	}
	RespondJSON(w, http.StatusOK, user)
}

// POST /users
// Only admin should be able to create users directly
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	RespondError(w, http.StatusForbidden, "forbidden")
}
