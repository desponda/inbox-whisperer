package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/desponda/inbox-whisperer/internal/auth/middleware"
	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/desponda/inbox-whisperer/internal/common"
	"github.com/desponda/inbox-whisperer/internal/data"
	"github.com/desponda/inbox-whisperer/internal/models"
	"github.com/desponda/inbox-whisperer/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type UserHandler struct {
	svc            service.UserServiceInterface
	sessionManager session.Manager
}

func NewUserHandler(svc service.UserServiceInterface, sessionManager session.Manager) *UserHandler {
	return &UserHandler{
		svc:            svc,
		sessionManager: sessionManager,
	}
}

// requireSelfOrAdmin checks if the current user is the same as the target user or has the admin role
func requireSelfOrAdmin(r *http.Request, targetID uuid.UUID) bool {
	userID, ok := common.UserIDFromContext(r.Context())
	if !ok {
		return false
	}
	if userID == targetID {
		return true
	}
	role, _ := common.RoleFromContext(r.Context())
	return role == "admin"
}

// GetMe handles GET /api/users/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := common.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	user, err := h.svc.GetUser(r.Context(), userID)
	if err != nil {
		slog.Debug("GetMe: user not found", "user_id", userID, "error", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	slog.Debug("GetMe: user found", "user_id", userID)
	json.NewEncoder(w).Encode(user)
}

// ListUsers handles GET /users
// Only admin should be able to list all users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	role, _ := common.RoleFromContext(r.Context())
	if role != "admin" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	users, err := h.svc.ListUsers(r.Context())
	if err != nil {
		slog.Error("Failed to list users", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		slog.Error("Failed to encode users", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// UpdateUser handles PUT /users/{id}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	if !requireSelfOrAdmin(r, id) {
		slog.Debug("UpdateUser: forbidden", "requested_id", id)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	user.ID = id
	if err := h.svc.UpdateUser(r.Context(), &user); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}

// DeleteUser handles DELETE /users/{id}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	if !requireSelfOrAdmin(r, id) {
		slog.Debug("DeleteUser: forbidden", "requested_id", id)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if err := h.svc.DeactivateUser(r.Context(), id); err != nil {
		slog.Error("Failed to deactivate user", "user_id", id, "error", err)
		http.Error(w, "Failed to deactivate user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetUser handles GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	if !requireSelfOrAdmin(r, id) {
		slog.Debug("GetUser: forbidden", "requested_id", id)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	user, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		slog.Error("Failed to get user", "user_id", id, "error", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(user)
}

// CreateUser handles POST /users
// Only admin should be able to create users directly
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Forbidden", http.StatusForbidden)
}

// RegisterUserRoutes adds the User API endpoints
func RegisterUserRoutes(r chi.Router, sessionManager session.Manager, db *data.DB, oauthConfig *oauth2.Config) {
	h := NewUserHandler(service.NewUserService(db), sessionManager)

	// Define public paths
	publicPaths := []string{
		"/api/health/*", // Health check endpoints
		"/api/auth/*",   // Authentication endpoints
		"/api/docs/*",   // API documentation
	}

	// Create middleware chain
	sessionMiddleware := middleware.NewSessionMiddleware(sessionManager, publicPaths...)
	oauthMiddleware := middleware.NewOAuthMiddleware(sessionManager, db, oauthConfig)

	// Apply middleware to routes
	r.Group(func(r chi.Router) {
		r.Use(sessionMiddleware.Handler)
		r.Use(oauthMiddleware.Handler)
		r.Get("/api/users/me", h.GetMe)
		r.Get("/api/users", h.ListUsers)
		r.Post("/api/users", h.CreateUser)
		r.Get("/api/users/{id}", h.GetUser)
		r.Put("/api/users/{id}", h.UpdateUser)
		r.Delete("/api/users/{id}", h.DeleteUser)
	})
}
