package repositories

import (
	"context"

	"github.com/morikeli/golangrestapi/internal/store"
)

// UserRepository defines the database operations required by
// the application's user/authentication services.
type UserRepository interface {
	CreateUser(ctx context.Context, arg store.CreateUserParams) (store.CreateUserRow, error)
	GetUserById(ctx context.Context, id int64) (store.GetUserByIdRow, error)
	GetUserByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (store.GetUserByUsernameOrEmailRow, error)
	ListUsers(ctx context.Context, arg store.ListUsersParams) ([]store.ListUsersRow, error)
	CountUsers(ctx context.Context) (int64, error)
	UpdateUserProfile(ctx context.Context, arg store.UpdateUserProfileParams) (store.UpdateUserProfileRow, error)
}

// Compile-time check that sqlc's Queries implements
// the repository interface.
var _ UserRepository = (*store.Queries)(nil)
