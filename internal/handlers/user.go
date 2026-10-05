package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/morikeli/golangrestapi/internal/dtos/requests"
	"github.com/morikeli/golangrestapi/internal/dtos/responses"
	"github.com/morikeli/golangrestapi/internal/middlewares"
	"github.com/morikeli/golangrestapi/internal/utils"
)

func (h *Handler) ListUsersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Parse limit with safe defaults (default: 10, max: 100)
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit <= 0 {
			limit = 10
		} else if limit > 100 {
			limit = 100
		}

		// Parse offset (default: 0)
		offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil || offset < 0 {
			offset = 0
		}

		// Fetch paginated users
		users, totalItems, err := h.UserService.ListUsers(ctx, limit, offset)
		if err != nil {
			log.Printf("failed to retrieve users: %v", err)
			responses.Error(w, http.StatusInternalServerError, "Failed to retrieve users!")
			return
		}

		// Map DB structs to response DTOs
		userResponses := make([]responses.UserResponse, 0, len(users))
		for _, u := range users {
			var photoURL *string
			if u.ProfilePhoto.Valid && u.ProfilePhoto.String != "" {
				photoURL = &u.ProfilePhoto.String // return null if there's no profile photo
			}

			userResponses = append(userResponses, responses.UserResponse{
				ID:             u.ID,
				Username:       u.Username,
				Email:          u.Email,
				ProfilePicture: photoURL,
				CreatedAt:      u.CreatedAt,
				UpdatedAt:      u.UpdatedAt,
			})
		}

		// Calculate current page & total pages safely
		currentPage := (offset / limit) + 1
		totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

		paginatedResult := responses.PaginatedUserResponse{
			Users: userResponses,
			Pagination: responses.PaginationMeta{
				Page:       currentPage,
				Limit:      limit,
				TotalItems: totalItems,
				TotalPages: totalPages,
			},
		}

		responses.Success(w, http.StatusOK, "Users retrieved successfully!", paginatedResult)
	}
}

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
		user, err := h.UserService.GetUser(ctx, userId)

		if err != nil {
			handleServiceError(w, err)
			return
		}

		var profilePic *string
		if user.ProfilePhoto.Valid && user.ProfilePhoto.String != "" {
			profilePic = &user.ProfilePhoto.String
		}

		fetchedUser := responses.UserResponse{
			ID:             user.ID,
			Username:       user.Username,
			Email:          user.Email,
			ProfilePicture: profilePic,
			CreatedAt:      user.CreatedAt,
			UpdatedAt:      user.UpdatedAt,
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
		const maxFileSize = utils.MaxProfileImageSize
		r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)
		if err := r.ParseMultipartForm(maxFileSize); err != nil {
			if utils.IsRequestBodyTooLarge(err) {
				responses.Error(w, http.StatusBadRequest, "Uploaded profile picture is too large!")
				return
			}
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

		file, fileHeader, fileErr := r.FormFile("profile_picture")

		// Ensure at least one field is provided to update user profile
		if req.Username == "" && fileErr != nil {
			responses.Error(w, http.StatusBadRequest, "Please provide a username or profile picture to update your profile!")
			return
		}

		var profilePhoto *string

		// Handle profile picture update if a file was attached
		if fileErr == nil {
			defer file.Close()

			// Validate profile image during upload
			_, err := utils.ValidateProfileImage(file, fileHeader.Size)
			if err != nil {
				switch {
				case errors.Is(err, utils.ErrImageTooLarge):
					responses.Error(w, http.StatusRequestEntityTooLarge, "Profile photo is too large! Max. size is 5 MB.")

				case errors.Is(err, utils.ErrUnsupportedImage):
					responses.Error(w, http.StatusUnsupportedMediaType, "Unsupported image format: only .jpeg, .png, .webp allowed!")

				default:
					responses.Error(w, http.StatusUnsupportedMediaType, "Invalid profile photo uploaded!")
				}
				return
			}

			uploadParams := uploader.UploadParams{
				Folder:         "profile_pictures",
				PublicID:       fmt.Sprintf("user_%d", userId),
				Overwrite:      api.Bool(true), // Overwrite previous picture for this user ID
				ResourceType:   "image",
				Transformation: "c_fill,g_face,w_400,h_400,r_max", // Square 400x400 auto-cropped to human face
			}

			uploadResult, err := h.Cloudinary.Upload.Upload(ctx, file, uploadParams)
			if err != nil {
				log.Printf("failed to upload image to cloud storage: %v", err)
				responses.Error(w, http.StatusInternalServerError, "Failed to upload image to cloud storage!")
				return
			}

			profilePhoto = &uploadResult.SecureURL
		}

		// update user profile
		updateUser, err := h.UserService.UpdateProfile(ctx, userId, req.Username, profilePhoto)
		if err != nil {
			handleServiceError(w, err)
			return
		}

		// Invalidate stale cache for this user
		cachedKey := fmt.Sprintf("user:%d", userId)
		_ = h.Redis.Del(ctx, cachedKey).Err()

		var profilePicture *string
		if updateUser.ProfilePhoto.Valid && updateUser.ProfilePhoto.String != "" {
			profilePicture = &updateUser.ProfilePhoto.String
		}

		// Return response DTO
		userResponse := responses.UserResponse{
			ID:             updateUser.ID,
			Username:       updateUser.Username,
			Email:          updateUser.Email,
			ProfilePicture: profilePicture,
			CreatedAt:      updateUser.CreatedAt,
			UpdatedAt:      updateUser.UpdatedAt,
		}

		responses.Success(w, http.StatusOK, "User profile updated successfully!", userResponse)

	}
}
