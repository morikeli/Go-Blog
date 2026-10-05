package routes

import (
	"net/http"

	"github.com/morikeli/golangrestapi/internal/handlers"
	"github.com/morikeli/golangrestapi/internal/middlewares"
	"github.com/redis/go-redis/v9"
)

func SetupAuthRoutes(mux *http.ServeMux, handler *handlers.Handler, redisClient *redis.Client) {
	authRateLimit := middlewares.AuthRateLimit(redisClient)
	mux.Handle("POST /auth/signup", authRateLimit(http.HandlerFunc(handler.SignupHandler())))
	mux.Handle("POST /auth/login", authRateLimit(http.HandlerFunc(handler.LoginHandler())))
	mux.Handle("POST /auth/token/refresh", authRateLimit(http.HandlerFunc(handler.RefreshTokenHandler())))
	mux.Handle("POST /auth/logout", authRateLimit(http.HandlerFunc(handler.LogoutHandler())))
}
