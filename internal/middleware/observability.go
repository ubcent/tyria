// Package middleware provides HTTP middleware for logging and tracing
package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/ubcent/edge.link/internal/logging"
	"github.com/ubcent/edge.link/internal/tracing"
)

// LoggingAndTracing creates middleware that combines structured logging with OpenTelemetry tracing
func LoggingAndTracing(logger *logging.Logger, tracingProvider *tracing.Provider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Wrap with OpenTelemetry HTTP instrumentation
		otelHandler := otelhttp.NewHandler(next, "http-request")

		// Add request ID middleware
		requestIDHandler := logging.RequestIDMiddleware(otelHandler)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get request ID from context
			requestID := logging.GetRequestIDFromContext(r.Context())

			// Get current span for tracing correlation
			span := trace.SpanFromContext(r.Context())

			// Create request-scoped logger with basic request info
			requestLogger := logger.WithRequest(r.Method, r.URL.Path, getClientIP(r)).
				WithRequestID(requestID)

			// Log the incoming request
			requestLogger.Info().
				Str("user_agent", r.UserAgent()).
				Str("referer", r.Referer()).
				Msg("incoming request")

			// Add tracing attributes for the HTTP request
			if span.IsRecording() {
				span.SetAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.user_agent", r.UserAgent()),
					attribute.String("request.id", requestID),
				)
			}

			// Continue with the request
			requestIDHandler.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP from request headers
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}