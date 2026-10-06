package middlewares

import (
	"context"
	"net/http"
	"uuid"
)

// Define custom typed contextKey to avoid collision with keys from other packages
type contextKey string

const (
	RequestIDKey contextKey = "requestID"
	UserIDKey    contextKey = "userID"
	UsernameKey  contextKey = "username"
)

// RequestID ensures every incoming request has a unique request tracking ID.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reuse upstream Request ID if passed (e.g., from Nginx/Cloudflare), or generate a new UUID
		requestID := r.Header.Get("X-Request-ID")

		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Echo back the Request ID in the HTTP response headers
		w.Header().Set("X-Request-ID", requestID)

		// Attach request ID to the context so handlers and loggers can access it
		ctx := WithRequestID(r.Context(), requestID)

		// Pass execution downstream with updated request context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// WithRequestID stores the request ID value inside the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID safely retrieves the request ID string from context.
func GetRequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(RequestIDKey).(string)
	return requestID
}
