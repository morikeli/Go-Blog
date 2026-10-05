package services

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/morikeli/golangrestapi/internal/repositories"
	"github.com/morikeli/golangrestapi/internal/store"
)

type UserService struct {
	UserRepository repositories.UserRepository
}

func NewUserService(userRepository repositories.UserRepository) *UserService {
	return &UserService{
		UserRepository: userRepository,
	}
}

// GetUser returns a user by ID.
func (s *UserService) GetUser(ctx context.Context, userID int64) (store.GetUserByIdRow, error) {
	user, err := s.UserRepository.GetUserById(ctx, userID)

	if err != nil {
		// [NOTE] In cases where the db is unavailable (e.g., network issue, db goes down), 
		// the User not found error may be returned but the user exists. 
		// 
		// Check if the error is due to no rows found to avoid returning a generic error
		if errors.Is(err, pgx.ErrNoRows) {
			return store.GetUserByIdRow{}, ErrUserNotFound
		}
		// Return the error if it's not due to no rows found
		return store.GetUserByIdRow{}, err
	}

	return user, nil
}

// ListUsers returns paginated users and the total count.
func (s *UserService) ListUsers(ctx context.Context, limit int, offset int) ([]store.ListUsersRow, int64, error) {
	users, err := s.UserRepository.ListUsers(
		ctx,
		store.ListUsersParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		},
	)

	if err != nil {
		return nil, 0, err
	}

	totalItems, err := s.UserRepository.CountUsers(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, totalItems, nil
}

// UpdateProfile updates the username and/or profile picture.
func (s *UserService) UpdateProfile(
	ctx context.Context,
	userID int64,
	username string,
	profilePhoto *string,
) (store.UpdateUserProfileRow, error) {
	params := store.UpdateUserProfileParams{
		ID: userID,
	}

	username = strings.TrimSpace(username)

	if username != "" {
		params.Username = pgtype.Text{String: username, Valid: true}
	}

	if profilePhoto != nil {
		params.ProfilePhoto = pgtype.Text{String: *profilePhoto, Valid: true}
	}

	user, err := s.UserRepository.UpdateUserProfile(ctx, params)

	if err != nil {
		var pgErr *pgconn.PgError

		// Check if the error is a duplicate username constraint violation
		if errors.As(err, pgErr) {
			if strings.Contains(pgErr.ConstraintName, "username_key") {
				return store.UpdateUserProfileRow{}, ErrDuplicateUsername
			}
		}

		// Return the error if it's not a duplicate username constraint violation
		return store.UpdateUserProfileRow{}, err
	}
	return user, nil
}
