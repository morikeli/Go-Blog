package models

type User struct {
	Id       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	ProfilePhoto *string `json:"profile_photo"`
	Password string `json:"-"`
	TimestampMixin
}
