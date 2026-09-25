package models

type Blog struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	UserId  int    `json:"user_id"`
	TimestampMixin
}
