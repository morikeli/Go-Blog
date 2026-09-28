package routes

import (
	"net/http"

	"github.com/morikeli/golangrestapi/internal/handlers"
	"github.com/morikeli/golangrestapi/internal/middlewares"
)

func SetupUserRoutes(mux *http.ServeMux, handler *handlers.Handler) {
	// instantiate auth middleware with token maker
	authMiddleware := middlewares.AuthMiddleware(handler.TokenMaker)

	// wrap user profile handler with auth middleware
	mux.Handle("GET /users", authMiddleware(http.HandlerFunc(handler.ListUsersHandler())))
	mux.Handle("GET /user/profile", authMiddleware(http.HandlerFunc(handler.UserProfileHandler())))
	mux.Handle("PUT /user/profile", authMiddleware(http.HandlerFunc(handler.UpdateUserProfileHandler())))
}
