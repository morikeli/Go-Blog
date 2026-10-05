package middlewares

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/morikeli/golangrestapi/internal/dtos/responses"
)

const (
	authRateLimit       = 5           // Maximum allowed requests within the time window
	authRateLimitWindow = time.Minute // Duration of the rate limit window
)

// rateLimitScript is a Redis Lua script executed atomically on the server.
// It increments a request counter and sets an expiration key on the first request.
// Lua scripts execute atomically in Redis, preventing race conditions.
var rateLimitScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])

-- If this is the first request for this key in the current window, set its TTL
if current == 1 then
    redis.call("EXPIRE", KEYS[1], ARGV[1])
end

return current
`)

// AuthRateLimit creates an HTTP middleware that limits incoming requests by client IP.
// It accepts a Redis client instance and returns a standard HTTP middleware handler.
func AuthRateLimit(rdb *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract client IP address from the incoming request
			clientIP := getClientIP(r)

			// Construct a unique Redis key for this IP address
			key := fmt.Sprintf("rate_limit:auth:ip:%s", clientIP)

			// Run the atomic Lua script in Redis.
			// Pass key as KEYS[1] and window duration (in seconds) as ARGV[1].
			count, err := rateLimitScript.Run(
				r.Context(),
				rdb,
				[]string{key},
				int(authRateLimitWindow.Seconds()),
			).Int()

			if err != nil {
				// Fail-Closed Strategy: If Redis is down or unreachable, reject requests.
				// This protects auth endpoints from brute-force attacks during Redis outages.
				responses.Error(
					w,
					http.StatusServiceUnavailable,
					"Authentication service temporarily unavailable!",
				)
				return
			}

			// Inform the client about the total request limit allowed per window
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", authRateLimit))

			// Check if client exceeds max allowed attempts
			if count > authRateLimit {
				// Tell the client how many seconds to wait before retrying
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(authRateLimitWindow.Seconds())))

				// Block the request with HTTP 429 Too Many Requests
				responses.Error(
					w,
					http.StatusTooManyRequests,
					"Too many authentication attempts. Please try again later.",
				)
				return
			}

			// Request is within allowed limits; forward it to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the IP address from the request's RemoteAddr field.
func getClientIP(r *http.Request) string {
	// Split the host and port from RemoteAddr (e.g., "192.168.1.1:12345" -> "192.168.1.1")
	host, _, err := net.SplitHostPort(r.RemoteAddr)

	if err == nil {
		return host
	}

	// Fall back to raw RemoteAddr if it does not contain a port
	return strings.TrimSpace(r.RemoteAddr)
}
