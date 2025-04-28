package gmail

import "context"

// Context keys
type (
	ctxKeyUserID              struct{}
	ctxKeyPaginationAfterID   struct{}
	ctxKeyPaginationAfterDate struct{}
	ctxKeyDebug               struct{}
)

var (
	// CtxKeyUserID is the context key for user ID
	CtxKeyUserID = ctxKeyUserID{}
	// CtxKeyPaginationAfterID is the context key for pagination after ID
	CtxKeyPaginationAfterID = ctxKeyPaginationAfterID{}
	// CtxKeyPaginationAfterDate is the context key for pagination after date
	CtxKeyPaginationAfterDate = ctxKeyPaginationAfterDate{}
	// keyDebug is the context key for debug flag
	keyDebug = ctxKeyDebug{}
)

// GetUserIDFromContext extracts the user ID from the context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(CtxKeyUserID).(string)
	if !ok || userID == "" {
		return "", ErrUserIDNotFound
	}
	return userID, nil
}

// GetPaginationFromContext extracts pagination parameters from the context
func GetPaginationFromContext(ctx context.Context) (string, int64) {
	afterID := ""
	if v := ctx.Value(CtxKeyPaginationAfterID); v != nil {
		afterID = v.(string)
	}
	afterInternalDate := int64(0)
	if v := ctx.Value(CtxKeyPaginationAfterDate); v != nil {
		afterInternalDate = v.(int64)
	}
	return afterID, afterInternalDate
}

// ContextWithUserID adds a user ID to context
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, CtxKeyUserID, userID)
}

// GetDebugFromContext extracts the debug flag from context
func GetDebugFromContext(ctx context.Context) bool {
	debug, _ := ctx.Value(keyDebug).(bool)
	return debug
}

// ContextWithDebug adds a debug flag to context
func ContextWithDebug(ctx context.Context, debug bool) context.Context {
	return context.WithValue(ctx, keyDebug, debug)
}

// ContextWithPagination adds pagination cursor to context
func ContextWithPagination(ctx context.Context, afterID string, afterInternalDate int64) context.Context {
	ctx = context.WithValue(ctx, CtxKeyPaginationAfterID, afterID)
	return context.WithValue(ctx, CtxKeyPaginationAfterDate, afterInternalDate)
}
