package responses

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

type PaginatedUserResponse struct {
	Users      []UserResponse `json:"users"`
	Pagination PaginationMeta `json:"pagination"`
}