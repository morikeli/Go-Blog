package responses

import "github.com/jackc/pgx/v5/pgtype"

type UserResponse struct {
	ID             int64              `json:"id"`
	Username       string             `json:"username"`
	Email          string             `json:"email"`
	ProfilePicture string             `json:"profile_picture"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	UpdatedAt      pgtype.Timestamptz `json:"updated_at"`
}
