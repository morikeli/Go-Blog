package middlewares

import (
	"net/http"
	"strings"
)

// CORS creates a middleware that manages Cross-Origin Resource Sharing based on an allowed origin list.
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	// Pre-build a set (map) for O(1) lookups during incoming HTTP requests
	allowed := make(map[string]struct{}, len(allowedOrigins))

	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)

		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Validate if the request origin matches the whitelist
			if _, ok := allowed[origin]; ok {
				// Reflect the validated origin back to allow credentials safely
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set(
					"Access-Control-Allow-Methods",
					"GET, POST, PATCH, DELETE, OPTIONS",
				)
				w.Header().Set(
					"Access-Control-Allow-Headers",
					"Content-Type, Authorization, X-Request-ID",
				)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				// Ensure upstream caches vary response headers based on origin
				w.Header().Add("Vary", "Origin")
			}

			// Intercept and resolve preflight OPTIONS requests without invoking downstream routes
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}