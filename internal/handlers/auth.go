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
			responses.Error(w, http.StatusUnauthorized, "Refresh token missing!")
			return
		}

		// Validate refresh token
		claims, err := h.TokenMaker.VerifyToken(cookie.Value)
		if err != nil {
			responses.Error(w, http.StatusUnauthorized, "Invalid or expired refresh token!")
			return
		}

		// Check Redis blacklist for token ID (claims.ID)
		blacklistKey := "blacklist:" + claims.ID
		exists, err := h.Redis.Exists(r.Context(), blacklistKey).Result()
		if err == nil && exists > 0 {
			responses.Error(w, http.StatusUnauthorized, "Refresh token has been revoked!")
			return
		}

		// Generate a fresh Access Token
		newAccessToken, err := h.TokenMaker.GenerateAccessToken(claims.UserId, claims.Username, 15*time.Minute)
		if err != nil {
			responses.Error(w, http.StatusInternalServerError, "Failed to issue new access token!")
			return
		}

		// Return new access token
		responses.Success(w, http.StatusOK, "Access token refreshed successfully!", map[string]string{
			"access_token": newAccessToken,
		})
	}
}

func (h *Handler) LogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		cookie, err := r.Cookie("refresh_token")
		if err == nil {
			claims, err := h.TokenMaker.VerifyToken(cookie.Value)
			
			if err == nil && claims.ID != "" {
				// Calculate remaining TTL until expiration
				remainingDuration := time.Until(claims.ExpiresAt.Time)
				
				if remainingDuration > 0 {
					// Store token ID in Redis with expiration matching token TTL
					blacklistKey := "blacklist:" + claims.ID
					_ = h.Redis.Set(ctx, blacklistKey, "revoked", remainingDuration).Err()
				}
			}
		}

		// Expire the cookie in user's browser
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",	// Set the value to an empty string to wipe out the actual JWT payload.
			Path:     "/auth/refreshToken",

			// Sets the cookie's expiration date to January 1, 1970 UTC (Unix Epoch).
    		// Because this timestamp is decades in the past, the browser immediately deletes the cookie.
			Expires:  time.Unix(0, 0),
			HttpOnly: true,	// JavaScript cannot read this cookie (XSS protection)
			Secure:   true,
			// Retains CSRF protection by controlling cross-site cookie transmission behavior.
			SameSite: http.SameSiteLaxMode,	// Prevents the cookie from being sent in cross-site requests.
		})

		responses.Success(w, http.StatusOK, "Successfully logged out!", nil)
	}
}