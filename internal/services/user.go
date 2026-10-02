package services

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/morikeli/golangrestapi/internal/repositories"
	"github.com/morikeli/golangrestapi/internal/store"
)

var (
	ErrUserNotFound = errors.New("User not found!")
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
		return store.GetUserByIdRow{}, ErrUserNotFound
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
		params.Username = pgtype.Text{
			String: username,
			Valid:  true,
		}
	}

	if profilePhoto != nil {
		params.ProfilePhoto = pgtype.Text{
			String: *profilePhoto,
			Valid:  true,
		}
	}

	return s.UserRepository.UpdateUserProfile(
		ctx,
		params,
	)
}
