package session

import (
	"context"
)

type contextKey string

const (
	sessionKey contextKey = "session"
)

// WithSession stores a session in context
func WithSession(ctx context.Context, session Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

// GetSession retrieves a session from context
func GetSession(ctx context.Context) (Session, bool) {
	session, ok := ctx.Value(sessionKey).(Session)
	return session, ok
}
