package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

		cachedKey := fmt.Sprintf("user:%d", userId)
		if cached, err := h.Redis.Get(ctx, cachedKey).Result(); err == nil {
			var cachedUser responses.UserResponse
			if err := json.Unmarshal([]byte(cached), &cachedUser); err == nil {
				responses.Success(w, http.StatusOK, "(From Cache) User profile fetched successfully!", cachedUser)
				return
			}
		}

		// Fetch user from database if not found in cache
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

		// Cache the sanitized user response with 15-minute TTL
		userJSON, _ := json.Marshal(fetchedUser)
		h.Redis.Set(ctx, cachedKey, userJSON, 15*time.Minute)
		
		responses.Success(w, http.StatusOK, "User profile fetched successfully!", fetchedUser)
	}
}
