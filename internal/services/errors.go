package services

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDuplicateEmail    = errors.New("email is already registered")
	ErrDuplicateUsername = errors.New("username is already registered")
	ErrUserNotFound      = errors.New("user not found")
)