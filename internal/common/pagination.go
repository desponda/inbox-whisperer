package common

import (
	"context"
	"strconv"
)

// PaginationParams represents standardized pagination parameters across all providers
type PaginationParams struct {
	Limit          int    // Number of items to return
	AfterID        string // Cursor-based pagination using message ID
	AfterTimestamp int64  // Cursor-based pagination using internal date
}

type paginationKey struct{}

// NewContextWithPagination creates a new context with pagination parameters
func NewContextWithPagination(ctx context.Context, params PaginationParams) context.Context {
	return context.WithValue(ctx, paginationKey{}, params)
}

// PaginationFromContext extracts pagination parameters from context
func PaginationFromContext(ctx context.Context) PaginationParams {
	if v := ctx.Value(paginationKey{}); v != nil {
		if params, ok := v.(PaginationParams); ok {
			return params
		}
	}
	return PaginationParams{
		Limit: 50, // Default limit if not specified
	}
}

// ParsePaginationFromQuery parses pagination parameters from HTTP query parameters
func ParsePaginationFromQuery(limit, afterID string, afterTimestamp string) PaginationParams {
	params := PaginationParams{
		Limit: 50, // Default limit
	}

	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			params.Limit = l
		}
	}

	if afterID != "" {
		params.AfterID = afterID
	}

	if afterTimestamp != "" {
		if ts, err := strconv.ParseInt(afterTimestamp, 10, 64); err == nil {
			params.AfterTimestamp = ts
		}
	}

	return params
}
