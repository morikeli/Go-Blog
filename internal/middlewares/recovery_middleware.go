package middlewares

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/morikeli/golangrestapi/internal/dtos/responses"
)

// Recovery catches unhandled panics inside downstream handlers to keep the server running smoothly.
// This ensures a panic in a handler doesn't bring down the whole HTTP server.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Defer a cleanup function that executes after the request pipeline finishes or panics
		defer func() {
			if recovered := recover(); recovered != nil {
				// Retrieve tracing context for correlation in logs[cite: 5, 6]
				requestID := GetRequestID(r.Context())

				// Log the panic details along with stack traces for debugging
				log.Printf(
					"panic recovered: request_id=%s method=%s path=%s panic=%v\n%s",
					requestID,
					r.Method,
					r.URL.Path,
					recovered,
					debug.Stack(),
				)

				// Return a safe 500 response without leaking internal crash details
				responses.Error(w, http.StatusInternalServerError, "Internal server error!")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
