package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/morikeli/golangrestapi/internal/dtos/responses"
	"github.com/morikeli/golangrestapi/internal/utils"
)

func AuthMiddleware(tokenMaker *utils.TokenMaker) func(http.Handler) http.Handler {
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

			// Inject user info into r.Context() and pass down execution chain
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserId)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
