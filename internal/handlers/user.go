package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/morikeli/golangrestapi/internal/dtos/requests"
	"github.com/morikeli/golangrestapi/internal/dtos/responses"
	"github.com/morikeli/golangrestapi/internal/store"
	"github.com/morikeli/golangrestapi/internal/utils"
)

func (h *Handler) SignupHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// create context
		ctx := r.Context()

		var req requests.SignupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		hashedPassword, err := utils.HashPassword(req.Password)

		if err != nil {
			responses.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		
		_, err = h.Queries.CreateUser(ctx, store.CreateUserParams{
			Username: req.Username,
			Email: req.Email,
			Password: hashedPassword,
		})

		if err != nil {
			responses.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		
		responses.Success(w, http.StatusCreated, "User account created successfully!", nil)
	}
}