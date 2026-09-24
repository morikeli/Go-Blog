package routes

import (
	"net/http"

	"github.com/morikeli/golangrestapi/internal/handlers"
)

func SetupAuthRoutes(mux *http.ServeMux, handler *handlers.Handler) {
	mux.HandleFunc("POST /auth/signup", handler.SignupHandler())
	mux.HandleFunc("POST /auth/login", handler.LoginHandler())
	mux.HandleFunc("GET /auth/refreshToken", handler.RefreshTokenHandler())

}
