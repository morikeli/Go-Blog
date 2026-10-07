package middlewares

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/morikeli/golangrestapi/internal/dtos/responses"
	"github.com/morikeli/golangrestapi/internal/utils"
	"github.com/redis/go-redis/v9"
)

func AuthMiddleware(tokenMaker *utils.TokenMaker, rdb *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				responses.Error(w, http.StatusUnauthorized, "Authorization header is required!")
				return
			}

			parts := strings.Fields(authHeader)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				responses.Error(w, http.StatusUnauthorized, "Invalid authorization header format!")
				return
			}

			tokenStr := parts[1]

			claims, err := tokenMaker.VerifyAccessToken(tokenStr)
			if err != nil {
				responses.Error(w, http.StatusUnauthorized, "Invalid or expired token!")
				return
			}

			// Reject access tokens that were revoked at logout
			revoked, err := rdb.Exists(r.Context(), utils.RefreshTokenBlacklistPrefix+claims.ID).Result()
			if err != nil {
				log.Printf("failed to check token blacklist: %v", err)
				responses.Error(w, http.StatusServiceUnavailable, "Unable to validate token!")
				return
			}

			if revoked > 0 {
				responses.Error(w, http.StatusUnauthorized, "Token has been revoked!")
				return
			}

			// Inject user info into r.Context() and pass down execution chain
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserId)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
