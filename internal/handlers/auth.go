package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/morikeli/golangrestapi/internal/dtos/requests"
	"github.com/morikeli/golangrestapi/internal/dtos/responses"
	"github.com/morikeli/golangrestapi/internal/store"
	"github.com/morikeli/golangrestapi/internal/utils"
)

func (h *Handler) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req requests.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			responses.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		// Validate the request
		if err := utils.ValidateRequest(req); err != nil {
			responses.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		// fetch user
		user, err := h.Queries.GetUserByUsernameOrEmail(ctx, req.Username)

		if err != nil {
			responses.Error(w, http.StatusUnauthorized, "Invalid credentials provided!")
			return
		}

		if !utils.ValidatePassword(user.Password, req.Password) {
			responses.Error(w, http.StatusUnauthorized, "Invalid credentials provided!")
			return
		}

		// Generate JWT token
		accessToken, err := h.TokenMaker.GenerateAccessToken(int64(user.ID), user.Username, 15*time.Minute)

		if err != nil {
			responses.Error(w, http.StatusInternalServerError, "An error occurred while generating access token!")
			return
		}

		refreshToken, err := h.TokenMaker.GenerateRefreshToken(int64(user.ID), user.Username, 7*24*time.Hour)

		if err != nil {
			responses.Error(w, http.StatusInternalServerError, "An error occurred while generating refresh token!")
			return
		}

		// Attach Refresh Token as an HttpOnly, Secure Cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Path:     "/auth/refreshToken", // Restrict cookie scope exclusively to the refresh endpoint
			Expires:  time.Now().Add(7 * 24 * time.Hour),
			HttpOnly: true,                 // JavaScript cannot read this cookie (XSS protection)
			Secure:   true,                 // Requires HTTPS in production
			SameSite: http.SameSiteLaxMode, // Protection against CSRF attacks
		})

		userResponse := responses.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		loginResponse := responses.LoginResponse{
			AccessToken: accessToken,
			User:        userResponse,
		}
		responses.Success(w, http.StatusOK, "Login successful! Welcome back.", loginResponse)
	}
}

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
			Email:    req.Email,
			Password: hashedPassword,
		})

		if err != nil {
			responses.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		responses.Success(w, http.StatusCreated, "User account created successfully!", nil)
	}
}

func (h *Handler) RefreshTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the cookie
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			responses.Error(w, http.StatusUnauthorized, "Refresh token missing")
			return
		}

		// Validate refresh token
		claims, err := h.TokenMaker.VerifyToken(cookie.Value)
		if err != nil {
			responses.Error(w, http.StatusUnauthorized, "Invalid or expired refresh token!")
			return
		}

		// Generate a fresh Access Token
		newAccessToken, err := h.TokenMaker.GenerateAccessToken(claims.UserId, claims.Username, 15*time.Minute)
		if err != nil {
			responses.Error(w, http.StatusInternalServerError, "Failed to issue new token")
			return
		}

		// Return new access token
		responses.Success(w, http.StatusOK, "Access token refreshed successfully!", map[string]string{
			"access_token": newAccessToken,
		})
	}
}
