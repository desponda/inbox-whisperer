package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/google/uuid"
	"time"
)

type UserHandler struct {
	Service service.UserServiceInterface
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Service.ListUsers(r.Context())
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondJSON(w, http.StatusOK, users)
}

// UpdateUser handles PUT /users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "missing id")
		return
	}
	var user models.User
	if err := DecodeJSON(r, &user); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user.ID = id
	if err := h.Service.UpdateUser(r.Context(), &user); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondJSON(w, http.StatusOK, user)
}

// DeleteUser handles DELETE /users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "missing id")
		return
	}
	if err := h.Service.DeleteUser(r.Context(), id); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func NewUserHandler(svc service.UserServiceInterface) *UserHandler {
	return &UserHandler{Service: svc}
}

// GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		RespondError(w, http.StatusBadRequest, "missing id")
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
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := DecodeJSON(r, &user); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if user.Email == "" {
		RespondError(w, http.StatusBadRequest, "missing email")
		return
	}
	if user.ID == "" {
		user.ID = uuid.NewString()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	} else {
		user.CreatedAt = user.CreatedAt.UTC()
	}
	err := h.Service.CreateUser(r.Context(), &user)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondJSON(w, http.StatusCreated, user)
}
