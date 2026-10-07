package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/morikeli/golangrestapi/internal/repositories"
	"github.com/morikeli/golangrestapi/internal/store"
	"github.com/morikeli/golangrestapi/internal/utils"
)

// AuthService contains authentication-related business logic.
type AuthService struct {
	UserRepository repositories.UserRepository
	TokenMaker     *utils.TokenMaker
}

func NewAuthService(userRepository repositories.UserRepository, tokenMaker *utils.TokenMaker) *AuthService {
	return &AuthService{
		UserRepository: userRepository,
		TokenMaker:     tokenMaker,
	}
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	User         store.GetUserByUsernameOrEmailRow
}

// Login authenticates a user and generates access/refresh tokens.
func (s *AuthService) Login(ctx context.Context, usernameOrEmail string, password string) (*LoginResult, error) {
	usernameOrEmail = utils.NormalizeUsernameOrEmail(usernameOrEmail)

	user, err := s.UserRepository.GetUserByUsernameOrEmail(ctx, usernameOrEmail)

	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !utils.ValidatePassword(user.Password, password) {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.TokenMaker.GenerateAccessToken(user.ID, user.Username, 15*time.Minute)

	if err != nil {
		return nil, err
	}

	refreshToken, err := s.TokenMaker.GenerateRefreshToken(user.ID, user.Username, 7*24*time.Hour)

	if err != nil {
		return nil, err
	}

	return &LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// Signup creates a new user.
func (s *AuthService) Signup(ctx context.Context, username string, email string, password string) error {
	username = utils.NormalizeUsernameOrEmail(username)
	email = utils.NormalizeUsernameOrEmail(email)

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	_, err = s.UserRepository.CreateUser(ctx, store.CreateUserParams{
		Username: username,
		Email:    email,
		Password: hashedPassword,
	})

	if err != nil {
		// Check if error is a Postgres unique constraint violation
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 = unique_violation
			if strings.Contains(pgErr.ConstraintName, "email_key") {
				return ErrDuplicateEmail
			}
			if strings.Contains(pgErr.ConstraintName, "username_key") {
				return ErrDuplicateUsername
			}
		}
		return err
	}

	return nil
}
