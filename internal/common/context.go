package common

import (
	"context"

	"github.com/google/uuid"
)

// Context keys for common values
type (
	UserIDKey     struct{}
	PaginationKey struct{}
	ProviderKey   struct{}
	OAuthTokenKey struct{}
	LimitKey      struct{}
	RoleKey       struct{}
)

// ContextWithUserID adds a user ID to the context (uuid.UUID)
func ContextWithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, UserIDKey{}, userID)
}

// UserIDFromContext extracts the user ID from context as uuid.UUID
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(UserIDKey{})
	if v == nil {
		return uuid.Nil, false
	}
	switch id := v.(type) {
	case uuid.UUID:
		return id, true
	case string:
		u, err := uuid.Parse(id)
		if err == nil {
			return u, true
		}
		return uuid.Nil, false
	default:
		return uuid.Nil, false
	}
}

// ContextWithProvider adds a provider name to the context
func ContextWithProvider(ctx context.Context, provider string) context.Context {
	return context.WithValue(ctx, ProviderKey{}, provider)
}

// ProviderFromContext extracts the provider from context
func ProviderFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ProviderKey{})
	if v == nil {
		return "", false
	}
	provider, ok := v.(string)
	return provider, ok
}

// GetLimitFromContext gets the pagination limit from context
func GetLimitFromContext(ctx context.Context) int {
	if v := ctx.Value(LimitKey{}); v != nil {
		if limit, ok := v.(int); ok && limit > 0 {
			return limit
		}
	}
	return 50 // default limit
}

// ContextWithRole adds a user role to the context
func ContextWithRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, RoleKey{}, role)
}

// RoleFromContext extracts the user role from context
func RoleFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(RoleKey{})
	if v == nil {
		return "", false
	}
	role, ok := v.(string)
	return role, ok
}
