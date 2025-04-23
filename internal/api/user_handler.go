package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/desponda/inbox-whisperer/internal/session"
)

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
	userID := r.Context().Value(ContextUserIDKey).(string)
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
	userID := r.Context().Value(ContextUserIDKey).(string)
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
	userID := r.Context().Value(ContextUserIDKey).(string)
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
