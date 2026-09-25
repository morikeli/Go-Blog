package handlers

import (
	"net/http"

	"github.com/morikeli/golangrestapi/internal/dtos/responses"
	"github.com/morikeli/golangrestapi/internal/middlewares"
)

func (h *Handler) UserProfileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userId, ok := ctx.Value(middlewares.UserIDKey).(int64)
		if !ok {
			responses.Error(w, http.StatusUnauthorized, "Please login to your account to view your profile!")
			return
		}

		user, err := h.Queries.GetUserById(ctx, userId)

		if err != nil {
			responses.Error(w, http.StatusNotFound, "User profile not found!")
			return
		}

		fetchedUser := responses.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		responses.Success(w, http.StatusOK, "User profile fetched successfully!", fetchedUser)
	}
}
