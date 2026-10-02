package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/morikeli/golangrestapi/internal/dtos/requests"
	"github.com/morikeli/golangrestapi/internal/dtos/responses"
	"github.com/morikeli/golangrestapi/internal/services"
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

		result, err := h.AuthService.Login(ctx, req.Username, req.Password)

		if err != nil {
			if errors.Is(err, services.ErrInvalidCredentials) {
				responses.Error(w, http.StatusUnauthorized, "Invalid credentials provided!")
				return
			}
			
			responses.Error(
				w,
				http.StatusInternalServerError,
				"Oh snap! We could not authenticate you at the moment. Please try again later.",
			)
			return
		}


		// Attach Refresh Token as an HttpOnly, Secure Cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    result.RefreshToken,
			Path:     "/auth/token/refresh", // Restrict cookie scope exclusively to the refresh endpoint
			Expires:  time.Now().Add(7 * 24 * time.Hour),
			HttpOnly: true,                 // JavaScript cannot read this cookie (XSS protection)
			Secure:   true,                 // Requires HTTPS in production
			SameSite: http.SameSiteLaxMode, // Protection against CSRF attacks
		})

		userResponse := responses.UserResponse{
			ID:        result.User.ID,
			Username:  result.User.Username,
			Email:     result.User.Email,
			CreatedAt: result.User.CreatedAt,
			UpdatedAt: result.User.UpdatedAt,
		}

		loginResponse := responses.LoginResponse{
			AccessToken: result.AccessToken,
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

		err := h.AuthService.Signup(ctx, req.Username, req.Email, req.Password)

		if err != nil {
			responses.Error(
				w, http.StatusInternalServerError,
				"Oh no! We could not create your account. Please try again later.",
			)
			return
		}

		responses.Success(w, http.StatusCreated, "User account created successfully!", nil)
	}
}

func (h *Handler) RefreshTokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Read the cookie
		cookie, err := r.Cookie("refresh_token")

		if err != nil {
			responses.Error(w, http.StatusUnauthorized, "Refresh token missing!")
			return
		}

		// Validate refresh token
		claims, err := h.TokenMaker.VerifyRefreshToken(cookie.Value)

		if err != nil {
			responses.Error(w, http.StatusUnauthorized, "Invalid or expired refresh token!")
			return
		}

		// check if the token type is refresh
		if claims.TokenType != utils.RefreshToken || claims.ID == "" {
			responses.Error(w, http.StatusUnauthorized, "Invalid refresh token!")
			return
		}

		// Check whether this refresh token has already been revoked
		blacklistKey := "blacklist:" + claims.ID
		exists, err := h.Redis.Exists(ctx, blacklistKey).Result()

		// If Redis fails, trigger an early exit with HTTP status 533 Service Unavailable
		if err != nil {
			responses.Error(w, http.StatusServiceUnavailable, "Unable to validate refresh token!")
			return
		}

		if exists > 0 {
			responses.Error(w, http.StatusUnauthorized, "Refresh token has been revoked!")
			return
		}

		// Revoke the current refresh token BEFORE issuing a new one. This makes the refresh token single-use.
		if err := utils.RevokeRefreshToken(ctx, h.Redis, claims); err != nil {
			responses.Error(w, http.StatusServiceUnavailable, "Unable to rotate refresh token!")
			return
		}

		// Generate a fresh Access Token
		newAccessToken, err := h.TokenMaker.GenerateAccessToken(claims.UserId, claims.Username, 15*time.Minute)
		if err != nil {
			responses.Error(w, http.StatusInternalServerError, "Failed to issue new access token!")
			return
		}

		// Generate a new refresh token.
		newRefreshToken, err := h.TokenMaker.GenerateRefreshToken(claims.UserId, claims.Username, 7*24*time.Hour)
		if err != nil {
			responses.Error(w, http.StatusInternalServerError, "Failed to issue new refresh token!")
			return
		}

		// Replace the old refresh-token cookie.
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    newRefreshToken,
			Path:     "/auth/token/refresh",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,

			// Optional but recommended:
			// Seconds: >0 (persistent), <0 (delete immediately), 0 (session/unspecified)
			MaxAge: 7 * 24 * 60 * 60, // 7 days in seconds
		})

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

			if err == nil && claims.TokenType == utils.RefreshToken {
				_ = utils.RevokeRefreshToken(ctx, h.Redis, claims)
			}
		}

		// Expire the cookie in user's browser
		http.SetCookie(w, &http.Cookie{
			Name:  "refresh_token",
			Value: "", // Set the value to an empty string to wipe out the actual JWT payload.
			Path:  "/auth/token/refresh",

			// Sets the cookie's expiration date to January 1, 1970 UTC (Unix Epoch).
			// Because this timestamp is decades in the past, the browser immediately deletes the cookie.
			Expires:  time.Unix(0, 0),
			HttpOnly: true, // JavaScript cannot read this cookie (XSS protection)
			Secure:   true,
			// Retains CSRF protection by controlling cross-site cookie transmission behavior.
			SameSite: http.SameSiteLaxMode, // Prevents the cookie from being sent in cross-site requests.
		})

		responses.Success(w, http.StatusOK, "Successfully logged out!", nil)
	}
}
