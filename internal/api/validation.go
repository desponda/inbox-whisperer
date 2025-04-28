package api

import (
	"errors"
	"net/http"

	"github.com/desponda/inbox-whisperer/internal/auth/session"
	"github.com/go-chi/chi/v5"
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
	session, ok := session.GetSession(r.Context())
	if !ok || session.UserID() == "" {
		return "", errors.New("not authenticated: no user session")
	}
	return session.UserID(), nil
}

// ValidateSession ensures a valid session exists and returns it, or an error.
func ValidateSession(r *http.Request) (session.Session, error) {
	session, ok := session.GetSession(r.Context())
	if !ok {
		return nil, errors.New("no session found")
	}
	return session, nil
}
