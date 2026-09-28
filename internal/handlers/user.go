package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/morikeli/golangrestapi/internal/dtos/requests"
	"github.com/morikeli/golangrestapi/internal/dtos/responses"
	"github.com/morikeli/golangrestapi/internal/middlewares"
	"github.com/morikeli/golangrestapi/internal/store"
	"github.com/morikeli/golangrestapi/internal/utils"
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

func (h *Handler) UpdateUserProfileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userId, ok := ctx.Value(middlewares.UserIDKey).(int64)
		if !ok {
			responses.Error(w, http.StatusUnauthorized, "Please login to your account to view your profile!")
			return
		}

		// upload profile picture
		// (10 << 20); Limit file upload to 10 MB
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			responses.Error(w, http.StatusBadRequest, "Failed to parse multipart form!")
			return
		}

		req := requests.UpdateUserProfileRequest{
			Username: r.FormValue("username"),
		}

		if err := utils.ValidateRequest(req); err != nil {
			responses.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		file, _, fileErr := r.FormFile("profile_picture")

		// Ensure at least one field is provided to update user profile
		if req.Username == "" && fileErr != nil {
			responses.Error(w, http.StatusBadRequest, "Please provide a username or profile picture to update your profile!")
			return
		}

		// Prepare sqlc params structure
		params := store.UpdateUserProfileParams{
			ID: userId,
		}

		// Handle profile picture update if a file was attached
		if req.Username != "" {
			params.Username = pgtype.Text{String: req.Username, Valid: true}
		}

		// Handle profile picture update if a file was attached
		if fileErr == nil {
			defer file.Close()

			cld, err := cloudinary.NewFromURL(h.Config.CloudinaryURL)
			if err != nil {
				responses.Error(w, http.StatusInternalServerError, "Failed to initialize cloud storage client!")
				return
			}

			uploadParams := uploader.UploadParams{
				Folder:         "profile_pictures",
				PublicID:       fmt.Sprintf("user_%d", userId),
				Overwrite:      api.Bool(true), // Overwrite previous picture for this user ID
				ResourceType:   "image",
				Transformation: "c_fill,g_face,w_400,h_400,r_max", // Square 400x400 auto-cropped to human face
			}

			uploadResult, err := cld.Upload.Upload(ctx, file, uploadParams)
			if err != nil {
				responses.Error(w, http.StatusInternalServerError, "Failed to upload image to cloud storage!")
				return
			}

			params.ProfilePhoto = pgtype.Text{String: uploadResult.SecureURL, Valid: true}
		}

		// update user profile
		updateUser, err := h.Queries.UpdateUserProfile(ctx, params)
		if err != nil {
			responses.Error(w, http.StatusInternalServerError, "Failed to update user profile!")
			return
		}

		// Invalidate stale cache for this user
		cachedKey := fmt.Sprintf("user:%d", userId)
		_ = h.Redis.Del(ctx, cachedKey).Err()

		// 6. Return response DTO
		userResponse := responses.UserResponse{
			ID:        updateUser.ID,
			Username:  updateUser.Username,
			Email:     updateUser.Email,
			ProfilePicture: updateUser.ProfilePhoto.String,
			CreatedAt: updateUser.CreatedAt,
			UpdatedAt: updateUser.UpdatedAt,
		}

		responses.Success(w, http.StatusOK, "User profile updated successfully!", userResponse)

	}
}
