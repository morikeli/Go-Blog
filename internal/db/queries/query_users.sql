-- name: CreateUser :one
INSERT INTO users (username, email, password, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, username, email, created_at, updated_at;

-- name: UpdateUserProfile :one
UPDATE users
SET 
    username = COALESCE(sqlc.narg('username'), username),
    profile_photo = COALESCE(sqlc.narg('profile_photo'), profile_photo),
    updated_at = NOW()
WHERE id = $1
RETURNING id, username, email, profile_photo, created_at, updated_at;

-- name: GetUserById :one
SELECT id, username, email, profile_photo, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByUsernameOrEmail :one
SELECT id, username, email, password, created_at, updated_at
FROM users
WHERE username = $1 OR email = $1;

-- name: ListUsers :many
SELECT id, username, email, created_at, updated_at
FROM users
ORDER BY id ASC;
