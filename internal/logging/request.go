package logging

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDKey is the context key for request IDs
type RequestIDKey string

const (
	// RequestIDContextKey is the key used to store request IDs in context
	RequestIDContextKey RequestIDKey = "request_id"
	// RequestIDHeader is the HTTP header name for request IDs
	RequestIDHeader = "X-Request-ID"
)

// GenerateRequestID generates a new unique request ID
func GenerateRequestID() string {
	return uuid.New().String()
}

// GetRequestIDFromContext retrieves the request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDContextKey).(string); ok {
		return id
	}
	return ""
}

// SetRequestIDInContext stores the request ID in context
func SetRequestIDInContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDContextKey, requestID)
}

// RequestIDMiddleware is a middleware that generates or extracts request IDs
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get request ID from header first
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			// Generate new request ID if not provided
			requestID = GenerateRequestID()
		}

		// Set request ID in response header
		w.Header().Set(RequestIDHeader, requestID)

		// Add request ID to context
		ctx := SetRequestIDInContext(r.Context(), requestID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}