package api

import (
	"errors"
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/desponda/inbox-whisperer/internal/session"
)

// ValidateIDParam checks if the chi URL param 'id' is present and returns it, or an error.
func ValidateIDParam(r *http.Request) (string, error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		return "", errors.New("missing id parameter")
	}
	return id, nil
}

// ValidateAuth ensures the user is authenticated and returns the userID, or an error.
func ValidateAuth(r *http.Request) (string, error) {
	userID := session.GetUserID(r.Context())
	if userID == "" {
		return "", errors.New("not authenticated: no user session")
	}
	return userID, nil
}
