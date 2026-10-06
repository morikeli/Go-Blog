package middlewares

import "net/http"

// SecurityHeaders injects standard defensive HTTP response headers to protect client browsers.
func SecurityHeadersMiddleware(isProd bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headers := w.Header()

			// Prevent browsers from MIME-sniffing response body types away from content-type
			headers.Set("X-Content-Type-Options", "nosniff")

			// Disallow rendering the site inside iframes to prevent clickjacking
			headers.Set("X-Frame-Options", "DENY")

			// Restrict referrer info sent on cross-origin requests
			headers.Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Disable unused browser capability features
			headers.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

			if isProd {
				// Force HTTPS connections (Strict-Transport-Security) for 1 year
				headers.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			next.ServeHTTP(w, r)
		})
	}
}
