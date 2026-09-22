package routes

import (
	"net/http"

	"github.com/morikeli/golangrestapi/internal/handlers"
)

func SetupHealthRoute(mux *http.ServeMux, handler *handlers.Handler) {
	mux.HandleFunc("/health", handler.CheckServerHealthHandler)
}
