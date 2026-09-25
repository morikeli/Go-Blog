package models

type TimestampMixin struct {
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
