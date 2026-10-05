package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/morikeli/golangrestapi/internal/dtos/responses"
	"github.com/morikeli/golangrestapi/internal/services"
)

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidCredentials):
		responses.Error(w, http.StatusUnauthorized, "Invalid credentials provided!")

	case errors.Is(err, services.ErrUserNotFound):
		responses.Error(w, http.StatusNotFound, "User profile not found!")

	case errors.Is(err, services.ErrDuplicateEmail),
		errors.Is(err, services.ErrDuplicateUsername):
		responses.Error(w, http.StatusConflict, "Username or email already taken!")

	default:
		// Log the actual internal error.
		log.Printf("internal service error: %v", err)

		// Never expose it to the client.
		responses.Error(
			w,
			http.StatusInternalServerError,
			"Oh snap! Something went wrong. Please try again later.",
		)
	}
}
